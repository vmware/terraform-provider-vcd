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
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of global role.",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Global role description",
			},
			"bundle_key": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Key used for internationalization",
			},
			"read_only": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this global role is read-only",
			},
			"rights": &schema.Schema{
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "list of rights assigned to this global role",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"publish_to_all_tenants": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "When true, publishes the global role to all tenants",
			},
			"tenants": &schema.Schema{
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
