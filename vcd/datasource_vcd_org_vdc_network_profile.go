package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxtOrgVdcNetworkProfile() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVcdNsxtOrgVdcNetworkProfileRead,

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"edge_cluster_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of NSX-T Edge Cluster (provider vApp networking services and DHCP capability for Isolated networks)",
			},
			"vdc_networks_default_segment_profile_template_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Default NSX-T Segment Profile for Org VDC networks",
			},
			"vapp_networks_default_segment_profile_template_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Default NSX-T Segment Profile for vApp networks",
			},
		},
	}
}

func dataSourceVcdNsxtOrgVdcNetworkProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDataSourceVcdNsxtOrgVdcNetworkProfileRead(ctx, d, meta, "datasource")
}
