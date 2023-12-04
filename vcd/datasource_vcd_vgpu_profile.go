package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdVgpuProfile() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdVgpuProfileRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Description: "ID of the vGPU profile",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the vGPU profile",
			},
			"tenant_facing_name": {
				Type:        schema.TypeString,
				Description: "The tenant facing name of the vGPU profile",
				Computed:    true,
			},
			"instructions": {
				Type:        schema.TypeString,
				Description: "The instructions for the vGPU profile",
				Computed:    true,
			},
		},
	}
}

func datasourceVcdVgpuProfileRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	profileName := d.Get("name").(string)
	vgpuProfile, err := vcdClient.GetVgpuProfileByName(profileName)
	if err != nil {
		return diag.FromErr(err)
	}

	dSet(d, "tenant_facing_name", vgpuProfile.VgpuProfile.TenantFacingName)
	dSet(d, "instructions", vgpuProfile.VgpuProfile.Instructions)
	d.SetId(vgpuProfile.VgpuProfile.Id)

	return nil
}
