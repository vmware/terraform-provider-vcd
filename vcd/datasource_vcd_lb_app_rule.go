package vcd

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
)

func datasourceVcdLBAppRule() *schema.Resource {
	return &schema.Resource{
		Read: datasourceVcdLBAppRuleRead,
		Schema: map[string]*schema.Schema{
			"org": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "vCD organization in which the Application Rule is located",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "vCD virtual datacenter in which the Application Rule is located",
			},
			"edge_gateway": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway name in which the Application Rule is located",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Application Rule name for lookup",
			},
			"script": &schema.Schema{
				Computed: true,
				Type:     schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The script for the Application Rule",
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

	readLBAppRule, err := edgeGateway.ReadLBAppRuleByName(d.Get("name").(string))
	if err != nil {
		return fmt.Errorf("unable to find load balancer application rule with Name %s: %s",
			d.Get("name").(string), err)
	}

	d.SetId(readLBAppRule.ID)
	return setLBAppRuleData(d, readLBAppRule)
}
