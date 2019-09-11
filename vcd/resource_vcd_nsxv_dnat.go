package vcd

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"

	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdNsxvDnat() *schema.Resource {
	return &schema.Resource{
		Create: natRuleCreator("dnat", setDnatRuleData, getDnatRuleType),
		Read:   natRuleReader("id", "dnat", setDnatRuleData),
		Update: natRuleUpdater("dnat", setDnatRuleData, getDnatRuleType),
		Delete: natRuleDeleter("dnat"),
		Importer: &schema.ResourceImporter{
			State: natRuleImporter("dnat"),
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
			"rule_type": &schema.Schema{ // read only field
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
					"the destination address for DNAT rules.",
			},
			"protocol": &schema.Schema{
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         false,
				DiffSuppressFunc: suppressWordToEmptyString("any"),
				ValidateFunc:     validateCase("lower"),
				Description:      "Protocol. Such as 'tcp', 'udp', 'icmp', 'any'",
			},
			"icmp_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     false,
				ValidateFunc: validateCase("lower"),
				Description:  "ICMP type. Only supported when protocol is ICMP",
			},
			"original_port": &schema.Schema{
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         false,
				DiffSuppressFunc: suppressWordToEmptyString("any"),
				Description:      "Original port. This is the destination port for DNAT rules",
			},
			"translated_address": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "Translated address or address range",
			},
			"translated_port": &schema.Schema{
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         false,
				DiffSuppressFunc: suppressWordToEmptyString("any"),
				Description:      "Translated port",
			},
		},
	}
}

// getDnatRuleType is responsible for getting types.EdgeNatRule for DNAT rule from Terraform
// configuration
func getDnatRuleType(d *schema.ResourceData, edgeGateway govcd.EdgeGateway) (*types.EdgeNatRule, error) {
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
		Protocol:          d.Get("protocol").(string),
		IcmpType:          d.Get("icmp_type").(string),
		OriginalPort:      d.Get("original_port").(string),
		TranslatedAddress: d.Get("translated_address").(string),
		TranslatedPort:    d.Get("translated_port").(string),
	}

	return natRule, nil
}

// setDnatRuleData is responsible for setting DNAT rule data into the statefile
func setDnatRuleData(d *schema.ResourceData, natRule *types.EdgeNatRule, edgeGateway govcd.EdgeGateway) error {
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
	_ = d.Set("original_address", natRule.OriginalAddress)
	_ = d.Set("protocol", natRule.Protocol)
	_ = d.Set("icmp_type", natRule.IcmpType)
	_ = d.Set("original_port", natRule.OriginalPort)
	_ = d.Set("translated_address", natRule.TranslatedAddress)
	_ = d.Set("translated_port", natRule.TranslatedPort)
	_ = d.Set("rule_type", natRule.RuleType)

	return nil
}
