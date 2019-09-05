package vcd

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceVcdNsxvNat() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdNsxvNatCreate,
		Read:   resourceVcdNsxvNatRead,
		Update: resourceVcdNsxvNatUpdate,
		Delete: resourceVcdNsxvNatDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdNsxvNatImport,
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
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Nat rule name",
			},

			"rule": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{ // read only field
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},

						"action": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true, // Should it be so? Can one change nat type
							ValidateFunc: validation.StringInSlice([]string{"snat", "dnat"}, false),
							Description:  "Type of NAT. 'snat' or 'dnat'",
						},

						"rule_type": &schema.Schema{ // read only field
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    false,
							Computed:    true,
							Description: "Read only. Possible values 'user', 'internal_high'.",
						},

						"description": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    false,
							Description: "NAT rule description",
						},

						"vnic": &schema.Schema{
							Type:        schema.TypeString,
							Required:    false,
							ForceNew:    false,
							Description: "Interface on which the translating isapplied.",
						},

						"original_address": &schema.Schema{
							Type:     schema.TypeString,
							Required: false,
							ForceNew: false,
							Description: "Original address or address range.This is the source " +
								"address for SNAT rules, and the destination addressfor DNAT rules.",
						},

						"translated_address": &schema.Schema{
							Type:        schema.TypeString,
							Required:    false,
							ForceNew:    false,
							Description: "Translated address or addressrange",
						},

						"enabled": &schema.Schema{
							Type:        schema.TypeBool,
							Required:    false,
							ForceNew:    false,
							Default:     true,
							Description: "Wether the rule should be enabled. Default 'true'",
						},

						"logging_enabled": &schema.Schema{
							Type:        schema.TypeBool,
							Required:    false,
							ForceNew:    false,
							Default:     false,
							Description: "Wether logging should be enabled for this rule. Default 'false'",
						},

						"protocol": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice([]string{"tcp", "udp", "icmp", "any"}, false),
							Description:  "Protocol. One of 'tcp', 'udp', 'icmp', 'any'",
						},

						"icmp_type": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "ICMP type. Only supported when protocol is ICMP",
						},

						"original_port": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "Original port. This is the source portfor SNAT rules, and the destinationport for DNAT rules.",
						},

						"translated_port": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "Translated port",
						},

						// SNAT related
						"snat_match_destination_address": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "Destination address to match inSNAT rules.",
						},
						"snat_match_destination_port": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "Destination port in SNAT rules.",
						},

						// DNAT related
						"dnat_match_source_address": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "Source address to match in DNATrules.",
						},
						"dnat_match_source_port": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "Source port in DNAT rules",
						},
					},
				},
			},
		},
	}
}

func resourceVcdNsxvNatCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	natRule, err := getNatRuleType(d)
	if err != nil {
		return fmt.Errorf("unable to make NAT rule query: %s", err)
	}

	createdNatRule, err := edgeGateway.CreateNsxvNatRule(natRule)
	if err != nil {
		return fmt.Errorf("error creating new NAT rule: %s", err)
	}

	d.SetId(createdNatRule.ID)
	return resourceVcdNsxvNatRead(d, meta)

	return nil
}

func resourceVcdNsxvNatRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

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

func resourceVcdNsxvNatUpdate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	updateNatRule, err := getNatRuleType(d)
	updateNatRule.ID = d.Id()

	if err != nil {
		return fmt.Errorf("could not create NAT rule type for update: %s", err)
	}

	updatedNatRule, err := edgeGateway.UpdateNsxvNatRule(updateNatRule)
	if err != nil {
		return fmt.Errorf("unable to update NAT rule with ID %s: %s", d.Id(), err)
	}

	return setNatRuleData(d, updatedNatRule)
}

func resourceVcdNsxvNatDelete(d *schema.ResourceData, meta interface{}) error {
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

func resourceVcdNsxvNatImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// vcdClient := meta.(*VCDClient)

	// vcdClient.lockParentEdgeGtw(d)
	// defer vcdClient.unLockParentEdgeGtw(d)

	// edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	// if err != nil {
	// 	return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	// }

	return nil, nil
}

func getNatRuleType(d *schema.ResourceData) (*types.EdgeNatRule, error) {

}

func setNatRuleData(d *schema.ResourceData, lBVirtualServer *types.EdgeNatRule) error {

}
