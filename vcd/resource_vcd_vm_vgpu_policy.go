package vcd

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/vmware/go-vcloud-director/v2/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdVmVgpuPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVmSizingPolicyCreate,
		DeleteContext: resourceVmSizingPolicyDelete,
		ReadContext:   resourceVmSizingPolicyRead,
		UpdateContext: resourceVmSizingPolicyUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVmSizingPolicyImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"vgpu_profile": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							ForceNew: true,
							Required: true,
						},
						"count": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: IsIntAndAtLeast(1),
						},
					},
				},
			},
			"cpu": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 0,
				MaxItems: 1,
				Elem:     sizingPolicyCpu,
			},
			"memory": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 0,
				MaxItems: 1,
				Elem:     sizingPolicyMemory,
			},
			"provider_vdc_scope": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     providerVdcScope,
			},
			"vm_group_scope": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     vmGroupScope,
			},
		},
	}
}

var providerVdcScope = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"provider_vdc_id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"cluster_names": {
			Type:     schema.TypeSet,
			Required: true,
			Elem:     schema.TypeString,
		},
	},
}

var vmGroupScope = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"provider_vdc_id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"vm_group_name": {
			Type:     schema.TypeString,
			Required: true,
		},
	},
}

func resourceVmVgpuPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	policyName := d.Get("name").(string)
	log.Printf("[TRACE] VM vGPU policy creation initiated: %s", policyName)

	vcdClient := meta.(*VCDClient)

	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("functionality requires System administrator privileges")
	}

	params, err := getVgpuPolicyInput(d, vcdClient)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Creating VM sizing policy: %#v", params)

	createdVmSizingPolicy, err := vcdClient.CreateVdcComputePolicyV2(params)
	if err != nil {
		log.Printf("[DEBUG] Error VM sizing policy: %s", err)
		return diag.Errorf("error VM sizing policy: %s", err)
	}

	d.SetId(createdVmSizingPolicy.VdcComputePolicyV2.ID)
	log.Printf("[TRACE] VM sizing policy created: %#v", createdVmSizingPolicy.VdcComputePolicyV2)

	return resourceVmSizingPolicyRead(ctx, d, meta)
}

// resourceVmSizingPolicyRead reads a resource VM Sizing Policy
func resourceVmVgpuPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdVmSizingPolicyRead(ctx, d, meta)
}

// Fetches information about an existing VM sizing policy for a data definition
func genericVcdVgpuPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	policyName := d.Get("name").(string)
	log.Printf("[TRACE] VM sizing policy read initiated: %s", policyName)

	vcdClient := meta.(*VCDClient)

	// The method variable stores the information about how we found the rule, for logging purposes
	method := "id"

	var policy *govcd.VdcComputePolicy
	var err error
	if d.Id() != "" {
		policy, err = vcdClient.Client.GetVdcComputePolicyById(d.Id())
		if err != nil {
			log.Printf("[DEBUG] Unable to find VM sizing policy %s. Removing from tfstate.", policyName)
			d.SetId("")
			return diag.Errorf("unable to find VM sizing policy %s, err: %s. Removing from tfstate", policyName, err)
		}
	}

	// The secondary method of retrieval is from name
	if d.Id() == "" {
		if policyName == "" {
			return diag.Errorf("both name and ID are empty")
		}
		method = "name"
		queryParams := url.Values{}
		queryParams.Add("filter", fmt.Sprintf("name==%s;isSizingOnly==true", policyName))
		filteredPoliciesByName, err := vcdClient.Client.GetAllVdcComputePolicies(queryParams)
		if err != nil {
			log.Printf("[DEBUG] Unable to find VM sizing policy %s. Removing from tfstate.", policyName)
			d.SetId("")
			return diag.Errorf("unable to find VM sizing policy %s, err: %s. Removing from tfstate", policyName, err)
		}
		if len(filteredPoliciesByName) != 1 {
			log.Printf("[DEBUG] Unable to find VM sizing policy %s . Found Policies by name: %d. Removing from tfstate.", policyName, len(filteredPoliciesByName))
			d.SetId("")
			return diag.Errorf("[DEBUG] Unable to find VM sizing policy %s, err: %s. Found Policies by name: %d. Removing from tfstate", policyName, govcd.ErrorEntityNotFound, len(filteredPoliciesByName))
		}
		policy = filteredPoliciesByName[0]
		d.SetId(policy.VdcComputePolicy.ID)
	}

	// Fix coverity warning
	if policy == nil {
		return diag.Errorf("[genericVcdVmSizingPolicyRead] error defining sizing policy")
	}
	util.Logger.Printf("[TRACE] [get VM sizing policy] Retrieved by %s\n", method)
	return setVmSizingPolicy(ctx, d, *policy.VdcComputePolicy)
}

// resourceVmVgpuPolicyUpdate function updates resource with found configurations changes
func resourceVmVgpuPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	policyName := d.Get("name").(string)
	log.Printf("[TRACE] VM sizing policy update initiated: %s", policyName)

	vcdClient := meta.(*VCDClient)

	policy, err := vcdClient.Client.GetVdcComputePolicyById(d.Id())
	if err != nil {
		log.Printf("[DEBUG] Unable to find VM sizing policy %s", policyName)
		return diag.Errorf("unable to find VM sizing policy %s, error:  %s", policyName, err)
	}

	changedPolicy, err := getUpdatedVmSizingPolicyInput(d, policy)
	if err != nil {
		log.Printf("[DEBUG] Error updating VM sizing policy %s with error %s", policyName, err)
		return diag.Errorf("error updating VM sizing policy %s, err: %s", policyName, err)
	}

	_, err = changedPolicy.Update()
	if err != nil {
		log.Printf("[DEBUG] Error updating VM sizing policy %s with error %s", policyName, err)
		return diag.Errorf("error updating VM sizing policy %s, err: %s", policyName, err)
	}

	log.Printf("[TRACE] VM sizing policy update completed: %s", policyName)
	return resourceVmSizingPolicyRead(ctx, d, meta)
}

