package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxtOrgVdcNetworkSegmentProfileTemplate() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVcdNsxtOrgVdcNetworkSegmentProfileRead,

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"org_network_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the Organization Network that uses the Segment Profile Template",
			},
			"segment_profile_template_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Segment Profile Template ID",
			},
			"segment_profile_template_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Segment Profile Template Name",
			},
			// Individual Segment Profiles
			"ip_discovery_profile_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T IP Discovery Profile",
			},
			"mac_discovery_profile_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T Mac Discovery Profile",
			},
			"spoof_guard_profile_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T Spoof Guard Profile",
			},
			"qos_profile_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T QoS Profile",
			},
			"segment_security_profile_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T Segment Security Profile",
			},
		},
	}
}

func dataSourceVcdNsxtOrgVdcNetworkSegmentProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDataSourceVcdNsxtOrgVdcNetworkSegmentProfileRead(ctx, d, meta, "datasource")
}
