package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// datasourceVcdProviderVdc defines the data source for a Provider VDC.
func datasourceVcdProviderVdc() *schema.Resource {
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
			"storage_profile_names": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
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
				Deprecated:  "Use metadata_entry instead",
			},
			"metadata_entry": metadataEntryDatasourceSchema("Provider VDC"),
		},
	}
}

func datasourceVcdProviderVdcRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericResourceVcdProviderVdcRead(ctx, d, meta, "datasource")
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
