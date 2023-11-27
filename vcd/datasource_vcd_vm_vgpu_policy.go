package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdVmVgpuPolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdVmVgpuPolicyRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vgpu_profile": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Defines the vGPU profile configuration.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The identifier of the vGPU profile.",
						},
						"count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Specifies the number of vGPU profiles. ",
						},
					},
				},
			},
			"cpu": {
				Computed: true,
				Type:     schema.TypeList,
				Elem:     sizingPolicyCpuDS,
			},
			"memory": {
				Computed: true,
				Type:     schema.TypeList,
				Elem:     sizingPolicyMemoryDS,
			},
			"provider_vdc_scope": {
				Computed: true,
				Type:     schema.TypeSet,
				Elem:     providerVdcScopeDS,
			},
		},
	}
}

var providerVdcScopeDS = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"provider_vdc_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Identifier for the provider virtual data center.",
		},
		"cluster_names": {
			Type:        schema.TypeSet,
			Computed:    true,
			Description: "Set of cluster names within the provider virtual data center.",
			Elem: &schema.Schema{
				MinItems: 1,
				Type:     schema.TypeString,
			},
		},
		"vm_group_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Identifier for a VM group within the provider VDC scope.",
		},
	},
}

// datasourceVcdVmVgpuPolicyRead reads a data source VM vGPU policy
func datasourceVcdVmVgpuPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdVgpuPolicyRead(ctx, d, meta)
}
