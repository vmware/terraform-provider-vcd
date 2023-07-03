package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"log"
	"net/url"
)

// This internal schema defines the Root Capacity of the Provider VDC.
func providerVdcRootCapacityUsage(typeOfCapacity string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: fmt.Sprintf("Single-element list with an indicator of %s capacity available in the Provider VDC", typeOfCapacity),
		// MaxItems: 1 - A computed field can't use "MaxItems", this is a reminder that this is a single-element list.
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"allocation": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: fmt.Sprintf("Allocated %s for this Provider VDC", typeOfCapacity),
				},
				"overhead": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: fmt.Sprintf("%s overhead for this Provider VDC", typeOfCapacity),
				},
				"reserved": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: fmt.Sprintf("Reserved %s for this Provider VDC", typeOfCapacity),
				},
				"total": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: fmt.Sprintf("Total %s for this Provider VDC", typeOfCapacity),
				},
				"units": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: fmt.Sprintf("Units for the %s of this Provider VDC", typeOfCapacity),
				},
				"used": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: fmt.Sprintf("Used %s in this Provider VDC", typeOfCapacity),
				},
			},
		},
	}
}

// resourceVcdProviderVdc defines the resource for a Provider VDC.
func resourceVcdProviderVdc() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdProviderVdcCreate,
		ReadContext:   resourceVcdProviderVdcRead,
		UpdateContext: resourceVcdProviderVdcUpdate,
		DeleteContext: resourceVcdProviderVdcDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceProviderVdcImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Provider VDC",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional description of the Provider VDC",
			},
			"status": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Status of the Provider VDC, it can be -1 (creation failed), 0 (not ready), 1 (ready), 2 (unknown) or 3 (unrecognized)",
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "True if this Provider VDC is enabled and can provide resources to organization VDCs. A Provider VDC is always enabled on creation",
			},
			"capabilities": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
				Description: "Set of virtual hardware versions supported by this Provider VDC",
			},
			"compute_capacity": {
				Type:     schema.TypeList,
				Computed: true,
				// MaxItems: 1 - A computed field can't use "MaxItems", this is a reminder that this is a single-element list.
				Description: "Single-element list with an indicator of CPU and memory capacity",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu":    providerVdcRootCapacityUsage("CPU"),
						"memory": providerVdcRootCapacityUsage("Memory"),
						"is_elastic": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "True if compute capacity can grow or shrink based on demand",
						},
						"is_ha": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "True if compute capacity is highly available",
						},
					},
				},
			},
			"compute_provider_scope": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Represents the compute fault domain for this Provider VDC. This value is a tenant-facing tag that is shown to tenants when viewing fault domains of the child Organization VDCs (for example, a VDC Group)",
			},
			"highest_supported_hardware_version": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The highest virtual hardware version supported by this Provider VDC",
			},
			"nsxt_manager_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the registered NSX-T Manager that backs networking operations for this Provider VDC",
			},
			"storage_container_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
				Description: "Set of IDs of the vSphere datastores backing this provider VDC",
			},
			"external_network_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
				Description: "Set of IDs of external networks",
			},
			"storage_profile_names": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required:    true,
				Description: "Set of storage profile names used to create this Provider VDC",
			},
			"storage_profile_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
				Description: "Set of IDs to the storage profiles available to this Provider VDC",
			},
			"resource_pool_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required:    true,
				Description: "Set of IDs of the resource pools backing this provider VDC",
			},
			"network_pool_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Computed:    true,
				Description: "Set IDs of the network pools used by this Provider VDC",
			},
			"universal_network_pool_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the universal network reference",
			},
			"host_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
				Description: "Set with all the hosts which are connected to VC server",
			},
			"vcenter_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the vCenter server that provides the resource pools and datastores",
			},
			// TODO: metadata handling to be added after refactoring of conflicting fields "metadata" and "metadata_entry"
			//"metadata": {
			//	Type:        schema.TypeMap,
			//	Computed:    true,
			//	Description: "Key and value pairs for Provider VDC metadata",
			//	Deprecated:  "Use metadata_entry instead",
			//},
			//"metadata_entry": getMetadataEntrySchema("Provider VDC", false),
		},
	}
}

func resourceVcdProviderVdcCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	providerVdcName := d.Get("name").(string)
	providerVdcDescription := d.Get("description").(string)
	vcenterId := d.Get("vcenter_id").(string)
	managerid := d.Get("nsxt_manager_id").(string)
	hwVersion := d.Get("highest_supported_hardware_version").(string)
	rawNetworkPoolIds := d.Get("network_pool_ids").(*schema.Set)
	rawResourcePoolIds := d.Get("resource_pool_ids").(*schema.Set)
	managerHref, err := url.JoinPath(vcdClient.Client.VCDHREF.String(), "admin", "extension", "nsxtManagers", managerid)
	if err != nil {
		return diag.FromErr(err)
	}

	vcenter, err := vcdClient.GetVcenterById(vcenterId)
	if err != nil {
		return diag.FromErr(err)
	}
	vcenterUrl, err := vcenter.GetVimServerUrl()
	if err != nil {
		return diag.FromErr(err)
	}
	if rawNetworkPoolIds.Len() == 0 {
		return diag.Errorf("no network pool was provided")
	}
	if rawNetworkPoolIds.Len() > 1 {
		return diag.Errorf("only one network pool can be used to create a Provider VDC")
	}
	if rawResourcePoolIds.Len() == 0 {
		return diag.Errorf("no resource pool was provided")
	}
	if rawResourcePoolIds.Len() > 1 {
		return diag.Errorf("only one resource pool can be used to create a Provider VDC")
	}
	resourcePoolId := rawResourcePoolIds.List()[0].(string)
	networkPoolId := rawNetworkPoolIds.List()[0].(string)

	resourcePool, err := vcenter.GetResourcePoolById(resourcePoolId)
	if err != nil {
		return diag.FromErr(err)
	}
	defaultHwVersion, err := resourcePool.GetDefaultHardwareVersion()
	if err != nil {
		return diag.FromErr(err)
	}
	if hwVersion == "" {
		hwVersion = defaultHwVersion
	}
	rawStorageProfileNames := d.Get("storage_profile_names").(*schema.Set)
	var wantedStorageProfiles []string
	for _, sp := range rawStorageProfileNames.List() {
		wantedStorageProfiles = append(wantedStorageProfiles, sp.(string))
	}
	if len(wantedStorageProfiles) == 0 {
		return diag.Errorf("no storage profiles were indicated")
	}

	nsxtManagers, err := vcdClient.QueryNsxtManagerByHref(managerHref)
	if err != nil {
		return diag.FromErr(err)
	}
	if len(nsxtManagers) > 1 {
		return diag.Errorf("more than one NSX-T manager found with ID %s", managerid)
	}
	networkPool, err := vcdClient.GetNetworkPoolById(networkPoolId)
	if err != nil {
		return diag.FromErr(err)
	}
	networkPoolHref, err := networkPool.GetOpenApiUrl()
	if err != nil {
		return diag.FromErr(err)
	}

	providerVdcCreation := types.ProviderVdcCreation{
		Name:                            providerVdcName,
		Description:                     providerVdcDescription,
		HighestSupportedHardwareVersion: hwVersion,
		IsEnabled:                       d.Get("is_enabled").(bool),
		VimServer: []*types.Reference{
			{
				HREF: vcenterUrl,
				ID:   extractUuid(vcenter.VSphereVcenter.VcId),
				Name: vcenter.VSphereVcenter.Name,
			},
		},
		ResourcePoolRefs: &types.VimObjectRefs{
			VimObjectRef: []*types.VimObjectRef{
				{
					VimServerRef: &types.Reference{
						HREF: vcenterUrl,
						ID:   extractUuid(vcenter.VSphereVcenter.VcId),
						Name: vcenter.VSphereVcenter.Name,
					},
					MoRef:         resourcePool.ResourcePool.Moref,
					VimObjectType: "RESOURCE_POOL",
				},
			},
		},
		StorageProfile: wantedStorageProfiles,
		NsxTManagerReference: types.Reference{
			HREF: nsxtManagers[0].HREF,
			ID:   extractUuid(nsxtManagers[0].HREF),
			Name: nsxtManagers[0].Name,
		},
		NetworkPool: types.Reference{
			HREF: networkPoolHref,
			Name: networkPool.NetworkPool.Name,
			ID:   extractUuid(networkPool.NetworkPool.Id),
			Type: networkPool.NetworkPool.PoolType,
		},
		AutoCreateNetworkPool: false,
	}
	providerVdc, err := vcdClient.CreateProviderVdc(&providerVdcCreation)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(providerVdc.VMWProviderVdc.ID)
	return resourceVcdProviderVdcRead(ctx, d, meta)
}

func resourceVcdProviderVdcRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericResourceVcdProviderVdcRead(ctx, d, meta, "resource")
}

func genericResourceVcdProviderVdcRead(ctx context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	providerVdcName := d.Get("name").(string)
	providerVdcId := d.Id()
	var extendedProviderVdc *govcd.ProviderVdcExtended
	var err error
	if extractUuid(d.Id()) != "" {
		extendedProviderVdc, err = vcdClient.GetProviderVdcExtendedById(providerVdcId)
	} else {
		extendedProviderVdc, err = vcdClient.GetProviderVdcExtendedByName(providerVdcName)
	}
	if err != nil {
		log.Printf("[DEBUG] (%s) Could not find any extended Provider VDC with name %s: %s", origin, providerVdcName, err)
		if origin == "datasource" {
			return diag.Errorf("could not find any extended Provider VDC with name %s: %s", providerVdcName, err)
		}
		d.SetId("")
		return nil
	}
	providerVdc, err := extendedProviderVdc.ToProviderVdc()
	if err != nil {
		log.Printf("[DEBUG] Could not find any Provider VDC with name %s: %s", providerVdcName, err)
		return diag.Errorf("could not find any Provider VDC with name %s: %s", providerVdcName, err)
	}

	dSet(d, "name", extendedProviderVdc.VMWProviderVdc.Name)
	dSet(d, "description", extendedProviderVdc.VMWProviderVdc.Description)
	dSet(d, "status", extendedProviderVdc.VMWProviderVdc.Status)
	dSet(d, "is_enabled", extendedProviderVdc.VMWProviderVdc.IsEnabled)
	dSet(d, "compute_provider_scope", extendedProviderVdc.VMWProviderVdc.ComputeProviderScope)
	dSet(d, "highest_supported_hardware_version", extendedProviderVdc.VMWProviderVdc.HighestSupportedHardwareVersion)

	if extendedProviderVdc.VMWProviderVdc.NsxTManagerReference != nil {
		dSet(d, "nsxt_manager_id", extendedProviderVdc.VMWProviderVdc.NsxTManagerReference.ID)
	}

	if extendedProviderVdc.VMWProviderVdc.AvailableNetworks != nil {
		if err = d.Set("external_network_ids", extractIdsFromReferences(extendedProviderVdc.VMWProviderVdc.AvailableNetworks.Network)); err != nil {
			return diag.Errorf("error setting external_network_ids: %s", err)
		}
	}

	if extendedProviderVdc.VMWProviderVdc.DataStoreRefs.VimObjectRef != nil {
		ids := ObjectMap[*types.VimObjectRef, string](extendedProviderVdc.VMWProviderVdc.DataStoreRefs.VimObjectRef,
			vimObjectRefToMoref)
		if err = d.Set("storage_container_ids", ids); err != nil {
			return diag.Errorf("error setting storage_container_ids: %s", err)
		}
	}

	if extendedProviderVdc.VMWProviderVdc.StorageProfiles != nil {
		ids := ObjectMap[*types.Reference, string](extendedProviderVdc.VMWProviderVdc.StorageProfiles.ProviderVdcStorageProfile,
			referenceToId)
		if err = d.Set("storage_profile_ids", ids); err != nil {
			return diag.Errorf("error setting storage_profile_ids: %s", err)
		}

		names := ObjectMap[*types.Reference, string](extendedProviderVdc.VMWProviderVdc.StorageProfiles.ProviderVdcStorageProfile,
			referenceToName)
		if err = d.Set("storage_profile_names", names); err != nil {
			return diag.Errorf("error setting storage_profile_names: %s", err)
		}
	}

	if extendedProviderVdc.VMWProviderVdc.ResourcePoolRefs != nil {
		ids := ObjectMap[*types.VimObjectRef, string](extendedProviderVdc.VMWProviderVdc.ResourcePoolRefs.VimObjectRef,
			vimObjectRefToMoref)
		if err = d.Set("resource_pool_ids", ids); err != nil {
			return diag.Errorf("error setting resource_pool_ids: %s", err)
		}
	}

	// Network pool IDs cannot be read safely from the provider VDC.
	// During creation, if more than one network pool is available from the designated NSX-T manager,
	// the system takes all of them, even if only one network pool was selected. As a result, 'terraform plan' will fail.
	// Given that the network pools cannot be changed, it is safe to let "network_pool_ids" maintain the values given at
	// creation.
	// The only case when this arrangement may fail is during import.

	rawNetworkPoolSet := d.Get("network_pool_ids")
	// We only collect network pool data if we have the information from schema.
	// (This operation will cause a plan failure if it happens after an import)
	if rawNetworkPoolSet != nil && rawNetworkPoolSet.(*schema.Set).Len() != 0 {
		networkPoolIds := rawNetworkPoolSet.(*schema.Set).List()
		existingIds := extractIdsFromReferences(extendedProviderVdc.VMWProviderVdc.NetworkPoolReferences.NetworkPoolReference)
		// Safe case: the number of existing network pools and the requested network pools is the same
		if len(networkPoolIds) == len(extendedProviderVdc.VMWProviderVdc.NetworkPoolReferences.NetworkPoolReference) {
			if err = d.Set("network_pool_ids", existingIds); err != nil {
				return diag.Errorf("error setting network_pool_ids: %s", err)
			}
		} else {
			// Unsafe case: there are more network pools in the provider VDC than the ones requested at creation
			// We only get the common IDs, to keep the plan clean
			var commonIds []string
			for _, id := range existingIds {
				if rawNetworkPoolSet.(*schema.Set).Contains(id) {
					commonIds = append(commonIds, id)
				}
			}
			if err = d.Set("network_pool_ids", commonIds); err != nil {
				return diag.Errorf("error setting network_pool_ids: %s", err)
			}
		}
	}
	var items []string
	if extendedProviderVdc.VMWProviderVdc.Capabilities != nil && extendedProviderVdc.VMWProviderVdc.Capabilities.SupportedHardwareVersions != nil {
		for _, item := range extendedProviderVdc.VMWProviderVdc.Capabilities.SupportedHardwareVersions.SupportedHardwareVersion {
			items = append(items, item.Name)
		}
	}
	if err = d.Set("capabilities", items); err != nil {
		return diag.Errorf("error setting capabilities: %s", err)
	}

	if extendedProviderVdc.VMWProviderVdc.HostReferences != nil {
		if err = d.Set("host_ids", extractIdsFromReferences(extendedProviderVdc.VMWProviderVdc.HostReferences.HostReference)); err != nil {
			return diag.Errorf("error setting host_ids: %s", err)
		}
	}

	if extendedProviderVdc.VMWProviderVdc.ComputeCapacity != nil {
		if err = d.Set("compute_capacity", getComputeCapacityForProviderVdc(extendedProviderVdc.VMWProviderVdc.ComputeCapacity)); err != nil {
			return diag.Errorf("error setting compute_capacity: %s", err)
		}
	}

	if extendedProviderVdc.VMWProviderVdc.AvailableUniversalNetworkPool != nil {
		dSet(d, "universal_network_pool_id", extendedProviderVdc.VMWProviderVdc.AvailableUniversalNetworkPool.ID)
	}
	if extendedProviderVdc.VMWProviderVdc.VimServer != nil {
		dSet(d, "vcenter_id", extendedProviderVdc.VMWProviderVdc.VimServer[0].ID)
	}

	// TODO: metadata handling to be added after refactoring of conflicting fields "metadata" and "metadata_entry"
	//metadata, err := providerVdc.GetMetadata()
	//if err != nil {
	//	log.Printf("[DEBUG] Error retrieving metadata for Provider VDC: %s", err)
	//	return diag.Errorf("error retrieving metadata for Provider VDC %s: %s", providerVdcName, err)
	//}
	//if len(metadata.MetadataEntry) > 0 {
	//	if err = d.Set("metadata", getMetadataStruct(metadata.MetadataEntry)); err != nil {
	//		return diag.Errorf("There was an issue when setting metadata into the schema - %s", err)
	//	}
	//}

	d.SetId(providerVdc.ProviderVdc.ID)
	return nil
}

func resourceVcdProviderVdcUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	providerVdcId := d.Id()
	providerVdcName := d.Get("name").(string)
	//extendedProviderVdc, err := vcdClient.GetProviderVdcExtendedById(providerVdcId)
	pvdc, err := vcdClient.GetProviderVdcExtendedById(providerVdcId)
	if err != nil {
		log.Printf("[DEBUG] Could not find any extended Provider VDC with name %s: %s", providerVdcName, err)
		return diag.Errorf("could not find any extended Provider VDC with name %s: %s", providerVdcName, err)
	}

	if d.HasChanges("name", "description") {
		err = pvdc.Rename(d.Get("name").(string), d.Get("description").(string))
		if err != nil {
			return diag.Errorf("error renaming provider VDC '%s': %s", providerVdcName, err)
		}
	}
	if d.HasChange("is_enabled") {
		wantEnabled := d.Get("is_enabled").(bool)
		requested := "en"
		if wantEnabled {
			err = pvdc.Enable()
		} else {
			requested = "dis"
			err = pvdc.Disable()
		}
		if err != nil {
			return diag.Errorf("error %sabling provider VDC '%s': %s", requested, providerVdcName, err)
		}
	}
	if d.HasChange("resource_pool_ids") {
		err = updateProviderVdcResourcePools(vcdClient, pvdc, d)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChange("storage_profile_names") {
		err = updateProviderVdcStorageProfiles(vcdClient, pvdc, d)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return resourceVcdProviderVdcRead(ctx, d, meta)
}

func resourceVcdProviderVdcDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	providerVdcId := d.Id()
	providerVdcName := d.Get("name").(string)
	extendedProviderVdc, err := vcdClient.GetProviderVdcExtendedById(providerVdcId)
	if err != nil {
		log.Printf("[DEBUG] Could not find any extended Provider VDC with name %s: %s", providerVdcName, err)
		return diag.Errorf("could not find any extended Provider VDC with name %s: %s", providerVdcName, err)
	}
	if extendedProviderVdc.IsEnabled() {
		err = extendedProviderVdc.Disable()
		if err != nil {
			return diag.Errorf("error disabling provider VDC %s: %s", providerVdcName, err)
		}
	}
	task, err := extendedProviderVdc.Delete()
	if err != nil {
		return diag.FromErr(err)
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func updateProviderVdcResourcePools(client *VCDClient, pvdc *govcd.ProviderVdcExtended, d *schema.ResourceData) error {
	existingResourcePoolsIds := ObjectMap[*types.VimObjectRef, string](pvdc.VMWProviderVdc.ResourcePoolRefs.VimObjectRef, vimObjectRefToMoref)
	rawResourcePoolIds := d.Get("resource_pool_ids").(*schema.Set)
	var requestedResourcePoolIds []string
	for _, rpId := range rawResourcePoolIds.List() {
		requestedResourcePoolIds = append(requestedResourcePoolIds, rpId.(string))
	}

	var toBeAdded []string
	var toBeDeleted []string

	for _, wanted := range requestedResourcePoolIds {
		if contains(existingResourcePoolsIds, wanted) {
			// The same ID is in the provider VDC and in the request : skipping
			continue
		}
		// Not in the provider VDC: going to the adding list
		toBeAdded = append(toBeAdded, wanted)
	}
	for _, existing := range existingResourcePoolsIds {
		if contains(requestedResourcePoolIds, existing) {
			// The same ID is in provider VDC and the request: skipping
			continue
		}
		// It is in the provider VDC, but not in the request: will be deleted
		toBeDeleted = append(toBeDeleted, existing)
	}

	removed := false
	added := false
	if len(toBeAdded) > 0 {
		newResourcePools, err := client.ResourcePoolsFromIds(toBeAdded)
		if err != nil {
			return fmt.Errorf("error getting resource pools from list of IDs: %s", err)
		}
		err = pvdc.AddResourcePools(newResourcePools)
		if err != nil {
			return fmt.Errorf("error adding resource pools to provider VDC %s: %s", pvdc.VMWProviderVdc.Name, err)
		}
		added = true
	}
	if len(toBeDeleted) > 0 {
		selectedResourcePools, err := client.ResourcePoolsFromIds(toBeDeleted)
		if err != nil {
			return fmt.Errorf("error getting resource pools from list of IDs: %s", err)
		}
		err = pvdc.DeleteResourcePools(selectedResourcePools)
		if err != nil {
			return fmt.Errorf("error removing resource pools from provider VDC %s: %s", pvdc.VMWProviderVdc.Name, err)
		}
		removed = true
	}

	if !removed && !added {
		return fmt.Errorf("changes requested but none were performed. A likely explanation is a mismatch or a duplicate in the resource pool IDs ")
	}
	return nil
}
func updateProviderVdcStorageProfiles(client *VCDClient, pvdc *govcd.ProviderVdcExtended, d *schema.ResourceData) error {
	existingStorageProfiles := ObjectMap[*types.Reference, string](pvdc.VMWProviderVdc.StorageProfiles.ProviderVdcStorageProfile, referenceToName)
	rawStorageProfiles := d.Get("storage_profile_names").(*schema.Set)
	var requestedStorageProfileNames []string
	for _, spName := range rawStorageProfiles.List() {
		requestedStorageProfileNames = append(requestedStorageProfileNames, spName.(string))
	}

	var toBeAdded []string
	var toBeDeleted []string

	for _, wanted := range requestedStorageProfileNames {
		if contains(existingStorageProfiles, wanted) {
			// The same storage profile is in the provider VDC and in the request : skipping
			continue
		}
		// Not in the provider VDC: going to the adding list
		toBeAdded = append(toBeAdded, wanted)
	}
	for _, existing := range existingStorageProfiles {
		if contains(requestedStorageProfileNames, existing) {
			// The same storage profile name is in provider VDC and the request: skipping
			continue
		}
		// It is in the provider VDC, but not in the request: will be deleted
		toBeDeleted = append(toBeDeleted, existing)
	}

	removed := false
	added := false
	if len(toBeAdded) > 0 {
		err := pvdc.AddStorageProfiles(toBeAdded)
		if err != nil {
			return fmt.Errorf("error adding storage profiles to provider VDC %s: %s", pvdc.VMWProviderVdc.Name, err)
		}
		added = true
	}
	if len(toBeDeleted) > 0 {
		err := pvdc.DeleteStorageProfiles(toBeDeleted)
		if err != nil {
			return fmt.Errorf("error removing storage profiles from provider VDC %s: %s", pvdc.VMWProviderVdc.Name, err)
		}
		removed = true
	}

	if !removed && !added {
		return fmt.Errorf("changes requested but none were performed. A likely explanation is a mismatch or a duplicate in the storage profile names ")
	}
	return nil
}

func resourceProviderVdcImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)
	identifier := d.Id()
	if identifier == "" {
		return nil, fmt.Errorf("[provider VDC import] no identifier given. The name or the ID of the provider VDC should be given")
	}

	var pvdc *govcd.ProviderVdcExtended
	var err error
	if extractUuid(identifier) != "" {
		pvdc, err = vcdClient.GetProviderVdcExtendedById(identifier)
	} else {
		pvdc, err = vcdClient.GetProviderVdcExtendedByName(identifier)
	}
	if err != nil {
		return nil, fmt.Errorf("[provider VDC import] error retrieving provider VDC '%s'", identifier)
	}
	d.SetId(pvdc.VMWProviderVdc.ID)
	dSet(d, "name", pvdc.VMWProviderVdc.Name)

	return []*schema.ResourceData{d}, nil
}
