package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdExternalNetwork() *schema.Resource {
	return &schema.Resource{
		Read: resourceVcdExternalNetworkRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_scope": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A list of IP scopes for the network",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gateway": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Gateway of the network",
						},
						"netmask": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Network mask",
						},
						"dns1": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Primary DNS server",
						},
						"dns2": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Secondary DNS server",
						},
						"dns_suffix": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "DNS suffix",
						},
						"static_ip_pool": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "IP ranges used for static pool allocation in the network",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"start_address": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Start address of the IP range",
									},
									"end_address": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "End address of the IP range",
									},
								},
							},
						},
					},
				},
			},
			"vsphere_network": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A list of port groups that back this network. Each referenced DV_PORTGROUP or NETWORK must exist on a vCenter server registered with the system.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vcenter": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The vCenter server name",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the port group",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The vSphere port group type. One of: DV_PORTGROUP (distributed virtual port group), NETWORK",
						},
					},
				},
			},
			"retain_net_info_across_deployments": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies whether the network resources such as IP/MAC of router will be retained across deployments. Default is false.",
			},
		},
	}
}
