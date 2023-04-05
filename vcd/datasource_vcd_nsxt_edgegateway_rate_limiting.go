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
				Description: "Edge gateway ID for rate limiting Configuration",
			},
			"ingress_policy_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Ingress policy ID for rate limiting Configuration",
			},
			"egress_policy_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Egress policy ID for rate limiting Configuration",
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
		return diag.Errorf("[rate limiting (qos) DS read] error retrieving NSX-T Edge Gateway rate limiting (qos): %s", err)
	}

	qosPolicy, err := nsxtEdge.GetQoS()
	if err != nil {
		return diag.Errorf("[rate limiting (qos) DS read] error retrieving NSX-T Edge Gateway rate limiting (qos): %s", err)
	}

	// Rate limiting does not have its own ID - it is a part of Edge Gateway
	d.SetId(edgeGatewayId)
	setNsxtEdgeGatewayQosData(d, qosPolicy)

	return nil
}
