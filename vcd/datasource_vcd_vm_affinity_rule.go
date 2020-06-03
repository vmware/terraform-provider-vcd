package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceVcdVmAffinityRule() *schema.Resource {
	return &schema.Resource{
		Read: datasourceVcdVmAffinityRuleRead,
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
			"name": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"name", "rule_id"},
				Description:  "VM affinity rule name. Used to retrieve a rule only when the name is unique",
			},
			"rule_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "VM affinity rule ID. It's the preferred way of identifying a rule",
			},
			"polarity": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "One of 'Affinity', 'Anti-Affinity'",
			},
			"required": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
				Description: "True if this affinity rule is required. When a rule is mandatory, " +
					"a host failover will not power on the VM if doing so would violate the rule",
			},
			"enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if this affinity rule is enabled",
			},
			"virtual_machine_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Set of VM IDs assigned to this rule",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

// datasourceVcdVmAffinityRuleRead reads a data source VM affinity rule
func datasourceVcdVmAffinityRuleRead(d *schema.ResourceData, meta interface{}) error {
	return genericVcdVmAffinityRuleRead(d, meta, "datasource")
}
