package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdGlobalRole() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceGlobalRoleRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of global role.",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Global role description",
			},
			"bundle_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Key used for internationalization",
			},
			"read_only": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this global role is read-only",
			},
			"rights": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "list of rights assigned to this global role",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"publish_to_all_tenants": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "When true, publishes the global role to all tenants",
			},
			"tenants": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "list of tenants to which this global role is published",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func datasourceGlobalRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericGlobalRoleRead(ctx, d, meta, "datasource", "read")
}
