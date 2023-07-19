package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxtEdgegatewayDhcpForwarding() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxtEdgegatewayDhcpForwardingRead,
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"edge_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Edge gateway ID for DHCP forwarding configuration",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Status of DHCP Forwarding for the Edge Gateway",
			},
			"dhcp_servers": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "IP addresses of the DHCP servers",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func datasourceVcdNsxtEdgegatewayDhcpForwardingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdNsxtEdgegatewayDhcpForwardingRead(ctx, d, meta, "datasource")
}
