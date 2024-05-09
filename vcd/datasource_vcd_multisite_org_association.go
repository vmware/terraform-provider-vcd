package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdMultisiteOrgAssociation() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdOrgAssociationRead,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Organization ID",
			},
			"associated_org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the associated Organization",
			},
			"associated_org_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the associated Organization",
			},
			"associated_site_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the site to which the associated Organization belongs",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the Org association",
			},
		},
	}
}

func datasourceVcdOrgAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdOrgAssociationRead(ctx, d, meta, "datasource")
}
