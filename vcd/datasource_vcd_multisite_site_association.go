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
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the site to which the associated site belongs",
			},
			"associated_site_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the site to which the associated site belongs",
			},
		},
	}
}

func datasourceVcdSiteAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdSiteAssociationRead(ctx, d, meta, "datasource")
}
