package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxvDnat() *schema.Resource {
	return &schema.Resource{
		Read: natRuleRead("rule_id", "dnat", setDnatRuleData),
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
				Description: "Possible values 'user', 'internal_high'",
			},
			"rule_tag": &schema.Schema{
				Type:        schema.TypeInt,
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
					"the destination address for DNAT rules.",
			},
			"protocol": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Protocol. One of 'tcp', 'udp', 'icmp', 'any'",
			},
			"icmp_type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
				Description: "ICMP type. Only supported when protocol is ICMP. One of `any`, " +
					"`address-mask-request`, `address-mask-reply`, `destination-unreachable`, `echo-request`, " +
					"`echo-reply`, `parameter-problem`, `redirect`, `router-advertisement`, `router-solicitation`, " +
					"`source-quench`, `time-exceeded`, `timestamp-request`, `timestamp-reply`. Default `any`",
			},
			"original_port": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Original port. This is the destinationport for DNAT rules",
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
		},
	}
}
