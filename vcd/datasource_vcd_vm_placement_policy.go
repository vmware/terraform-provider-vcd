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
				Type:     schema.TypeString,
				Required: true,
				Description: "Name of the VM Placement Policy",
			},
			"pvdc_id": {
				Type:     schema.TypeString,
				Required: true,
				Description: "ID of the Provider VDC to which the VM Placement Policy belongs",
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "Description of the VM Placement Policy",
			},
			"vm_groups": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Collection of VMs with similar host requirements",
			},
			"logical_vm_groups": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "One or more Logical VM Groups defined in this VM Placement policy. There is an AND relationship among all the entries fetched to this attribute",
			},
		},
	}
}

func datasourceVcdVmPlacementPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
