package vcd

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"

	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdNsxvSnat() *schema.Resource {
	return &schema.Resource{
		Create: natRuleCreator("snat", setSnatRuleData, getSnatRuleType),
		Read:   natRuleReader("id", "snat", setSnatRuleData),
		Update: natRuleUpdater("snat", setSnatRuleData, getSnatRuleType),
		Delete: natRuleDeleter("snat"),
		Importer: &schema.ResourceImporter{
			State: natRuleImporter("snat"),
		},

		Schema: map[string]*schema.Schema{
			"org": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "vCD organization in which NAT Rule is located",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "vCD virtual datacenter in which NAT Rule is located",
			},
			"edge_gateway": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway name in which NAT Rule is located",
			},
			"network_name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Edge gateway name in which NAT Rule is located",
			},
			"network_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "org",
				ValidateFunc: validation.StringInSlice([]string{"ext", "org"}, false),
			},
			"rule_type": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Computed:    true,
				Description: "Read only. Possible values 'user', 'internal_high'",
			},
			"rule_tag": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
				Description: "Optional. Allows to set rule custom rule ID",
			},
			"enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Default:     true,
				Description: "Whether the rule should be enabled. Default 'true'",
			},
			"logging_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Default:     false,
				Description: "Whether logging should be enabled for this rule. Default 'false'",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "NAT rule description",
			},
			"original_address": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
				Description: "Original address or address range. This is the " +
					"the source address for SNAT rules",
			},
			"translated_address": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "Translated address or address range",
			},
		},
	}
}

// getSnatRuleType is responsible for getting types.EdgeNatRule for SNAT rule from Terraform
// configuration
func getSnatRuleType(d *schema.ResourceData, edgeGateway govcd.EdgeGateway) (*types.EdgeNatRule, error) {
	networkName := d.Get("network_name").(string)
	networkType := d.Get("network_type").(string)

	vnicIndex, err := getvNicIndexFromNetworkNameType(networkName, networkType, edgeGateway)
	if err != nil {
		return nil, err
	}

	natRule := &types.EdgeNatRule{
		RuleTag:           d.Get("rule_tag").(string),
		Enabled:           d.Get("enabled").(bool),
		LoggingEnabled:    d.Get("logging_enabled").(bool),
		Description:       d.Get("description").(string),
		Vnic:              vnicIndex,
		OriginalAddress:   d.Get("original_address").(string),
		TranslatedAddress: d.Get("translated_address").(string),
	}

	return natRule, nil
}

// setSnatRuleData is responsible for setting SNAT rule data into the statefile
func setSnatRuleData(d *schema.ResourceData, natRule *types.EdgeNatRule, edgeGateway govcd.EdgeGateway) error {
	networkName, resourceNetworkType, err := getNetworkNameTypeFromVnicIndex(*natRule.Vnic, edgeGateway)
	if err != nil {
		return err
	}

	_ = d.Set("network_type", resourceNetworkType)
	_ = d.Set("network_name", networkName)
	_ = d.Set("rule_tag", natRule.RuleTag)
	_ = d.Set("enabled", natRule.Enabled)
	_ = d.Set("logging_enabled", natRule.LoggingEnabled)
	_ = d.Set("description", natRule.Description)
	_ = d.Set("vnic", natRule.Vnic)
	_ = d.Set("original_address", natRule.OriginalAddress)
	_ = d.Set("translated_address", natRule.TranslatedAddress)
	_ = d.Set("rule_type", natRule.RuleType)

	return nil
}
