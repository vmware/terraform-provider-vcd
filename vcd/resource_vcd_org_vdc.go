package vcd

//lint:file-ignore SA1019 ignore deprecated functions
import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

func resourceVcdOrgVdc() *schema.Resource {
	capacityWithUsage := schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"allocated": {
					Type:        schema.TypeInt,
					Optional:    true,
					Computed:    true,
					Description: "Capacity that is committed to be available. Value in MB or MHz. Used with AllocationPool (Allocation pool) and ReservationPool (Reservation pool).",
				},
				"limit": {
					Type:        schema.TypeInt,
					Optional:    true,
					Computed:    true,
					Description: "Capacity limit relative to the value specified for Allocation. It must not be less than that value. If it is greater than that value, it implies over provisioning. A value of 0 specifies unlimited units. Value in MB or MHz. Used with AllocationVApp (Pay as you go).",
				},
				"reserved": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"used": {
					Type:     schema.TypeInt,
					Computed: true,
				},
			},
		},
	}

	return &schema.Resource{
		CreateContext: resourceVcdVdcCreate,
		DeleteContext: resourceVcdVdcDelete,
		ReadContext:   resourceVcdVdcRead,
		UpdateContext: resourceVcdVdcUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdOrgVdcImport,
		},
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"allocation_model": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"AllocationVApp", "AllocationPool", "ReservationPool", "Flex"}, false),
				Description:  "The allocation model used by this VDC; must be one of {AllocationVApp, AllocationPool, ReservationPool, Flex}",
			},
			"compute_capacity": {
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu":    &capacityWithUsage,
						"memory": &capacityWithUsage,
					},
				},
				Description: "The compute capacity allocated to this VDC.",
			},
			"nic_quota": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Maximum number of virtual NICs allowed in this VDC. Defaults to 0, which specifies an unlimited number.",
			},
			"network_quota": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Maximum number of network objects that can be deployed in this VDC. Defaults to 0, which means no networks can be deployed.",
			},
			"vm_quota": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The maximum number of VMs that can be created in this VDC. Includes deployed and undeployed VMs in vApps and vApp templates. Defaults to 0, which specifies an unlimited number.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "True if this VDC is enabled for use by the organization VDCs. Default is true.",
			},
			"storage_profile": {
				Type:        schema.TypeSet,
				Required:    true,
				ForceNew:    false,
				MinItems:    1,
				Description: "Storage profiles supported by this VDC.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of Provider VDC storage profile.",
						},
						"enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "True if this storage profile is enabled for use in the VDC.",
						},
						"limit": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Maximum number of MB allocated for this storage profile. A value of 0 specifies unlimited MB.",
						},
						"default": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "True if this is default storage profile for this VDC. The default storage profile is used when an object that can specify a storage profile is created with no storage profile specified.",
						},
						"storage_used_in_mb": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Storage used in MB",
						},
					},
				},
			},
			"memory_guaranteed": {
				Type:     schema.TypeFloat,
				Computed: true,
				Optional: true,
				Description: "Percentage of allocated memory resources guaranteed to vApps deployed in this VDC. " +
					"For example, if this value is 0.75, then 75% of allocated resources are guaranteed. " +
					"Required when AllocationModel is AllocationVApp or AllocationPool. When Allocation model is AllocationPool minimum value is 0.2. If the element is empty, vCD sets a value.",
			},
			"cpu_guaranteed": {
				Type:     schema.TypeFloat,
				Optional: true,
				Computed: true,
				Description: "Percentage of allocated CPU resources guaranteed to vApps deployed in this VDC. " +
					"For example, if this value is 0.75, then 75% of allocated resources are guaranteed. " +
					"Required when AllocationModel is AllocationVApp or AllocationPool. If the element is empty, vCD sets a value",
			},
			"cpu_speed": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Specifies the clock frequency, in Megahertz, for any virtual CPU that is allocated to a VM. A VM with 2 vCPUs will consume twice as much of this value. Ignored for ReservationPool. Required when AllocationModel is AllocationVApp or AllocationPool, and may not be less than 256 MHz. Defaults to 1000 MHz if the element is empty or missing.",
			},
			"enable_thin_provisioning": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Boolean to request thin provisioning. Request will be honored only if the underlying datastore supports it. Thin provisioning saves storage space by committing it on demand. This allows over-allocation of storage.",
			},
			"network_pool_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of a network pool in the Provider VDC. Required if this VDC will contain routed or isolated networks.",
			},
			"provider_vdc_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "A reference to the Provider VDC from which this organization VDC is provisioned.",
			},
			"enable_fast_provisioning": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Request for fast provisioning. Request will be honored only if the underlying datas tore supports it. Fast provisioning can reduce the time it takes to create virtual machines by using vSphere linked clones. If you disable fast provisioning, all provisioning operations will result in full clones.",
			},
			//  Always null in the response to a GET request. On update, set to false to disallow the update if the AllocationModel is AllocationPool or ReservationPool
			//  and the ComputeCapacity you specified is greater than what the backing Provider VDC can supply. Defaults to true if empty or missing.
			"allow_over_commit": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Set to false to disallow creation of the VDC if the AllocationModel is AllocationPool or ReservationPool and the ComputeCapacity you specified is greater than what the backing Provider VDC can supply. Default is true.",
			},
			"enable_vm_discovery": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "True if discovery of vCenter VMs is enabled for resource pools backing this VDC. If left unspecified, the actual behaviour depends on enablement at the organization level and at the system level.",
			},
			"elasticity": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Set to true to indicate if the Flex VDC is to be elastic.",
			},
			"include_vm_memory_overhead": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Set to true to indicate if the Flex VDC is to include memory overhead into its accounting for admission control.",
			},
			"delete_force": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "When destroying use delete_force=True to remove a VDC and any objects it contains, regardless of their state.",
			},
			"delete_recursive": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "When destroying use delete_recursive=True to remove the VDC and any objects it contains that are in a state that normally allows removal.",
			},
			"metadata": {
				Type:          schema.TypeMap,
				Optional:      true,
				Computed:      true, // To be compatible with `metadata_entry`
				Description:   "Key and value pairs for Org VDC metadata",
				Deprecated:    "Use metadata_entry instead",
				ConflictsWith: []string{"metadata_entry"},
			},
			"metadata_entry": metadataEntryResourceSchema("VDC"),
			"vm_sizing_policy_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Description: "Set of VM Sizing Policy IDs",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"vm_placement_policy_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Description: "Set of VM Placement Policy IDs",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"default_vm_sizing_policy_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "ID of default VM Compute policy, which can be a VM Sizing Policy, VM Placement Policy or vGPU Policy",
				ConflictsWith: []string{"default_compute_policy_id"},
				Deprecated:    "Use `default_compute_policy_id` attribute instead, which can support VM Sizing Policies, VM Placement Policies and vGPU Policies",
			},
			"default_compute_policy_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "ID of default Compute policy for this VDC, which can be a VM Sizing Policy, VM Placement Policy or vGPU Policy",
				ConflictsWith: []string{"default_vm_sizing_policy_id"},
			},
			"edge_cluster_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of NSX-T Edge Cluster (provider vApp networking services and DHCP capability for Isolated networks)",
			},
			"enable_nsxv_distributed_firewall": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Set to true to enable distributed firewall - Only applies to NSX-V VDCs",
			},
		},
	}
}

