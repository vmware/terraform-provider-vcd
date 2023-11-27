package vcd

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/vmware/go-vcloud-director/v2/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdVmVgpuPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdVmVgpuPolicyCreate,
		DeleteContext: resourceVcdVmVgpuPolicyDelete,
		ReadContext:   resourceVcdVmVgpuPolicyRead,
		UpdateContext: resourceVcdVmVgpuPolicyUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdVmVgpuPolicyImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The unique name of the vGPU policy.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the vGPU policy.",
			},
			"vgpu_profile": {
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				MaxItems:    1,
				Description: "Defines the vGPU profile configuration. Only one profile is allowed.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							ForceNew:    true,
							Required:    true,
							Description: "The identifier of the vGPU profile.",
						},
						"count": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(1),
							Description:  "Specifies the number of vGPU profiles. Must be at least 1.",
						},
					},
				},
			},
			"cpu": {
				Type:        schema.TypeList,
				Optional:    true,
				MinItems:    0,
				MaxItems:    1,
				ForceNew:    true,
				Description: "Configuration options for CPU resources.",
				Elem:        sizingPolicyCpu,
			},
			"memory": {
				Type:        schema.TypeList,
				Optional:    true,
				MinItems:    0,
				MaxItems:    1,
				ForceNew:    true,
				Description: "Memory resource configuration settings.",
				Elem:        sizingPolicyMemory,
			},
			"provider_vdc_scope": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Defines the scope of the policy within provider virtual data centers.",
				Elem:        providerVdcScope,
			},
		},
	}
}

var providerVdcScope = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"provider_vdc_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Identifier for the provider virtual data center.",
		},
		"cluster_names": {
			Type:        schema.TypeSet,
			Optional:    true,
			Computed:    true,
			Description: "Set of cluster names within the provider virtual data center.",
			Elem: &schema.Schema{
				MinItems: 1,
				Type:     schema.TypeString,
			},
		},
		"vm_group_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Optional identifier for a VM group within the provider VDC scope.",
		},
	},
}

func resourceVcdVmVgpuPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	log.Printf("[DEBUG] Creating VM vGPU policy: %#v", params)

	createdVmVgpuPolicy, err := vcdClient.CreateVdcComputePolicyV2(params)
	if err != nil {
		log.Printf("[DEBUG] Error VM vGPU policy: %s", err)
		return diag.Errorf("error VM vGPU policy: %s", err)
	}

	d.SetId(createdVmVgpuPolicy.VdcComputePolicyV2.ID)
	log.Printf("[TRACE] VM vGPU policy created: %#v", createdVmVgpuPolicy.VdcComputePolicyV2)

	return resourceVcdVmVgpuPolicyRead(ctx, d, meta)
}

// resourceVmSizingPolicyRead reads a resource VM Sizing Policy
func resourceVcdVmVgpuPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdVgpuPolicyRead(ctx, d, meta)
}

// Fetches information about an existing VM vGPU policy for a data definition
func genericVcdVgpuPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	policyName := d.Get("name").(string)
	log.Printf("[TRACE] VM vGPU policy read initiated: %s", policyName)

	vcdClient := meta.(*VCDClient)

	// The method variable stores the information about how we found the rule, for logging purposes
	method := "id"

	var policy *govcd.VdcComputePolicyV2
	var err error
	if d.Id() != "" {
		policy, err = vcdClient.GetVdcComputePolicyV2ById(d.Id())
		if err != nil {
			log.Printf("[DEBUG] Unable to find VM vGPU policy %s. Removing from tfstate.", policyName)
			d.SetId("")
			return diag.Errorf("unable to find VM vGPU policy %s, err: %s. Removing from tfstate", policyName, err)
		}
	}

	// The secondary method of retrieval is from name
	if d.Id() == "" {
		if policyName == "" {
			return diag.Errorf("both name and ID are empty")
		}
		method = "name"
		queryParams := url.Values{}
		queryParams.Add("filter", fmt.Sprintf("name==%s;isVgpuPolicy==true", policyName))
		filteredPoliciesByName, err := vcdClient.GetAllVdcComputePoliciesV2(queryParams)
		if err != nil {
			log.Printf("[DEBUG] Unable to find VM vGPU policy %s. Removing from tfstate.", policyName)
			d.SetId("")
			return diag.Errorf("unable to find VM vGPU policy %s, err: %s. Removing from tfstate", policyName, err)
		}
		if len(filteredPoliciesByName) != 1 {
			log.Printf("[DEBUG] Unable to find VM vGPU policy %s . Found Policies by name: %d. Removing from tfstate.", policyName, len(filteredPoliciesByName))
			d.SetId("")
			return diag.Errorf("[DEBUG] Unable to find VM vGPU policy %s, err: %s. Found Policies by name: %d. Removing from tfstate", policyName, govcd.ErrorEntityNotFound, len(filteredPoliciesByName))
		}
		policy = filteredPoliciesByName[0]
		d.SetId(policy.VdcComputePolicyV2.ID)
	}

	// Fix coverity warning
	if policy == nil {
		return diag.Errorf("[genericVcdVgpuPolicyRead] error defining vGPU policy")
	}
	util.Logger.Printf("[TRACE] [get VM vGPU policy] Retrieved by %s\n", method)
	return setVgpuPolicy(ctx, d, policy.VdcComputePolicyV2)
}