// Deletes a VM vGPU policy
func resourceVmVgpuPolicyDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	policyName := d.Get("name").(string)
	log.Printf("[TRACE] VM sizing policy delete started: %s", policyName)

	vcdClient := meta.(*VCDClient)

	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("functionality requires System administrator privileges")
	}

	policy, err := vcdClient.Client.GetVdcComputePolicyById(d.Id())
	if err != nil {
		log.Printf("[DEBUG] Unable to find VM sizing policy %s. Removing from tfstate", policyName)
		d.SetId("")
		return nil
	}

	err = policy.Delete()
	if err != nil {
		log.Printf("[DEBUG] Error removing VM sizing policy %s, err: %s", policyName, err)
		return diag.Errorf("error removing VM sizing policy %s, err: %s", policyName, err)
	}

	log.Printf("[TRACE] VM sizing policy delete completed: %s", policyName)
	return nil
}

func getVgpuPolicyInput(d *schema.ResourceData, vcdClient *VCDClient) (*types.VdcComputePolicyV2, error) {
	params := &types.VdcComputePolicy{
		Name:        d.Get("name").(string),
		Description: getStringAttributeAsPointer(d, "description"),
	}

	cpuPart := d.Get("cpu").([]interface{})
	if len(cpuPart) == 1 {
		var err error
		params, err = getCpuInput(cpuPart, params)
		if err != nil {
			return nil, err
		}
	}

	memoryPart := d.Get("memory").([]interface{})
	if len(memoryPart) == 1 {
		var err error
		params, err = getMemoryInput(memoryPart, params)
		if err != nil {
			return nil, err
		}
	}

	vgpuProfilePart := d.Get("vgpu_profile").([]interface{})
	vgpuProfile, err := getVgpuProfile(vgpuProfilePart, vcdClient)
	if err != nil {
		return nil, err
	}

	policy := &types.VdcComputePolicyV2{
		VdcComputePolicy:     *params,
		PolicyType:           "VcdVmPolicy",
		IsVgpuPolicy:         true,
		PvdcNamedVmGroupsMap: []types.PvdcNamedVmGroupsMap{},
		PvdcVgpuClustersMap:  []types.PvdcVgpuClustersMap{},
		VgpuProfiles: []types.VgpuProfile{
			*vgpuProfile,
		},
	}

	return policy, nil
}

func getVgpuProfile(vgpuProfile []interface{}, vcdClient *VCDClient) (*types.VgpuProfile, error) {
	profileMap := vgpuProfile[0].(map[string]interface{})
	profileId := profileMap["id"].(string)

	profile, err := vcdClient.GetVgpuProfileById(profileId)
	if err != nil {
		return nil, err
	}
	profile.VgpuProfile.Count = profileMap["count"].(int)

	return profile.VgpuProfile, nil
}

func getVgpuNamedVmGroups(d *schema.ResourceData, vcdClient *VCDClient) ([]types.PvdcNamedVmGroupsMap, error) {
	vmGroupSet := d.Get("vm_group_scope").(*schema.Set)
	vmGroups := make([]types.PvdcNamedVmGroupsMap, len(vmGroupSet.List()))
	for rangeIndex, vmGroup := range vmGroupSet.List() {
		vmGroupDefinition := vmGroup.(map[string]interface{})
		groupName := vmGroupDefinition["vm_group_name"].(string)
		pvdcId := vmGroupDefinition["provider_vdc_id"].(string)
		group, err := vcdClient.GetVmGroupByNameAndProviderVdcUrn(groupName, pvdcId)
		if err != nil {
			return nil, err
		}
		oneVmGroup := types.PvdcNamedVmGroupsMap{
			NamedVmGroups: []types.OpenApiReferences{
				{
					{
						Name: group.VmGroup.Name,
						ID:   group.VmGroup.ID,
					},
				},
			},
			Pvdc: types.OpenApiReference{
				ID: pvdcId,
			},
		}
		vmGroups[rangeIndex] = oneVmGroup
	}

	return vmGroups, nil
}

func getVgpuClustersMap(d *schema.ResourceData, vcdClient *VCDClient) ([]types.PvdcVgpuClustersMap, error) {
	vgpuClusterSet := d.Get("provider_vdc_scope").(*schema.Set)
	vgpuClusters := make([]types.PvdcVgpuClustersMap, len(vgpuClusterSet.List()))
	for rangeIndex, vgpuCluster := range vgpuClusterSet.List() {
		vgpuClusterDefinition := vgpuCluster.(map[string]interface{})

		pvdcId := vgpuClusterDefinition["provider_vdc_id"].(string)

		clusterNameSet := vgpuClusterDefinition["cluster_names"].(*schema.Set)
		clusterNames := convertSchemaSetToSliceOfStrings(clusterNameSet)
		providerVdc, err := vcdClient.GetProviderVdcById(pvdcId)
		if err != nil {
			return nil, err
		}
		providerVdc.
			resPools, err := vcdClient.GetAllResourcePools(nil)
		if err != nil {
			return nil, err
		}
		oneVmGroup := types.PvdcNamedVmGroupsMap{
			NamedVmGroups: []types.OpenApiReferences{
				{
					{
						Name: group.VmGroup.Name,
						ID:   group.VmGroup.ID,
					},
				},
			},
			Pvdc: types.OpenApiReference{
				ID: pvdcId,
			},
		}
		vgpuClusters[rangeIndex] = oneVmGroup
	}

	return vgpuClusters, nil
}
