package govcd

import (
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// VmAffinityRule is the govcd structure to deal with VM affinity rules
type VmAffinityRule struct {
	VmAffinityRule *types.VmAffinityRule
	client         *Client
}

// NewVmAffinityRule creates a new VM affinity rule
func NewVmAffinityRule(cli *Client) *VmAffinityRule {
	return &VmAffinityRule{
		VmAffinityRule: new(types.VmAffinityRule),
		client:         cli,
	}
}

// validPolarity validates the polarity passed as a string
// Accepted values are only 'Affinity' and 'Anti-Affinity'
func validPolarity(polarity string) bool {
	return polarity == types.PolarityAffinity || polarity == types.PolarityAntiAffinity
}

// GetAllVmAffinityRuleList retrieves all VM affinity and anti-affinity rules
func (vdc *Vdc) GetAllVmAffinityRuleList() ([]*types.VmAffinityRule, error) {

	affinityRules := new(types.VmAffinityRules)

	href := vdc.getLinkHref("down", "application/vnd.vmware.vcloud.vmaffinityrules+xml")
	if href == "" {
		return nil, fmt.Errorf("no link with VM affinity rule found in VDC %s", vdc.Vdc.Name)
	}
	_, err := vdc.client.ExecuteRequest(href, http.MethodGet,
		"", "error retrieving list of affinity rules: %s", nil, affinityRules)
	if err != nil {
		return nil, err
	}
	return affinityRules.VmAffinityRule, nil

}

// GetVmAffinityRuleList retrieves VM affinity rules
func (vdc *Vdc) GetVmAffinityRuleList() ([]*types.VmAffinityRule, error) {
	return vdc.getSpecificVmAffinityRuleList(types.PolarityAffinity)
}

// GetVmAntiAffinityRuleList retrieves VM anti-affinity rules
func (vdc *Vdc) GetVmAntiAffinityRuleList() ([]*types.VmAffinityRule, error) {
	return vdc.getSpecificVmAffinityRuleList(types.PolarityAntiAffinity)
}

// getSpecificVmAffinityRuleList retrieves specific VM affinity rules
func (vdc *Vdc) getSpecificVmAffinityRuleList(polarity string) ([]*types.VmAffinityRule, error) {
	fullList, err := vdc.GetAllVmAffinityRuleList()
	if err != nil {
		return nil, err
	}

	var returnList []*types.VmAffinityRule
	for _, rule := range fullList {
		if rule.Polarity == polarity {
			returnList = append(returnList, rule)
		}
	}

	return returnList, nil
}

// GetVmAffinityRuleByHref finds a VM affinity or anti-affinity rule by HREF
func (vdc *Vdc) GetVmAffinityRuleByHref(href string) (*VmAffinityRule, error) {

	affinityRule := NewVmAffinityRule(vdc.client)

	_, err := vdc.client.ExecuteRequest(href, http.MethodGet,
		"", "error retrieving affinity rule: %s", nil, affinityRule.VmAffinityRule)
	if err != nil {
		return nil, err
	}

	return affinityRule, nil
}

// GetVmAffinityRulesByName finds the rules with the given name
// Note that name does not have to be unique, so a search by name can match several items
// If polarity is indicated, the function retrieves only the rules with the given polarity
func (vdc *Vdc) GetVmAffinityRulesByName(name string, polarity string) ([]*VmAffinityRule, error) {

	var returnList []*VmAffinityRule
	ruleList, err := vdc.GetAllVmAffinityRuleList()
	if err != nil {
		return nil, err
	}
	for _, rule := range ruleList {
		if rule.Name == name {
			fullRule, err := vdc.GetVmAffinityRuleByHref(rule.HREF)
			if err != nil {
				return returnList, err
			}
			if (polarity != "" && polarity == rule.Polarity) || polarity == "" {
				returnList = append(returnList, fullRule)
			}
		}
	}
	return returnList, nil
}

// GetVmAffinityRuleById retrieves a VM affinity or anti-affinity rule by ID
func (vdc *Vdc) GetVmAffinityRuleById(id string) (*VmAffinityRule, error) {

	list, err := vdc.GetAllVmAffinityRuleList()
	if err != nil {
		return nil, err
	}
	for _, rule := range list {
		if equalIds(id, rule.ID, rule.HREF) {
			return vdc.GetVmAffinityRuleByHref(rule.HREF)
		}
	}
	return nil, ErrorEntityNotFound
}

// GetVmAffinityRuleByNameOrId retrieves an affinity or anti-affinity rule by name or ID
// Given the possibility of a name identifying multiple items, this function may also fail
// when the search by name returns more than one item.
func (vdc *Vdc) GetVmAffinityRuleByNameOrId(identifier string) (*VmAffinityRule, error) {
	getByName := func(name string, refresh bool) (interface{}, error) {
		list, err := vdc.GetVmAffinityRulesByName(name, "")
		if err != nil {
			return nil, err
		}
		if len(list) == 0 {
			return nil, ErrorEntityNotFound
		}
		if len(list) == 1 {
			return list[0], nil
		}
		return nil, fmt.Errorf("more than one item matches the name '%s'", name)
	}
	getById := func(id string, refresh bool) (interface{}, error) { return vdc.GetVmAffinityRuleById(id) }
	entity, err := getEntityByNameOrId(getByName, getById, identifier, false)
	if entity == nil {
		return nil, err
	}
	return entity.(*VmAffinityRule), err
}

// validateAffinityRule checks that a VM affinity rule has all the needed properties
// If checkVMs is true, then the function checks that all VMs in the internal list exist.
// The usual workflow is:
// 1. validation without VM checking
// 2. creation or update
// 3. if no error -> end
// 4. if error, validation with VM checks
//    4a. if validation error, it was a VM issue: return combined original error + validation error
//    4b. if no validation error, the failure was due to something else: return only original error
func validateAffinityRule(client *Client, affinityRuleDef *types.VmAffinityRule, checkVMs bool) (*types.VmAffinityRule, error) {
	if affinityRuleDef == nil {
		return nil, fmt.Errorf("empty definition given for a VM affinity rule")
	}
	if affinityRuleDef.Name == "" {
		return nil, fmt.Errorf("no name given for a VM affinity rule")
	}
	if affinityRuleDef.Polarity == "" {
		return nil, fmt.Errorf("no polarity given for a VM affinity rule")
	}
	if !validPolarity(affinityRuleDef.Polarity) {
		return nil, fmt.Errorf("illegal polarity given (%s) for a VM affinity rule", affinityRuleDef.Polarity)
	}
	// Ensure the VMs in the list are different
	var seenVms = make(map[string]bool)
	var allVmMap = make(map[string]bool)
	if checkVMs {
		vmList, err := client.QueryVmList(types.VmQueryFilterOnlyDeployed)
		if err != nil {
			return nil, fmt.Errorf("error getting VM list : %s", err)
		}
		for _, vm := range vmList {
			allVmMap[extractUuid(vm.HREF)] = true
		}
	}
	for _, vmr := range affinityRuleDef.VmReferences {
		if len(vmr.VMReference) == 0 {
			continue
		}
		for _, vm := range vmr.VMReference {
			if vm == nil {
				continue
			}
			// The only mandatory field is the HREF
			if vm.HREF == "" {
				return nil, fmt.Errorf("empty VM HREF provided in VM list")
			}
			_, seen := seenVms[vm.HREF]
			if seen {
				return nil, fmt.Errorf("VM HREF %s used more than once", vm.HREF)
			}
			seenVms[vm.HREF] = true

			if checkVMs {
				// Checking that the VMs indicated exist.
				// Without this check, if any of the VMs do not exist, we would get an ugly error that doesn't easily explain
				//  the nature of the problem, such as
				//   > "error instantiating a new VM affinity rule: API Error: 403: [ ... ]
				//   > Either you need some or all of the following rights [ORG_VDC_VM_VM_AFFINITY_EDIT]
				//   > to perform operations [VAPP_VM_EDIT_AFFINITY_RULE] for $OP_ID or the target entity is invalid"

				_, vmInList := allVmMap[extractUuid(vm.HREF)]
				if !vmInList {
					return nil, fmt.Errorf("VM identified by '%s' not found ", vm.HREF)
				}
			}
		}
	}
	if len(seenVms) < 2 {
		return nil, fmt.Errorf("at least 2 VMs should be given for a VM Affinity Rule")
	}
	return affinityRuleDef, nil
}

// CreateVmAffinityRuleAsync creates a new VM affinity rule, and returns a task that handles the operation
func (vdc *Vdc) CreateVmAffinityRuleAsync(affinityRuleDef *types.VmAffinityRule) (Task, error) {

	var err error
	// We validate the input, without a strict check on the VMs
	affinityRuleDef, err = validateAffinityRule(vdc.client, affinityRuleDef, false)
	if err != nil {
		return Task{}, fmt.Errorf("[CreateVmAffinityRuleAsync] %s", err)
	}

	affinityRuleDef.Xmlns = types.XMLNamespaceVCloud

	href := vdc.getLinkHref("add", "application/vnd.vmware.vcloud.vmaffinityrule+xml")
	if href == "" {
		return Task{}, fmt.Errorf("no link with VM affinity rule found in VDC %s", vdc.Vdc.Name)
	}

	task, err := vdc.client.ExecuteTaskRequest(href, http.MethodPost,
		"application/vnd.vmware.vcloud.vmaffinityrule+xml", "error instantiating a new VM affinity rule: %s", affinityRuleDef)
	if err != nil {
		// if we get any error, we repeat the validation
		// with a strict check on VM existence.
		_, validationErr := validateAffinityRule(vdc.client, affinityRuleDef, true)
		if validationErr != nil {
			// If we get any error from the validation now, it should be an invalid VM,
			// so we combine the original error with the validation error
			return Task{}, fmt.Errorf("%s - %s", err, validationErr)
		}
		// If the validation error is nil, we return just the original error
		return Task{}, err
	}
	return task, err
}

// CreateVmAffinityRule is a wrap around CreateVmAffinityRuleAsync that handles the task and returns the finished object
func (vdc *Vdc) CreateVmAffinityRule(affinityRuleDef *types.VmAffinityRule) (*VmAffinityRule, error) {

	task, err := vdc.CreateVmAffinityRuleAsync(affinityRuleDef)
	if err != nil {
		return nil, err
	}
	// The rule ID is the ID of the task owner (see Task definition in types.go)
	ruleId := task.Task.Owner.ID

	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, err
	}

	// Retrieving the newly created rule using the ID from the task
	vmAffinityRule, err := vdc.GetVmAffinityRuleById(ruleId)
	if err != nil {
		return nil, fmt.Errorf("error retrieving VmAffinityRule %s using ID %s: %s", affinityRuleDef.Name, ruleId, err)
	}
	return vmAffinityRule, nil
}