// Creates a new VDC from a resource definition
func resourceVcdVdcCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	orgVdcName := d.Get("name").(string)
	log.Printf("[TRACE] VDC creation initiated: %s", orgVdcName)

	vcdClient := meta.(*VCDClient)

	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("functionality requires System administrator privileges")
	}

	// check that elasticity and include_vm_memory_overhead are used only for Flex
	_, elasticityConfigured := d.GetOkExists("elasticity")
	_, vmMemoryOverheadConfigured := d.GetOkExists("include_vm_memory_overhead")
	if d.Get("allocation_model").(string) != "Flex" && (elasticityConfigured || vmMemoryOverheadConfigured) {
		return diag.Errorf("`elasticity` and `include_vm_memory_overhead` can be used only with Flex allocation model (vCD 9.7+)")
	}

	// VDC creation is accessible only in administrator API part
	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	orgVdc, err := adminOrg.GetVDCByName(orgVdcName, false)
	if orgVdc != nil || err == nil {
		return diag.Errorf("org VDC with such name already exists: %s", orgVdcName)
	}

	params, err := getVcdVdcInput(d, vcdClient)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Creating VDC: %#v", params)

	vdc, err := adminOrg.CreateOrgVdc(params)
	if err != nil {
		log.Printf("[DEBUG] Error creating VDC: %s", err)
		return diag.Errorf("error creating VDC: %s", err)
	}

	d.SetId(vdc.Vdc.ID)
	log.Printf("[TRACE] VDC created: %#v", vdc)

	err = createOrUpdateOrgMetadata(d, meta)
	if err != nil {
		return diag.Errorf("error adding metadata to VDC: %s", err)
	}

	err = addAssignedComputePolicies(d, meta)
	if err != nil {
		return diag.Errorf("error assigning VM Compute Policies to VDC: %s", err)
	}

	// Edge Cluster ID uses different endpoint (VDC Network Profiles endpoint) and it shouldn't be
	// set on create if it is not present
	edgeClusterId := d.Get("edge_cluster_id").(string)
	if edgeClusterId != "" {
		err = setVdcEdgeCluster(d, vdc)
		if err != nil {
			return diag.Errorf("error setting Edge Cluster: %s", err)
		}
	}
	if d.Get("enable_nsxv_distributed_firewall").(bool) {
		if !vdc.IsNsxv() {
			return diag.Errorf("VDC '%s' is not a NSX-V VDC and the property 'enable_nsxv_distributed_firewall' can't be used", orgVdcName)
		}
		dfw := govcd.NewNsxvDistributedFirewall(&vcdClient.Client, vdc.Vdc.ID)
		err = dfw.Enable()
		if err != nil {
			return diag.Errorf("error enabling NSX-V distributed firewall for VDC '%s': %s", orgVdcName, err)
		}
		dSet(d, "enable_nsxv_distributed_firewall", true)
	}

	return resourceVcdVdcRead(ctx, d, meta)
}

func resourceVcdVdcRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vdcName := d.Get("name").(string)
	log.Printf("[TRACE] VDC read initiated: %s", vdcName)

	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	adminVdc, err := adminOrg.GetAdminVDCByName(vdcName, false)
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		log.Printf("[DEBUG] Unable to find VDC %s", vdcName)
		return diag.Errorf("unable to find VDC %s, err: %s", vdcName, err)
	}

	diagErr := setOrgVdcData(d, vcdClient, adminVdc)
	if diagErr != nil {
		return diagErr
	}

	err = setEdgeClusterData(d, adminVdc, "vdc_org_vdc")
	if err != nil {
		return diag.FromErr(err)
	}
	dSet(d, "enable_nsxv_distributed_firewall", false)
	if adminVdc.IsNsxv() {
		dfw := govcd.NewNsxvDistributedFirewall(&vcdClient.Client, adminVdc.AdminVdc.ID)
		enabled, err := dfw.IsEnabled()
		if err != nil {
			return diag.Errorf("error retrieving NSX-V distributed firewall state for VDC '%s': %s", vdcName, err)
		}
		dSet(d, "enable_nsxv_distributed_firewall", enabled)
	}
	return nil
}

