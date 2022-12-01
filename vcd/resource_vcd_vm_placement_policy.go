package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/vmware/go-vcloud-director/v2/util"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdVmPlacementPolicy() *schema.Resource {

	return &schema.Resource{
		CreateContext: resourceVmPlacementPolicyCreate,
		ReadContext:   resourceVmPlacementPolicyRead,
		UpdateContext: resourceVmPlacementPolicyUpdate,
		DeleteContext: resourceVmPlacementPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVmPlacementPolicyImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the VM Placement Policy",
			},
			"provider_vdc_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the Provider VDC to which the VM Placement Policy belongs",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the VM Placement Policy",
			},
			"vm_group_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:     true,
				Description:  "IDs of the collection of VMs with similar host requirements",
				AtLeastOneOf: []string{"vm_group_ids", "logical_vm_group_ids"},
			},
			"logical_vm_group_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:     true,
				Description:  "IDs of one or more Logical VM Groups to define this VM Placement Policy. There is an AND relationship among all the entries set in this attribute",
				AtLeastOneOf: []string{"vm_group_ids", "logical_vm_group_ids"},
			},
		},
	}
}

func resourceVmPlacementPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vmGroupIds := d.Get("vm_group_ids").(*schema.Set)
	logicalVmGroupIds := d.Get("logical_vm_group_ids").(*schema.Set)
	if len(vmGroupIds.List()) == 0 && len(logicalVmGroupIds.List()) == 0 {
		return diag.Errorf("either `vm_group_ids` or `logical_vm_group_ids` must have a")
	}

	log.Printf("[TRACE] VM Placement Policy creation initiated: %s in pVDC %s", d.Get("name").(string), d.Get("provider_vdc_id").(string))
	vcdClient := meta.(*VCDClient)

	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("functionality requires System administrator privileges")
	}

	pVdc, err := vcdClient.GetProviderVdcById(d.Get("provider_vdc_id").(string))
	if err != nil {
		return diag.Errorf("could not retrieve required Provider VDC: %s", err)
	}

	computePolicy := &types.VdcComputePolicyV2{
		VdcComputePolicy: types.VdcComputePolicy{
			Name:         d.Get("name").(string),
			Description:  getStringAttributeAsPointer(d, "description"),
			IsSizingOnly: false,
		},
		PolicyType: "VdcVmPolicy",
	}

	vmGroups, err := getPvdcNamedVmGroupsMap(d, vcdClient, pVdc)
	if err != nil {
		return diag.FromErr(err)
	}
	computePolicy.PvdcNamedVmGroupsMap = vmGroups

	logicalVmGroups, err := getPvdcLogicalVmGroupsMap(d, vcdClient, pVdc)
	if err != nil {
		return diag.FromErr(err)
	}
	computePolicy.PvdcLogicalVmGroupsMap = logicalVmGroups

	log.Printf("[DEBUG] Creating VM Placement Policy: %#v", computePolicy)

	createdVmSizingPolicy, err := vcdClient.CreateVdcComputePolicyV2(computePolicy)
	if err != nil {
		log.Printf("[DEBUG] Error creating VM Placement Policy: %s", err)
		return diag.Errorf("error creating VM Placement Policy: %s", err)
	}

	d.SetId(createdVmSizingPolicy.VdcComputePolicyV2.ID)
	log.Printf("[TRACE] VM Placement Policy created: %#v", createdVmSizingPolicy.VdcComputePolicyV2)

	return sharedVcdVmPlacementPolicyRead(ctx, d, meta, true)
}

func resourceVmPlacementPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return sharedVcdVmPlacementPolicyRead(ctx, d, meta, true)
}

