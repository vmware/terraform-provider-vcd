package vcd

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
)

func datasourceVcdNsxvSnat() *schema.Resource {
	return &schema.Resource{
		Read: datasourceVcdNsxvDnatRead,
		Schema: map[string]*schema.Schema{
			"org": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "vCD organization in which the NAT rule is located",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "vCD virtual datacenter in which NAT rule is located",
			},
			"edge_gateway": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway name in which the NAT rule is located",
			},
			"rule_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "LB Application Rule name for lookup",
			},

			"rule_type": &schema.Schema{ // read only field
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Read only. Possible values 'user', 'internal_high'.",
			},
			"rule_tag": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Optional. Allows to set rule custom rule ID.",
			},
			"enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Wether the rule should be enabled. Default 'true'",
			},
			"logging_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Wether logging should be enabled for this rule. Default 'false'",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NAT rule description",
			},
			"original_address": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
				Description: "Original address or address range. This is the " +
					"the source address for SNAT rules",
			},
			"translated_address": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Translated address or address range",
			},
			// SNAT related undocumented
			"snat_match_destination_address": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Destination address to match in SNAT rules",
			},
			"snat_match_destination_port": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Destination port to match in SNAT rules",
			},
		},
	}
}

func datasourceVcdNsxvSnatRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	readNatRule, err := edgeGateway.GetNsxvNatRuleById(d.Get("rule_id").(string))
	if err != nil {
		d.SetId("")
		return fmt.Errorf("unable to find NAT rule with ID %s: %s", d.Id(), err)
	}

	d.SetId(readNatRule.ID)
	return setSnatRuleData(d, readNatRule, edgeGateway)
}