// resourceVmVgpuPolicyUpdate function updates resource with found configurations changes
func resourceVcdVmVgpuPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	policyName := d.Get("name").(string)
	log.Printf("[TRACE] VM vGPU policy update initiated: %s", policyName)

	vcdClient := meta.(*VCDClient)

	policy, err := vcdClient.GetVdcComputePolicyV2ById(d.Id())
	if err != nil {
		log.Printf("[DEBUG] Unable to find VM vGPU policy %s", policyName)
		return diag.Errorf("unable to find VM vGPU policy %s, error:  %s", policyName, err)
	}

	changedPolicy, err := getUpdatedVgpuPolicyInput(d, vcdClient, policy)
	if err != nil {
		log.Printf("[DEBUG] Error updating VM vGPU policy %s with error %s", policyName, err)
		return diag.Errorf("error updating VM vGPU policy %s, err: %s", policyName, err)
	}

	_, err = changedPolicy.Update()
	if err != nil {
		log.Printf("[DEBUG] Error updating VM vGPU policy %s with error %s", policyName, err)
		return diag.Errorf("error updating VM vGPU policy %s, err: %s", policyName, err)
	}

	log.Printf("[TRACE] VM vGPU policy update completed: %s", policyName)
	return resourceVcdVmVgpuPolicyRead(ctx, d, meta)
}

// Deletes a VM vGPU policy
func resourceVcdVmVgpuPolicyDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	policyName := d.Get("name").(string)
	log.Printf("[TRACE] VM vGPU policy delete started: %s", policyName)

	vcdClient := meta.(*VCDClient)

	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("functionality requires System administrator privileges")
	}

	policy, err := vcdClient.GetVdcComputePolicyV2ById(d.Id())
	if err != nil {
		log.Printf("[DEBUG] Unable to find VM vGPU policy %s. Removing from tfstate", policyName)
		d.SetId("")
		return nil
	}

	err = policy.Delete()
	if err != nil {
		log.Printf("[DEBUG] Error removing VM vGPU policy %s, err: %s", policyName, err)
		return diag.Errorf("error removing VM vGPU policy %s, err: %s", policyName, err)
	}

	log.Printf("[TRACE] VM vGPU policy delete completed: %s", policyName)
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

	clustersPart, vmGroupsPart, err := getVgpuClustersAndVmGroups(d, vcdClient)
	if err != nil {
		return nil, fmt.Errorf("error getting vgpu clusters map: %s", err)
	}

	policy := &types.VdcComputePolicyV2{
		VdcComputePolicy:     *params,
		PolicyType:           "VdcVmPolicy",
		PvdcNamedVmGroupsMap: vmGroupsPart,
		PvdcVgpuClustersMap:  clustersPart,
		VgpuProfiles: []types.VgpuProfile{
			*vgpuProfile,
		},
	}

	return policy, nil
}

