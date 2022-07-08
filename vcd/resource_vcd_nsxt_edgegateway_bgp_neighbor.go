package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdEdgeBgpNeighbor() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdEdgeBgpNeighborCreate,
		ReadContext:   resourceVcdEdgeBgpNeighborRead,
		UpdateContext: resourceVcdEdgeBgpNeighborUpdate,
		DeleteContext: resourceVcdEdgeBgpNeighborDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdEdgeBgpNeighborImport,
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
				Description: "Edge gateway ID for BGP Neighbor Configuration",
			},
			"ip_address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "BGP Neighbor IP address (IPv4 or IPv6)",
			},
			"remote_as_number": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Remote Autonomous System (AS) number",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Neighbor password",
			},
			"keep_alive_timer": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Time interval (in seconds) between sending keep alive messages to a peer",
			},
			"hold_down_timer": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Time interval (in seconds) before declaring a peer dead",
			},
			"graceful_restart_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "One of 'DISABLE', 'HELPER_ONLY', 'GRACEFUL_AND_HELPER'",
				ValidateFunc: validation.StringInSlice([]string{"DISABLE", "HELPER_ONLY", "GRACEFUL_AND_HELPER"}, false),
			},
			"allow_as_in": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "A flag indicating whether BGP neighbors can receive routes with same Autonomous System (AS) (default 'false')",
			},
			"bfd_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "BFD configuration for failure detection",
			},
			"bfd_interval": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Time interval (in milliseconds) between heartbeat packets",
			},
			"bfd_dead_multiple": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Number of times a heartbeat packet is missed before BFD declares that the neighbor is down",
			},
			"route_filtering": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "One of 'DISABLED', 'IPV4', 'IPV6'",
				ValidateFunc: validation.StringInSlice([]string{"DISABLED", "IPV4", "IPV6"}, false),
			},
			"in_filter_ip_prefix_list_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "An optional IP Prefix List ID for filtering 'IN' direction.",
			},
			"out_filter_ip_prefix_list_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "An optional IP Prefix List ID for filtering 'OUT' direction.",
			},
		},
	}
}

func resourceVcdEdgeBgpNeighborCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	// Handling locks for BGP configuration is conditional. There are two scenarios:
	// * When the parent Edge Gateway is in a VDC - a lock on parent Edge Gateway must be acquired
	// * When the parent Edge Gateway is in a VDC Group - a lock on parent VDC Group must be acquired
	// To find out parent lock object, Edge Gateway must be looked up and its OwnerRef must be checked
	// Note. It is not safe to do multiple locks in the same resource as it can result in a deadlock
	parentEdgeGatewayOwnerId, _, err := getParentEdgeGatewayOwnerId(vcdClient, d)
	if err != nil {
		return diag.Errorf("[bgp neighbor create] error finding parent Edge Gateway: %s", err)
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
		return diag.Errorf("[bgp neighbor create] error retrieving Edge Gateway: %s", err)
	}

	bgpNeighbor := getEdgeBgpNeighborType(d)

	createdbgpNeighbor, err := nsxtEdge.CreateBgpNeighbor(bgpNeighbor)
	if err != nil {
		return diag.Errorf("[bgp neighbor create] error updating NSX-T Edge Gateway BGP IP Prefix List: %s", err)
	}

	d.SetId(createdbgpNeighbor.EdgeBgpNeighbor.ID)
	return resourceVcdEdgeBgpNeighborRead(ctx, d, meta)
}

func resourceVcdEdgeBgpNeighborUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	// Handling locks for BGP configuration is conditional. There are two scenarios:
	// * When the parent Edge Gateway is in a VDC - a lock on parent Edge Gateway must be acquired
	// * When the parent Edge Gateway is in a VDC Group - a lock on parent VDC Group must be acquired
	// To find out parent lock object, Edge Gateway must be looked up and its OwnerRef must be checked
	// Note. It is not safe to do multiple locks in the same resource as it can result in a deadlock
	parentEdgeGatewayOwnerId, _, err := getParentEdgeGatewayOwnerId(vcdClient, d)
	if err != nil {
		return diag.Errorf("[bgp configuration update] error finding parent Edge Gateway: %s", err)
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
		return diag.Errorf("[bgp neighbor update] error retrieving Edge Gateway: %s", err)
	}

	bgpNeighbor, err := nsxtEdge.GetBgpNeighborById(d.Id())
	if err != nil {
		return diag.Errorf("[bgp neighbor update] error retrieving NSX-T Edge Gateway BGP IP Prefix Lis: %s", err)
	}

	bgpNeighbor.EdgeBgpNeighbor = getEdgeBgpNeighborType(d)
	bgpNeighbor.EdgeBgpNeighbor.ID = d.Id()

	_, err = bgpNeighbor.Update(bgpNeighbor.EdgeBgpNeighbor)
	if err != nil {
		return diag.Errorf("[bgp neighbor update] error updating NSX-T Edge Gateway BGP IP Prefix List: %s", err)
	}

	return resourceVcdEdgeBgpNeighborRead(ctx, d, meta)
}

func resourceVcdEdgeBgpNeighborRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("[bgp neighbor read] error retrieving NSX-T Edge Gateway BGP IP Prefix List: %s", err)
	}

	bgpNeighbor, err := nsxtEdge.GetBgpNeighborById(d.Id())
	if err != nil {
		return diag.Errorf("[bgp neighbor read] error retrieving NSX-T Edge Gateway BGP IP Prefix List: %s", err)
	}

	err = setEdgeBgpNeighborData(d, bgpNeighbor.EdgeBgpNeighbor)
	if err != nil {
		return diag.Errorf("[bgp neighbor read] error storing entity into schema: %s", err)
	}

	return nil
}

func resourceVcdEdgeBgpNeighborDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	// Handling locks for BGP configuration is conditional. There are two scenarios:
	// * When the parent Edge Gateway is in a VDC - a lock on parent Edge Gateway must be acquired
	// * When the parent Edge Gateway is in a VDC Group - a lock on parent VDC Group must be acquired
	// To find out parent lock object, Edge Gateway must be looked up and its OwnerRef must be checked
	// Note. It is not safe to do multiple locks in the same resource as it can result in a deadlock
	parentEdgeGatewayOwnerId, _, err := getParentEdgeGatewayOwnerId(vcdClient, d)
	if err != nil {
		return diag.Errorf("[bgp neighbor delete] error finding parent Edge Gateway: %s", err)
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
		return diag.Errorf("[bgp neighbor delete] error retrieving NSX-T Edge Gateway: %s", err)
	}

	bgpNeighbor, err := nsxtEdge.GetBgpNeighborById(d.Id())
	if err != nil {
		return diag.Errorf("[bgp neighbor delete] error retrieving NSX-T Edge Gateway BGP Configuration: %s", err)
	}

	err = bgpNeighbor.Delete()
	if err != nil {
		return diag.Errorf("[bgp neighbor delete] error deleting NSX-T Edge Gateway BGP Configuration: %s", err)
	}

	return nil
}

func resourceVcdEdgeBgpNeighborImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.SplitN(d.Id(), ImportSeparator, 4)
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-or-vdc-group-name.edge_gateway_name.bgp_neighbor_ip, got '%s'", d.Id())
	}
	orgName, vdcOrVdcGroupName, edgeGatewayName, bgpNeighborIp := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)
	vdcOrVdcGroup, err := lookupVdcOrVdcGroup(vcdClient, orgName, vdcOrVdcGroupName)
	if err != nil {
		return nil, err
	}

	edgeGateway, err := vdcOrVdcGroup.GetNsxtEdgeGatewayByName(edgeGatewayName)
	if err != nil {
		return nil, fmt.Errorf("[bgp neighbor import] unable to find Edge Gateway '%s': %s", edgeGatewayName, err)
	}

	bgpPrefixList, err := edgeGateway.GetBgpNeighborByIp(bgpNeighborIp)
	if err != nil {
		return nil, fmt.Errorf("[bgp neighbor import] unable to find BGP Neighbor with Name '%s': %s", bgpNeighborIp, err)
	}

	dSet(d, "org", orgName)
	dSet(d, "edge_gateway_id", edgeGateway.EdgeGateway.ID)
	d.SetId(bgpPrefixList.EdgeBgpNeighbor.ID)

	return []*schema.ResourceData{d}, nil
}

