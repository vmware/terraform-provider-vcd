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
				Description: "Name of Segment Profile Template",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of Segment Profile Template",
			},
			"nsxt_manager_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T Manager ID",
			},
			"ip_discovery_profile_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Segment IP Discovery Profile ID",
			},
			"mac_discovery_profile_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Segment MAC Discovery Profile ID",
			},
			"spoof_guard_profile_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Segment Spoof Guard Profile ID",
			},
			"qos_profile_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Segment QoS Profile ID",
			},
			"segment_security_profile_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Segment Security Profile ID",
			},
		},
	}
}

func datasourceVcdSegmentProfileTemplateRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	segmentProfileTemplate, err := vcdClient.GetSegmentProfileTemplateByName(d.Get("name").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	setNsxtSegmentProfileTemplateData(d, segmentProfileTemplate.NsxtSegmentProfileTemplate)
	d.SetId(segmentProfileTemplate.NsxtSegmentProfileTemplate.ID)

	return nil
}
