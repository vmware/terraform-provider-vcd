package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
			"edge_gateway": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway name for DHCP relay settings",
			},
			"ip_addresses": {
				Computed:    true,
				Type:        schema.TypeSet,
				Description: "IP addresses ",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"domain_names": {
				Computed:    true,
				Type:        schema.TypeSet,
				Description: "IP addresses ",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"ip_sets": {
				Computed:    true,
				Type:        schema.TypeSet,
				Description: "IP addresses ",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"relay_agent": {
				Computed: true,
				Type:     schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"org_network": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"gateway_ip_address": {
							Computed: true,
							Type:     schema.TypeString,
						},
					},
				},
			},
		},
	}
}
