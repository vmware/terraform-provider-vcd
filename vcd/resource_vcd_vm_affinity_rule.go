package vcd

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

func resourceVcdVmAffinityRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdVmAffinityRuleCreate,
		Read:   resourceVcdVmAffinityRuleRead,
		Update: resourceVcdVmAffinityRuleUpdate,
		Delete: resourceVcdVmAffinityRuleDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdVmAffinityRuleImport,
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
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "VM affinity rule name",
			},
			"polarity": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				// Polarity can't change. If we want to, we need to create a new rule
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{types.PolarityAffinity, types.PolarityAntiAffinity}, false),
				Description:  "One of 'Affinity', 'Anti-Affinity'",
			},
			"required": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				Description: "True if this affinity rule is required. When a rule is mandatory, " +
					"a host failover will not power on the VM if doing so would violate the rule",
			},
			"enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "True if this affinity rule is enabled",
			},
			"virtual_machine_ids": {
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    2,
				Description: "Set of VM IDs assigned to this rule",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

// resourceToAffinityRule prepares a VM affinity rule definition from the data in the resource
func resourceToAffinityRule(d *schema.ResourceData, meta interface{}) (*types.VmAffinityRule, error) {
	vcdClient := meta.(*VCDClient)

	name := d.Get("name").(string)
	polarity := d.Get("polarity").(string)
	required := d.Get("required").(bool)
	enabled := d.Get("enabled").(bool)
	rawVms := d.Get("virtual_machine_ids").(*schema.Set)
	vmIdList := convertSchemaSetToSliceOfStrings(rawVms)

	fullVmList, err := vcdClient.Client.QueryVmList(types.VmQueryFilterOnlyDeployed)
	if err != nil {
		return nil, err
	}
	var vmReferences []*types.Reference

	var invalidEntries = make(map[string]bool)
	var foundEntries = make(map[string]bool)
	for _, vmId := range vmIdList {
		for _, vm := range fullVmList {
			uuid := extractUuid(vmId)
			if uuid != "" {
				if extractUuid(vmId) == extractUuid(vm.HREF) {
					vmReferences = append(vmReferences, &types.Reference{HREF: vm.HREF})
					foundEntries[vmId] = true
				}
			} else {
				invalidEntries[vmId] = true
			}
		}
	}
	if len(invalidEntries) > 0 {
		var invalidItems []string
		for k := range invalidEntries {
			invalidItems = append(invalidItems, k)
		}
		return nil, fmt.Errorf("invalid entries (not a VM ID) detected: %v", invalidItems)
	}
	if len(vmIdList) > len(foundEntries) {

		var notExistingVms []string
		for _, vmId := range vmIdList {
			_, exists := foundEntries[vmId]
			if !exists {
				notExistingVms = append(notExistingVms, vmId)
			}
		}
		return nil, fmt.Errorf("not existing VMs detected: %v", notExistingVms)
	}

	var vmAffinityRuleDef = &types.VmAffinityRule{
		Name:        name,
		IsEnabled:   takeBoolPointer(enabled),
		IsMandatory: takeBoolPointer(required),
		Polarity:    polarity,
		VmReferences: []*types.VMs{
			&types.VMs{
				VMReference: vmReferences,
			},
		},
	}
	return vmAffinityRuleDef, nil
}

// resourceVcdVmAffinityRuleCreate creates a VM affinity rule from the definition in the resource data
func resourceVcdVmAffinityRuleCreate(d *schema.ResourceData, meta interface{}) error {
	util.Logger.Printf("[TRACE] VM affinity rule creation")

	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return err
	}
	vmAffinityRuleDef, err := resourceToAffinityRule(d, meta)
	if err != nil {
		return err
	}

	vmAffinityRule, err := vdc.CreateVmAffinityRule(vmAffinityRuleDef)
	if err != nil {
		return err
	}
	d.SetId(vmAffinityRule.VmAffinityRule.ID)

	return resourceVcdVmAffinityRuleRead(d, meta)
}