// setOrgVdcData sets object state from *govcd.AdminVdc
func setOrgVdcData(d *schema.ResourceData, vcdClient *VCDClient, adminVdc *govcd.AdminVdc) diag.Diagnostics {

	dSet(d, "allocation_model", adminVdc.AdminVdc.AllocationModel)
	if adminVdc.AdminVdc.ResourceGuaranteedCpu != nil {
		dSet(d, "cpu_guaranteed", *adminVdc.AdminVdc.ResourceGuaranteedCpu)
	}
	if adminVdc.AdminVdc.VCpuInMhz != nil {
		dSet(d, "cpu_speed", int(*adminVdc.AdminVdc.VCpuInMhz))
	}
	dSet(d, "description", adminVdc.AdminVdc.Description)
	if adminVdc.AdminVdc.UsesFastProvisioning != nil {
		dSet(d, "enable_fast_provisioning", *adminVdc.AdminVdc.UsesFastProvisioning)
	}
	if adminVdc.AdminVdc.IsThinProvision != nil {
		dSet(d, "enable_thin_provisioning", *adminVdc.AdminVdc.IsThinProvision)
	}
	dSet(d, "enable_vm_discovery", adminVdc.AdminVdc.VmDiscoveryEnabled)
	dSet(d, "enabled", adminVdc.AdminVdc.IsEnabled)
	if adminVdc.AdminVdc.ResourceGuaranteedMemory != nil {
		dSet(d, "memory_guaranteed", *adminVdc.AdminVdc.ResourceGuaranteedMemory)
	}
	dSet(d, "name", adminVdc.AdminVdc.Name)

	if adminVdc.AdminVdc.NetworkPoolReference != nil {
		networkPool, err := govcd.GetNetworkPoolByHREF(vcdClient.VCDClient, adminVdc.AdminVdc.NetworkPoolReference.HREF)
		if err != nil {
			return diag.Errorf("error retrieving network pool: %s", err)
		}
		dSet(d, "network_pool_name", networkPool.Name)
	}

	dSet(d, "network_quota", adminVdc.AdminVdc.NetworkQuota)
	dSet(d, "nic_quota", adminVdc.AdminVdc.Vdc.NicQuota)
	if adminVdc.AdminVdc.ProviderVdcReference != nil {
		dSet(d, "provider_vdc_name", adminVdc.AdminVdc.ProviderVdcReference.Name)
	}
	dSet(d, "vm_quota", adminVdc.AdminVdc.Vdc.VMQuota)

	if err := d.Set("compute_capacity", getComputeCapacities(adminVdc.AdminVdc.ComputeCapacity)); err != nil {
		return diag.Errorf("error setting compute_capacity: %s", err)
	}

	if adminVdc.AdminVdc.VdcStorageProfiles != nil {

		storageProfileStateData, err := getComputeStorageProfiles(vcdClient, adminVdc.AdminVdc.VdcStorageProfiles)
		if err != nil {
			return diag.Errorf("error preparing storage profile data: %s", err)
		}

		if err := d.Set("storage_profile", storageProfileStateData); err != nil {
			return diag.Errorf("error setting compute_capacity: %s", err)
		}
	}

	if adminVdc.AdminVdc.IsElastic != nil {
		dSet(d, "elasticity", *adminVdc.AdminVdc.IsElastic)
	}

	if adminVdc.AdminVdc.IncludeMemoryOverhead != nil {
		dSet(d, "include_vm_memory_overhead", *adminVdc.AdminVdc.IncludeMemoryOverhead)
	}

	dSet(d, "default_vm_sizing_policy_id", adminVdc.AdminVdc.DefaultComputePolicy.ID) // Deprecated, populating for compatibility
	dSet(d, "default_compute_policy_id", adminVdc.AdminVdc.DefaultComputePolicy.ID)

	assignedVmComputePolicies, err := adminVdc.GetAllAssignedVdcComputePoliciesV2(url.Values{
		"filter": []string{fmt.Sprintf("%spolicyType==VdcVmPolicy", getVgpuFilterToPrepend(vcdClient, false))}, // Filtering out vGPU Policies as there's no attribute support yet.
	})
	if err != nil {
		log.Printf("[DEBUG] Unable to get assigned VM Compute policies")
		return diag.Errorf("unable to get assigned VM Compute policies %s", err)
	}
	var sizingPolicyIds []string
	var placementPolicyIds []string
	for _, policy := range assignedVmComputePolicies {
		if policy.VdcComputePolicyV2.IsSizingOnly {
			sizingPolicyIds = append(sizingPolicyIds, policy.VdcComputePolicyV2.ID)
		} else {
			placementPolicyIds = append(placementPolicyIds, policy.VdcComputePolicyV2.ID)
		}
	}

	vmSizingPoliciesSet := convertStringsToTypeSet(sizingPolicyIds)
	vmPlacementPoliciesSet := convertStringsToTypeSet(placementPolicyIds)

	err = d.Set("vm_sizing_policy_ids", vmSizingPoliciesSet)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("vm_placement_policy_ids", vmPlacementPoliciesSet)
	if err != nil {
		return diag.FromErr(err)
	}

	diagErr := updateMetadataInState(d, vcdClient, "vcd_org_vdc", adminVdc)
	if diagErr != nil {
		log.Printf("[DEBUG] Unable to set VDC metadata")
		return diagErr
	}

	log.Printf("[TRACE] vdc read completed: %#v", adminVdc.AdminVdc)
	return nil
}

// getComputeStorageProfiles constructs specific struct to be saved in Terraform state file.
// Expected E.g.
func getComputeStorageProfiles(vcdClient *VCDClient, profile *types.VdcStorageProfiles) ([]map[string]interface{}, error) {
	root := make([]map[string]interface{}, 0)

	for _, vdcStorageProfile := range profile.VdcStorageProfile {
		vdcStorageProfileDetails, err := vcdClient.GetStorageProfileByHref(vdcStorageProfile.HREF)
		if err != nil {
			return nil, err
		}
		storageProfileData := make(map[string]interface{})
		storageProfileData["limit"] = vdcStorageProfileDetails.Limit
		storageProfileData["default"] = vdcStorageProfileDetails.Default
		storageProfileData["enabled"] = vdcStorageProfileDetails.Enabled
		storageProfileData["name"] = vdcStorageProfileDetails.Name

		storageProfileData["storage_used_in_mb"] = vdcStorageProfileDetails.StorageUsedMB
		root = append(root, storageProfileData)
	}

	return root, nil
}

// getComputeCapacities constructs specific struct to be saved in Terraform state file.
// Expected E.g. &[]map[string]interface {}
// {map[string]interface {}{"cpu":(*[]map[string]interface {})
// ({"allocated":8000, "limit":8000, "overhead":0, "reserved":4000, "used":0}),
// "memory":(*[]map[string]interface {})
// ({"allocated":7168, "limit":7168, "overhead":0, "reserved":3584, "used":0})},
func getComputeCapacities(capacities []*types.ComputeCapacity) *[]map[string]interface{} {
	rootInternal := map[string]interface{}{}
	var root []map[string]interface{}

	for _, capacity := range capacities {
		cpuValueMap := map[string]interface{}{}
		cpuValueMap["limit"] = int(capacity.CPU.Limit)
		cpuValueMap["allocated"] = int(capacity.CPU.Allocated)
		cpuValueMap["reserved"] = int(capacity.CPU.Reserved)
		cpuValueMap["used"] = int(capacity.CPU.Used)

		memoryValueMap := map[string]interface{}{}
		memoryValueMap["limit"] = int(capacity.Memory.Limit)
		memoryValueMap["allocated"] = int(capacity.Memory.Allocated)
		memoryValueMap["reserved"] = int(capacity.Memory.Reserved)
		memoryValueMap["used"] = int(capacity.Memory.Used)

		var memoryCapacityArray []map[string]interface{}
		memoryCapacityArray = append(memoryCapacityArray, memoryValueMap)
		var cpuCapacityArray []map[string]interface{}
		cpuCapacityArray = append(cpuCapacityArray, cpuValueMap)

		rootInternal["cpu"] = &cpuCapacityArray
		rootInternal["memory"] = &memoryCapacityArray

		root = append(root, rootInternal)
	}

	return &root
}

// Converts to terraform understandable structure
func getMetadataStruct(metadata []*types.MetadataEntry) StringMap {
	metadataMap := make(StringMap, len(metadata))
	for _, metadataEntry := range metadata {
		metadataMap[metadataEntry.Key] = metadataEntry.TypedValue.Value
	}
	return metadataMap
}

