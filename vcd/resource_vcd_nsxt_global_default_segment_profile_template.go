package vcd

import (
	"context"
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdGlobalDefaultSegmentProfileTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdGlobalDefaultSegmentProfileTemplateCreateUpdate,
		ReadContext:   resourceDataSourceVcdGlobalDefaultSegmentProfileTemplateRead,
		UpdateContext: resourceVcdGlobalDefaultSegmentProfileTemplateCreateUpdate,
		DeleteContext: resourceVcdGlobalDefaultSegmentProfileTemplateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdGlobalDefaultSegmentProfileTemplateImport,
		},

		Schema: map[string]*schema.Schema{
			"vdc_networks_default_segment_profile_template_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Global default NSX-T Segment Profile for Org VDC networks",
			},
			"vapp_networks_default_segment_profile_template_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Global default NSX-T Segment Profile for vApp networks",
			},
		},
	}
}

func resourceVcdGlobalDefaultSegmentProfileTemplateCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	globalDefaultSegmentProfileConfig := &types.NsxtGlobalDefaultSegmentProfileTemplate{
		VappNetworkSegmentProfileTemplateRef: &types.OpenApiReference{ID: d.Get("vapp_networks_default_segment_profile_template_id").(string)},
		VdcNetworkSegmentProfileTemplateRef:  &types.OpenApiReference{ID: d.Get("vdc_networks_default_segment_profile_template_id").(string)},
	}

	_, err := vcdClient.UpdateGlobalDefaultSegmentProfileTemplates(globalDefaultSegmentProfileConfig)
	if err != nil {
		return diag.Errorf("error updating Global Default Segment Profile Template configuration: %s", err)
	}

	d.SetId("global-default-segment-profile")

	return resourceDataSourceVcdGlobalDefaultSegmentProfileTemplateRead(ctx, d, meta)
}

func resourceDataSourceVcdGlobalDefaultSegmentProfileTemplateRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	defaults, err := vcdClient.GetGlobalDefaultSegmentProfileTemplates()
	if err != nil {
		return diag.Errorf("error reading Global Default Segment Profile Template configuration: %s", err)
	}

	dSet(d, "vdc_networks_default_segment_profile_template_id", "")
	if defaults.VdcNetworkSegmentProfileTemplateRef != nil {
		dSet(d, "vdc_networks_default_segment_profile_template_id", defaults.VdcNetworkSegmentProfileTemplateRef.ID)
	}

	dSet(d, "vapp_networks_default_segment_profile_template_id", "")
	if defaults.VappNetworkSegmentProfileTemplateRef != nil {
		dSet(d, "vapp_networks_default_segment_profile_template_id", defaults.VappNetworkSegmentProfileTemplateRef.ID)
	}

	d.SetId("global-default-segment-profile")

	return nil
}

func resourceVcdGlobalDefaultSegmentProfileTemplateDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	_, err := vcdClient.UpdateGlobalDefaultSegmentProfileTemplates(&types.NsxtGlobalDefaultSegmentProfileTemplate{})
	if err != nil {
		return diag.Errorf("error deleting Global Default Segment Profile Template configuration: %s", err)
	}

	return nil
}

func resourceVcdGlobalDefaultSegmentProfileTemplateImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	_, err := vcdClient.GetGlobalDefaultSegmentProfileTemplates()
	if err != nil {
		return nil, fmt.Errorf("error finding Global Segment Profile Template: %s", err)
	}

	d.SetId("global-default-segment-profile")
	return []*schema.ResourceData{d}, nil
}
