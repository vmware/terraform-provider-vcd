package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdSegmentProfileTemplate() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdSegmentProfileTemplateRead,

		Schema: map[string]*schema.Schema{

			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "NSX-T Segment Profile Template name",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T Segment Profile Template name",
			},
			"nsxt_manager_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T Segment Profile Template name",
			},
			"ip_discovery_profile_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T Segment Profile Template name",
			},
			"mac_discovery_profile_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T Segment Profile Template name",
			},
			"spoof_guard_profile_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T Segment Profile Template name",
			},
			"qos_profile_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T Segment Profile Template name",
			},
			"segment_security_profile_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T Segment Profile Template name",
			},
		},
	}
}

func datasourceVcdSegmentProfileTemplateRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("this resource is only supported for Providers")
	}

	segmentProfileTemplate, err := vcdClient.GetSegmentProfileTemplateByName(d.Get("name").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	setNsxtSegmentProfileTemplateData(d, segmentProfileTemplate.NsxtSegmentProfileTemplate)

	d.SetId(segmentProfileTemplate.NsxtSegmentProfileTemplate.ID)

	return nil
}
