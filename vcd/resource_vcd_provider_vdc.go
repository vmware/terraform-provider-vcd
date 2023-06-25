package vcd

import (
	"context"
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"log"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// datasourceVcdProviderVdc defines the data source for a Provider VDC.
func resourceVcdProviderVdc() *schema.Resource {
	// This internal schema defines the Root Capacity of the Provider VDC.
	rootCapacityUsage := func(typeOfCapacity string) *schema.Schema {
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

	return &schema.Resource{
		ReadContext:   resourceVcdProviderVdcRead,
		CreateContext: resourceVcdProviderVdcCreate,
		UpdateContext: resourceVcdProviderVdcUpdate,
		DeleteContext: resourceVcdProviderVdcDelete,
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
						"cpu":    rootCapacityUsage("CPU"),
						"memory": rootCapacityUsage("Memory"),
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
				Optional:    true,
				Description: "The highest virtual hardware version supported by this Provider VDC",
			},
			"nsxt_manager_id": {
				Type:        schema.TypeString,
				Optional:    true,
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
			"storage_profiles": {
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
			"resource_pool_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the resource pool needed to instantiate the provider VDC.",
			},
			"resource_pool_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
				Description: "Set of IDs of the resource pools backing this provider VDC",
			},
			"network_pool_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the network pool used to create this Provider VDC",
			},
			"network_pool_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
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
			"metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Key and value pairs for Provider VDC metadata",
				Deprecated:  "Use metadata_entry instead",
			},
			"metadata_entry": getMetadataEntrySchema("Provider VDC", false),
		},
	}
}

func resourceVcdProviderVdcCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	vcdClient := meta.(*VCDClient)
	providerVdcName := d.Get("name").(string)
	providerVdcDescription := d.Get("description").(string)
	vcenterId := d.Get("vcenter_id").(string)
	resourcePoolName := d.Get("resource_pool_name").(string)
	managerid := d.Get("nsxt_manager_id").(string)
	networkPoolId := d.Get("network_pool_id").(string)
	hwVersion := d.Get("highest_supported_hardware_version").(string)
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
	resourcePool, err := vcenter.GetAvailableResourcePoolByName(resourcePoolName)
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
	rawStorageProfiles := d.Get("storage_profiles").(*schema.Set)
	var wantedStorageProfiles []string
	for _, sp := range rawStorageProfiles.List() {
		wantedStorageProfiles = append(wantedStorageProfiles, sp.(string))
	}
	allStorageProfiles, err := vcdClient.Client.QueryAllProviderVdcStorageProfiles()
	if err != nil {
		return diag.FromErr(err)
	}
	var foundStorageProfiles []string

	var used = make(map[string]bool)
	for _, sp := range allStorageProfiles {
		if contains(wantedStorageProfiles, sp.Name) {
			seen, ok := used[sp.Name]
			if ok && seen {
				continue
			}
			foundStorageProfiles = append(foundStorageProfiles, sp.Name)
			used[sp.Name] = true
		}
	}
	if len(foundStorageProfiles) != len(wantedStorageProfiles) {
		return diag.Errorf(" %d storage profiles were requested, but %d were found", len(wantedStorageProfiles), len(foundStorageProfiles))
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
		StorageProfile: foundStorageProfiles,
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

func resourceVcdProviderVdcRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	providerVdcName := d.Get("name").(string)
	extendedProviderVdc, err := vcdClient.GetProviderVdcExtendedByName(providerVdcName)
	if err != nil {
		log.Printf("[DEBUG] Could not find any extended Provider VDC with name %s: %s", providerVdcName, err)
		return diag.Errorf("could not find any extended Provider VDC with name %s: %s", providerVdcName, err)
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

	if extendedProviderVdc.VMWProviderVdc.DataStoreRefs != nil {
		if err = d.Set("storage_container_ids", extractIdsFromVimObjectRefs(extendedProviderVdc.VMWProviderVdc.DataStoreRefs.VimObjectRef)); err != nil {
			return diag.Errorf("error setting storage_container_ids: %s", err)
		}
	}

	if extendedProviderVdc.VMWProviderVdc.StorageProfiles != nil {
		if err = d.Set("storage_profile_ids", extractIdsFromReferences(extendedProviderVdc.VMWProviderVdc.StorageProfiles.ProviderVdcStorageProfile)); err != nil {
			return diag.Errorf("error setting storage_profile_ids: %s", err)
		}
	}

	if extendedProviderVdc.VMWProviderVdc.ResourcePoolRefs != nil {
		if err = d.Set("resource_pool_ids", extractIdsFromVimObjectRefs(extendedProviderVdc.VMWProviderVdc.ResourcePoolRefs.VimObjectRef)); err != nil {
			return diag.Errorf("error setting resource_pool_ids: %s", err)
		}
	}

	if extendedProviderVdc.VMWProviderVdc.NetworkPoolReferences != nil {
		if err = d.Set("network_pool_ids", extractIdsFromReferences(extendedProviderVdc.VMWProviderVdc.NetworkPoolReferences.NetworkPoolReference)); err != nil {
			return diag.Errorf("error setting network_pool_ids: %s", err)
		}
	}

	var items []string
	if extendedProviderVdc.VMWProviderVdc.Capabilities != nil && extendedProviderVdc.VMWProviderVdc.Capabilities.SupportedHardwareVersions != nil {
		items = append(items, extendedProviderVdc.VMWProviderVdc.Capabilities.SupportedHardwareVersions.SupportedHardwareVersion...)
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
		dSet(d, "vcenter_id", extendedProviderVdc.VMWProviderVdc.VimServer.ID)
	}

	metadata, err := providerVdc.GetMetadata()
	if err != nil {
		log.Printf("[DEBUG] Error retrieving metadata for Provider VDC: %s", err)
		return diag.Errorf("error retrieving metadata for Provider VDC %s: %s", providerVdcName, err)
	}
	if len(metadata.MetadataEntry) > 0 {
		if err = d.Set("metadata", getMetadataStruct(metadata.MetadataEntry)); err != nil {
			return diag.Errorf("There was an issue when setting metadata into the schema - %s", err)
		}
	}

	d.SetId(providerVdc.ProviderVdc.ID)
	return nil
}

func resourceVcdProviderVdcUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	providerVdcId := d.Id()
	providerVdcName := d.Get("name").(string)
	//extendedProviderVdc, err := vcdClient.GetProviderVdcExtendedById(providerVdcId)
	_, err := vcdClient.GetProviderVdcExtendedById(providerVdcId)
	if err != nil {
		log.Printf("[DEBUG] Could not find any extended Provider VDC with name %s: %s", providerVdcName, err)
		return diag.Errorf("could not find any extended Provider VDC with name %s: %s", providerVdcName, err)
	}

	return diag.Errorf("not implemented yet")
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
