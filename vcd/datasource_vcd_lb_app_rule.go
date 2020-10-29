package vcd

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdLBAppRule() *schema.Resource {
	return &schema.Resource{
		Read: datasourceVcdLBAppRuleRead,
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
				Description: "Edge gateway name in which the LB Application Rule is located",
			},
			"name": &schema.Schema{
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

func datasourceVcdLBAppRuleRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	readLBAppRule, err := edgeGateway.GetLbAppRuleByName(d.Get("name").(string))
	if err != nil {
		return fmt.Errorf("unable to find load balancer application rule with Name %s: %s",
			d.Get("name").(string), err)
	}

	d.SetId(readLBAppRule.ID)
	return setLBAppRuleData(d, readLBAppRule)
}
