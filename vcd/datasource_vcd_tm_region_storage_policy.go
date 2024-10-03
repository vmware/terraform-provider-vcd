package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdTmRegionStoragePolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdTmRegionStoragePolicyRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Region Storage Policy name",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Region Storage Policy description",
			},
		},
	}
}

func datasourceVcdTmRegionStoragePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdTmRegionStoragePolicyRead(ctx, d, meta, "datasource")
}
