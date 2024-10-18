package vcd

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

func resourceVcdDynamicSecurityGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdDynamicSecurityGroupCreate,
		ReadContext:   resourceVcdDynamicSecurityGroupRead,
		UpdateContext: resourceVcdDynamicSecurityGroupUpdate,
		DeleteContext: resourceVcdDynamicSecurityGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdDynamicSecurityGroupImport,
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
				Description: "VDC Group ID in which Dynamic Security Group is located",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Dynamic Security Group name",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Dynamic Security Group description",
			},
			"criteria": {
				Type:        schema.TypeSet,
				Description: "Up to 3 criteria to be used to define the Dynamic Security Group (VCD 10.2, 10.3)",
				// Up to 3 criteria can be defined as per current documentation, but API errors are
				// human readable so not hard-enforcing it as these limits may change in future VCD
				// versions
				// MaxItems: 3,
				Optional: true,
				Elem:     criteria,
			},
			"member_vms": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Set of VM IDs",
				Elem:        nsxtFirewallGroupMemberVms,
			},
		},
	}
}

var criteria = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"rule": {
			Type:        schema.TypeSet,
			Description: "Up to 4 rules can be used to define single criteria (VCD 10.2, 10.3)",
			// Up to 4 rules can be used to define single criteria as per documentation, but API
			// error is human readable and this might change in future so not enforcing max of 4
			// rules
			// MaxItems:    4,
			Optional: true,
			Elem:     criteriaRule,
		},
	},
}

var criteriaRule = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"type": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Type of object matching 'VM_TAG' or 'VM_NAME'",
		},
		"operator": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Operator can be one of 'EQUALS', 'CONTAINS', 'STARTS_WITH', 'ENDS_WITH'",
		},
		"value": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Filter value",
		},
	},
}

func resourceVcdDynamicSecurityGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	vdcGroupId := d.Get("vdc_group_id").(string)
	vcdClient.lockById(vdcGroupId)
	defer vcdClient.unlockById(vdcGroupId)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[nsxt dynamic security group create] error retrieving Org: %s", err)
	}

	vdcGroup, err := org.GetVdcGroupById(vdcGroupId)
	if err != nil {
		return diag.Errorf("error retrieving VDC Group: %s", err)
	}

	securityGroup := getNsxtDynamicSecurityGroupType(d)
	createdFwGroup, err := vdcGroup.CreateNsxtFirewallGroup(securityGroup)
	if err != nil {
		return diag.Errorf("[nsxt dynamic security group create] error creating NSX-T dynamic security group '%s': %s", securityGroup.Name, err)
	}

	d.SetId(createdFwGroup.NsxtFirewallGroup.ID)

	return resourceVcdDynamicSecurityGroupRead(ctx, d, meta)
}

func resourceVcdDynamicSecurityGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	vdcGroupId := d.Get("vdc_group_id").(string)
	vcdClient.lockById(vdcGroupId)
	defer vcdClient.unlockById(vdcGroupId)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[nsxt dynamic security group create] error retrieving Org: %s", err)
	}

	vdcGroup, err := org.GetVdcGroupById(vdcGroupId)
	if err != nil {
		return diag.Errorf("error retrieving VDC Group: %s", err)
	}

	updateSecurityGroup := getNsxtDynamicSecurityGroupType(d)
	securityGroup, err := vdcGroup.GetNsxtFirewallGroupById(d.Id())
	if err != nil {
		return diag.Errorf("[nsxt dynamic security group update] error getting NSX-T dynamic security group: %s", err)
	}

	// Inject existing ID for update
	updateSecurityGroup.ID = d.Id()

	_, err = securityGroup.Update(updateSecurityGroup)
	if err != nil {
		return diag.Errorf("[nsxt dynamic security group update] error updating NSX-T dynamic security group '%s': %s", securityGroup.NsxtFirewallGroup.Name, err)
	}

	return resourceVcdDynamicSecurityGroupRead(ctx, d, meta)
}

func resourceVcdDynamicSecurityGroupRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vdcGroupId := d.Get("vdc_group_id").(string)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[nsxt dynamic security group read] error retrieving Org: %s", err)
	}

	vdcGroup, err := org.GetVdcGroupById(vdcGroupId)
	if err != nil {
		return diag.Errorf("error retrieving VDC Group: %s", err)
	}

	securityGroup, err := vdcGroup.GetNsxtFirewallGroupById(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("[nsxt dynamic security group read] error getting NSX-T dynamic security group: %s", err)
	}

	err = setNsxtDynamicSecurityGroupData(d, securityGroup.NsxtFirewallGroup)
	if err != nil {
		return diag.Errorf("[nsxt dynamic security group resource read] error storing NSX-T dynamic security group: %s", err)
	}

	// A separate GET call is required to get all associated VMs
	associatedVms, err := securityGroup.GetAssociatedVms()
	if err != nil {
		return diag.Errorf("[nsxt dynamic security group resource read] error getting associated VMs for Security Group '%s': %s", securityGroup.NsxtFirewallGroup.Name, err)
	}

	err = setNsxtSecurityGroupAssociatedVmsData(d, associatedVms)
	if err != nil {
		return diag.Errorf("[nsxt dynamic security group resource read] error getting associated VMs for Security Group '%s': %s", securityGroup.NsxtFirewallGroup.Name, err)
	}

	return nil
}

func resourceVcdDynamicSecurityGroupDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vdcGroupId := d.Get("vdc_group_id").(string)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[nsxt dynamic security group read] error retrieving Org: %s", err)
	}

	vdcGroup, err := org.GetVdcGroupById(vdcGroupId)
	if err != nil {
		return diag.Errorf("error retrieving VDC Group: %s", err)
	}

	securityGroup, err := vdcGroup.GetNsxtFirewallGroupById(d.Id())
	if err != nil {
		return diag.Errorf("[nsxt dynamic security group delete] error getting NSX-T dynamic security group: %s", err)
	}

	err = securityGroup.Delete()
	if err != nil {
		return diag.Errorf("[nsxt dynamic security group resource delete] error deleting NSX-T dynamic security group: %s", err)
	}

	d.SetId("")

	return nil
}

func resourceVcdDynamicSecurityGroupImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-group-name.security_group_name")
	}
	orgName, vdcGroupName, securityGroupName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return nil, fmt.Errorf("[nsxt dynamic security group read] error retrieving Org: %s", err)
	}

	vdcGroup, err := org.GetVdcGroupByName(vdcGroupName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving VDC Group: %s", err)
	}

	securityGroup, err := vdcGroup.GetNsxtFirewallGroupByName(securityGroupName, types.FirewallGroupTypeVmCriteria)
	if err != nil {
		return nil, fmt.Errorf("[nsxt dynamic security group read] error getting NSX-T dynamic security group: %s", err)
	}

	dSet(d, "org", orgName)
	dSet(d, "vdc_group_id", securityGroup.NsxtFirewallGroup.OwnerRef.ID)
	d.SetId(securityGroup.NsxtFirewallGroup.ID)

	return []*schema.ResourceData{d}, nil
}

func getNsxtDynamicSecurityGroupType(d *schema.ResourceData) *types.NsxtFirewallGroup {
	fwGroup := &types.NsxtFirewallGroup{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		OwnerRef:    &types.OpenApiReference{ID: d.Get("vdc_group_id").(string)},
		TypeValue:   types.FirewallGroupTypeVmCriteria,
	}

	criteriaSet := d.Get("criteria").(*schema.Set)
	criteriaSlice := make([]types.NsxtFirewallGroupVmCriteria, len(criteriaSet.List()))
	for criteriaIndex, criteria := range criteriaSet.List() {
		criteriaMap := criteria.(map[string]interface{})

		// One more level deeper
		criteriaRuleSet := criteriaMap["rule"].(*schema.Set)
		criteriaRuleSlice := make([]types.NsxtFirewallGroupVmCriteriaRule, len(criteriaRuleSet.List()))
		for criteriaRuleIndex, criteriaRule := range criteriaRuleSet.List() {
			criteriaRuleMap := criteriaRule.(map[string]interface{})

			criteriaRuleSlice[criteriaRuleIndex] = types.NsxtFirewallGroupVmCriteriaRule{
				AttributeType:  criteriaRuleMap["type"].(string),
				Operator:       criteriaRuleMap["operator"].(string),
				AttributeValue: criteriaRuleMap["value"].(string),
			}
		}
		criteriaSlice[criteriaIndex] = types.NsxtFirewallGroupVmCriteria{
			VmCriteriaRule: criteriaRuleSlice,
		}
	}

	fwGroup.VmCriteria = criteriaSlice

	return fwGroup
}

func setNsxtDynamicSecurityGroupData(d *schema.ResourceData, fw *types.NsxtFirewallGroup) error {
	dSet(d, "name", fw.Name)
	dSet(d, "description", fw.Description)
	dSet(d, "vdc_group_id", fw.OwnerRef.ID)

	if len(fw.VmCriteria) > 0 {
		criteriaSlice := make([]interface{}, len(fw.VmCriteria))
		for criteriaIndex, criteria := range fw.VmCriteria {
			criteriaMap := make(map[string]interface{})

			criteriaRuleSlice := make([]interface{}, len(criteria.VmCriteriaRule))
			for ruleIndex, rule := range criteria.VmCriteriaRule {
				criteriaRuleMap := make(map[string]interface{})
				criteriaRuleMap["type"] = rule.AttributeType
				criteriaRuleMap["operator"] = rule.Operator
				criteriaRuleMap["value"] = rule.AttributeValue

				criteriaRuleSlice[ruleIndex] = criteriaRuleMap
			}
			ruleSet := schema.NewSet(schema.HashResource(criteriaRule), criteriaRuleSlice)
			criteriaMap["rule"] = ruleSet

			criteriaSlice[criteriaIndex] = criteriaMap
		}

		criteriaSet := schema.NewSet(schema.HashResource(criteria), criteriaSlice)
		err := d.Set("criteria", criteriaSet)
		if err != nil {
			return fmt.Errorf("error setting 'criteria' block: %s", err)
		}
	}

	return nil
}
