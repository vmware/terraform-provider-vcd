package vcd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdEdgeBgpConfig() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdEdgeBgpConfigCreateUpdate,
		ReadContext:   resourceVcdEdgeBgpConfigRead,
		UpdateContext: resourceVcdEdgeBgpConfigCreateUpdate,
		DeleteContext: resourceVcdEdgeBgpConfigDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdEdgeBgpConfigImport,
		},

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"edge_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway ID for BGP Configuration",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Defines if BGP service is enabled",
			},
			"local_as_number": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Autonomous system number",
			},
			"ecmp_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Defines if ECMP (Equal-cost multi-path routing) is enabled",
			},
			"graceful_restart_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "Graceful restart configuration on Edge Gateway. One of 'DISABLE', 'HELPER_ONLY', 'GRACEFUL_AND_HELPER'",
				ValidateFunc: validation.StringInSlice([]string{"DISABLE", "HELPER_ONLY", "GRACEFUL_AND_HELPER"}, false),
			},
			"graceful_restart_timer": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Maximum time taken (in seconds) for a BGP session to be established after a restart",
			},
			"stale_route_timer": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Maximum time (in seconds) before stale routes are removed when BGP restarts",
			},
		},
	}
}

func resourceVcdEdgeBgpConfigCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	// Handling locks on for BGP configuration is conditional. There are two scenarios:
	// * When the parent Edge Gateway is in a VDC - a lock on parent Edge Gateway must be acquired
	// * When the parent Edge Gateway is in a VDC Group - a lock on parent VDC Group must be acquired
	// To find out parent lock object, Edge Gateway must be looked up and its OwnerRef must be checked
	// Note. It is not safe to do multiple locks in the same resource as it can result in a deadlock
	parentEdgeGatewayOwnerId, _, err := getParentEdgeGatewayOwnerId(vcdClient, d)
	if err != nil {
		return diag.Errorf("[bgp configuration create/update] error finding parent Edge Gateway: %s", err)
	}

	if govcd.OwnerIsVdcGroup(parentEdgeGatewayOwnerId) {
		vcdClient.lockById(parentEdgeGatewayOwnerId)
		defer vcdClient.unlockById(parentEdgeGatewayOwnerId)
	} else {
		vcdClient.lockParentEdgeGtw(d)
		defer vcdClient.unLockParentEdgeGtw(d)
	}

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("error retrieving Edge Gateway: %s", err)
	}

	bgpConfig := getEdgeBgpConfigType(d)

	// If there is no user provided GracefulRestart config, reading BGP configuration as some VCD
	// (10.5.0+) versions require supplying existing values
	if bgpConfig.GracefulRestart == nil {
		existingBgpConfig, err := nsxtEdge.GetBgpConfiguration()
		if err != nil {
			return diag.Errorf("error reading BGP config before creation: %s", err)
		}
		bgpConfig.GracefulRestart = existingBgpConfig.GracefulRestart
	}

	_, err = nsxtEdge.UpdateBgpConfiguration(bgpConfig)
	if err != nil {
		return diag.Errorf("error updating NSX-T Edge Gateway BGP Configuration: %s", err)
	}

	d.SetId(nsxtEdge.EdgeGateway.ID)

	return resourceVcdEdgeBgpConfigRead(ctx, d, meta)
}

func resourceVcdEdgeBgpConfigRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error retrieving NSX-T Edge Gateway: %s", err)
	}

	bgpConfig, err := nsxtEdge.GetBgpConfiguration()
	if err != nil {
		return diag.Errorf("error retrieving NSX-T Edge Gateway BGP Configuration: %s", err)
	}

	setEdgeBgpConfigData(d, bgpConfig)

	return nil
}

func resourceVcdEdgeBgpConfigDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	// Handling locks on for BGP configuration is conditional. There are two scenarios:
	// * When the parent Edge Gateway is in a VDC - a lock on parent Edge Gateway must be acquired
	// * When the parent Edge Gateway is in a VDC Group - a lock on parent VDC Group must be acquired
	// To find out parent lock object, Edge Gateway must be looked up and its OwnerRef must be checked
	// Note. It is not safe to do multiple locks in the same resource as it can result in a deadlock
	parentEdgeGatewayOwnerId, _, err := getParentEdgeGatewayOwnerId(vcdClient, d)
	if err != nil {
		return diag.Errorf("[route advertisement delete] error finding parent Edge Gateway: %s", err)
	}

	if govcd.OwnerIsVdcGroup(parentEdgeGatewayOwnerId) {
		vcdClient.lockById(parentEdgeGatewayOwnerId)
		defer vcdClient.unlockById(parentEdgeGatewayOwnerId)
	} else {
		vcdClient.lockParentEdgeGtw(d)
		defer vcdClient.unLockParentEdgeGtw(d)
	}

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)

	if err != nil {
		return diag.Errorf("error retrieving NSX-T Edge Gateway: %s", err)
	}

	err = nsxtEdge.DisableBgpConfiguration()
	if err != nil {
		return diag.Errorf("error disableing NSX-T Edge Gateway BGP Configuration: %s", err)
	}

	return nil
}

func resourceVcdEdgeBgpConfigImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] NSX-T Edge Gateway BGP Configuration import initiated")

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-or-vdc-group-name.nsxt-edge-gw-name")
	}
	orgName, vdcOrVdcGroupName, edgeName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)
	vdcOrVdcGroup, err := lookupVdcOrVdcGroup(vcdClient, orgName, vdcOrVdcGroupName)
	if err != nil {
		return nil, err
	}

	edge, err := vdcOrVdcGroup.GetNsxtEdgeGatewayByName(edgeName)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve NSX-T edge gateway with ID '%s': %s", d.Id(), err)
	}

	dSet(d, "org", orgName)
	dSet(d, "edge_gateway_id", edge.EdgeGateway.ID)
	d.SetId(edge.EdgeGateway.ID)

	return []*schema.ResourceData{d}, nil
}

func getEdgeBgpConfigType(d *schema.ResourceData) *types.EdgeBgpConfig {
	bgpConfig := &types.EdgeBgpConfig{
		Enabled: d.Get("enabled").(bool),
		Ecmp:    d.Get("ecmp_enabled").(bool),
		// Version is required, but it is automatically handled by Go Cloud Director SDK in function
		// UpdateBgpConfiguration
	}

	// bgpConfig.GracefulRestart can only be specified when Edge Gateway is backed by Tier-0
	// gateway, not a VRF. For that reason types.EdgeBgpGracefulRestartConfig must only be sent if
	// the user configured it.
	graceFullRestartMode, graceFullRestartModeExists := d.GetOk("graceful_restart_mode")
	graceFullRestartTimer, graceFullRestartTimerExists := d.GetOk("graceful_restart_timer")
	staleRouteTimer, staleRouteTimerExists := d.GetOk("stale_route_timer")

	if graceFullRestartModeExists || graceFullRestartTimerExists || staleRouteTimerExists {
		bgpConfig.GracefulRestart = &types.EdgeBgpGracefulRestartConfig{
			Mode:            graceFullRestartMode.(string),
			RestartTimer:    graceFullRestartTimer.(int),
			StaleRouteTimer: staleRouteTimer.(int),
		}
	}

	localAsNumber, localAsNumberExists := d.GetOk("local_as_number")
	if localAsNumberExists {
		bgpConfig.LocalASNumber = localAsNumber.(string)
	}

	return bgpConfig
}

func setEdgeBgpConfigData(d *schema.ResourceData, bgpConfig *types.EdgeBgpConfig) {
	dSet(d, "enabled", bgpConfig.Enabled)
	dSet(d, "ecmp_enabled", bgpConfig.Ecmp)
	dSet(d, "local_as_number", bgpConfig.LocalASNumber)
	dSet(d, "graceful_restart_mode", bgpConfig.GracefulRestart.Mode)
	dSet(d, "graceful_restart_timer", bgpConfig.GracefulRestart.RestartTimer)
	dSet(d, "stale_route_timer", bgpConfig.GracefulRestart.StaleRouteTimer)
}
