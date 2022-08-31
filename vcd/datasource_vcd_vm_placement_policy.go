package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"log"
)

func datasourceVcdVmPlacementPolicy() *schema.Resource {

	return &schema.Resource{
		ReadContext: datasourceVcdVmPlacementPolicyRead,
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
				Computed:    true,
				Description: "Description of the VM Placement Policy",
			},
			"vm_group_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
				Description: "IDs of the collection of VMs with similar host requirements",
			},
			"logical_vm_group_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
				Description: "IDs of one or more Logical VM Groups defined in this VM Placement policy. There is an AND relationship among all the entries fetched to this attribute",
			},
		},
	}
}

func datasourceVcdVmPlacementPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdVmPlacementPolicyRead(ctx, d, meta)
}

// TODO: Probably we should move this function to the Resource when it is created, to follow same code style as other resource-datasource pairs.
// setVmPlacementPolicy sets object state from *govcd.VdcComputePolicy
func setVmPlacementPolicy(_ context.Context, d *schema.ResourceData, policy types.VdcComputePolicy) diag.Diagnostics {

	dSet(d, "name", policy.Name)
	dSet(d, "description", policy.Description)
	var vmGroupIds []string
	for _, namedVmGroupPerPvdc := range policy.NamedVMGroups {
		for _, namedVmGroup := range namedVmGroupPerPvdc {
			vmGroupIds = append(vmGroupIds, namedVmGroup.ID)
		}
	}
	dSet(d, "vm_group_ids", vmGroupIds)
	vmGroupIds = []string{}
	for _, namedVmGroup := range policy.LogicalVMGroupReferences {
		vmGroupIds = append(vmGroupIds, namedVmGroup.ID)
	}
	dSet(d, "logical_vm_group_ids", vmGroupIds)

	log.Printf("[TRACE] VM Placement Policy read completed: %s", policy.Name)
	return nil
}
