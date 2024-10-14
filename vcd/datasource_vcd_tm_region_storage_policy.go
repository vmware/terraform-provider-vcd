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
				Description: "Description of the Region Storage Policy",
			},
			"region_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Region that this Region Storage Policy belongs to",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation status of the Region Storage Policy. Can be [NOT_READY, READY]",
			},
			"storage_capacity_mb": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Storage capacity in megabytes for this Region Storage Policy",
			},
			"storage_consumed_mb": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Consumed storage in megabytes for this Region Storage Policy",
			},
		},
	}
}

func datasourceVcdTmRegionStoragePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdTmRegionStoragePolicyRead(ctx, d, meta, "datasource")
}
