package vcd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdNsxtDistributedFirewall() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdNsxtDistributedFirewallCreateUpdate,
		ReadContext:   resourceVcdNsxtDistributedFirewallRead,
		UpdateContext: resourceVcdNsxtDistributedFirewallCreateUpdate,
		DeleteContext: resourceVcdNsxtDistributedFirewallDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdNsxtDistributedFirewallImport,
		},

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
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"rule": {
				Type:        schema.TypeList, // Firewall rule order matters
				Required:    true,
				MinItems:    1,
				Description: "Ordered list of firewall rules",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Firewall Rule ID",
						},
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Firewall Rule name",
						},
						"description": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Description is not shown in UI",
						},
						"comment": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Comment that is shown next to rule in UI (VCD 10.3.2+)",
						},
						"direction": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Direction on which Firewall Rule applies (One of 'IN', 'OUT', 'IN_OUT')",
							Default:      "IN_OUT",
							ValidateFunc: validation.StringInSlice([]string{"IN", "OUT", "IN_OUT"}, false),
						},
						"ip_protocol": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Firewall Rule Protocol (One of 'IPV4', 'IPV6', 'IPV4_IPV6')",
							Default:      "IPV4_IPV6",
							ValidateFunc: validation.StringInSlice([]string{"IPV4", "IPV6", "IPV4_IPV6"}, false),
						},
						"action": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Defines if the rule should 'ALLOW', 'DROP', 'REJECT' matching traffic",
							ValidateFunc: validation.StringInSlice([]string{"ALLOW", "DROP", "REJECT"}, false),
						},
						"enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Defined if Firewall Rule is active",
						},
						"logging": {
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
						"network_context_profile_ids": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "A set of Network Context Profile IDs. Leaving it empty means 'Any'",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"source_groups_excluded": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Reverses firewall matching for to match all except Source Groups specified in 'source_ids' (VCD 10.3.2+)",
						},
						"destination_groups_excluded": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Reverses firewall matching for to match all except Destinations Groups specified in 'destination_ids' (VCD 10.3.2+)",
						},
					},
				},
			},
		},
	}
}

// resourceVcdNsxtDistributedFirewallCreateUpdate is used in both Create and Update cases because
// firewall rules don't have separate create or update methods. Firewall endpoint only uses HTTP PUT
// for update.
func resourceVcdNsxtDistributedFirewallCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentVdcGroup(d)
	defer vcdClient.unlockParentVdcGroup(d)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[Distributed Firewall Create/Update] error retriving Org: %s", err)
	}

	vdcGroup, err := org.GetVdcGroupById(d.Get("vdc_group_id").(string))
	if err != nil {
		return diag.Errorf("[Distributed Firewall Create/Update] error retrieving VDC Group: %s", err)
	}

	firewallRulesType, err := getDistributedFirewallTypes(vcdClient, d)
	if err != nil {
		return diag.Errorf("[Distributed Firewall Create/Update] error getting Distributed Firewall type: %s", err)
	}
	_, err = vdcGroup.UpdateDistributedFirewall(firewallRulesType)
	if err != nil {
		return diag.Errorf("[Distributed Firewall Create/Update] error setting Distributed Firewall: %s", err)
	}

	d.SetId(vdcGroup.VdcGroup.Id)

	return resourceVcdNsxtDistributedFirewallRead(ctx, d, meta)
}

func resourceVcdNsxtDistributedFirewallRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[Distributed Firewall Read] error retriving Org: %s", err)
	}

	vdcGroupId := d.Id()
	vdcGroup, err := org.GetVdcGroupById(vdcGroupId)
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("[Distributed Firewall Read] error retrieving VDC Group (%s): %s", vdcGroupId, err)
	}

	fwRules, err := vdcGroup.GetDistributedFirewall()
	if err != nil {
		return diag.Errorf("[Distributed Firewall Read] error retrieving NSX-T Firewall Rules: %s", err)
	}

	err = setDistributedFirewallData(vcdClient, fwRules.DistributedFirewallRuleContainer, d, vdcGroup.VdcGroup.Id)
	if err != nil {
		return diag.Errorf("[Distributed Firewall Read] error storing NSX-T Firewall data to schema: %s", err)
	}

	return nil
}

func resourceVcdNsxtDistributedFirewallDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentVdcGroup(d)
	defer vcdClient.unlockParentVdcGroup(d)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[Distributed Firewall Delete] error retriving Org: %s", err)
	}

	vdcGroup, err := org.GetVdcGroupById(d.Get("vdc_group_id").(string))
	if err != nil {
		return diag.Errorf("[Distributed Firewall Delete] error retrieving VDC Group: %s", err)
	}

	err = vdcGroup.DeleteAllDistributedFirewallRules()
	if err != nil {
		return diag.Errorf("[Distributed Firewall Delete]  error deleting Distributed Firewall rules: %s", err)
	}

	return nil
}

func resourceVcdNsxtDistributedFirewallImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] NSX-T Distributed Firewall import initiated")

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-group-name")
	}

	orgName, vdcGroupName := resourceURI[0], resourceURI[1]

	vcdClient := meta.(*VCDClient)
	adminOrg, err := vcdClient.GetAdminOrg(orgName)
	if err != nil {
		return nil, fmt.Errorf("[Distributed Firewall Import] error retrieving org %s: %s", orgName, err)
	}

	vdcGroup, err := adminOrg.GetVdcGroupByName(vdcGroupName)
	if err != nil {
		return nil, fmt.Errorf("[Distributed Firewall Import] error importing VDC group item: %s", err)
	}

	d.SetId(vdcGroup.VdcGroup.Id)
	dSet(d, "org", orgName)

	return []*schema.ResourceData{d}, nil
}

func setDistributedFirewallData(vcdClient *VCDClient, dfwRules *types.DistributedFirewallRules, d *schema.ResourceData, vdcGroupId string) error {
	dSet(d, "vdc_group_id", vdcGroupId)

	result := make([]interface{}, len(dfwRules.Values))
	for index, value := range dfwRules.Values {
		sourceSlice := extractIdsFromOpenApiReferences(value.SourceFirewallGroups)
		sourceSet := convertStringsTotTypeSet(sourceSlice)

		destinationSlice := extractIdsFromOpenApiReferences(value.DestinationFirewallGroups)
		destinationSet := convertStringsTotTypeSet(destinationSlice)

		appPortProfileSlice := extractIdsFromOpenApiReferences(value.ApplicationPortProfiles)
		appPortProfileSet := convertStringsTotTypeSet(appPortProfileSlice)

		netContextProfileSlice := extractIdsFromOpenApiReferences(value.NetworkContextProfiles)
		netPortProfileSet := convertStringsTotTypeSet(netContextProfileSlice)

		var actionFieldValue string
		if vcdClient.Client.APIVCDMaxVersionIs(">= 35.2") {
			actionFieldValue = value.ActionValue
		} else { // TODO remove this when VCD 10.2 support is dropped
			actionFieldValue = value.Action
		}

		result[index] = map[string]interface{}{
			"id":                          value.ID,
			"name":                        value.Name,
			"description":                 value.Description,
			"comment":                     value.Comments,
			"action":                      actionFieldValue,
			"enabled":                     value.Enabled,
			"ip_protocol":                 value.IpProtocol,
			"direction":                   value.Direction,
			"logging":                     value.Logging,
			"source_ids":                  sourceSet,
			"destination_ids":             destinationSet,
			"app_port_profile_ids":        appPortProfileSet,
			"network_context_profile_ids": netPortProfileSet,
			"source_groups_excluded":      value.SourceGroupsExcluded,
			"destination_groups_excluded": value.DestinationGroupsExcluded,
		}
	}

	return d.Set("rule", result)
}