func getUpdatedVgpuPolicyInput(d *schema.ResourceData, vcdClient *VCDClient, policy *govcd.VdcComputePolicyV2) (*govcd.VdcComputePolicyV2, error) {
	policy.VdcComputePolicyV2.Name = d.Get("name").(string)
	policy.VdcComputePolicyV2.Description = getStringAttributeAsPointer(d, "description")

	clustersPart, vmGroupsPart, err := getVgpuClustersAndVmGroups(d, vcdClient)
	if err != nil {
		return nil, fmt.Errorf("error getting vgpu clusters map: %s", err)
	}
	policy.VdcComputePolicyV2.NamedVMGroups = nil
	policy.VdcComputePolicyV2.PvdcVgpuClustersMap = clustersPart
	policy.VdcComputePolicyV2.PvdcNamedVmGroupsMap = vmGroupsPart

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

func setVgpuPolicy(ctx context.Context, d *schema.ResourceData, vgpuPolicy *types.VdcComputePolicyV2) diag.Diagnostics {
	var diags diag.Diagnostics
	diags = append(diags, setVmSizingPolicy(ctx, d, vgpuPolicy.VdcComputePolicy)...)

	diags = append(diags, setVgpuProfile(d, vgpuPolicy.VgpuProfiles)...)

	diags = append(diags, setPvdcClusterScope(d, vgpuPolicy)...)

	if len(diags) != 0 {
		return nil
	}
	return nil
}

func setVgpuProfile(d *schema.ResourceData, vgpuProfiles []types.VgpuProfile) diag.Diagnostics {
	var vgpuProfileList []map[string]interface{}
	vgpuProfileMap := make(map[string]interface{})

	vgpuProfileMap["id"] = vgpuProfiles[0].Id
	vgpuProfileMap["count"] = vgpuProfiles[0].Count

	vgpuProfileList = append(vgpuProfileList, vgpuProfileMap)
	err := d.Set("vgpu_profile", vgpuProfileList)
	if err != nil {
		return diag.Errorf("error setting vgpu profile: %s", err)
	}

	return nil
}

func setPvdcClusterScope(d *schema.ResourceData, vgpuPolicy *types.VdcComputePolicyV2) diag.Diagnostics {
	pvdcClusters := make([]interface{}, len(vgpuPolicy.PvdcVgpuClustersMap))
	for index, cluster := range vgpuPolicy.PvdcVgpuClustersMap {
		singleScope := make(map[string]interface{})
		singleScope["provider_vdc_id"] = cluster.Pvdc.ID
		clusterSet := convertStringsToTypeSet(cluster.Clusters)
		singleScope["cluster_names"] = clusterSet

		for _, vmGroup := range vgpuPolicy.PvdcNamedVmGroupsMap {
			if vmGroup.Pvdc.ID == cluster.Pvdc.ID {
				singleScope["vm_group_id"] = vmGroup.NamedVmGroups[0][0].ID
			}
		}

		pvdcClusters[index] = singleScope
	}

	err := d.Set("provider_vdc_scope", pvdcClusters)
	if err != nil {
		return diag.Errorf("error setting provider_vdc_scope: %s", err)
	}

	return nil
}

func getVgpuClustersAndVmGroups(d *schema.ResourceData, vcdClient *VCDClient) ([]types.PvdcVgpuClustersMap, []types.PvdcNamedVmGroupsMap, error) {
	vgpuClusterSet := d.Get("provider_vdc_scope").(*schema.Set)
	vgpuClusters := make([]types.PvdcVgpuClustersMap, len(vgpuClusterSet.List()))
	var namedVmGroups []types.PvdcNamedVmGroupsMap
	for rangeIndex, vgpuCluster := range vgpuClusterSet.List() {
		vgpuClusterDefinition := vgpuCluster.(map[string]interface{})

		pvdcId := vgpuClusterDefinition["provider_vdc_id"].(string)

		clusterNameSet := vgpuClusterDefinition["cluster_names"].(*schema.Set)
		clusterNames := convertSchemaSetToSliceOfStrings(clusterNameSet)
		providerVdc, err := vcdClient.GetProviderVdcById(pvdcId)
		if err != nil {
			return nil, nil, err
		}

		// Helper function so the cluster names field isn't mandatory. If none are provided, VCD chooses one by default
		if len(clusterNames) == 0 {
			vgpuProfile, err := vcdClient.GetVgpuProfileById(d.Get("vgpu_profile.0.id").(string))
			if err != nil {
				return nil, nil, err
			}

			queryParams := url.Values{}
			queryParams.Add("filter", fmt.Sprintf("(vgpuProfileName==%s;name==%s)", vgpuProfile.VgpuProfile.Name, providerVdc.ProviderVdc.Name))
			vcenters, err := vcdClient.GetAllVCenters(nil)
			if err != nil {
				return nil, nil, err
			}
			var resourcePools []*govcd.ResourcePool
			for _, vcenter := range vcenters {
				rps, err := vcenter.GetAllResourcePools(queryParams)
				if err != nil {
					return nil, nil, err
				}
				resourcePools = append(resourcePools, rps...)
			}
			if len(resourcePools) == 0 {
				return nil, nil, fmt.Errorf("no resource pools for the provider vdc were found")
			}

			if len(resourcePools) >= 2 {
				return nil, nil, fmt.Errorf("more than one resource pool for the provider vdc name was found")
			}

			clusterNames = append(clusterNames, resourcePools[0].ResourcePool.ClusterMoref)
		}

		onePvdcCluster := types.PvdcVgpuClustersMap{
			Clusters: clusterNames,
			Pvdc: types.OpenApiReference{
				ID:   pvdcId,
				Name: providerVdc.ProviderVdc.Name,
			},
		}

		vmGroupId := vgpuClusterDefinition["vm_group_id"].(string)
		if vmGroupId != "" {
			vmGroup, err := vcdClient.GetVmGroupById(vmGroupId)
			if err != nil {
				return nil, nil, err
			}
			namedVmGroups = append(namedVmGroups, types.PvdcNamedVmGroupsMap{
				Pvdc: types.OpenApiReference{
					ID:   pvdcId,
					Name: providerVdc.ProviderVdc.Name,
				},
				NamedVmGroups: []types.OpenApiReferences{
					{
						{
							Name: vmGroup.VmGroup.Name,
							ID:   "urn:vcloud:namedVmGroup:" + vmGroup.VmGroup.ID,
						},
					},
				},
			})

		}
		vgpuClusters[rangeIndex] = onePvdcCluster
	}

	return vgpuClusters, namedVmGroups, nil
}

var errHelpVmVgpuPolicyImport = fmt.Errorf(`resource id must be specified in one of these formats:
'vm-vgpu-policy-name', 'vm-vgpu-policy-id' or 'list@' to get a list of VM vgpu policies with their IDs`)

func resourceVcdVmVgpuPolicyImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)

	log.Printf("[DEBUG] importing VM vGPU policy resource with provided id %s", d.Id())

	if len(resourceURI) != 1 {
		return nil, errHelpVmVgpuPolicyImport
	}
	if strings.Contains(d.Id(), "list@") {
		return listComputePoliciesForImport(meta, "vcd_vm_vgpu_policy", "vgpu")
	} else {
		policyId := resourceURI[0]
		return getVmVgpuPolicy(d, meta, policyId)
	}
}

func getVmVgpuPolicy(d *schema.ResourceData, meta interface{}, policyId string) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	var computePolicy *govcd.VdcComputePolicyV2
	var err error
	computePolicy, err = vcdClient.GetVdcComputePolicyV2ById(policyId)
	if err != nil {
		queryParams := url.Values{}
		queryParams.Add("filter", fmt.Sprintf("name==%s;isVgpuPolicy==true", policyId))
		computePolicies, err := vcdClient.GetAllVdcComputePoliciesV2(queryParams)
		if err != nil {
			log.Printf("[DEBUG] Unable to find VM vGPU Policy %s", policyId)
			return nil, fmt.Errorf("unable to find VM vGPU Policy %s, err: %s", policyId, err)
		}
		if len(computePolicies) != 1 {
			log.Printf("[DEBUG] Unable to find unique VM vGPU Policy %s", policyId)
			return nil, fmt.Errorf("unable to find unique VM vGPU Policy %s, err: %s", policyId, err)
		}
		computePolicy = computePolicies[0]
	}

	dSet(d, "name", computePolicy.VdcComputePolicyV2.Name)
	d.SetId(computePolicy.VdcComputePolicyV2.ID)

	return []*schema.ResourceData{d}, nil
}
