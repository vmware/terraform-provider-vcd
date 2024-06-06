package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdMultisiteSiteAssociation() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdSiteAssociationRead,
		Schema: map[string]*schema.Schema{
			"associated_site_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "ID of the site to which the associated site belongs",
				ExactlyOneOf: []string{"associated_site_id", "association_data_file"},
			},
			"associated_site_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the site to which the associated site belongs",
			},
			"associated_site_href": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL of the associated site",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the site association",
			},
			"association_data_file": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"associated_site_id", "association_data_file"},
				Description:  "Name of the file filled with association data for this Site. Used when user doesn't have associated site ID",
			},
		},
	}
}

func datasourceVcdSiteAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdSiteAssociationRead(ctx, d, meta, "datasource")
}
