package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxtDistributedFirewallRule() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxtDistributedFirewallRuleRead,

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of VDC Group for Distributed Firewall",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Firewall Rule name",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description is not shown in UI",
			},
			"comment": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Comment that is shown next to rule in UI (VCD 10.3.2+)",
			},
			"direction": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Direction on which Firewall Rule applies (One of 'IN', 'OUT', 'IN_OUT')",
			},
			"ip_protocol": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Firewall Rule Protocol (One of 'IPV4', 'IPV6', 'IPV4_IPV6')",
			},
			"action": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Defines if the rule should 'ALLOW', 'DROP', 'REJECT' matching traffic",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defined if Firewall Rule is active",
			},
			"logging": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defines if matching traffic should be logged",
			},
			"source_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of Source Firewall Group IDs (IP Sets or Security Groups).Empty means 'Any'",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"destination_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of Destination Firewall Group IDs (IP Sets or Security Groups). Empty means 'Any'",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"app_port_profile_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of Application Port Profile IDs. Empty means 'Any'",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"network_context_profile_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of Network Context Profile IDs. Empty means 'Any'",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"source_groups_excluded": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Reverses firewall matching for to match all except Source Groups specified in 'source_ids' (VCD 10.3.2+)",
			},
			"destination_groups_excluded": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Reverses firewall matching for to match all except Destinations Groups specified in 'destination_ids' (VCD 10.3.2+)",
			},
		},
	}
}

func datasourceVcdNsxtDistributedFirewallRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[Distributed Firewall Rule DS read] error retriving Org: %s", err)
	}

	vdcGroup, err := org.GetVdcGroupById(d.Get("vdc_group_id").(string))
	if err != nil {
		return diag.Errorf("[Distributed Firewall Rule DS read] error retrieving VDC Group: %s", err)
	}

	name := d.Get("name").(string)
	rule, err := vdcGroup.GetDistributedFirewallRuleByName(name)
	if err != nil {
		return diag.Errorf("[Distributed Firewall Rule read] error retrieving Firewall Rule By Name '%s': %s", name, err)
	}

	d.SetId(rule.Rule.ID)
	err = setDistributedFirewallRuleData(rule.Rule, d)
	if err != nil {
		return diag.Errorf("error storing data to state: %s", err)
	}

	return nil
}
