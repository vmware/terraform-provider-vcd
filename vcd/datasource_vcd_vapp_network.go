package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceVcdVappNetwork() *schema.Resource {
	return &schema.Resource{
		Read: datasourceVappNetworkRead,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"vapp_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Optional description for the network",
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

			"guest_vlan_allowed": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"org_network": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "org network name to which vapp network is connected",
			},
			"firewall_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "firewall service enabled or disabled. Default - true",
			},
			"nat_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "NAT service enabled or disabled. Default - true",
			},
			"retain_ip_mac_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "NAT service enabled or disabled. Default - true",
			},
			"dhcp_pool": &schema.Schema{
				Type:     schema.TypeSet,
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

						"enabled": &schema.Schema{
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
				Set: resourceVcdNetworkIPAddressHash,
			},
			"static_ip_pool": &schema.Schema{
				Type:     schema.TypeSet,
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
				Set: resourceVcdNetworkIPAddressHash,
			},
		},
	}
}

func datasourceVappNetworkRead(d *schema.ResourceData, meta interface{}) error {
	return genericVappNetworkRead(d, meta, "datasource")
}