// resourceVcdVdcUpdate function updates resource with found configurations changes
func resourceVcdVdcUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vdcName := d.Get("name").(string)
	log.Printf("[TRACE] VDC update initiated: %s", vdcName)

	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	if d.HasChange("name") {
		oldValue, _ := d.GetChange("name")
		vdcName = oldValue.(string)
	}

	adminVdc, err := adminOrg.GetAdminVDCByName(vdcName, false)
	if err != nil {
		log.Printf("[DEBUG] Unable to find VDC %s", vdcName)
		return diag.Errorf("unable to find VDC %s, error:  %s", vdcName, err)
	}

	changedAdminVdc, err := getUpdatedVdcInput(d, vcdClient, adminVdc)
	if err != nil {
		log.Printf("[DEBUG] Error updating VDC %s with error %s", vdcName, err)
		return diag.Errorf("error updating VDC %s, err: %s", vdcName, err)
	}

	updatedAdminVdc, err := changedAdminVdc.Update()
	if err != nil {
		log.Printf("[DEBUG] Error updating VDC %s with error %s", vdcName, err)
		return diag.Errorf("error updating VDC %s, err: %s", vdcName, err)
	}

	err = createOrUpdateOrgMetadata(d, meta)
	if err != nil {
		return diag.Errorf("error updating VDC metadata: %s", err)
	}

	err = updateAssignedVmComputePolicies(d, meta, changedAdminVdc)
	if err != nil {
		return diag.Errorf("error assigning VM sizing policies to VDC: %s", err)
	}

	if d.HasChange("storage_profile") {
		vdcStorageProfilesConfigurations := d.Get("storage_profile").(*schema.Set)
		err = updateStorageProfiles(vdcStorageProfilesConfigurations, vcdClient, adminVdc, d.Get("provider_vdc_name").(string))
		if err != nil {
			return diag.Errorf("[VDC update] error updating storage profiles: %s", err)
		}
	}

	if d.HasChange("edge_cluster_id") {
		orgVdc, err := adminOrg.GetVDCByName(updatedAdminVdc.AdminVdc.Name, false)
		if orgVdc == nil || err != nil {
			return diag.Errorf("error retrieving Org VDC from Admin VDC '%s': %s", updatedAdminVdc.AdminVdc.Name, err)
		}
		err = setVdcEdgeCluster(d, orgVdc)
		if err != nil {
			return diag.Errorf("error updating 'edge_cluster_id': %s", err)
		}

	}

	if adminVdc.IsNsxv() && d.HasChange("enable_nsxv_distributed_firewall") {
		dfw := govcd.NewNsxvDistributedFirewall(&vcdClient.Client, adminVdc.AdminVdc.ID)
		enablementState := d.Get("enable_nsxv_distributed_firewall").(bool)
		if enablementState {
			err = dfw.Enable()
		} else {
			err = dfw.Disable()
		}
		if err != nil {
			return diag.Errorf("error setting NSX-V distributed firewall state for VDC '%s': %s", vdcName, err)
		}
		dSet(d, "enable_nsxv_distributed_firewall", enablementState)
	}

	log.Printf("[TRACE] VDC update completed: %s", adminVdc.AdminVdc.Name)
	return resourceVcdVdcRead(ctx, d, meta)
}

func updateStorageProfileDetails(vcdClient *VCDClient, adminVdc *govcd.AdminVdc, storageProfile *types.Reference, storageConfiguration map[string]interface{}) error {
	util.Logger.Printf("updating storage profile %#v", storageProfile)
	uuid, err := govcd.GetUuidFromHref(storageProfile.HREF, true)
	if err != nil {
		return fmt.Errorf("error parsing VDC storage profile ID : %s", err)
	}
	vdcStorageProfileDetails, err := vcdClient.GetStorageProfileByHref(storageProfile.HREF)
	if err != nil {
		return fmt.Errorf("error getting VDC storage profile: %s", err)
	}
	_, err = adminVdc.UpdateStorageProfile(uuid, &types.AdminVdcStorageProfile{
		Name:         storageConfiguration["name"].(string),
		IopsSettings: nil,
		Units:        "MB", // only this value is supported
		Limit:        int64(storageConfiguration["limit"].(int)),
		Default:      storageConfiguration["default"].(bool),
		Enabled:      addrOf(storageConfiguration["enabled"].(bool)),
		ProviderVdcStorageProfile: &types.Reference{
			HREF: vdcStorageProfileDetails.ProviderVdcStorageProfile.HREF,
		},
	})
	if err != nil {
		return fmt.Errorf("error updating VDC storage profile '%s': %s", storageConfiguration["name"].(string), err)
	}
	return nil
}

func updateStorageProfiles(set *schema.Set, client *VCDClient, adminVdc *govcd.AdminVdc, providerVdcName string) error {

	type storageProfileCombo struct {
		configuration map[string]interface{}
		reference     *types.Reference
	}
	var (
		existingStorageProfiles []storageProfileCombo
		newStorageProfiles      []storageProfileCombo
		removeStorageProfiles   []storageProfileCombo
		defaultSp               = make(map[string]bool)
	)

	var isDefaultStorageProfileNew = false
	// 1. find existing storage profiles: SP are both in the definition and in the VDC
	for _, storageConfigurationValues := range set.List() {
		storageConfiguration := storageConfigurationValues.(map[string]interface{})
		for _, vdcStorageProfile := range adminVdc.AdminVdc.VdcStorageProfiles.VdcStorageProfile {
			if storageConfiguration["name"].(string) == vdcStorageProfile.Name {
				if storageConfiguration["default"].(bool) {
					defaultSp[storageConfiguration["name"].(string)] = true
				}
				existingStorageProfiles = append(existingStorageProfiles, storageProfileCombo{
					configuration: storageConfiguration,
					reference:     vdcStorageProfile,
				})
			}
		}
	}

	// 2. find new storage profiles: SP are in the definition, but not in the VDC
	for _, storageConfigurationValues := range set.List() {
		storageConfiguration := storageConfigurationValues.(map[string]interface{})
		found := false
		for _, vdcStorageProfile := range adminVdc.AdminVdc.VdcStorageProfiles.VdcStorageProfile {
			if storageConfiguration["name"].(string) == vdcStorageProfile.Name {
				found = true
			}
		}
		if !found {
			if storageConfiguration["default"].(bool) {
				defaultSp[storageConfiguration["name"].(string)] = true
				isDefaultStorageProfileNew = true
			}
			newStorageProfiles = append(newStorageProfiles, storageProfileCombo{
				configuration: storageConfiguration,
				reference:     nil,
			})
		}
	}

	// 3 find removed storage profiles: SP are in the VDC, but not in the definition
	for _, vdcStorageProfile := range adminVdc.AdminVdc.VdcStorageProfiles.VdcStorageProfile {
		found := false
		for _, storageConfigurationValues := range set.List() {
			storageConfiguration := storageConfigurationValues.(map[string]interface{})
			if storageConfiguration["name"].(string) == vdcStorageProfile.Name {
				found = true
			}
		}
		if !found {
			_, isDefault := defaultSp[vdcStorageProfile.Name]
			if isDefault {
				delete(defaultSp, vdcStorageProfile.Name)
			}
			removeStorageProfiles = append(removeStorageProfiles, storageProfileCombo{
				configuration: nil,
				reference:     vdcStorageProfile,
			})
		}
	}

	// 4. Check that there is one and only one default element
	if len(defaultSp) == 0 {
		return fmt.Errorf("updateStorageProfiles] no default storage profile left after update")
	}
	if len(defaultSp) > 1 {
		defaultItems := ""
		for d := range defaultSp {
			defaultItems += " " + d
		}
		return fmt.Errorf("updateStorageProfiles] more than one default storage profile defined [%s]", defaultItems)
	}

	// 5. Set the default storage profile early
	if !isDefaultStorageProfileNew {
		defaultSpName := ""
		for name := range defaultSp {
			defaultSpName = name
			break
		}
		err := adminVdc.SetDefaultStorageProfile(defaultSpName)
		if err != nil {
			return fmt.Errorf("[updateStorageProfiles] error setting default storage profile '%s': %s", defaultSpName, err)
		}
	}

	// 6. Add new storage profiles
	for _, spCombo := range newStorageProfiles {
		storageProfile, err := client.QueryProviderVdcStorageProfileByName(spCombo.configuration["name"].(string), adminVdc.AdminVdc.ProviderVdcReference.HREF)
		if err != nil {
			return fmt.Errorf("[updateStorageProfiles] error retrieving storage profile '%s' from provider VDC '%s': %s", spCombo.configuration["name"].(string), providerVdcName, err)
		}
		err = adminVdc.AddStorageProfileWait(&types.VdcStorageProfileConfiguration{
			Enabled: addrOf(spCombo.configuration["enabled"].(bool)),
			Units:   "MB",
			Limit:   int64(spCombo.configuration["limit"].(int)),
			Default: spCombo.configuration["default"].(bool),
			ProviderVdcStorageProfile: &types.Reference{
				HREF: storageProfile.HREF,
				Name: storageProfile.Name,
			},
		}, "")
		if err != nil {
			return fmt.Errorf("[updateStorageProfiles] error adding new storage profile: %s", err)
		}
	}

	// 7. Update existing storage profiles
	for _, spCombo := range existingStorageProfiles {
		err := updateStorageProfileDetails(client, adminVdc, spCombo.reference, spCombo.configuration)
		if err != nil {
			return fmt.Errorf("[updateStorageProfiles] error updating storage profile '%s': %s", spCombo.reference.Name, err)
		}
	}

	// 8. Delete unwanted storage profiles
	for _, spCombo := range removeStorageProfiles {
		err := adminVdc.RemoveStorageProfileWait(spCombo.reference.Name)
		if err != nil {
			return fmt.Errorf("[updateStorageProfiles] error removing storage profile %s: %s", spCombo.reference.Name, err)
		}
	}

	return nil
}

