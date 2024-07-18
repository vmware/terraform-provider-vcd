package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdExternalEndpoint() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdExternalEndpointRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the External Endpoint",
			},
			"vendor": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Vendor of the External Endpoint",
			},
			"version": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Version of the External Endpoint",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the External Endpoint is enabled or not",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the External Endpoint",
			},
			"root_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URL which requests are redirected to",
			},
		},
	}
}

func datasourceVcdExternalEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdExternalEndpointRead(ctx, d, meta, "datasource")
}
