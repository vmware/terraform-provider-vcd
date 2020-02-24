package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceVcdVappNetwork() *schema.Resource {
	return &schema.Resource{
		Read: datasourceVappNetworkRead,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "vApp network name",
			},
			"vapp_name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "vApp to use",
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
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Netmask address for a subnet",
			},
			"gateway": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Gateway of the network",
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

			"guest_vlan_allowed": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if Network allows guest VLAN tagging",
			},
			"org_network": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "org network name to which vapp network is connected",
			},
			"firewall_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "firewall service enabled or disabled. Default is true",
			},
			"nat_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "NAT service enabled or disabled. Default is true",
			},
			"retain_ip_mac_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "NAT service enabled or disabled. Default is true",
			},
			"dhcp_pool": &schema.Schema{
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A range of IPs to issue to virtual machines that don't have a static IP",
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
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A range of IPs permitted to be used as static IPs for virtual machines",
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
