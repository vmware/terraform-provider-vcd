package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxtEdgeGatewayStaticRoute() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxtEdgeGatewayStaticRouteRead,

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Required: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"edge_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Edge gateway ID for DHCP forwarding configuration",
			},
		},
	}
}

func datasourceVcdNsxtEdgeGatewayStaticRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
