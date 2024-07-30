package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNetworkPool() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceNetworkPoolRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of network pool.",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of the network pool (one of `GENEVE`, `VLAN`, `PORTGROUP_BACKED`)",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the network pool",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the network pool",
			},
			"promiscuous_mode": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the network pool is in promiscuous mode",
			},
			"total_backings_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total number of backings",
			},
			"used_backings_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of used backings",
			},
			"network_provider_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Id of the network provider (either vCenter or NSX-T manager)",
			},
			"network_provider_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the network provider",
			},
			"network_provider_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of network provider",
			},
			"backing": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The components used by the network pool",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"transport_zone": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Transport Zone Backing",
							Elem:        resourceNetworkPoolBacking("resource"),
						},
						"port_group": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Port Group backing",
							Elem:        resourceNetworkPoolBacking("resource"),
						},
						"distributed_switch": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Distributed switch backing",
							Elem:        resourceNetworkPoolBacking("resource"),
						},
						"range_id": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Distributed Switch ID ranges (used with VLAN backing)",
							Elem:        resourceNetworkPoolVlanIdRange,
						},
					},
				},
			},
		},
	}
}

func datasourceNetworkPoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericNetworkPoolRead(ctx, d, meta, "datasource")
}