// getVmAffinityRule searches a VM affinity rule using the data passed in the resource
// It can retrieve a rule either by ID or by name (if it was unique)
func getVmAffinityRule(d *schema.ResourceData, meta interface{}) (*govcd.VmAffinityRule, error) {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	name := d.Get("name").(string)

	// If ID is set, we use it
	identifier := d.Id()

	// The method variable stores the information about how we found the rule, for logging purposes
	method := "id"

	// The secondary method of retrieval is from the internal 'rule_id' field
	if identifier == "" {
		ruleId, ok := d.GetOk("rule_id")
		if ok {
			identifier = ruleId.(string)
			method = "rule_id"
		}
	}

	// The last method of retrieval is by name
	if identifier == "" {
		if name == "" {
			return nil, fmt.Errorf("both name and ID are empty")
		}
		identifier = name
		method = "name"
	}

	vmAffinityRule, err := vdc.GetVmAffinityRuleByNameOrId(identifier)
	if err != nil {
		return nil, fmt.Errorf("error retrieving rule by %s (%s): %s", method, identifier, err)
	}

	util.Logger.Printf("[TRACE] [get vm affinity rule] Retrieved by %s\n", method)
	return vmAffinityRule, nil
}

// resourceVcdVmAffinityRuleRead reads a resource VM affinity rule
func resourceVcdVmAffinityRuleRead(d *schema.ResourceData, meta interface{}) error {
	return genericVcdVmAffinityRuleRead(d, meta, "resource")
}

// genericVcdVmAffinityRuleRead retrieve an affinity rule using the resource or data source data
func genericVcdVmAffinityRuleRead(d *schema.ResourceData, meta interface{}, origin string) error {
	util.Logger.Printf("[TRACE] VM affinity rule Read")

	name := d.Get("name").(string)

	vmAffinityRule, err := getVmAffinityRule(d, meta)
	if err != nil {
		if origin == "resource" {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("[VM affinity rule read] error retrieving VM affinity rule %s: %s", name, err)
	}

	_ = d.Set("name", vmAffinityRule.VmAffinityRule.Name)
	if origin == "datasource" {
		_ = d.Set("rule_id", vmAffinityRule.VmAffinityRule.ID)
	}
	_ = d.Set("required", *vmAffinityRule.VmAffinityRule.IsMandatory)
	_ = d.Set("enabled", *vmAffinityRule.VmAffinityRule.IsEnabled)
	_ = d.Set("polarity", vmAffinityRule.VmAffinityRule.Polarity)
	var endpointVMs []string
	for _, vmr := range vmAffinityRule.VmAffinityRule.VmReferences {
		for _, ref := range vmr.VMReference {
			endpointVMs = append(endpointVMs, normalizeId("urn:vcloud:vm:", ref.ID))
		}
	}
	endpointVmSlice := convertToTypeSet(endpointVMs)
	endpointVmSet := schema.NewSet(schema.HashSchema(&schema.Schema{Type: schema.TypeString}), endpointVmSlice)
	err = d.Set("virtual_machine_ids", endpointVmSet)
	if err != nil {
		return fmt.Errorf("[VM affinity rule read] error setting the list of VM IDs: %s ", err)
	}

	d.SetId(vmAffinityRule.VmAffinityRule.ID)

	return nil
}

// resourceVcdVmAffinityRuleUpdate updates a VM affinity rule, including changing its name
func resourceVcdVmAffinityRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	util.Logger.Printf("[TRACE] VM affinity rule Update")
	vmAffinityRuleDef, err := resourceToAffinityRule(d, meta)
	if err != nil {
		return err
	}

	vmAffinityRule, err := getVmAffinityRule(d, meta)
	if err != nil {
		d.SetId("")
		return nil
	}
	if vmAffinityRuleDef.Polarity != vmAffinityRule.VmAffinityRule.Polarity {
		return fmt.Errorf("[VM affinity rule update] polarity cannot be changed")
	}
	vmAffinityRule.VmAffinityRule.Name = vmAffinityRuleDef.Name
	vmAffinityRule.VmAffinityRule.IsMandatory = vmAffinityRuleDef.IsMandatory
	vmAffinityRule.VmAffinityRule.IsEnabled = vmAffinityRuleDef.IsEnabled
	vmAffinityRule.VmAffinityRule.VmReferences = vmAffinityRuleDef.VmReferences

	err = vmAffinityRule.Update()
	if err != nil {
		return fmt.Errorf("[VM affinity rule update] error running the update: %s", err)
	}

	return resourceVcdVmAffinityRuleRead(d, meta)
}

// resourceVcdVmAffinityRuleDelete removes a VM affinity rule
func resourceVcdVmAffinityRuleDelete(d *schema.ResourceData, meta interface{}) error {
	util.Logger.Printf("[TRACE] VM affinity rule Delete")
	vmAffinityRule, err := getVmAffinityRule(d, meta)
	if err != nil {
		return fmt.Errorf("[VM affinity rule delete] error retrieving VM affinity rule %s: %s", d.Get("name"), err)
	}
	return vmAffinityRule.Delete()
}