// sharedVcdVmPlacementPolicyRead is a Read function shared between this resource and the corresponding data source.
func sharedVcdVmPlacementPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}, isResource bool) diag.Diagnostics {
	policyName := d.Get("name").(string)
	pVdcId := d.Get("provider_vdc_id").(string)
	vdcId := ""
	if !isResource {
		vdcId = d.Get("vdc_id").(string)
		log.Printf("[TRACE] VM Placement Policy read initiated: %s in VDC with ID %s", policyName, vdcId)
	} else {
		log.Printf("[TRACE] VM Placement Policy read initiated: %s in Provider VDC with ID %s", policyName, pVdcId)
	}

	if pVdcId == "" && vdcId == "" {
		return diag.Errorf("either `provider_vdc_id` or `vdc_id` needs to be set")
	}

	vcdClient := meta.(*VCDClient)
	// The method variable stores the information about how we found the rule, for logging purposes
	method := "id"

	var policy *govcd.VdcComputePolicyV2
	var err error
	if d.Id() != "" {
		policy, err = vcdClient.GetVdcComputePolicyV2ById(d.Id())
		if err != nil {
			if isResource && govcd.ContainsNotFound(err) {
				d.SetId("")
				return nil
			}
			return diag.Errorf("unable to find VM Placement Policy %s: %s", policyName, err)
		}
	}

	// The secondary method of retrieval is from name
	if d.Id() == "" {
		if policyName == "" {
			return diag.Errorf("both Placement Policy name and ID are empty")
		}
		method = "name"
		queryParams := url.Values{}

		var foundPolicies []*govcd.VdcComputePolicyV2
		var err error
		// Fetches the VM Placement Policy with Provider VDC information, intended for System admins
		if vdcId == "" {
			queryParams.Add("filter", fmt.Sprintf("%spolicyType==VdcVmPolicy;isSizingOnly==false;name==%s;pvdcId==%s", getVgpuFilterToPrepend(vcdClient, false), policyName, pVdcId))
			foundPolicies, err = vcdClient.GetAllVdcComputePoliciesV2(queryParams)
		} else {
			var adminOrg *govcd.AdminOrg
			// Fetches the VM Placement Policy with VDC information, intended for tenants
			adminOrg, err = vcdClient.GetAdminOrgFromResource(d)
			if err != nil {
				return diag.Errorf("error retrieving Org: %s", err)
			}

			var adminVdc *govcd.AdminVdc
			adminVdc, err = adminOrg.GetAdminVDCById(vdcId, false)
			if err != nil {
				return diag.Errorf("unable to get the VDC with ID %s: %s", vdcId, err)
			}
			queryParams.Add("filter", fmt.Sprintf("%spolicyType==VdcVmPolicy;isSizingOnly==false;name==%s", getVgpuFilterToPrepend(vcdClient, false), policyName))
			foundPolicies, err = adminVdc.GetAllAssignedVdcComputePoliciesV2(queryParams)
		}
		if err != nil {
			return diag.Errorf("error getting VM Placement Policy %s: %s. Removing from tfstate", policyName, err)
		}
		if len(foundPolicies) == 0 {
			return diag.Errorf("unable to find VM Placement Policy %s of Provider VDC %s, err: %s", policyName, pVdcId, govcd.ErrorEntityNotFound)
		}
		if len(foundPolicies) > 1 {
			return diag.Errorf("found %d VM Placement Policies with name %s", len(foundPolicies), policyName)
		}
		policy = foundPolicies[0]
		d.SetId(policy.VdcComputePolicyV2.ID)
	}

	if policy == nil {
		return diag.Errorf("error fetching VM Placement Policy with name %s", policyName)
	}
	util.Logger.Printf("[TRACE] [get VM Placement Policy] Retrieved by %s\n", method)
	return setVmPlacementPolicy(ctx, d, vcdClient, *policy.VdcComputePolicyV2)
}

func resourceVmPlacementPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	policyName := d.Get("name").(string)
	log.Printf("[TRACE] VM sizing policy update initiated: %s", policyName)

	vcdClient := meta.(*VCDClient)

	policy, err := vcdClient.GetVdcComputePolicyV2ById(d.Id())
	if err != nil {
		log.Printf("[DEBUG] Unable to find VM Placement Policy %s", policyName)
		return diag.Errorf("unable to find VM Placement Policy %s, error:  %s", policyName, err)
	}

	if d.HasChange("name") {
		policy.VdcComputePolicyV2.Name = d.Get("name").(string)
	}

	if d.HasChange("description") {
		policy.VdcComputePolicyV2.Description = getStringAttributeAsPointer(d, "description")
	}

	if d.HasChange("vm_group_ids") {
		vmGroups, err := getPvdcNamedVmGroupsMap(d, vcdClient, nil)
		if err != nil {
			return diag.FromErr(err)
		}
		policy.VdcComputePolicyV2.PvdcNamedVmGroupsMap = vmGroups
	}
	if d.HasChange("logical_vm_group_ids") {
		logicalVmGroups, err := getPvdcLogicalVmGroupsMap(d, vcdClient, nil)
		if err != nil {
			return diag.FromErr(err)
		}
		policy.VdcComputePolicyV2.PvdcLogicalVmGroupsMap = logicalVmGroups
	}

	_, err = policy.Update()
	if err != nil {
		log.Printf("[DEBUG] Error updating VM Placement Policy %s with error %s", policyName, err)
		return diag.Errorf("error updating VM Placement Policy %s, err: %s", policyName, err)
	}

	log.Printf("[TRACE] VM Placement Policy update completed: %s", policyName)
	return resourceVmPlacementPolicyRead(ctx, d, meta)
}

func resourceVmPlacementPolicyDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	policyName := d.Get("name").(string)
	log.Printf("[TRACE] VM Placement Policy delete started: %s", policyName)

	vcdClient := meta.(*VCDClient)

	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("functionality requires System administrator privileges")
	}

	policy, err := vcdClient.GetVdcComputePolicyV2ById(d.Id())
	if err != nil {
		log.Printf("[DEBUG] Unable to find VM Placement Policy %s. Removing from tfstate", policyName)
		d.SetId("")
		return nil
	}

	err = policy.Delete()
	if err != nil {
		log.Printf("[DEBUG] Error deleting VM Placement Policy %s, err: %s", policyName, err)
		return diag.Errorf("error deleting VM Placement Policy %s, err: %s", policyName, err)
	}

	log.Printf("[TRACE] VM Placement Policy delete completed: %s", policyName)
	return nil
}

var errHelpVmPlacementPolicyImport = fmt.Errorf(`resource id must be specified in one of these formats:
'vm-placement-policy-name', 'vm-placement-policy-id' or 'list@' to get a list of VM placement policies with their IDs`)

// resourceVmPlacementPolicyImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it set's the ID field for `_resource_name_` resource in state file
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_vm_placement_policy.my_existing_policy_name
// Example import path (_the_id_string_): my_existing_vm_placement_policy_id
// Example list path (_the_id_string_): list@
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func resourceVmPlacementPolicyImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)

	log.Printf("[DEBUG] importing VM Placement Policy resource with provided id %s", d.Id())

	if len(resourceURI) != 1 {
		return nil, errHelpVmPlacementPolicyImport
	}
	if strings.Contains(d.Id(), "list@") {

		return listComputePoliciesForImport(meta, "vcd_vm_placement_policy", "placement")
	} else {
		policyId := resourceURI[0]
		return getVmPlacementPolicy(d, meta, policyId)
	}
}

// getPvdcNamedVmGroupsMap fetches the vm_group_ids attribute and retrieves the associated []types.PvdcNamedVmGroupsMap
func getPvdcNamedVmGroupsMap(d *schema.ResourceData, vcdClient *VCDClient, pVdc *govcd.ProviderVdc) ([]types.PvdcNamedVmGroupsMap, error) {
	vmGroupIdsSet, isPopulated := d.GetOk("vm_group_ids")
	if !isPopulated {
		return []types.PvdcNamedVmGroupsMap{}, nil
	}

	vmGroupIdsList := vmGroupIdsSet.(*schema.Set).List()
	if len(vmGroupIdsList) == 0 {
		return []types.PvdcNamedVmGroupsMap{}, nil
	}

	// It is assumed that is a single-item list as there's one pVDC at a time:
	pvdcNamedVmGroupsMap := []types.PvdcNamedVmGroupsMap{
		{
			NamedVmGroups: []types.OpenApiReferences{{}},
			Pvdc: types.OpenApiReference{
				Name: pVdc.ProviderVdc.Name,
				ID:   pVdc.ProviderVdc.ID,
			},
		},
	}
	for _, vmGroupId := range vmGroupIdsList {
		vmGroup, err := vcdClient.GetVmGroupById(vmGroupId.(string))
		if err != nil {
			return nil, fmt.Errorf("error retrieving the associated name of VM Group %s: %s", vmGroupId, err)
		}
		// VM Placement policies use Named VM ID, not the normal ID
		pvdcNamedVmGroupsMap[0].NamedVmGroups[0] = append(pvdcNamedVmGroupsMap[0].NamedVmGroups[0], types.OpenApiReference{
			ID:   fmt.Sprintf("urn:vcloud:namedVmGroup:%s", vmGroup.VmGroup.NamedVmGroupId),
			Name: vmGroup.VmGroup.Name,
		})
	}
	return pvdcNamedVmGroupsMap, nil
}

