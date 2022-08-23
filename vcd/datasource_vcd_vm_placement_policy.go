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
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vm_groups": {
				Type:     schema.TypeSet,
				Computed: true,
				Description: "Collection of VMs with similar host requirements",
			},
			"logical_vm_groups": {
				Type:     schema.TypeSet,
				Computed: true,
				Description: "One or more Logical VM Groups defined in this VM Placement policy. There is an AND relationship among all the entries fetched to this attribute",
			},
		},
	}
}

// datasourceVcdVmPlacementPolicyRead reads a data source VM Placement policy
func datasourceVcdVmPlacementPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdVmPlacementPolicyRead(ctx, d, meta)
}