// resourceVcdVmAffinityRuleImport is responsible for importing a VM affinity rule into state
// It requires a string composed of org name + separator + VDC name + separator + identifier
// where separator is '.' by default but can be customized in the Provider
// The affinity rule identifier can be either the name or the ID (preferred)
// If the name is unique, it is used to retrieve the list, otherwise an error is returned, containing the list
// of the rules that correspond to such name, and their IDs
// If the import string starts with the special token "list@", then a list of ALL the affinity rules is returned
// as an error message
//
// examples:
// (1)
//  terraform import vcd_vm_affinity_rule.unknown my-org.my-vdc.my-afr
// if the name is unique, import is executed. If it is not, an error with the IDs of the rules named "my-afr" is returned
// (2)
//  terraform import vcd_vm_affinity_rule.unknown my-org.my-vdc.cf73a7aa-ddd2-4d11-aca6-1917816065cc
// If the ID is valid, the import gets performed
// (3)
//  terraform import vcd_vm_affinity_rule.unknown list@my-org.my-vdc.any_string
// Returns an error with all the VM affinity rules (name + ID for each)
func resourceVcdVmAffinityRuleImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("[VM affinity rule import] resource identifier must be specified as org.vdc.my-affinity-rule")
	}
	listRequested := false
	orgName, vdcName, affinityRuleIdentifier := resourceURI[0], resourceURI[1], resourceURI[2]
	if strings.Contains(orgName, "list@") {
		listRequested = true
		orgNameList := strings.Split(orgName, "@")
		if len(orgNameList) < 2 {
			return nil, fmt.Errorf("[VM affinity rule import] empty Org name provided with @list request")
		}
		orgName = orgNameList[1]
	}
	if orgName == "" {
		return nil, fmt.Errorf("[VM affinity rule import] empty org name provided")
	}
	if vdcName == "" {
		return nil, fmt.Errorf("[VM affinity rule import] empty VDC name provided")
	}
	if affinityRuleIdentifier == "" {
		return nil, fmt.Errorf("[VM affinity rule import] empty VM affinity rule identifier provided")
	}

	lookingForId := isUuid(affinityRuleIdentifier)
	vcdClient := meta.(*VCDClient)
	adminOrg, err := vcdClient.GetAdminOrgByName(orgName)
	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrg, orgName)
	}

	vdc, err := adminOrg.GetVDCByName(vdcName, false)
	if err != nil {
		return nil, govcd.ErrorEntityNotFound
	}

	ruleList, err := vdc.GetAllVmAffinityRuleList()
	if err != nil {
		return nil, fmt.Errorf("[VM affinity rule import] error retrieving VM affinity rule list: %s", err)
	}
	if listRequested {
		return nil, fmt.Errorf("[VM affinity rule import] list of all VM affinity rules:\n%s", formatVmAffinityRulesList(ruleList))
	}
	var foundRules []*types.VmAffinityRule

	for _, rule := range ruleList {
		if lookingForId {
			if rule.ID == affinityRuleIdentifier {
				foundRules = append(foundRules, rule)
				break
			}
		} else {
			if rule.Name == affinityRuleIdentifier {
				foundRules = append(foundRules, rule)
			}
		}
	}

	if len(foundRules) == 0 {
		return nil, fmt.Errorf("[VM affinity rule import] no VM affinity rule found with identifier %s", affinityRuleIdentifier)
	}
	if len(foundRules) > 1 {
		return nil, fmt.Errorf("[VM affinity rule import] more than one VM affinity rule matches the name %s\n%s",
			affinityRuleIdentifier, formatVmAffinityRulesList(foundRules))
	}
	vmAffinityRule := foundRules[0]
	_ = d.Set("org", orgName)
	_ = d.Set("vdc", vdcName)
	_ = d.Set("name", vmAffinityRule.Name)
	d.SetId(vmAffinityRule.ID)

	return []*schema.ResourceData{d}, nil
}

// formatVmAffinityRulesList returns a formatted string containing the names and IDs of the
// affinity rules passed as input
func formatVmAffinityRulesList(list []*types.VmAffinityRule) string {
	result := ""
	for i, rule := range list {
		result += fmt.Sprintf("%3d %-30s %-13s %s\n", i, rule.Name, rule.Polarity, rule.ID)
	}
	return result
}
