package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdGlobalDefaultSegmentProfileTemplate() *schema.Resource {
	return &schema.Resource{
		ReadContext: resourceDataSourceVcdGlobalDefaultSegmentProfileTemplateRead,
		Schema: map[string]*schema.Schema{
			"vdc_networks_default_segment_profile_template_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Global default NSX-T Segment Profile for Org VDC networks",
			},
			"vapp_networks_default_segment_profile_template_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Global default NSX-T Segment Profile for vApp networks",
			},
		},
	}
}
