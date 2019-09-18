package vcd

import "github.com/hashicorp/terraform/helper/schema"

func datasourceVcdNetworkRouted() *schema.Resource {
	return &schema.Resource{
		Read: resourceVcdNetworkRoutedRead,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"org": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"vdc": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"edge_gateway": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"netmask": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"gateway": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"dns1": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"dns2": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"dns_suffix": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"href": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"shared": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},

			"dhcp_pool": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_address": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"end_address": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"default_lease_time": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},

						"max_lease_time": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"static_ip_pool": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_address": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"end_address": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}
