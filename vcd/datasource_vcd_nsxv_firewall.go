package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxvFirewallRule() *schema.Resource {
	return &schema.Resource{
		Read: resourceVcdNsxvFirewallRuleRead,
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"edge_gateway": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Edge gateway name in which the firewall rule is located",
			},
			"rule_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Firewall rule ID for lookup",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Firewall rule name",
			},
			"rule_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Read only. Possible values 'user', 'internal_high'",
			},
			"rule_tag": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Optional. Allows to set custom rule tag",
			},
			"action": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "'accept' or 'deny'. Default 'accept'",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the rule should be enabled. Default 'true'",
			},
			"logging_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether logging should be enabled for this rule. Default 'false'",
			},
			"source": {
				Computed: true,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"exclude": {
							Type:     schema.TypeBool,
							Computed: true,
							Description: "Rule is applied to traffic coming from all sources " +
								"except for the excluded source. Default 'false'",
						},
						"ip_addresses": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "IP address, CIDR, an IP range, or the keyword 'any'",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"gateway_interfaces": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "'vse', 'internal', 'external' or network name",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"vm_ids": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "Set of VM IDs",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"org_networks": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "Set of org network names",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"ip_sets": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "Set of IP set names",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						// TODO - uncomment once security groups are supported
						// "security_groups": {
						// 	Type:        schema.TypeSet,
						// 	Computed:    true,
						// 	Description: "Set of security group names",
						// 	Elem: &schema.Schema{
						// 		Type: schema.TypeString,
						// 	},
						// },
					},
				},
			},
			"destination": {
				Computed: true,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"exclude": {
							Type:     schema.TypeBool,
							Computed: true,
							Description: "Rule is applied to traffic going to any destinations " +
								"except for the excluded destination. Default 'false'",
						},
						"ip_addresses": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "IP address, CIDR, an IP range, or the keyword 'any'",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"gateway_interfaces": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "'vse', 'internal', 'external' or network name",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"vm_ids": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "Set of VM IDs",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"org_networks": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "Set of org network names",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"ip_sets": {
							Computed:    true,
							Type:        schema.TypeSet,
							Description: "Set of IP set names",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						// TODO - uncomment once security groups are supported
						// "security_groups": {
						// 	Optional:    true,
						// 	Type:        schema.TypeSet,
						// 	Description: "Set of security group names",
						// 	Elem: &schema.Schema{
						// 		Type: schema.TypeString,
						// 	},
						// },
					},
				},
			},
			"service": {
				Computed: true,
				Type:     schema.TypeSet,
				Set:      resourceVcdNsxvFirewallRuleServiceHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"protocol": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"port": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"source_port": {
							Computed: true,
							Type:     schema.TypeString,
						},
					},
				},
			},
		},
	}
}