func getEdgeBgpNeighborType(d *schema.ResourceData) *types.EdgeBgpNeighbor {
	bgpNeighborConfig := &types.EdgeBgpNeighbor{
		NeighborAddress:        d.Get("ip_address").(string),
		RemoteASNumber:         d.Get("remote_as_number").(string),
		KeepAliveTimer:         d.Get("keep_alive_timer").(int),
		HoldDownTimer:          d.Get("hold_down_timer").(int),
		NeighborPassword:       d.Get("password").(string),
		AllowASIn:              d.Get("allow_as_in").(bool),
		GracefulRestartMode:    d.Get("graceful_restart_mode").(string),
		IpAddressTypeFiltering: d.Get("route_filtering").(string),
	}

	bgpNeighborConfig.Bfd = &types.EdgeBgpNeighborBfd{
		Enabled:             d.Get("bfd_enabled").(bool),
		BfdInterval:         d.Get("bfd_interval").(int),
		DeclareDeadMultiple: d.Get("bfd_dead_multiple").(int),
	}

	if inRoutesInterface, exists := d.GetOk("in_filter_ip_prefix_list_id"); exists {
		bgpNeighborConfig.InRoutesFilterRef = &types.OpenApiReference{ID: inRoutesInterface.(string)}
	}

	if outRoutesInterface, exists := d.GetOk("out_filter_ip_prefix_list_id"); exists {
		bgpNeighborConfig.OutRoutesFilterRef = &types.OpenApiReference{ID: outRoutesInterface.(string)}
	}

	return bgpNeighborConfig
}

func setEdgeBgpNeighborData(d *schema.ResourceData, bgpNeighborConfig *types.EdgeBgpNeighbor) error {
	dSet(d, "ip_address", bgpNeighborConfig.NeighborAddress)
	dSet(d, "remote_as_number", bgpNeighborConfig.RemoteASNumber)

	dSet(d, "keep_alive_timer", bgpNeighborConfig.KeepAliveTimer)
	dSet(d, "remote_as_number", bgpNeighborConfig.RemoteASNumber)
	dSet(d, "hold_down_timer", bgpNeighborConfig.HoldDownTimer)
	// Password is a "write-only" field. It cannot be read afterwards.
	//dSet(d, "password", bgpNeighborConfig.NeighborPassword)
	dSet(d, "allow_as_in", bgpNeighborConfig.AllowASIn)

	dSet(d, "graceful_restart_mode", bgpNeighborConfig.GracefulRestartMode)
	dSet(d, "route_filtering", bgpNeighborConfig.IpAddressTypeFiltering)

	if bgpNeighborConfig.Bfd != nil {
		dSet(d, "bfd_enabled", bgpNeighborConfig.Bfd.Enabled)
		dSet(d, "bfd_interval", bgpNeighborConfig.Bfd.BfdInterval)
		dSet(d, "bfd_dead_multiple", bgpNeighborConfig.Bfd.DeclareDeadMultiple)
	}

	if bgpNeighborConfig.InRoutesFilterRef != nil {
		dSet(d, "in_filter_ip_prefix_list_id", bgpNeighborConfig.InRoutesFilterRef.ID)
	} else {
		dSet(d, "in_filter_ip_prefix_list_id", "")
	}
	if bgpNeighborConfig.OutRoutesFilterRef != nil {
		dSet(d, "out_filter_ip_prefix_list_id", bgpNeighborConfig.OutRoutesFilterRef.ID)
	} else {
		dSet(d, "out_filter_ip_prefix_list_id", "")
	}

	return nil
}
