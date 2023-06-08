package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// datasourceVcdOrgSaml shows Org SAML settings
func datasourceVcdOrgSaml() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdOrgSamlRead,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Organization ID",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Enable SAML authentication",
			},
			"entity_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Your service provider entity ID.",
			},
			"email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Optional email attribute name",
			},
			"user_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Optional username attribute name",
			},
			"first_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Optional first name attribute name",
			},
			"surname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Optional surname attribute name",
			},
			"full_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Optional full name attribute name",
			},
			"group": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Optional group attribute name",
			},
			"role": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Optional role attribute name",
			},
		},
	}
}

func datasourceVcdOrgSamlRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdOrgSamlRead(ctx, d, meta, "datasource")
}
