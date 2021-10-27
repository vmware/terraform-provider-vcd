package vcd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdNsxtNatRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdNsxtNatRuleCreate,
		ReadContext:   resourceVcdNsxtNatRuleRead,
		UpdateContext: resourceVcdNsxtNatRuleUpdate,
		DeleteContext: resourceVcdNsxtNatRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdNsxtNatRuleImport,
		},

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
			"edge_gateway_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway name in which NAT Rule is located",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of NAT rule",
			},
			"rule_type": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "Rule type - one of 'DNAT', 'NO_DNAT', 'SNAT', 'NO_SNAT', 'REFLEXIVE'",
				ValidateFunc: validation.StringInSlice([]string{"DNAT", "NO_DNAT", "SNAT", "NO_SNAT", "REFLEXIVE"}, false),
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of NAT rule",
			},
			"external_address": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "IP address or CIDR of external network",
			},
			"internal_address": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "IP address or CIDR of the virtual machines for which you are configuring NAT",
			},
			"app_port_profile_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Application Port Profile to apply for this rule",
			},
			"dnat_external_port": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "For DNAT only. Enter a port into which the DNAT rule is translating for the packets inbound to the virtual machines.",
			},
			"snat_destination_address": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "For SNAT only. If you want the rule to apply only for traffic to a specific domain, enter an IP address for this domain or an IP address range in CIDR format.",
			},
			"logging": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable logging when this rule is applied",
			},
			"enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Enables or disables this rule",
			},
			"firewall_match": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "VCD 10.2.2+ Determines how the firewall matches the address during NATing if firewall stage is not skipped. One of 'MATCH_INTERNAL_ADDRESS', 'MATCH_EXTERNAL_ADDRESS', 'BYPASS'",
				ValidateFunc: validation.StringInSlice([]string{"MATCH_INTERNAL_ADDRESS", "MATCH_EXTERNAL_ADDRESS", "BYPASS"}, false),
			},
			"priority": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "VCD 10.2.2+ If an address has multiple NAT rules, the rule with the highest priority is applied. A lower value means a higher precedence for this rule.",
			},
		},
	}
}

func resourceVcdNsxtNatRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	orgName := d.Get("org").(string)
	vdcName := d.Get("vdc").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, vdcName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("error retrieving Edge Gateway: %s", err)
	}

	nsxtNatRule, err := getNsxtNatType(d, vcdClient)
	if err != nil {
		return diag.Errorf("error getting NSX-T NAT rule type: %s", err)
	}

	rule, err := nsxtEdge.CreateNatRule(nsxtNatRule)
	if err != nil {

		return diag.Errorf("error creating NSX-T NAT rule: %s", err)
	}
	d.SetId(rule.NsxtNatRule.ID)

	return resourceVcdNsxtNatRuleRead(ctx, d, meta)
}

func resourceVcdNsxtNatRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	orgName := d.Get("org").(string)
	vdcName := d.Get("vdc").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, vdcName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("error retrieving Edge Gateway: %s", err)
	}

	nsxtNatRule, err := getNsxtNatType(d, vcdClient)
	if err != nil {
		return diag.Errorf("error getting NSX-T NAT rule type: %s", err)
	}

	existingRule, err := nsxtEdge.GetNatRuleById(d.Id())
	if err != nil {
		return diag.Errorf("unable to find NSX-T NAT rule: %s", err)
	}

	// Inject ID for update
	nsxtNatRule.ID = existingRule.NsxtNatRule.ID
	_, err = existingRule.Update(nsxtNatRule)
	if err != nil {
		return diag.Errorf("error updating NSX-T NAT rule: %s", err)
	}

	return resourceVcdNsxtNatRuleRead(ctx, d, meta)
}

func resourceVcdNsxtNatRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgName := d.Get("org").(string)
	vdcName := d.Get("vdc").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, vdcName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("error retrieving Edge Gateway: %s", err)
	}

	existingRule, err := nsxtEdge.GetNatRuleById(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
		}
		return diag.Errorf("unable to find NSX-T NAT rule: %s", err)
	}

	err = setNsxtNatRuleData(existingRule.NsxtNatRule, d, vcdClient)
	if err != nil {
		return diag.Errorf("error storing NSX-T NAT rule in statefile: %s", err)
	}

	return nil
}

func resourceVcdNsxtNatRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	orgName := d.Get("org").(string)
	vdcName := d.Get("vdc").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, vdcName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("error retrieving Edge Gateway: %s", err)
	}

	rule, err := nsxtEdge.GetNatRuleById(d.Id())
	if err != nil {
		return diag.Errorf("error finding NSX-T NAT Rule: %s", err)
	}

	err = rule.Delete()
	if err != nil {
		return diag.Errorf("error deleting NSX-T NAT rule: %s", err)
	}

	d.SetId("")

	return nil
}

func resourceVcdNsxtNatRuleImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-name.edge_gateway_name.nat_rule_name")
	}
	orgName, vdcName, edgeGatewayName, natRuleIdentifier := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)
	org, err := vcdClient.GetAdminOrg(orgName)
	if err != nil {
		return nil, fmt.Errorf("unable to find Org %s: %s", orgName, err)
	}
	vdc, err := org.GetVDCByName(vdcName, false)
	if err != nil {
		return nil, fmt.Errorf("unable to find VDC %s: %s", vdcName, err)
	}

	if !vdc.IsNsxt() {
		return nil, errors.New("vcd_nsxt_nat_rule is only supported by NSX-T VDCs")
	}

	edgeGateway, err := vdc.GetNsxtEdgeGatewayByName(edgeGatewayName)
	if err != nil {
		return nil, fmt.Errorf("unable to find Edge Gateway '%s': %s", edgeGatewayName, err)
	}

	natRule, err := edgeGateway.GetNatRuleByName(natRuleIdentifier)
	if govcd.ContainsNotFound(err) {
		natRule, err = edgeGateway.GetNatRuleById(natRuleIdentifier)
	}

	// Error occurred and it is not ErrorEntityNotFound. This means - more than one rule found and we should dump a list
	// of rules with their IDs so that one can pick ID
	if err != nil && !govcd.ContainsNotFound(err) {
		allRules, err2 := edgeGateway.GetAllNatRules(nil)
		if err2 != nil {
			return nil, fmt.Errorf("error getting list of all NAT rules: %s", err)
		}
		dumpNatRulesToScreen(natRuleIdentifier, allRules)
	}

	if err != nil {
		return nil, fmt.Errorf("unable to find NAT Rule '%s': %s", natRuleIdentifier, err)
	}

	dSet(d, "org", orgName)
	dSet(d, "vdc", vdcName)
	dSet(d, "edge_gateway_id", edgeGateway.EdgeGateway.ID)
	d.SetId(natRule.NsxtNatRule.ID)

	return []*schema.ResourceData{d}, nil
}