// Delete removes a VM affinity rule from vCD
func (vmar *VmAffinityRule) Delete() error {

	if vmar == nil || vmar.VmAffinityRule == nil {
		return fmt.Errorf("nil VM Affinity Rule passed for deletion")
	}

	if vmar.VmAffinityRule.HREF == "" {
		return fmt.Errorf("VM Affinity Rule passed for deletion has no HREF")
	}

	deleteHref := vmar.VmAffinityRule.HREF
	linkHref := vmar.getLinkHref("remove")
	if linkHref != "" {
		deleteHref = linkHref
	}

	deleteTask, err := vmar.client.ExecuteTaskRequest(deleteHref, http.MethodDelete,
		"", "error removing VM Affinity Rule : %s", nil)
	if err != nil {
		return err
	}
	return deleteTask.WaitTaskCompletion()
}

// getLinkHref returns an HREF for a given value of Rel
func (vmar *VmAffinityRule) getLinkHref(rel string) string {
	if vmar.VmAffinityRule.Link != nil {
		for _, link := range vmar.VmAffinityRule.Link {
			if link.Rel == rel {
				return link.HREF
			}
		}
	}
	return ""
}

// Update modifies a VM affinity rule using as input
// the entity's internal data.
func (vmar *VmAffinityRule) Update() error {
	var err error
	var affinityRuleDef *types.VmAffinityRule

	if vmar == nil || vmar.VmAffinityRule == nil {
		return fmt.Errorf("nil VM Affinity Rule passed for update")
	}
	if vmar.VmAffinityRule.HREF == "" {
		return fmt.Errorf("VM Affinity Rule passed for update has no HREF")
	}

	// We validate the input, without a strict check on the VMs
	affinityRuleDef, err = validateAffinityRule(vmar.client, vmar.VmAffinityRule, false)
	if err != nil {
		return fmt.Errorf("[Update] %s", err)
	}
	vmar.VmAffinityRule = affinityRuleDef

	updateRef := vmar.VmAffinityRule.HREF
	linkHref := vmar.getLinkHref("edit")
	if linkHref != "" {
		updateRef = linkHref
	}

	vmar.VmAffinityRule.Link = nil
	vmar.VmAffinityRule.VCloudExtension = nil
	updateTask, err := vmar.client.ExecuteTaskRequest(updateRef, http.MethodPut,
		"", "error updating VM Affinity Rule : %s", vmar.VmAffinityRule)
	if err != nil {
		// if we get any error, we repeat the validation
		// with a strict check on VM existence.
		_, validationErr := validateAffinityRule(vmar.client, affinityRuleDef, true)
		// If we get any error from the validation now, it should be an invalid VM,
		// so we combine the original error with the validation error
		if validationErr != nil {
			return fmt.Errorf("%s - %s", err, validationErr)
		}
		// If the validation error is nil, we return just the original error
		return err
	}
	err = updateTask.WaitTaskCompletion()
	if err != nil {
		return err
	}
	return vmar.Refresh()
}