// getPvdcLogicalVmGroupsMap fetches the logical_vm_group_ids attribute and retrieves the associated []types.PvdcLogicalVmGroupsMap
func getPvdcLogicalVmGroupsMap(d *schema.ResourceData, vcdClient *VCDClient, pVdc *govcd.ProviderVdc) ([]types.PvdcLogicalVmGroupsMap, error) {
	vmGroupIdsSet, isPopulated := d.GetOk("logical_vm_group_ids")
	if !isPopulated {
		return []types.PvdcLogicalVmGroupsMap{}, nil
	}

	vmGroupIdsList := vmGroupIdsSet.(*schema.Set).List()
	if len(vmGroupIdsList) == 0 {
		return []types.PvdcLogicalVmGroupsMap{}, nil
	}

	// It is assumed that is a single-item list as there's one pVDC at a time:
	logicalVmGroupReferences := []types.PvdcLogicalVmGroupsMap{
		{
			LogicalVmGroups: types.OpenApiReferences{},
			Pvdc: types.OpenApiReference{
				Name: pVdc.ProviderVdc.Name,
				ID:   pVdc.ProviderVdc.ID,
			},
		},
	}
	for _, vmGroupId := range vmGroupIdsList {
		logicalVmGroup, err := vcdClient.GetLogicalVmGroupById(vmGroupId.(string))
		if err != nil {
			return nil, fmt.Errorf("error retrieving the associated name of Logical VM Group %s: %s", vmGroupId, err)
		}
		// It is assumed that is a single-item list as there's one pVDC at a time:
		logicalVmGroupReferences[0].LogicalVmGroups = append(logicalVmGroupReferences[0].LogicalVmGroups, types.OpenApiReference{
			ID:   vmGroupId.(string),
			Name: logicalVmGroup.LogicalVmGroup.Name,
		})
	}
	return logicalVmGroupReferences, nil
}

// getVmPlacementPolicy reads the corresponding VM Placement Policy from the resource.
func getVmPlacementPolicy(d *schema.ResourceData, meta interface{}, policyId string) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	var computePolicy *govcd.VdcComputePolicyV2
	var err error
	computePolicy, err = vcdClient.GetVdcComputePolicyV2ById(policyId)
	if err != nil {
		queryParams := url.Values{}
		queryParams.Add("filter", fmt.Sprintf("%sname==%s;isSizingOnly==false", getVgpuFilterToPrepend(vcdClient, false), policyId))
		computePolicies, err := vcdClient.GetAllVdcComputePoliciesV2(queryParams)
		if err != nil {
			log.Printf("[DEBUG] Unable to find VM Placement Policy %s", policyId)
			return nil, fmt.Errorf("unable to find VM Placement Policy %s, err: %s", policyId, err)
		}
		if len(computePolicies) != 1 {
			log.Printf("[DEBUG] Unable to find unique VM Placement Policy %s", policyId)
			return nil, fmt.Errorf("unable to find unique VM Placement Policy %s, err: %s", policyId, err)
		}
		computePolicy = computePolicies[0]
	}

	dSet(d, "name", computePolicy.VdcComputePolicyV2.Name)
	dSet(d, "provider_vdc_id", computePolicy.VdcComputePolicyV2.PvdcID)
	var vmGroupIds []string
	for _, pvdcNamedVmGroupsMap := range computePolicy.VdcComputePolicyV2.PvdcNamedVmGroupsMap {
		for _, namedVmGroups := range pvdcNamedVmGroupsMap.NamedVmGroups {
			for _, namedVmGroup := range namedVmGroups {
				vmGroupIds = append(vmGroupIds, namedVmGroup.ID)
			}
		}
	}
	if err = d.Set("vm_group_ids", vmGroupIds); err != nil {
		return nil, fmt.Errorf("error setting vm_group_ids: %s", err)
	}

	vmGroupIds = []string{}
	for _, pvdcLogicalVmGroupsMap := range computePolicy.VdcComputePolicyV2.PvdcLogicalVmGroupsMap {
		for _, logicalVmGroup := range pvdcLogicalVmGroupsMap.LogicalVmGroups {
			vmGroupIds = append(vmGroupIds, logicalVmGroup.ID)
		}
	}
	if err = d.Set("logical_vm_group_ids", vmGroupIds); err != nil {
		return nil, fmt.Errorf("error setting logical_vm_group_ids: %s", err)
	}

	d.SetId(computePolicy.VdcComputePolicyV2.ID)

	return []*schema.ResourceData{d}, nil
}

