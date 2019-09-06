package vcd

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdNsxvSnat() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdNsxvSnatCreate,
		Read:   resourceVcdNsxvSnatRead,
		Update: resourceVcdNsxvSnatUpdate,
		Delete: resourceVcdNsxvSnatDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdNsxvSnatImport,
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
					"the source address for SNAT rules",
			},
			"translated_address": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "Translated address or address range",
			},
			// SNAT related undocumented
			"snat_match_destination_address": &schema.Schema{
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         false,
				DiffSuppressFunc: suppressWordToEmptyString("any"),
				Description:      "Source address to match in DNAT rules",
			},
			"snat_match_destination_port": &schema.Schema{
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppressWordToEmptyString("any"),
				Description:      "Source port to match in DNAT rules",
			},
		},
	}
}

func resourceVcdNsxvSnatCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	natRule := getSnatRuleType(d)

	natRule.Action = "snat"

	createdNatRule, err := edgeGateway.CreateNsxvNatRule(natRule)
	if err != nil {
		return fmt.Errorf("error creating new NAT rule: %s", err)
	}

	d.SetId(createdNatRule.ID)
	return resourceVcdNsxvSnatRead(d, meta)
}

func resourceVcdNsxvSnatRead(d *schema.ResourceData, meta interface{}) error {
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

	return setSnatRuleData(d, readNatRule)
}

func resourceVcdNsxvSnatUpdate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	updateNatRule := getSnatRuleType(d)
	updateNatRule.ID = d.Id()

	updateNatRule.Action = "snat"

	updatedNatRule, err := edgeGateway.UpdateNsxvNatRule(updateNatRule)
	if err != nil {
		return fmt.Errorf("unable to update NAT rule with ID %s: %s", d.Id(), err)
	}

	return setSnatRuleData(d, updatedNatRule)
}

func resourceVcdNsxvSnatDelete(d *schema.ResourceData, meta interface{}) error {
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

func resourceVcdNsxvSnatImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

func getSnatRuleType(d *schema.ResourceData) *types.EdgeNatRule {
	natRule := &types.EdgeNatRule{
		RuleTag:                     d.Get("rule_tag").(string),
		Enabled:                     d.Get("enabled").(bool),
		LoggingEnabled:              d.Get("logging_enabled").(bool),
		Description:                 d.Get("description").(string),
		Vnic:                        d.Get("vnic").(string),
		OriginalAddress:             d.Get("original_address").(string),
		TranslatedAddress:           d.Get("translated_address").(string),
		SnatMatchDestinationAddress: d.Get("snat_match_destination_address").(string),
		SnatMatchDestinationPort:    d.Get("snat_match_destination_port").(string),
	}

	return natRule
}

func setSnatRuleData(d *schema.ResourceData, natRule *types.EdgeNatRule) error {
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

	err = d.Set("translated_address", natRule.TranslatedAddress)
	if err != nil {
		return fmt.Errorf("unable to set 'translated_address'")
	}

	err = d.Set("rule_type", natRule.RuleType)
	if err != nil {
		return fmt.Errorf("unable to set 'rule_type'")
	}

	err = d.Set("snat_match_destination_port", natRule.SnatMatchDestinationPort)
	if err != nil {
		return fmt.Errorf("unable to set 'snat_match_destination_port'")
	}

	err = d.Set("snat_match_destination_address", natRule.SnatMatchDestinationAddress)
	if err != nil {
		return fmt.Errorf("unable to set 'snat_match_destination_address'")
	}

	return nil
}
