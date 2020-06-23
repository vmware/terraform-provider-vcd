package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceVcdNetworkIsolated() *schema.Resource {
	return &schema.Resource{
		Read: datasourceVcdNetworkIsolatedRead,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"name", "filter"},
				Description:  "A unique name for this network (optional if 'filter' is used)",
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
				Description: "The netmask for the new network",
			},
			"gateway": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The gateway for this network",
			},
			"dns1": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "First DNS server to use",
			},
			"dns2": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Second DNS server to use",
			},
			"dns_suffix": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A FQDN for the virtual machines on this network",
			},
			"href": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Network Hyper Reference",
			},
			"shared": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defines if this network is shared between multiple VDCs in the Org",
			},
			"dhcp_pool": &schema.Schema{
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A range of IPs to issue to virtual machines that don't have a static IP",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_address": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The first address in the IP Range",
						},
						"end_address": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The final address in the IP Range",
						},
						"default_lease_time": &schema.Schema{
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The default DHCP lease time to use",
						},
						"max_lease_time": &schema.Schema{
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The maximum DHCP lease time to use",
						},
					},
				},
				Set: resourceVcdNetworkIsolatedDhcpPoolHash,
			},
			"static_ip_pool": &schema.Schema{
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A range of IPs permitted to be used as static IPs for virtual machines",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_address": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The first address in the IP Range",
						},
						"end_address": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The final address in the IP Range",
						},
					},
				},
				Set: resourceVcdNetworkStaticIpPoolHash,
			},
			"filter": &schema.Schema{
				Type:        schema.TypeList,
				MaxItems:    1,
				MinItems:    1,
				Optional:    true,
				Description: "Criteria for retrieving a network by various attributes",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name_regex": elementNameRegex,
						"ip":         elementIp,
						"metadata":   elementMetadata,
					},
				},
			},
		},
	}
}

func datasourceVcdNetworkIsolatedRead(d *schema.ResourceData, meta interface{}) error {
	return genericVcdNetworkIsolatedRead(d, meta, "datasource")
}
