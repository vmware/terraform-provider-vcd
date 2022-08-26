package vcd

import (
	"context"
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdProviderVdc() *schema.Resource {
	rootCapacityUsage := func(typeOfCapacity string) *schema.Schema {
		return &schema.Schema{
			Type:     schema.TypeList,
			Computed: true,
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
		ReadContext: datasourceVcdProviderVdcRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Provider VDC",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Optional description of the Provider VDC",
			},
			"status": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Status of the Provider VDC, it can be -1 (creation failed), 0 (not ready), 1 (ready), 2 (unknown) or 3 (unrecognized)",
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if this Provider VDC is enabled and can provide resources to organization VDCs. A Provider VDC is always enabled on creation",
			},
			"capabilities": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
				Description: "Set of virtual hardware versions supported by this Provider VDC.",
			},
			"compute_capacity": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu":    rootCapacityUsage("CPU"),
						"memory": rootCapacityUsage("Memory"),
						"is_elastic": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "True if this storage profile is enabled for use in the VDC.",
						},
						"is_ha": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Maximum number of MB allocated for this storage profile. A value of 0 specifies unlimited MB.",
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
				Computed:    true,
				Description: "The highest virtual hardware version supported by this Provider VDC",
			},
			"nsxt_manager_id": { // FIXME: Should we use name??
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the registered NSX-T Manager that backs networking operations for this Provider VDC",
			},
			"vdc_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
				Description: "Set of organization VDCs backed by this Provider VDC",
			},
			"storage_containers_ids": {
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
			"storage_profile_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
				Description: "Set of IDs to the storage profiles available to this Provider VDC.",
			},
			"resource_pool_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
				Description: "Resource pools backing this provider VDC",
			},
			"network_pool_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
				Description: "Set of network pools used by this Provider VDC",
			},
			"universal_network_pool_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the universal network reference.",
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
				Computed:    true,
				Description: "ID of the vCenter server that provides the resource pools and datastores",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Key and value pairs for Provider VDC metadata",
			},
		},
	}
}

func datasourceVcdProviderVdcRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	dSet(d, "nsxt_manager_id", extendedProviderVdc.VMWProviderVdc.NsxTManagerReference.ID)

	var items []string
	for _, vdcId := range extendedProviderVdc.VMWProviderVdc.Vdcs {
		if vdcId != nil {
			items = append(items, vdcId.ID)
		}
	}
	dSet(d, "vdc_ids", items)

	items = []string{}
	if extendedProviderVdc.VMWProviderVdc.AvailableNetworks != nil {
		for _, network := range extendedProviderVdc.VMWProviderVdc.AvailableNetworks.Network {
			if network != nil {
				items = append(items, network.ID)
			}
		}
	}
	dSet(d, "external_network_ids", items)

	items = []string{}
	if extendedProviderVdc.VMWProviderVdc.DataStoreRefs != nil {
		for _, vimObject := range extendedProviderVdc.VMWProviderVdc.DataStoreRefs.VimObjectRef {
			if vimObject != nil && vimObject.VimServerRef != nil {
				items = append(items, vimObject.VimServerRef.ID)
			}
		}
	}
	dSet(d, "storage_containers_ids", items)

	items = []string{}
	if extendedProviderVdc.VMWProviderVdc.StorageProfiles != nil {
		for _, storageProfile := range extendedProviderVdc.VMWProviderVdc.StorageProfiles.ProviderVdcStorageProfile {
			if storageProfile != nil {
				items = append(items, storageProfile.ID)
			}
		}
	}
	dSet(d, "storage_profile_ids", items)

	items = []string{}
	if extendedProviderVdc.VMWProviderVdc.ResourcePoolRefs != nil {
		for _, vimObject := range extendedProviderVdc.VMWProviderVdc.ResourcePoolRefs.VimObjectRef {
			if vimObject != nil && vimObject.VimServerRef != nil {
				items = append(items, vimObject.VimServerRef.ID)
			}
		}
	}
	dSet(d, "resource_pool_ids", items)

	items = []string{}
	if extendedProviderVdc.VMWProviderVdc.NetworkPoolReferences != nil {
		for _, networkPool := range extendedProviderVdc.VMWProviderVdc.NetworkPoolReferences.NetworkPoolReference {
			if networkPool != nil {
				items = append(items, networkPool.ID)
			}
		}
	}
	dSet(d, "network_pool_ids", items)

	items = []string{}
	if extendedProviderVdc.VMWProviderVdc.Capabilities != nil && extendedProviderVdc.VMWProviderVdc.Capabilities.SupportedHardwareVersions != nil {
		items = append(items, extendedProviderVdc.VMWProviderVdc.Capabilities.SupportedHardwareVersions.SupportedHardwareVersion...)
	}
	dSet(d, "capabilities", items)

	items = []string{}
	if extendedProviderVdc.VMWProviderVdc.HostReferences != nil {
		for _, host := range extendedProviderVdc.VMWProviderVdc.HostReferences.HostReference {
			items = append(items, host.ID)
		}
	}
	dSet(d, "host_ids", items)

	if extendedProviderVdc.VMWProviderVdc.ComputeCapacity != nil {
		computeCapacity := getComputeCapacityForProviderVdc(extendedProviderVdc.VMWProviderVdc.ComputeCapacity)
		if err = d.Set("compute_capacity", computeCapacity); err != nil {
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
	err = d.Set("metadata", getMetadataStruct(metadata.MetadataEntry))
	if err != nil {
		return diag.Errorf("There was an issue when setting metadata into the schema - %s", err)
	}

	d.SetId(providerVdc.ProviderVdc.ID)
	return nil
}

// getComputeCapacityForProviderVdc constructs a specific struct for `compute_capacity` attribute in the `vcd_provider_vdc` Terraform state.
func getComputeCapacityForProviderVdc(computeCapacity *types.RootComputeCapacity) *[]map[string]interface{} {
	rootInternal := map[string]interface{}{}
	var root []map[string]interface{}

	cpuValueMap := map[string]interface{}{}
	if computeCapacity.Cpu != nil {
		cpuValueMap["cpu_allocation"] = *computeCapacity.Cpu.Allocation
		cpuValueMap["cpu_total"] = computeCapacity.Cpu.Total
		cpuValueMap["cpu_overhead"] = *computeCapacity.Cpu.Overhead
		cpuValueMap["cpu_used"] = *computeCapacity.Cpu.Used
		cpuValueMap["cpu_units"] = computeCapacity.Cpu.Units
		cpuValueMap["cpu_reserved"] = *computeCapacity.Cpu.Reserved
	}
	memoryValueMap := map[string]interface{}{}
	if computeCapacity.Memory != nil {
		memoryValueMap["memory_allocation"] = *computeCapacity.Memory.Allocation
		memoryValueMap["memory_total"] = computeCapacity.Memory.Total
		memoryValueMap["memory_overhead"] = *computeCapacity.Memory.Overhead
		memoryValueMap["memory_used"] = *computeCapacity.Memory.Used
		memoryValueMap["memory_units"] = computeCapacity.Memory.Units
		memoryValueMap["memory_reserved"] = *computeCapacity.Memory.Reserved
	}
	var memoryCapacityArray []map[string]interface{}
	memoryCapacityArray = append(memoryCapacityArray, memoryValueMap)
	var cpuCapacityArray []map[string]interface{}
	cpuCapacityArray = append(cpuCapacityArray, cpuValueMap)

	if computeCapacity.IsElastic != nil {
		rootInternal["is_elastic"] = *computeCapacity.IsElastic
	}
	if computeCapacity.IsHA != nil {
		rootInternal["is_elastic"] = *computeCapacity.IsHA
	}
	rootInternal["cpu"] = &cpuCapacityArray
	rootInternal["memory"] = &memoryCapacityArray

	root = append(root, rootInternal)
	return &root
}
