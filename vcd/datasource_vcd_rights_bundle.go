package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdRightsBundle() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceRightsBundleRead,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of rights bundle.",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Rights bundle description",
			},
			"bundle_key": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Key used for internationalization",
			},
			"read_only": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this rights bundle is read-only",
			},
			"rights": &schema.Schema{
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "list of rights assigned to this rights bundle",
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
				Description: "list of tenants to which this rights bundle is published",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func datasourceRightsBundleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericRightsBundleRead(ctx, d, meta, "datasource", "read")
}
