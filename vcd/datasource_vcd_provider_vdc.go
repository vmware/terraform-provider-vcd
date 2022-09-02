package vcd

import (
	"context"
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// datasourceVcdProviderVdc defines the data source for a Provider VDC.
func datasourceVcdProviderVdc() *schema.Resource {
	// This internal schema defines the Root Capacity of the Provider VDC.
	rootCapacityUsage := func(typeOfCapacity string) *schema.Schema {
		return &schema.Schema{
			Type:     schema.TypeList,
			Computed: true,
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
				Description: "Set of virtual hardware versions supported by this Provider VDC",
			},
			"compute_capacity": {
				Type:        schema.TypeList,
				Computed:    true,
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
				Computed:    true,
				Description: "The highest virtual hardware version supported by this Provider VDC",
			},
			"nsxt_manager_id": {
				Type:        schema.TypeString,
				Computed:    true,
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
				Computed:    true,
				Description: "Set of IDs of the resource pools backing this provider VDC",
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

	if extendedProviderVdc.VMWProviderVdc.AvailableNetworks != nil {
		dSet(d, "external_network_ids", extractIdsFromReferences(extendedProviderVdc.VMWProviderVdc.AvailableNetworks.Network))
	}

	if extendedProviderVdc.VMWProviderVdc.DataStoreRefs != nil {
		dSet(d, "storage_container_ids", extractIdsFromVimObjectRefs(extendedProviderVdc.VMWProviderVdc.DataStoreRefs.VimObjectRef))
	}

	if extendedProviderVdc.VMWProviderVdc.StorageProfiles != nil {
		dSet(d, "storage_profile_ids", extractIdsFromReferences(extendedProviderVdc.VMWProviderVdc.StorageProfiles.ProviderVdcStorageProfile))
	}

	if extendedProviderVdc.VMWProviderVdc.ResourcePoolRefs != nil {
		dSet(d, "resource_pool_ids", extractIdsFromVimObjectRefs(extendedProviderVdc.VMWProviderVdc.ResourcePoolRefs.VimObjectRef))
	}

	if extendedProviderVdc.VMWProviderVdc.NetworkPoolReferences != nil {
		dSet(d, "network_pool_ids", extractIdsFromReferences(extendedProviderVdc.VMWProviderVdc.NetworkPoolReferences.NetworkPoolReference))
	}

	var items []string
	if extendedProviderVdc.VMWProviderVdc.Capabilities != nil && extendedProviderVdc.VMWProviderVdc.Capabilities.SupportedHardwareVersions != nil {
		items = append(items, extendedProviderVdc.VMWProviderVdc.Capabilities.SupportedHardwareVersions.SupportedHardwareVersion...)
	}
	dSet(d, "capabilities", items)

	if extendedProviderVdc.VMWProviderVdc.HostReferences != nil {
		dSet(d, "host_ids", extractIdsFromReferences(extendedProviderVdc.VMWProviderVdc.HostReferences.HostReference))
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
	err = d.Set("metadata", getMetadataStruct(metadata.MetadataEntry))
	if err != nil {
		return diag.Errorf("There was an issue when setting metadata into the schema - %s", err)
	}

	d.SetId(providerVdc.ProviderVdc.ID)
	return nil
}

// getComputeCapacityForProviderVdc constructs a specific struct for `compute_capacity` attribute in the `vcd_provider_vdc` Terraform state.
func getComputeCapacityForProviderVdc(computeCapacity *types.RootComputeCapacity) *[]map[string]interface{} {
	cpuValueMap := map[string]interface{}{}
	if computeCapacity.Cpu != nil {
		cpuValueMap["allocation"] = computeCapacity.Cpu.Allocation
		cpuValueMap["total"] = computeCapacity.Cpu.Total
		cpuValueMap["overhead"] = computeCapacity.Cpu.Overhead
		cpuValueMap["used"] = computeCapacity.Cpu.Used
		cpuValueMap["units"] = computeCapacity.Cpu.Units
		cpuValueMap["reserved"] = computeCapacity.Cpu.Reserved
	}
	memoryValueMap := map[string]interface{}{}
	if computeCapacity.Memory != nil {
		memoryValueMap["allocation"] = computeCapacity.Memory.Allocation
		memoryValueMap["total"] = computeCapacity.Memory.Total
		memoryValueMap["overhead"] = computeCapacity.Memory.Overhead
		memoryValueMap["used"] = computeCapacity.Memory.Used
		memoryValueMap["units"] = computeCapacity.Memory.Units
		memoryValueMap["reserved"] = computeCapacity.Memory.Reserved
	}
	var memoryCapacityArray []map[string]interface{}
	memoryCapacityArray = append(memoryCapacityArray, memoryValueMap)
	var cpuCapacityArray []map[string]interface{}
	cpuCapacityArray = append(cpuCapacityArray, cpuValueMap)

	rootInternal := map[string]interface{}{}
	rootInternal["cpu"] = &cpuCapacityArray
	rootInternal["memory"] = &memoryCapacityArray
	rootInternal["is_elastic"] = computeCapacity.IsElastic
	rootInternal["is_ha"] = computeCapacity.IsHA

	var root []map[string]interface{}
	root = append(root, rootInternal)
	return &root
}
