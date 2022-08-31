package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	return sharedVcdVmPlacementPolicyRead(ctx, d, meta)
}
