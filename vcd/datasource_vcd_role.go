package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdRole() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceRoleRead,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of Role.",
			},
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Role description",
			},
			"bundle_key": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Key used for internationalization",
			},
			"read_only": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this role is read-only",
			},
			"rights": &schema.Schema{
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Set of rights assigned to this role",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func datasourceRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericRoleRead(ctx, d, meta, "datasource", "read")
}
