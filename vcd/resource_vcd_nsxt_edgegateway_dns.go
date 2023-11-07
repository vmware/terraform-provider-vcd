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

func resourceVcdNsxtEdgegatewayDns() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceVcdNsxtEdgegatewayDnsRead,
		CreateContext: resourceVcdNsxtEdgegatewayDnsCreate,
		UpdateContext: resourceVcdNsxtEdgegatewayDnsUpdate,
		DeleteContext: resourceVcdNsxtEdgegatewayDnsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdNsxtEdgegatewayDnsImport,
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
				Description: "Edge gateway ID for DNS configuration",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Status of the DNS Forwarder. Defaults to `true`",
			},
			"listener_ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: "IP of the DNS forwarder. " +
					"Can be modified only if the Edge Gateway has a dedicated external network.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ValidateFunc: validation.IsIPAddress,
			},
			"default_forwarder_zone": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "The default forwarder zone.",
				Elem:        defaultForwarderZone,
			},
			"conditional_forwarder_zone": {
				Type:        schema.TypeSet,
				Optional:    true,
				MaxItems:    5,
				Description: "Conditional forwarder zone",
				Elem:        conditionalForwarderZone,
			},
		},
	}
}

var defaultForwarderZone = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"zone_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Unique ID of the forwarder zone.",
		},
		"display_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Name of the forwarder zone.",
		},
		"upstream_servers": {
			Type:        schema.TypeSet,
			Required:    true,
			MaxItems:    3,
			Description: "Servers to which DNS requests should be forwarded to.",
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.IsIPAddress,
			},
		},
	},
}

var conditionalForwarderZone = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"zone_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Unique ID of the forwarder zone.",
		},
		"upstream_servers": {
			Type:        schema.TypeSet,
			Required:    true,
			MaxItems:    3,
			Description: "Servers to which DNS requests should be forwarded to.",
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.IsIPAddress,
			},
		},
		"domain_names": {
			Type:        schema.TypeSet,
			Required:    true,
			Description: "List of domain names on which conditional forwarding is based.",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	},
}

func resourceVcdNsxtEdgegatewayDnsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdNsxtEdgegatewayDhcpV6CreateUpdate(ctx, d, meta, "create")
}

func resourceVcdNsxtEdgegatewayDnsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdNsxtEdgegatewayDnsCreateUpdate(ctx, d, meta, "update")
}

func resourceVcdNsxtEdgegatewayDnsCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	unlock, err := vcdClient.lockParentVdcGroupOrEdgeGateway(d)
	if err != nil {
		return diag.Errorf("[dhcpv6 (SLAAC Profile) %s] %s", origin, err)
	}
	defer unlock()

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[dhcpv6 (SLAAC Profile) %s] error retrieving Edge Gateway: %s", origin, err)
	}

	dhcpv6Config, err := getNsxtEdgeGatewaySlaacProfileType(d)
	if err != nil {
		return diag.Errorf("[dhcpv6 (SLAAC Profile) %s] error getting DHCPv6 configuration: %s", origin, err)
	}

	_, err = nsxtEdge.UpdateSlaacProfile(dhcpv6Config)
	if err != nil {
		return diag.Errorf("[dhcpv6 (SLAAC Profile) %s] error updating DHCPv6 configuration: %s", origin, err)
	}

	d.SetId(edgeGatewayId)

	return resourceVcdNsxtEdgegatewayDhcpV6Read(ctx, d, meta)
}

func resourceVcdNsxtEdgegatewayDnsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		if govcd.ContainsNotFound(err) {
			// When parent Edge Gateway is not found - this resource is also not found and should be
			// removed from state
			d.SetId("")
			return nil
		}
		return diag.Errorf("[dhcpv6 (SLAAC Profile) read] error retrieving NSX-T Edge Gateway DHCPv6 (SLAAC Profile): %s", err)
	}

	slaacProfile, err := nsxtEdge.GetSlaacProfile()
	if err != nil {
		return diag.Errorf("[dhcpv6 (SLAAC Profile) read] error retrieving NSX-T Edge Gateway DHCPv6 (SLAAC Profile): %s", err)
	}

	err = setNsxtEdgeGatewaySlaacProfileData(d, slaacProfile)
	if err != nil {
		return diag.Errorf("error storing state: %s", err)
	}

	return nil
}

func resourceVcdNsxtEdgegatewayDnsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	unlock, err := vcdClient.lockParentVdcGroupOrEdgeGateway(d)
	if err != nil {
		return diag.Errorf("[dhcpv6 (SLAAC Profile) delete] %s", err)
	}
	defer unlock()

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[dhcpv6 (SLAAC Profile) delete] error retrieving Edge Gateway: %s", err)
	}

	// Disabling DHCPv6 configuration requires at least Mode field
	_, err = nsxtEdge.UpdateSlaacProfile(&types.NsxtEdgeGatewaySlaacProfile{Mode: "DISABLED", Enabled: false})
	if err != nil {
		return diag.Errorf("[dhcpv6 (SLAAC Profile) delete] error updating DHCPv6 Profile: %s", err)
	}

	return nil
}

func resourceVcdNsxtEdgegatewayDnsImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] NSX-T Edge Gateway DHCPv6 import initiated")

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-name.nsxt-edge-gw-name or org-name.vdc-group-name.nsxt-edge-gw-name")
	}
	orgName, vdcOrVdcGroupName, edgeName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)
	vdcOrVdcGroup, err := lookupVdcOrVdcGroup(vcdClient, orgName, vdcOrVdcGroupName)
	if err != nil {
		return nil, err
	}

	if !vdcOrVdcGroup.IsNsxt() {
		return nil, fmt.Errorf("please use 'vcd_edgegateway' for NSX-V backed VDC")
	}

	edge, err := vdcOrVdcGroup.GetNsxtEdgeGatewayByName(edgeName)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve NSX-T Edge Gateway with ID '%s': %s", d.Id(), err)
	}

	dSet(d, "org", orgName)
	dSet(d, "edge_gateway_id", edge.EdgeGateway.ID)

	// Storing Edge Gateway ID and Read will retrieve all other data
	d.SetId(edge.EdgeGateway.ID)

	return []*schema.ResourceData{d}, nil
}
