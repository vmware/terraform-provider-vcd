package vcd

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func datasourceVcdNsxvDnat() *schema.Resource {
	return &schema.Resource{
		Read: natRuleReader("rule_id", "dnat", setDnatRuleData),
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
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
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
			"vnic": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Interface on which the translation is applied.",
			},
			"original_address": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
				Description: "Original address or address range. This is the " +
					"the destination address for DNAT rules.",
			},
			"protocol": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Protocol. One of 'tcp', 'udp', 'icmp', 'any'",
			},
			"icmp_type": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ICMP type. Only supported when protocol is ICMP",
			},
			"original_port": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Original port. This is the source portfor SNAT rules, and the destinationport for DNAT rules.",
			},
			"translated_address": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Translated address or address range",
			},
			"translated_port": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Translated port",
			},
			"dnat_match_source_address": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Source address to match in DNAT rules",
			},
			"dnat_match_source_port": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Source port to match in DNAT rules",
			},
		},
	}
}