func getNsxtNatType(d *schema.ResourceData, client *VCDClient) (*types.NsxtNatRule, error) {
	// REFLEXIVE rule_type is only supported in VCD 10.3+
	if d.Get("rule_type").(string) == types.NsxtNatRuleTypeReflexive && client.Client.APIVCDMaxVersionIs("< 36.0") {
		return nil, fmt.Errorf("rule_type 'REFLEXIVE' can only be used for VCD 10.3+")
	}

	firewallMatch, firewallMatchOk := d.GetOk("firewall_match")
	priority, priorityOk := d.GetOk("priority")

	// Only supported in VCD 10.2.2+ (API V35.2) and throw immediate error if used with older versions as API error is
	// opaque
	if (firewallMatchOk || priorityOk) && client.Client.APIVCDMaxVersionIs("< 35.2") {
		return nil, fmt.Errorf("firewall_match and priority fields can only be set for VCD 10.2.2+")
	}

	nsxtNatRule := &types.NsxtNatRule{
		Name:                     d.Get("name").(string),
		Description:              d.Get("description").(string),
		Enabled:                  d.Get("enabled").(bool),
		ExternalAddresses:        d.Get("external_address").(string),
		InternalAddresses:        d.Get("internal_address").(string),
		SnatDestinationAddresses: d.Get("snat_destination_address").(string),
		Logging:                  d.Get("logging").(bool),
	}

	if client.Client.APIVCDMaxVersionIs(">= 36.0") {
		nsxtNatRule.Type = d.Get("rule_type").(string)
	} else {
		nsxtNatRule.RuleType = d.Get("rule_type").(string)
	}

	// DnatExternalPort field was replacement for InternalPort field. When using API < 35.2 InternalPort field should be
	// used.
	if client.Client.APIVCDMaxVersionIs("< 35.2") {
		nsxtNatRule.InternalPort = d.Get("dnat_external_port").(string)
	} else {
		nsxtNatRule.DnatExternalPort = d.Get("dnat_external_port").(string)
	}

	if appPortProf, ok := d.GetOk("app_port_profile_id"); ok {
		nsxtNatRule.ApplicationPortProfile = &types.OpenApiReference{ID: appPortProf.(string)}
	}

	if firewallMatchOk {
		nsxtNatRule.FirewallMatch = firewallMatch.(string)
	}

	if priorityOk {
		nsxtNatRule.Priority = takeIntPointer(priority.(int))
	}
	return nsxtNatRule, nil
}

func setNsxtNatRuleData(rule *types.NsxtNatRule, d *schema.ResourceData, client *VCDClient) error {
	dSet(d, "name", rule.Name)
	dSet(d, "description", rule.Description)
	dSet(d, "external_address", rule.ExternalAddresses)
	dSet(d, "internal_address", rule.InternalAddresses)
	dSet(d, "snat_destination_address", rule.SnatDestinationAddresses)
	dSet(d, "logging", rule.Logging)
	dSet(d, "enabled", rule.Enabled)
	if rule.ApplicationPortProfile != nil {
		dSet(d, "app_port_profile_id", rule.ApplicationPortProfile.ID)
	}

	// Some specific changes in API V35.2 (VCD 10.2.2+)
	if client.Client.APIVCDMaxVersionIs(">= 35.2") {
		// firewall_match and priority are new fields introduces in VCD 10.2.2
		dSet(d, "firewall_match", rule.FirewallMatch)
		dSet(d, "priority", rule.Priority)
		// Before v35.2 the field to use is rule.InternalPort
		dSet(d, "dnat_external_port", rule.DnatExternalPort)
	}

	if client.Client.APIVCDMaxVersionIs("< 35.2") {
		dSet(d, "dnat_external_port", rule.InternalPort)
	}

	// API V36.0+ uses Type field instead of RuleType
	if client.Client.APIVCDMaxVersionIs(">= 36.0") {
		dSet(d, "rule_type", rule.Type)
	} else {
		dSet(d, "rule_type", rule.RuleType)
	}

	return nil
}

// dumpNatRulesToScreen is a helper for import. NAT rules don't enforce name uniqueness therefore it may be that user
// specifies a rule with the same name. In that case NAT rule details and their IDs are listed and the one will be able
// to import by using ID.
func dumpNatRulesToScreen(name string, allRules []*govcd.NsxtNatRule) {
	stdout := getTerraformStdout()

	fmt.Fprintf(stdout, "# The following NAT rules with Name '%s' are available\n", name)
	fmt.Fprintf(stdout, "# Please use ID instead of Name in import path to pick exact rule\n")

	w := tabwriter.NewWriter(stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "ID\tName\tRule Type\tInternal Address\tExternal Address")
	for _, rule := range allRules {
		if rule.NsxtNatRule.Name != name {
			continue
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			rule.NsxtNatRule.ID, rule.NsxtNatRule.Name, rule.NsxtNatRule.RuleType, rule.NsxtNatRule.InternalAddresses,
			rule.NsxtNatRule.ExternalAddresses)
	}

	w.Flush()
}
