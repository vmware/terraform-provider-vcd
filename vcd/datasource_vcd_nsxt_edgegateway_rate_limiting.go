package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxtEdgegatewayRateLimiting() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxtEdgegatewayRateLimitingRead,

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
				Description: "Edge gateway ID for Rate Limiting (QoS) configuration",
			},
			"ingress_profile_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Ingress profile ID for Rate Limiting (QoS) configuration",
			},
			"egress_profile_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Egress profile ID for Rate Limiting (QoS) configuration",
			},
		},
	}
}

func datasourceVcdNsxtEdgegatewayRateLimitingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[rate limiting (QoS) DS read] error retrieving NSX-T Edge Gateway Rate Limiting (QoS): %s", err)
	}

	rateLimitConfig, err := nsxtEdge.GetQoS()
	if err != nil {
		return diag.Errorf("[rate limiting (QoS) DS read] error retrieving NSX-T Edge Gateway Rate Limiting (QoS): %s", err)
	}

	// Rate limiting does not have its own ID - it is a part of Edge Gateway
	d.SetId(edgeGatewayId)
	setNsxtEdgeGatewayRateLimitingData(d, rateLimitConfig)

	return nil
}
