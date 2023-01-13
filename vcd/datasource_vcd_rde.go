package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdRde() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdRdeRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the Runtime Defined Entity",
			},
			"metadata_entry": getMetadataEntrySchema("Runtime Defined Entity", true, true),
		},
	}
}

func datasourceVcdRdeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdRdeRead(ctx, d, meta, "datasource")
}