// Deletes a VDC, optionally removing all objects in it as well
func resourceVcdVdcDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vdcName := d.Get("name").(string)
	log.Printf("[TRACE] VDC delete started: %s", vdcName)

	vcdClient := meta.(*VCDClient)

	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("functionality requires System administrator privileges")
	}

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	vdc, err := adminOrg.GetVDCByName(vdcName, false)
	if err != nil {
		log.Printf("[DEBUG] Unable to find VDC %s. Removing from tfstate", vdcName)
		d.SetId("")
		return nil
	}

	err = vdc.DeleteWait(d.Get("delete_force").(bool), d.Get("delete_recursive").(bool))
	if err != nil {
		log.Printf("[DEBUG] Error removing VDC %s, err: %s", vdcName, err)
		return diag.Errorf("error removing VDC %s, err: %s", vdcName, err)
	}

	_, err = adminOrg.GetVDCByName(vdcName, true)
	if err == nil {
		return diag.Errorf("vdc %s still found after deletion", vdcName)
	}
	log.Printf("[TRACE] VDC delete completed: %s", vdcName)
	return nil
}

// updateAssignedVmComputePolicies handles VM compute policies.
func updateAssignedVmComputePolicies(d *schema.ResourceData, meta interface{}, vdc *govcd.AdminVdc) error {
	vcdClient := meta.(*VCDClient)
	defaultPolicyId, vcdComputePolicyHref, err := getDefaultPolicyIdAndComputePolicyHref(d, vcdClient)
	if err != nil {
		return err
	}
	if vcdComputePolicyHref == nil || defaultPolicyId == "" {
		return nil
	}

	arePoliciesChanged := d.HasChange("vm_sizing_policy_ids") || d.HasChange("vm_placement_policy_ids")
	isDefaultPolicyChanged := d.HasChange("default_compute_policy_id") || d.HasChange("default_vm_sizing_policy_id")

	// Compatibility patch: Remove deprecated `default_vm_sizing_policy_id` from this conditional when the attribute is removed.
	// We can do this as `default_compute_policy_id` contains the same value as `default_vm_sizing_policy_id`.
	if isDefaultPolicyChanged && !arePoliciesChanged {
		vdc.AdminVdc.DefaultComputePolicy = &types.Reference{HREF: vcdComputePolicyHref.String() + defaultPolicyId, ID: defaultPolicyId}
		_, err := vdc.Update()
		if err != nil {
			return fmt.Errorf("error setting default VM Compute Policy. %s", err)
		}
		return nil
	}

	if !isDefaultPolicyChanged && arePoliciesChanged {
		var vmComputePolicyIds []string
		computePolicyAttributes := []string{"vm_sizing_policy_ids", "vm_placement_policy_ids"}
		for _, attribute := range computePolicyAttributes {
			vmComputePolicyIds = append(vmComputePolicyIds, convertSchemaSetToSliceOfStrings(d.Get(attribute).(*schema.Set))...)
		}
		if !contains(vmComputePolicyIds, defaultPolicyId) {
			return fmt.Errorf("`default_compute_policy_id` %s is not present in any of `%v`", defaultPolicyId, computePolicyAttributes)
		}

		var vdcComputePolicyReferenceList []*types.Reference
		for _, policyId := range vmComputePolicyIds {
			vdcComputePolicyReferenceList = append(vdcComputePolicyReferenceList, &types.Reference{HREF: vcdComputePolicyHref.String() + policyId})
		}

		policyReferences := types.VdcComputePolicyReferences{}
		policyReferences.VdcComputePolicyReference = vdcComputePolicyReferenceList
		_, err = vdc.SetAssignedComputePolicies(policyReferences)
		if err != nil {
			return fmt.Errorf("error setting VM Compute Policies. %s", err)
		}
		return nil
	}

	err = changeComputePoliciesAndDefaultId(d, vcdClient, vcdComputePolicyHref.String(), vdc)
	if err != nil {
		return err
	}
	return nil
}