// setVmPlacementPolicy sets the Terraform state from the Compute Policy input parameter
func setVmPlacementPolicy(_ context.Context, d *schema.ResourceData, vcdClient *VCDClient, policy types.VdcComputePolicyV2) diag.Diagnostics {
	dSet(d, "name", policy.Name)
	dSet(d, "description", policy.Description)

	var vmGroupIds []string

	for _, namedVmGroupPerPvdc := range policy.NamedVMGroups {
		for _, namedVmGroup := range namedVmGroupPerPvdc {
			// The Policy has "Named VM Group IDs" in its attributes, but we need "VM Group IDs" which are unique
			vmGroup, err := vcdClient.VCDClient.GetVmGroupByNamedVmGroupIdAndProviderVdcUrn(namedVmGroup.ID, policy.PvdcID)
			if err != nil {
				return diag.Errorf("could not get VM Group associated to Named VM Group ID %s", namedVmGroup.ID)
			}
			vmGroupIds = append(vmGroupIds, vmGroup.VmGroup.ID)
		}
	}
	if err := d.Set("vm_group_ids", vmGroupIds); err != nil {
		return diag.Errorf("error setting vm_group_ids: %s", err)
	}

	vmGroupIds = []string{}
	for _, namedVmGroup := range policy.LogicalVMGroupReferences {
		vmGroupIds = append(vmGroupIds, namedVmGroup.ID)
	}
	if err := d.Set("logical_vm_group_ids", vmGroupIds); err != nil {
		return diag.Errorf("error setting logical_vm_group_ids: %s", err)
	}

	log.Printf("[TRACE] VM Placement Policy read completed: %s", policy.Name)
	return nil
}

// getVgpuFilterToPrepend gets a vGPU Policy filter set to `isVgpu` if API version of target VCD is greater than 36.2 (VCD 10.3.2).
// The semicolon is placed to the right so the returned filter must be prepended to an existing one.
// Returns an empty string otherwise.
// Note: This function should not be needed anymore once VCD 10.3.0 and 10.3.1 are discontinued.
func getVgpuFilterToPrepend(vcdClient *VCDClient, isVgpu bool) string {
	if vcdClient.Client.APIVCDMaxVersionIs(">= 36.2") {
		return fmt.Sprintf("isVgpuPolicy==%s;", strconv.FormatBool(isVgpu))
	}
	return ""
}

// getVgpuFilter gets a vGPU Policy filter set to `isVgpu` if API version of target VCD is greater than 36.2 (VCD 10.3.2).
// The filter option is returned WITHOUT any semicolon.
// Returns an empty string otherwise.
// Note: This function should not be needed anymore once VCD 10.3.0 and 10.3.1 are discontinued.
func getVgpuFilter(vcdClient *VCDClient, isVgpu bool) string {
	vgpuFilterToPrepend := getVgpuFilterToPrepend(vcdClient, isVgpu)
	if vgpuFilterToPrepend != "" {
		return vgpuFilterToPrepend[:len(vgpuFilterToPrepend)-1] // Removes semicolon to the right
	}
	return ""
}
