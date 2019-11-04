package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceVcdExternalNetwork() *schema.Resource {
	return &schema.Resource{
		Read: resourceVcdExternalNetworkRead,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_scope": &schema.Schema{
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A list of IP scopes for the network",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gateway": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Gateway of the network",
						},
						"netmask": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Network mask",
						},
						"dns1": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Primary DNS server",
						},
						"dns2": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Secondary DNS server",
						},
						"dns_suffix": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "DNS suffix",
						},
						"static_ip_pool": &schema.Schema{
							Type:        schema.TypeList,
							Computed:    true,
							Description: "IP ranges used for static pool allocation in the network",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"start_address": &schema.Schema{
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Start address of the IP range",
									},
									"end_address": &schema.Schema{
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
			"vsphere_network": &schema.Schema{
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A list of port groups that back this network. Each referenced DV_PORTGROUP or NETWORK must exist on a vCenter server registered with the system.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vcenter": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The vCenter server name",
						},
						"name": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the port group",
						},
						"type": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The vSphere port group type. One of: DV_PORTGROUP (distributed virtual port group), NETWORK",
						},
					},
				},
			},
			"retain_net_info_across_deployments": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies whether the network resources such as IP/MAC of router will be retained across deployments. Default is false.",
			},
		},
	}
}
