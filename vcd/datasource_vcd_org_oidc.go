package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// datasourceVcdOrgOidc defines the data source that reads Open ID Connect (OIDC) settings from an existing Organization
func datasourceVcdOrgOidc() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdOrgOidcRead,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Organization ID from where the OpenID Connect settings are read",
			},
		},
	}
}

func datasourceVcdOrgOidcRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdOrgOidcRead(ctx, d, meta, "datasource")
}
