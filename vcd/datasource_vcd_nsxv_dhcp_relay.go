package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxvDhcpRelay() *schema.Resource {
	return &schema.Resource{
		Read: resourceVcdNsxvDhcpRelayRead,
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"edge_gateway": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway name for DHCP relay settings",
			},
			"ip_addresses": {
				Computed:    true,
				Type:        schema.TypeSet,
				Description: "A set of IP address of DHCP servers",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"domain_names": {
				Computed:    true,
				Type:        schema.TypeSet,
				Description: "A set of IP domain names of DHCP servers",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"ip_sets": {
				Computed:    true,
				Type:        schema.TypeSet,
				Description: "A set of IP set names which consist DHCP servers",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"relay_agent": {
				Computed: true,
				Type:     schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network_name": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Org network which is to be used for relaying DHCP message to specified servers",
						},
						"gateway_ip_address": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Optional gateway IP address of org network which is to be used for relaying DHCP message to specified servers",
						},
					},
				},
			},
		},
	}
}