// changeComputePoliciesAndDefaultId handles Compute policies. Created VDC generates default Compute policy which requires additional handling.
// Assigning and setting default Compute policies requires different API calls. Default policy can't be removed, as result
// we approach this with adding new policies, set new default, remove all old policies.
func changeComputePoliciesAndDefaultId(d *schema.ResourceData, vcdClient *VCDClient, vcdComputePolicyHref string, vdc *govcd.AdminVdc) error {
	arePoliciesChanged := d.HasChange("vm_sizing_policy_ids") || d.HasChange("vm_placement_policy_ids")
	isDefaultPolicyChanged := d.HasChange("default_compute_policy_id") || d.HasChange("default_vm_sizing_policy_id")
	if !arePoliciesChanged && !isDefaultPolicyChanged {
		return nil
	}

	// Deprecation compatibility: If `default_compute_policy_id` is not set, fallback to deprecated one.
	// We can do this as `default_compute_policy_id` contains the same value as `default_vm_sizing_policy_id`.
	defaultPolicyId, defaultPolicyIsSet := d.GetOk("default_compute_policy_id")
	if !defaultPolicyIsSet {
		defaultPolicyId, _ = d.GetOk("default_vm_sizing_policy_id")
	}

	var vmComputePolicyIds []string
	computePolicyAttributes := []string{"vm_sizing_policy_ids", "vm_placement_policy_ids"}
	for _, attribute := range computePolicyAttributes {
		vmComputePolicyIds = append(vmComputePolicyIds, convertSchemaSetToSliceOfStrings(d.Get(attribute).(*schema.Set))...)
	}
	if !contains(vmComputePolicyIds, defaultPolicyId.(string)) {
		return fmt.Errorf("`default_compute_policy_id` %s is not present in any of `%v`", defaultPolicyId.(string), computePolicyAttributes)
	}

	var vdcComputePolicyReferenceList []*types.Reference
	for _, policyId := range vmComputePolicyIds {
		vdcComputePolicyReferenceList = append(vdcComputePolicyReferenceList, &types.Reference{HREF: vcdComputePolicyHref + policyId})
	}

	existingPolicies, err := vdc.GetAllAssignedVdcComputePoliciesV2(url.Values{
		"filter": []string{fmt.Sprintf("%spolicyType==VdcVmPolicy", getVgpuFilterToPrepend(vcdClient, false))}, // Filtering out vGPU Policies as there's no attribute support yet.
	})
	if err != nil {
		return fmt.Errorf("error getting Compute Policies. %s", err)
	}
	for _, existingPolicy := range existingPolicies {
		vdcComputePolicyReferenceList = append(vdcComputePolicyReferenceList, &types.Reference{HREF: vcdComputePolicyHref + existingPolicy.VdcComputePolicyV2.ID})
	}

	policyReferences := types.VdcComputePolicyReferences{}
	policyReferences.VdcComputePolicyReference = vdcComputePolicyReferenceList
	_, err = vdc.SetAssignedComputePolicies(policyReferences)
	if err != nil {
		return fmt.Errorf("error setting Compute Policies. %s", err)
	}

	// set default Compute Policy
	vdc.AdminVdc.DefaultComputePolicy = &types.Reference{HREF: vcdComputePolicyHref + defaultPolicyId.(string), ID: defaultPolicyId.(string)}
	updatedVdc, err := vdc.Update()
	if err != nil {
		return fmt.Errorf("error setting default Compute Policy. %s", err)
	}

	// Now we can remove previously existing policies as default policy changed
	vdcComputePolicyReferenceList = []*types.Reference{}
	for _, policyId := range vmComputePolicyIds {
		vdcComputePolicyReferenceList = append(vdcComputePolicyReferenceList, &types.Reference{HREF: vcdComputePolicyHref + policyId})
	}
	policyReferences.VdcComputePolicyReference = vdcComputePolicyReferenceList

	_, err = updatedVdc.SetAssignedComputePolicies(policyReferences)
	if err != nil {
		return fmt.Errorf("error setting Compute Policies. %s", err)
	}
	return nil
}

// getDefaultPolicyIdAndComputePolicyHref gets the default compute policy ID and returns the Compute Policy API endpoint.
func getDefaultPolicyIdAndComputePolicyHref(d *schema.ResourceData, vcdClient *VCDClient) (string, *url.URL, error) {
	// Deprecation compatibility: If `default_compute_policy_id` is not set, fallback to deprecated one.
	// We can do this as `default_compute_policy_id` contains the same value as `default_vm_sizing_policy_id`.
	defaultPolicyId, defaultPolicyIsSet := d.GetOk("default_compute_policy_id")
	if !defaultPolicyIsSet {
		defaultPolicyId, defaultPolicyIsSet = d.GetOk("default_vm_sizing_policy_id")
	}

	_, sizingOk := d.GetOk("vm_sizing_policy_ids")
	_, placementOk := d.GetOk("vm_placement_policy_ids")

	if defaultPolicyIsSet && !sizingOk && !placementOk {
		return "", nil, fmt.Errorf("when `default_compute_policy_id` is used, it requires also `vm_sizing_policy_ids` or `vm_placement_policy_ids`")
	}

	arePoliciesChanged := d.HasChange("vm_sizing_policy_ids") || d.HasChange("vm_placement_policy_ids")
	isDefaultPolicyChanged := d.HasChange("default_compute_policy_id") || d.HasChange("default_vm_sizing_policy_id")

	// Early return
	if !isDefaultPolicyChanged && !arePoliciesChanged {
		return "", nil, nil
	}

	vcdComputePolicyHref, err := vcdClient.Client.OpenApiBuildEndpoint(types.OpenApiPathVersion2_0_0, types.OpenApiEndpointVdcComputePolicies)
	if err != nil {
		return "", nil, fmt.Errorf("error constructing HREF for Compute Policy")
	}
	return defaultPolicyId.(string), vcdComputePolicyHref, nil
}

// addAssignedComputePolicies handles Compute policies. Created VDC generates default Compute policy which requires additional handling.
// Assigning and setting default Compute policies requires different API calls. Default approach is add new policies, set new default, remove all policies.
func addAssignedComputePolicies(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	_, vcdComputePolicyHref, err := getDefaultPolicyIdAndComputePolicyHref(d, vcdClient)
	if err != nil {
		return err
	}
	if vcdComputePolicyHref == nil {
		return nil
	}

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrg, err)
	}

	vdc, err := adminOrg.GetAdminVDCByName(d.Get("name").(string), false)
	if err != nil {
		return fmt.Errorf(errorRetrievingVdcFromOrg, d.Get("org").(string), d.Get("name").(string), err)
	}

	err = changeComputePoliciesAndDefaultId(d, vcdClient, vcdComputePolicyHref.String(), vdc)
	if err != nil {
		return err
	}
	return nil
}

func createOrUpdateOrgMetadata(d *schema.ResourceData, meta interface{}) error {

	log.Printf("[TRACE] adding/updating metadata to VDC")

	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrg, err)
	}

	adminVdc, err := adminOrg.GetAdminVDCByName(d.Get("name").(string), false)
	if err != nil {
		return fmt.Errorf(errorRetrievingVdcFromOrg, d.Get("org").(string), d.Get("name").(string), err)
	}

	return createOrUpdateMetadata(d, adminVdc, "metadata")
}

// helper for transforming the compute capacity section of the resource input into the VdcConfiguration structure
func capacityWithUsage(d map[string]interface{}, units string) *types.CapacityWithUsage {
	capacity := &types.CapacityWithUsage{
		Units: units,
	}

	if allocated, ok := d["allocated"]; ok {
		capacity.Allocated = int64(allocated.(int))
	}

	if limit, ok := d["limit"]; ok {
		capacity.Limit = int64(limit.(int))
	}

	return capacity
}

