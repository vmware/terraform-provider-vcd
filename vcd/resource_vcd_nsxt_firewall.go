package vcd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdNsxtFirewall() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdNsxtFirewallCreateUpdate,
		ReadContext:   resourceVcdNsxtFirewallRead,
		UpdateContext: resourceVcdNsxtFirewallCreateUpdate, // Update is exactly the same operation as create
		DeleteContext: resourceVcdNsxtFirewallDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdNsxtFirewallImport,
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
				Description: "Edge Gateway ID in which Firewall Rule are located",
			},
			"rule": &schema.Schema{
				Type:        schema.TypeList, // Firewall rule order matters
				Required:    true,
				MinItems:    1,
				Description: "Ordered list of firewall rules",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Firewall Rule ID",
						},
						"name": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "Firewall Rule name",
						},
						"direction": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Direction on which Firewall Rule applies (One of 'IN', 'OUT', 'IN_OUT')",
							ValidateFunc: validation.StringInSlice([]string{"IN", "OUT", "IN_OUT"}, false),
						},
						"ip_protocol": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Firewall Rule Protocol (One of 'IPV4', 'IPV6', 'IPV4_IPV6')",
							ValidateFunc: validation.StringInSlice([]string{"IPV4", "IPV6", "IPV4_IPV6"}, false),
						},
						"action": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Defines if the rule should 'ALLOW' or 'DROP' matching traffic",
							ValidateFunc: validation.StringInSlice([]string{"ALLOW", "DROP"}, false),
						},
						"enabled": &schema.Schema{
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Defined if Firewall Rule is active",
						},
						"logging": &schema.Schema{
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Defines if matching traffic should be logged",
						},
						"source_ids": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "A set of Source Firewall Group IDs (IP Sets or Security Groups). Leaving it empty means 'Any'",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"destination_ids": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "A set of Destination Firewall Group IDs (IP Sets or Security Groups). Leaving it empty means 'Any'",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"app_port_profile_ids": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "A set of Application Port Profile IDs. Leaving it empty means 'Any'",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

// resourceVcdNsxtFirewallCreateUpdate is the same function used for both - Create and Update
func resourceVcdNsxtFirewallCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	firewallRulesType := getNsxtFirewallTypes(d)
	firewallContainer := &types.NsxtFirewallRuleContainer{
		UserDefinedRules: firewallRulesType,
	}

	_, err = nsxtEdge.UpdateNsxtFirewall(firewallContainer)
	if err != nil {
		return diag.Errorf("error creating NSX-T Firewall Rules: %s", err)
	}

	// ID is stored as Edge Gateway ID - because this is a "container" for all firewall rules at once and each child
	// TypeSet element will have a computed ID field for each rule
	d.SetId(edgeGatewayId)

	return resourceVcdNsxtFirewallRead(ctx, d, meta)
}

func resourceVcdNsxtFirewallRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgName := d.Get("org").(string)
	vdcName := d.Get("vdc").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, vdcName, edgeGatewayId)
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error retrieving NSX-T Edge Gateway: %s", err)
	}

	fwRules, err := nsxtEdge.GetNsxtFirewall()
	if err != nil {
		return diag.Errorf("error retrieving NSX-T Firewall Rules: %s", err)
	}

	err = setNsxtFirewallData(fwRules.NsxtFirewallRuleContainer.UserDefinedRules, d, edgeGatewayId)
	if err != nil {
		return diag.Errorf("error storing NSX-T Firewall data to schema: %s", err)
	}

	return nil
}

func resourceVcdNsxtFirewallDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	orgName := d.Get("org").(string)
	vdcName := d.Get("vdc").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, vdcName, edgeGatewayId)

	if err != nil {
		return diag.Errorf("error retrieving NSX-T Edge Gateway: %s", err)
	}

	allRules, err := nsxtEdge.GetNsxtFirewall()
	if err != nil {
		return diag.Errorf("error retrieving all NSX-T Firewall Rules: %s", err)
	}

	err = allRules.DeleteAllRules()
	if err != nil {
		return diag.Errorf("error deleting NSX-T Firewall Rules : %s", err)
	}

	d.SetId("")

	return nil
}

func resourceVcdNsxtFirewallImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] NSX-T Edge Gateway Firewall Rule import initiated")

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-name.nsxt-edge-gw-name")
	}
	orgName, vdcName, edgeName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("unable to find Org %s: %s", vdcName, err)
	}

	if vdc.IsNsxv() {
		return nil, fmt.Errorf("please use 'vcd_nsxv_firewall_rule' for NSX-V backed firewall rules")
	}

	edge, err := vdc.GetNsxtEdgeGatewayByName(edgeName)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve NSX-T edge gateway with ID '%s': %s", d.Id(), err)
	}

	_ = d.Set("org", orgName)
	_ = d.Set("vdc", vdcName)

	_ = d.Set("edge_gateway_id", edge.EdgeGateway.ID)
	d.SetId(edge.EdgeGateway.ID)

	return []*schema.ResourceData{d}, nil
}

func setNsxtFirewallData(fwRules []*types.NsxtFirewallRule, d *schema.ResourceData, edgeGatewayId string) error {

	_ = d.Set("edge_gateway_id", edgeGatewayId)

	result := make([]interface{}, len(fwRules))

	for index, value := range fwRules {
		sourceSlice := extractIdsFromOpenApiReferences(value.SourceFirewallGroups)
		sourceInterface := convertToTypeSet(sourceSlice)
		sourceSet := schema.NewSet(schema.HashSchema(&schema.Schema{Type: schema.TypeString}), sourceInterface)

		destinationSlice := extractIdsFromOpenApiReferences(value.DestinationFirewallGroups)
		destinationInterface := convertToTypeSet(destinationSlice)
		destinationSet := schema.NewSet(schema.HashSchema(&schema.Schema{Type: schema.TypeString}), destinationInterface)

		appPortProfileSlice := extractIdsFromOpenApiReferences(value.ApplicationPortProfiles)
		appPortProfileInterface := convertToTypeSet(appPortProfileSlice)
		appPortProfileSet := schema.NewSet(schema.HashSchema(&schema.Schema{Type: schema.TypeString}), appPortProfileInterface)

		result[index] = map[string]interface{}{
			"id":                   value.ID,
			"name":                 value.Name,
			"action":               value.Action,
			"enabled":              value.Enabled,
			"ip_protocol":          value.IpProtocol,
			"direction":            value.Direction,
			"logging":              value.Logging,
			"source_ids":           sourceSet,
			"destination_ids":      destinationSet,
			"app_port_profile_ids": appPortProfileSet,
		}
	}

	return d.Set("rule", result)
}

func getNsxtFirewallTypes(d *schema.ResourceData) []*types.NsxtFirewallRule {
	ruleInterfaceSlice := d.Get("rule").([]interface{})
	if len(ruleInterfaceSlice) > 0 {
		result := make([]*types.NsxtFirewallRule, len(ruleInterfaceSlice))
		for index, oneRule := range ruleInterfaceSlice {
			oneRuleMapInterface := oneRule.(map[string]interface{})

			result[index] = &types.NsxtFirewallRule{
				Name:       oneRuleMapInterface["name"].(string),
				Action:     oneRuleMapInterface["action"].(string),
				Enabled:    oneRuleMapInterface["enabled"].(bool),
				IpProtocol: oneRuleMapInterface["ip_protocol"].(string),
				Logging:    oneRuleMapInterface["logging"].(bool),
				Direction:  oneRuleMapInterface["direction"].(string),
				Version:    nil,
			}

			if oneRuleMapInterface["source_ids"] != nil {
				sourceGroups := convertSchemaSetToSliceOfStrings(oneRuleMapInterface["source_ids"].(*schema.Set))
				result[index].SourceFirewallGroups = convertSliceOfStringsToOpenApiReferenceIds(sourceGroups)
			}

			if oneRuleMapInterface["destination_ids"] != nil {
				sourceGroups := convertSchemaSetToSliceOfStrings(oneRuleMapInterface["destination_ids"].(*schema.Set))
				result[index].DestinationFirewallGroups = convertSliceOfStringsToOpenApiReferenceIds(sourceGroups)
			}

			if oneRuleMapInterface["app_port_profile_ids"] != nil {
				sourceGroups := convertSchemaSetToSliceOfStrings(oneRuleMapInterface["app_port_profile_ids"].(*schema.Set))
				result[index].ApplicationPortProfiles = convertSliceOfStringsToOpenApiReferenceIds(sourceGroups)
			}
		}
		return result
	}

	return nil
}
