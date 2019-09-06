package vcd

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdNsxvDnat() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdNsxvDnatCreate,
		Read:   resourceVcdNsxvDnatRead,
		Update: resourceVcdNsxvDnatUpdate,
		Delete: resourceVcdNsxvDnatDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdNsxvDnatImport,
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
				Description: "Wether the rule should be enabled. Default 'true'",
			},
			"logging_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Default:     false,
				Description: "Wether logging should be enabled for this rule. Default 'false'",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "NAT rule description",
			},
			"vnic": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Computed:    true,
				Description: "Interface on which the translation is applied.",
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
				ValidateFunc:     validation.StringInSlice([]string{"tcp", "udp", "icmp", "any"}, false),
				Description:      "Protocol. One of 'tcp', 'udp', 'icmp', 'any'",
			},
			"icmp_type": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "ICMP type. Only supported when protocol is ICMP",
			},
			"original_port": &schema.Schema{
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
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
				ForceNew:         true,
				DiffSuppressFunc: suppressWordToEmptyString("any"),
				Description:      "Translated port",
			},
			// DNAT related undocumented
			"dnat_match_source_address": &schema.Schema{
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         false,
				DiffSuppressFunc: suppressWordToEmptyString("any"),
				Description:      "Source address to match in DNAT rules",
			},
			"dnat_match_source_port": &schema.Schema{
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppressWordToEmptyString("any"),
				Description:      "Source port to match in DNAT rules",
			},
		},
	}
}

// suppressWordToEmptyString is a DiffSuppressFunc which ignore the change from word to empty string "".
// This is useful when API returns some default value but it is not set (and not sent via API) in config.
func suppressWordToEmptyString(word string) schema.SchemaDiffSuppressFunc {
	return func(k string, old string, new string, d *schema.ResourceData) bool {
		if old == word && new == "" {
			return true
		}
		return false
	}
}

func resourceVcdNsxvDnatCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	natRule := getNatRuleType(d)

	natRule.Action = "dnat"

	createdNatRule, err := edgeGateway.CreateNsxvNatRule(natRule)
	if err != nil {
		return fmt.Errorf("error creating new NAT rule: %s", err)
	}

	d.SetId(createdNatRule.ID)
	return resourceVcdNsxvDnatRead(d, meta)
}

func resourceVcdNsxvDnatRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	readNatRule, err := edgeGateway.GetNsxvNatRuleById(d.Id())
	if err != nil {
		d.SetId("")
		return fmt.Errorf("unable to find NAT rule with ID %s: %s", d.Id(), err)
	}

	return setNatRuleData(d, readNatRule)
}

func resourceVcdNsxvDnatUpdate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	updateNatRule := getNatRuleType(d)
	updateNatRule.ID = d.Id()

	updateNatRule.Action = "dnat"

	updatedNatRule, err := edgeGateway.UpdateNsxvNatRule(updateNatRule)
	if err != nil {
		return fmt.Errorf("unable to update NAT rule with ID %s: %s", d.Id(), err)
	}

	return setNatRuleData(d, updatedNatRule)
}

func resourceVcdNsxvDnatDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	err = edgeGateway.DeleteNsxvNatRuleById(d.Id())
	if err != nil {
		return fmt.Errorf("error deleting NAT rule: %s", err)
	}

	d.SetId("")
	return nil
}

func resourceVcdNsxvDnatImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ".")
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("resource name must be specified in such way org.vdc.edge-gw.rule-id")
	}
	orgName, vdcName, edgeName, natRuleId := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)
	edgeGateway, err := vcdClient.GetEdgeGateway(orgName, vdcName, edgeName)
	if err != nil {
		return nil, fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	readNatRule, err := edgeGateway.GetNsxvNatRuleById(natRuleId)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("unable to find NAT rule with id %s: %s",
			d.Id(), err)
	}

	d.Set("org", orgName)
	d.Set("vdc", vdcName)
	d.Set("edge_gateway", edgeName)

	d.SetId(readNatRule.ID)
	return []*schema.ResourceData{d}, nil
}

func getNatRuleType(d *schema.ResourceData) *types.EdgeNatRule {
	natRule := &types.EdgeNatRule{
		RuleTag:                d.Get("rule_tag").(string),
		Enabled:                d.Get("enabled").(bool),
		LoggingEnabled:         d.Get("logging_enabled").(bool),
		Description:            d.Get("description").(string),
		Vnic:                   d.Get("vnic").(string),
		OriginalAddress:        d.Get("original_address").(string),
		Protocol:               d.Get("protocol").(string),
		IcmpType:               d.Get("icmp_type").(string),
		OriginalPort:           d.Get("original_port").(string),
		TranslatedAddress:      d.Get("translated_address").(string),
		TranslatedPort:         d.Get("translated_port").(string),
		DnatMatchSourceAddress: d.Get("dnat_match_source_address").(string),
		DnatMatchSourcePort:    d.Get("dnat_match_source_port").(string),
	}

	return natRule
}

func setNatRuleData(d *schema.ResourceData, natRule *types.EdgeNatRule) error {
	err := d.Set("rule_tag", natRule.RuleTag)
	if err != nil {
		return fmt.Errorf("unable to set 'rule_tag'")
	}

	err = d.Set("enabled", natRule.Enabled)
	if err != nil {
		return fmt.Errorf("unable to set 'enabled'")
	}

	err = d.Set("logging_enabled", natRule.LoggingEnabled)
	if err != nil {
		return fmt.Errorf("unable to set 'logging_enabled'")
	}

	err = d.Set("description", natRule.Description)
	if err != nil {
		return fmt.Errorf("unable to set 'description'")
	}

	err = d.Set("vnic", natRule.Vnic)
	if err != nil {
		return fmt.Errorf("unable to set 'vnic'")
	}

	err = d.Set("original_address", natRule.OriginalAddress)
	if err != nil {
		return fmt.Errorf("unable to set 'original_address'")
	}

	err = d.Set("protocol", natRule.Protocol)
	if err != nil {
		return fmt.Errorf("unable to set 'protocol'")
	}

	err = d.Set("icmp_type", natRule.IcmpType)
	if err != nil {
		return fmt.Errorf("unable to set 'icmp_type'")
	}

	err = d.Set("original_port", natRule.OriginalPort)
	if err != nil {
		return fmt.Errorf("unable to set 'original_port'")
	}

	err = d.Set("translated_address", natRule.TranslatedAddress)
	if err != nil {
		return fmt.Errorf("unable to set 'translated_address'")
	}

	err = d.Set("translated_port", natRule.TranslatedPort)
	if err != nil {
		return fmt.Errorf("unable to set 'translated_port'")
	}

	err = d.Set("rule_type", natRule.RuleType)
	if err != nil {
		return fmt.Errorf("unable to set 'rule_type'")
	}

	err = d.Set("dnat_match_source_port", natRule.DnatMatchSourcePort)
	if err != nil {
		return fmt.Errorf("unable to set 'dnat_match_source_port'")
	}

	err = d.Set("dnat_match_source_address", natRule.DnatMatchSourceAddress)
	if err != nil {
		return fmt.Errorf("unable to set 'dnat_match_source_address'")
	}

	return nil
}
