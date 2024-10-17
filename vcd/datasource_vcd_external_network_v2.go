package vcd

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/vmware/go-vcloud-director/v3/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdExternalNetworkV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdExternalNetworkV2Read,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dedicated_org_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of an Org that this network is dedicated to (VCD 10.4.1+)",
			},
			"use_ip_spaces": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defines if this network uses IP Spaces (VCD 10.4.1+)",
			},
			"nat_and_firewall_service_intention": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Defines different types of intentions to configure NAT and Firewall rules (only with IP Spaces, VCD 10.5.1+) One of `PROVIDER_GATEWAY`,`EDGE_GATEWAY`,`PROVIDER_AND_EDGE_GATEWAY`",
			},
			"route_advertisement_intention": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Defines intentions to configure route advertisement (only with IP Spaces, VCD 10.5.1+) One of `IP_SPACE_UPLINKS_ADVERTISED_STRICT`,`IP_SPACE_UPLINKS_ADVERTISED_FLEXIBLE`,`ALL_NETWORKS_ADVERTISED`",
			},
			"ip_scope": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of IP scopes for the network",
				Elem:        networkV2IpScope,
			},
			"vsphere_network": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A list of port groups that back this network. Each referenced DV_PORTGROUP or NETWORK must exist on a vCenter server registered with the system.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vcenter_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The vCenter server ID",
						},
						"portgroup_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The portgroup ID",
						},
					},
				},
			},
			"nsxt_network": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"nsxt_manager_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of NSX-T manager",
						},
						"nsxt_tier0_router_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of NSX-T Tier-0 router (for T0 gateway backed external network)",
						},
						"nsxt_segment_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of NSX-T segment (for NSX-T segment backed external network)",
						},
					},
				},
			},
		},
	}
}

func datasourceVcdExternalNetworkV2Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	log.Printf("[TRACE] external network V2 data source read initiated")

	name := d.Get("name").(string)

	extNet, err := govcd.GetExternalNetworkV2ByName(vcdClient.VCDClient, name)
	if err != nil {
		return diag.Errorf("could not find external network V2 by name '%s': %s", name, err)
	}

	d.SetId(extNet.ExternalNetwork.ID)

	err = setExternalNetworkV2Data(d, extNet.ExternalNetwork)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
