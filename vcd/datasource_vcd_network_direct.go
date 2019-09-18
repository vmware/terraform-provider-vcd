package vcd

import "github.com/hashicorp/terraform/helper/schema"

func datasourceVcdNetworkDirect() *schema.Resource {
	return &schema.Resource{
		Read: resourceVcdNetworkDirectRead,
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

			"external_network": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"external_network_gateway": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"external_network_netmask": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"external_network_dns1": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"external_network_dns2": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"external_network_dns_suffix": &schema.Schema{
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
		},
	}
}