// helper for transforming the resource input into the AdminVdc structure
func getUpdatedVdcInput(d *schema.ResourceData, vcdClient *VCDClient, vdc *govcd.AdminVdc) (*govcd.AdminVdc, error) {

	if d.HasChange("compute_capacity") {
		computeCapacityList := d.Get("compute_capacity").([]interface{})
		if len(computeCapacityList) == 0 {
			return &govcd.AdminVdc{}, errors.New("no compute_capacity field")
		}
		computeCapacity := computeCapacityList[0].(map[string]interface{})

		cpuCapacityList := computeCapacity["cpu"].([]interface{})
		if len(cpuCapacityList) == 0 {
			return &govcd.AdminVdc{}, errors.New("no cpu field in compute_capacity")
		}
		memoryCapacityList := computeCapacity["memory"].([]interface{})
		if len(memoryCapacityList) == 0 {
			return &govcd.AdminVdc{}, errors.New("no memory field in compute_capacity")
		}
		vdc.AdminVdc.ComputeCapacity[0].Memory = capacityWithUsage(memoryCapacityList[0].(map[string]interface{}), "MB")
		vdc.AdminVdc.ComputeCapacity[0].CPU = capacityWithUsage(cpuCapacityList[0].(map[string]interface{}), "MHz")
	}

	if d.HasChange("allocation_model") {
		vdc.AdminVdc.AllocationModel = d.Get("allocation_model").(string)
	}

	if d.HasChange("name") {
		vdc.AdminVdc.Name = d.Get("name").(string)
	}

	if d.HasChange("description") {
		vdc.AdminVdc.Description = d.Get("description").(string)
	}

	if d.HasChange("nic_quota") {
		vdc.AdminVdc.NicQuota = d.Get("nic_quota").(int)
	}

	if d.HasChange("network_quota") {
		vdc.AdminVdc.NetworkQuota = d.Get("network_quota").(int)
	}

	if d.HasChange("vm_quota") {
		vdc.AdminVdc.VMQuota = d.Get("vm_quota").(int)
	}

	if d.HasChange("enabled") {
		vdc.AdminVdc.IsEnabled = d.Get("enabled").(bool)
	}

	if d.HasChange("memory_guaranteed") {
		// only set 0 if value configured
		if value, ok := d.GetOkExists("memory_guaranteed"); ok {
			floatValue := value.(float64)
			vdc.AdminVdc.ResourceGuaranteedMemory = &floatValue
		}
	}

	if d.HasChange("cpu_guaranteed") {
		// only set 0 if value configured
		if value, ok := d.GetOkExists("cpu_guaranteed"); ok {
			floatValue := value.(float64)
			vdc.AdminVdc.ResourceGuaranteedCpu = &floatValue
		}
	}

	if d.HasChange("cpu_speed") {
		cpuSpeed := int64(d.Get("cpu_speed").(int))
		vdc.AdminVdc.VCpuInMhz = &cpuSpeed
	}

	if d.HasChange("enable_thin_provisioning") {
		thinProvisioned := d.Get("enable_thin_provisioning").(bool)
		vdc.AdminVdc.IsThinProvision = &thinProvisioned
	}

	if d.HasChange("network_pool_name") {
		networkPoolResults, err := govcd.QueryNetworkPoolByName(vcdClient.VCDClient, d.Get("network_pool_name").(string))
		if err != nil {
			return &govcd.AdminVdc{}, err
		}

		if len(networkPoolResults) == 0 {
			return &govcd.AdminVdc{}, fmt.Errorf("no network pool found with name %s", d.Get("network_pool_name"))
		}
		vdc.AdminVdc.NetworkPoolReference = &types.Reference{
			HREF: networkPoolResults[0].HREF,
		}
	}

	if d.HasChange("enable_fast_provisioning") {
		fastProvisioned := d.Get("enable_fast_provisioning").(bool)
		vdc.AdminVdc.UsesFastProvisioning = &fastProvisioned
	}

	if d.HasChange("allow_over_commit") {
		vdc.AdminVdc.OverCommitAllowed = d.Get("allow_over_commit").(bool)
	}

	if d.HasChange("enable_vm_discovery") {
		vdc.AdminVdc.VmDiscoveryEnabled = d.Get("enable_vm_discovery").(bool)
	}

	if d.HasChange("elasticity") {
		vdc.AdminVdc.IsElastic = addrOf(d.Get("elasticity").(bool))
	}

	if d.HasChange("include_vm_memory_overhead") {
		vdc.AdminVdc.IncludeMemoryOverhead = addrOf(d.Get("include_vm_memory_overhead").(bool))
	}

	//cleanup
	vdc.AdminVdc.Tasks = nil

	return vdc, nil
}

