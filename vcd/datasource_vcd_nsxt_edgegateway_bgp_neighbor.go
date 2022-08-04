package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdEdgeBgpNeighbor() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdEdgeBgpNeighborRead,

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
				Computed:    true,
				Description: "Remote Autonomous System (AS) number",
			},
			"password": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Neighbor password",
			},
			"keep_alive_timer": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Time interval (in seconds) between sending keep alive messages to a peer",
			},
			"hold_down_timer": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Time interval (in seconds) before declaring a peer dead",
			},
			"graceful_restart_mode": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "One of 'DISABLE', 'HELPER_ONLY', 'GRACEFUL_AND_HELPER'",
			},
			"allow_as_in": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "A flag indicating whether BGP neighbors can receive routes with same Autonomous System (AS)",
			},
			"bfd_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "BFD configuration for failure detection",
			},
			"bfd_interval": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Time interval (in milliseconds) between heartbeat packets",
			},
			"bfd_dead_multiple": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of times a heartbeat packet is missed before BFD declares that the neighbor is down",
			},
			"route_filtering": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "One of 'DISABLED', 'IPV4', 'IPV6'",
			},
			"in_filter_ip_prefix_list_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "An optional IP Prefix List ID for filtering 'IN' direction.",
			},
			"out_filter_ip_prefix_list_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "An optional IP Prefix List ID for filtering 'OUT' direction.",
			},
		},
	}
}

func datasourceVcdEdgeBgpNeighborRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[bgp ip prefix list read] error retrieving NSX-T Edge Gateway BGP Neighbor: %s", err)
	}

	bgpIpPrefixList, err := nsxtEdge.GetBgpNeighborByIp(d.Get("ip_address").(string))
	if err != nil {
		return diag.Errorf("[bgp ip prefix list read] error retrieving NSX-T Edge Gateway BGP Neighbor: %s", err)
	}

	err = setEdgeBgpNeighborData(d, bgpIpPrefixList.EdgeBgpNeighbor)
	if err != nil {
		return diag.Errorf("[bgp ip prefix list read] error storing entity into schema: %s", err)
	}

	d.SetId(bgpIpPrefixList.EdgeBgpNeighbor.ID)

	return nil
}
