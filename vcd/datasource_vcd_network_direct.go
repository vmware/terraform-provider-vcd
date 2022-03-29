package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNetworkDirect() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNetworkDirectRead,
		Schema: map[string]*schema.Schema{
			"name": {
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
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Optional description for the network",
			},
			"external_network": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the external network",
			},
			"external_network_gateway": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Gateway of the external network",
			},
			"external_network_netmask": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Net mask of the external network",
			},
			"external_network_dns1": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Main DNS of the external network",
			},
			"external_network_dns2": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Secondary DNS of the external network",
			},
			"external_network_dns_suffix": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "DNS suffix of the external network",
			},
			"href": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Network Hypertext Reference",
			},
			"shared": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defines if this network is shared between multiple VDCs in the Org",
			},
			"filter": {
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
			"metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Key value map of metadata assigned to this network. Key and value can be any string",
			},
		},
	}
}

func datasourceVcdNetworkDirectRead(c context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdNetworkDirectRead(c, d, meta, "datasource")
}