func getDistributedFirewallTypes(vcdClient *VCDClient, d *schema.ResourceData) (*types.DistributedFirewallRules, error) {
	ruleInterfaceSlice := d.Get("rule").([]interface{})
	if len(ruleInterfaceSlice) > 0 {
		sliceOfRules := make([]*types.DistributedFirewallRule, len(ruleInterfaceSlice))
		for index, oneRule := range ruleInterfaceSlice {
			oneRuleMapInterface := oneRule.(map[string]interface{})

			sliceOfRules[index] = &types.DistributedFirewallRule{
				Name:        oneRuleMapInterface["name"].(string),
				Description: oneRuleMapInterface["description"].(string),
				ActionValue: oneRuleMapInterface["action"].(string),
				Enabled:     oneRuleMapInterface["enabled"].(bool),
				IpProtocol:  oneRuleMapInterface["ip_protocol"].(string),
				Logging:     oneRuleMapInterface["logging"].(bool),
				Direction:   oneRuleMapInterface["direction"].(string),
				Version:     nil,
			}

			if oneRuleMapInterface["source_ids"] != nil {
				sourceGroupIds := convertSchemaSetToSliceOfStrings(oneRuleMapInterface["source_ids"].(*schema.Set))
				sliceOfRules[index].SourceFirewallGroups = convertSliceOfStringsToOpenApiReferenceIds(sourceGroupIds)
			}

			if oneRuleMapInterface["destination_ids"] != nil {
				destinationGroupIds := convertSchemaSetToSliceOfStrings(oneRuleMapInterface["destination_ids"].(*schema.Set))
				sliceOfRules[index].DestinationFirewallGroups = convertSliceOfStringsToOpenApiReferenceIds(destinationGroupIds)
			}

			if oneRuleMapInterface["app_port_profile_ids"] != nil {
				appPortProfileIds := convertSchemaSetToSliceOfStrings(oneRuleMapInterface["app_port_profile_ids"].(*schema.Set))
				sliceOfRules[index].ApplicationPortProfiles = convertSliceOfStringsToOpenApiReferenceIds(appPortProfileIds)
			}

			if oneRuleMapInterface["network_context_profile_ids"] != nil {
				networkContextPortProfileIds := convertSchemaSetToSliceOfStrings(oneRuleMapInterface["network_context_profile_ids"].(*schema.Set))
				sliceOfRules[index].NetworkContextProfiles = convertSliceOfStringsToOpenApiReferenceIds(networkContextPortProfileIds)
			}

			// Perform version specific conversion
			// TODO remove when VCD 10.2 is not supported anymore

			// ActionValue was introduced in API V35.2 (VCD 10.2.2), for 10.2.0 Action field must
			// still be used
			if vcdClient.Client.APIVCDMaxVersionIs("< 35.2") {
				sliceOfRules[index].Action = sliceOfRules[index].ActionValue
				sliceOfRules[index].ActionValue = ""
			}

			// Fields requiring 10.3.2+
			// TODO remove when VCD 10.3 is not supported anymore
			comment := oneRuleMapInterface["comment"].(string)
			sourceGroupsExcluded := oneRuleMapInterface["source_groups_excluded"].(bool)
			destinationGroupsExcluded := oneRuleMapInterface["destination_groups_excluded"].(bool)
			if vcdClient.Client.APIVCDMaxVersionIs(">= 36.2") {
				sliceOfRules[index].Comments = comment

				if sourceGroupsExcluded {
					sliceOfRules[index].SourceGroupsExcluded = &sourceGroupsExcluded
				}

				if destinationGroupsExcluded {
					sliceOfRules[index].DestinationGroupsExcluded = &destinationGroupsExcluded
				}
			} else {
				if comment != "" {
					return nil, fmt.Errorf("field 'comment' can only be set in VCD 10.3.2+")
				}

				// Two below checks will only throw an error if 'true' value has been set. False
				// will be ignored (when either set, or not set at all), because the only somewhat
				// reliable way is to use d.GetOkExists which has been deprecated in SDK with no
				// reliable replacement. There is no real need to use it here so just leaving one
				// less place to fix in future if d.GetOkExists is removed from SDK.
				if sourceGroupsExcluded {
					return nil, fmt.Errorf("field 'source_groups_excluded' can only be enabled in VCD 10.3.2+")
				}

				if destinationGroupsExcluded {
					return nil, fmt.Errorf("field 'source_groups_excluded' can only be enabled in VCD 10.3.2+")
				}
			}

		}

		return &types.DistributedFirewallRules{Values: sliceOfRules}, nil
	}

	return nil, nil
}
