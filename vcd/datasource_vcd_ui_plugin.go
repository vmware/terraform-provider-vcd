package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdUIPlugin() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdUIPluginRead,
		Schema: map[string]*schema.Schema{
			"vendor": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The UI Plugin vendor name. Combination of `vendor`, `name` and `version` must be unique",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The UI Plugin name. Combination of `vendor`, `name` and `version` must be unique",
			},
			"version": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The version of the UI Plugin. Combination of `vendor`, `name` and `version` must be unique",
			},
			"license": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The license of the UI Plugin",
			},
			"link": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The website of the UI Plugin",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The description of the UI Plugin",
			},
			"provider_scoped": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "'true' if the UI Plugin scope is the service provider. 'false' if not",
			},
			"tenant_scoped": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "true if the UI Plugin scope is the tenants (organizations)",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "true if the UI Plugin is enabled. 'false' if not",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the UI Plugin",
			},
			"tenant_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
				Description: "Set of Organization IDs where the UI Plugin is published to",
			},
		},
	}
}

func datasourceVcdUIPluginRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdUIPluginRead(ctx, d, meta, "datasource")
}
