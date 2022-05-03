package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxvSnat() *schema.Resource {
	return &schema.Resource{
		Read: natRuleRead("rule_id", "snat", setSnatRuleData),
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
				Description: "Edge gateway name in which the NAT rule is located",
			},
			"rule_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "NAT rule ID for lookup",
			},
			"network_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Org or external network name",
			},
			"network_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Network type. One of 'ext', 'org'",
			},
			"rule_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Possible values 'user', 'internal_high'.",
			},
			"rule_tag": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Custom rule tag. Contains rule ID if tag was not set",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defines if the rule is enabled",
			},
			"logging_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defines if logging is enabled for the rule",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NAT rule description",
			},
			"original_address": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "Original address or address range. This is the " +
					"the source address for SNAT rules",
			},
			"translated_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Translated address or address range",
			},
		},
	}
}