// Refresh gets a fresh copy of the VM affinity rule from vCD
func (vmar *VmAffinityRule) Refresh() error {
	var newVmAffinityRule types.VmAffinityRule
	_, err := vmar.client.ExecuteRequest(vmar.VmAffinityRule.HREF, http.MethodGet,
		"", "error retrieving affinity rule: %v", nil, &newVmAffinityRule)
	if err != nil {
		return err
	}
	vmar.VmAffinityRule = &newVmAffinityRule
	return nil
}

// SetEnabled is a shortcut to update only the IsEnabled property of a VM affinity rule
func (vmar *VmAffinityRule) SetEnabled(value bool) error {
	if vmar.VmAffinityRule.IsEnabled != nil {
		currentValue := *vmar.VmAffinityRule.IsEnabled
		if currentValue == value {
			return nil
		}
	}
	vmar.VmAffinityRule.IsEnabled = takeBoolPointer(value)
	return vmar.Update()
}

// SetMandatory is a shortcut to update only the IsMandatory property of a VM affinity rule
func (vmar *VmAffinityRule) SetMandatory(value bool) error {
	if vmar.VmAffinityRule.IsMandatory != nil {
		currentValue := *vmar.VmAffinityRule.IsMandatory
		if currentValue == value {
			return nil
		}
	}
	vmar.VmAffinityRule.IsMandatory = takeBoolPointer(value)
	return vmar.Update()
}
