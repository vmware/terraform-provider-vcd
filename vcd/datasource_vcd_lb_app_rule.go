package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdLBAppRule() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdLBAppRuleRead,
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
				Description: "Edge gateway name in which the LB Application Rule is located",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "LB Application Rule name for lookup",
			},
			"script": {
				Computed: true,
				Type:     schema.TypeString,
				Description: "The script for the LB Application Rule. Each line will be " +
					"terminated by newlines (\n)",
			},
		},
	}
}

func datasourceVcdLBAppRuleRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return diag.Errorf(errorUnableToFindEdgeGateway, err)
	}

	readLBAppRule, err := edgeGateway.GetLbAppRuleByName(d.Get("name").(string))
	if err != nil {
		return diag.Errorf("unable to find load balancer application rule with Name %s: %s",
			d.Get("name").(string), err)
	}

	d.SetId(readLBAppRule.ID)
	err = setLBAppRuleData(d, readLBAppRule)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
