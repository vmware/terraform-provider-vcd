package vcd

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

func datasourceVcdNetworkDirect() *schema.Resource {
	return &schema.Resource{
		Read: datasourceVcdNetworkDirectRead,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A unique name for this network (optional if 'filter' is used)",
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
			"external_network": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the external network",
			},
			"external_network_gateway": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Gateway of the external network",
			},
			"external_network_netmask": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Net mask of the external network",
			},
			"external_network_dns1": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Main DNS of the external network",
			},
			"external_network_dns2": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Secondary DNS of the external network",
			},
			"external_network_dns_suffix": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "DNS suffix of the external network",
			},
			"href": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Network Hypertext Reference",
			},
			"shared": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defines if this network is shared between multiple VDCs in the Org",
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

func datasourceVcdNetworkDirectRead(d *schema.ResourceData, meta interface{}) error {
	return genericVcdNetworkDirectRead(d, meta, "datasource")
}