// helper for transforming the resource input into the VdcConfiguration structure
// any cast operations or default values should be done here so that the create method is simple
func getVcdVdcInput(d *schema.ResourceData, vcdClient *VCDClient) (*types.VdcConfiguration, error) {
	computeCapacityList := d.Get("compute_capacity").([]interface{})
	if len(computeCapacityList) == 0 {
		return &types.VdcConfiguration{}, errors.New("no compute_capacity field")
	}
	computeCapacity := computeCapacityList[0].(map[string]interface{})

	vdcStorageProfilesConfigurations := d.Get("storage_profile").(*schema.Set)
	if len(vdcStorageProfilesConfigurations.List()) == 0 {
		return &types.VdcConfiguration{}, errors.New("no storage_profile field")
	}

	cpuCapacityList := computeCapacity["cpu"].([]interface{})
	if len(cpuCapacityList) == 0 {
		return &types.VdcConfiguration{}, errors.New("no cpu field in compute_capacity")
	}
	memoryCapacityList := computeCapacity["memory"].([]interface{})
	if len(memoryCapacityList) == 0 {
		return &types.VdcConfiguration{}, errors.New("no memory field in compute_capacity")
	}

	providerVdcName := d.Get("provider_vdc_name").(string)
	providerVdcResults, err := govcd.QueryProviderVdcByName(vcdClient.VCDClient, providerVdcName)
	if err != nil {
		return &types.VdcConfiguration{}, err
	}
	if len(providerVdcResults) == 0 {
		return &types.VdcConfiguration{}, fmt.Errorf("no provider VDC found with name %s", providerVdcName)
	}

	params := &types.VdcConfiguration{
		Name:            d.Get("name").(string),
		Xmlns:           "http://www.vmware.com/vcloud/v1.5",
		AllocationModel: d.Get("allocation_model").(string),
		ComputeCapacity: []*types.ComputeCapacity{
			{
				CPU:    capacityWithUsage(cpuCapacityList[0].(map[string]interface{}), "MHz"),
				Memory: capacityWithUsage(memoryCapacityList[0].(map[string]interface{}), "MB"),
			},
		},
		ProviderVdcReference: &types.Reference{
			HREF: providerVdcResults[0].HREF,
		},
	}

	var vdcStorageProfiles []*types.VdcStorageProfileConfiguration
	for _, storageConfigurationValues := range vdcStorageProfilesConfigurations.List() {
		storageConfiguration := storageConfigurationValues.(map[string]interface{})

		sp, err := vcdClient.QueryProviderVdcStorageProfileByName(storageConfiguration["name"].(string), providerVdcResults[0].HREF)
		if err != nil {
			return &types.VdcConfiguration{}, fmt.Errorf("[getVcdVdcInput] error retrieving storage profile '%s' from provider VDC '%s': %s", storageConfiguration["name"].(string), providerVdcResults[0].Name, err)
		}

		vdcStorageProfile := &types.VdcStorageProfileConfiguration{
			Units:   "MB", // only this value is supported
			Limit:   int64(storageConfiguration["limit"].(int)),
			Default: storageConfiguration["default"].(bool),
			Enabled: addrOf(storageConfiguration["enabled"].(bool)),
			ProviderVdcStorageProfile: &types.Reference{
				HREF: sp.HREF,
			},
		}
		vdcStorageProfiles = append(vdcStorageProfiles, vdcStorageProfile)
	}

	params.VdcStorageProfile = vdcStorageProfiles

	if description, ok := d.GetOk("description"); ok {
		params.Description = description.(string)
	}

	if nicQuota, ok := d.GetOk("nic_quota"); ok {
		params.NicQuota = nicQuota.(int)
	}

	if networkQuota, ok := d.GetOk("network_quota"); ok {
		params.NetworkQuota = networkQuota.(int)
	}

	if vmQuota, ok := d.GetOk("vm_quota"); ok {
		params.VmQuota = vmQuota.(int)
	}

	if isEnabled, ok := d.GetOk("enabled"); ok {
		params.IsEnabled = isEnabled.(bool)
	}

	// only set 0 if value configured
	if resourceGuaranteedMemory, ok := d.GetOkExists("memory_guaranteed"); ok {
		value := resourceGuaranteedMemory.(float64)
		params.ResourceGuaranteedMemory = &value
	}

	if resourceGuaranteedCpu, ok := d.GetOkExists("cpu_guaranteed"); ok {
		value := resourceGuaranteedCpu.(float64)
		params.ResourceGuaranteedCpu = &value
	}

	if vCpuInMhz, ok := d.GetOk("cpu_speed"); ok {
		params.VCpuInMhz = int64(vCpuInMhz.(int))
	}

	if enableThinProvision, ok := d.GetOk("enable_thin_provisioning"); ok {
		params.IsThinProvision = enableThinProvision.(bool)
	}

	if networkPoolName, ok := d.GetOk("network_pool_name"); ok {
		networkPoolResults, err := govcd.QueryNetworkPoolByName(vcdClient.VCDClient, networkPoolName.(string))
		if err != nil {
			return &types.VdcConfiguration{}, err
		}

		if len(networkPoolResults) == 0 {
			return &types.VdcConfiguration{}, fmt.Errorf("no network pool found with name %s", networkPoolName)
		}
		params.NetworkPoolReference = &types.Reference{
			HREF: networkPoolResults[0].HREF,
		}
	}

	if usesFastProvisioning, ok := d.GetOk("enable_fast_provisioning"); ok {
		params.UsesFastProvisioning = usesFastProvisioning.(bool)
	}

	if overCommitAllowed, ok := d.GetOk("allow_over_commit"); ok {
		params.OverCommitAllowed = overCommitAllowed.(bool)
	}

	if vmDiscoveryEnabled, ok := d.GetOk("enable_vm_discovery"); ok {
		params.VmDiscoveryEnabled = vmDiscoveryEnabled.(bool)
	}

	if elasticity, ok := d.GetOkExists("elasticity"); ok {
		elasticityPt := elasticity.(bool)
		params.IsElastic = &elasticityPt
	}

	if vmMemoryOverhead, ok := d.GetOkExists("include_vm_memory_overhead"); ok {
		vmMemoryOverheadPt := vmMemoryOverhead.(bool)
		params.IncludeMemoryOverhead = &vmMemoryOverheadPt
	}

	return params, nil
}

// resourceVcdOrgVdcImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it set's the ID field for `_resource_name_` resource in state file
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_org_vdc.my_existing_vdc
// Example import path (_the_id_string_): org.my_existing_vdc
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func resourceVcdOrgVdcImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as org.my_existing_vdc")
	}
	orgName, vdcName := resourceURI[0], resourceURI[1]

	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrg(orgName)
	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrg, err)
	}

	adminVdc, err := adminOrg.GetAdminVDCByName(vdcName, false)
	if err != nil {
		log.Printf("[DEBUG] Unable to find VDC %s", vdcName)
		return nil, fmt.Errorf("unable to find VDC %s, err: %s", vdcName, err)
	}

	dSet(d, "org", orgName)
	dSet(d, "name", vdcName)

	d.SetId(adminVdc.AdminVdc.ID)

	return []*schema.ResourceData{d}, nil
}

// setVdcEdgeCluster handles setting of Edge Cluster for NSX-T VDCs
//
// Note. VcdNetworkProfile structure contains more fields (it also depends on the VCD version used).
// The problem is that one must send all other already set fields to avoid reseting them. UI does
// this as well. To make this work well with Terraform and avoid removing other values, current
// state of the structure must always be retrieved and only the required value should be changed.
func setVdcEdgeCluster(d *schema.ResourceData, vdc *govcd.Vdc) error {
	// Set the value even if it is empty as this allows to remove it from configuration
	edgeClusterId := d.Get("edge_cluster_id").(string)

	if !vdc.IsNsxt() {
		return fmt.Errorf("'edge_cluster_id' is only applicable for NSX-T VDCs")
	}

	currentNetworkProfile, err := vdc.GetVdcNetworkProfile()
	if err != nil {
		return fmt.Errorf("error retrieving current VDC Network Profile: %s", err)
	}

	currentNetworkProfile.ServicesEdgeCluster = &types.VdcNetworkProfileServicesEdgeCluster{BackingID: edgeClusterId}

	_, err = vdc.UpdateVdcNetworkProfile(currentNetworkProfile)
	if err != nil {
		return fmt.Errorf("error setting 'edge_cluster_id' '%s' for VDC '%s': %s", edgeClusterId, vdc.Vdc.Name, err)
	}

	return nil
}

// setDataSourceEdgeClusterData is like setEdgeClusterData however it must handle the case where
// user has insufficient rights to retrieve VDC Network Profile. Resource itself is not affected
// by this problem because it requires provider user to create VDC.
func setEdgeClusterData(d *schema.ResourceData, adminVdc *govcd.AdminVdc, source string) error {
	vdcNetworkProfile, err := adminVdc.GetVdcNetworkProfile()
	if err != nil {
		// Conciously ignoring this error and logging it to output as it will most probably be
		// insufficient rights that the user has. It will work with System user but might not work
		// for users that got lower privileges.
		logForScreen(source, fmt.Sprintf("got error while attempting to retrieve Edge Cluster ID: %s", err))
		dSet(d, "edge_cluster_id", "")
		return nil
	}

	if vdcNetworkProfile != nil && vdcNetworkProfile.ServicesEdgeCluster != nil {
		dSet(d, "edge_cluster_id", vdcNetworkProfile.ServicesEdgeCluster.BackingID)
	} else {
		dSet(d, "edge_cluster_id", "")
	}
	return nil
}
