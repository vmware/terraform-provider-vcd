package vcd

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func datasourceVcdNsxvSnat() *schema.Resource {
	return &schema.Resource{
		Read: natRuleReader("rule_id", "snat", setSnatRuleData),
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
				Description: "NAT rule ID for lookup",
			},
			"network_name": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Org or external network name",
			},
			"network_type": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Network type. One of 'ext', 'org'",
			},
			"rule_type": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Possible values 'user', 'internal_high'.",
			},
			"rule_tag": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Custom rule tag. Contains rule ID if tag was not set",
			},
			"enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defines if the rule is enabled",
			},
			"logging_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defines if logging is enabled for the rule",
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
		},
	}
}
