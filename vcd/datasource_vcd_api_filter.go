package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdApiFilter() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdApiFilterRead,
		Schema: map[string]*schema.Schema{
			"api_filter_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the API Filter that unequivocally identifies it",
			},
			"external_endpoint_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the External Endpoint where this API Filter will process the requests to",
			},
			"url_matcher_pattern": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Request URL pattern, written as a regular expression pattern",
			},
			"url_matcher_scope": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Can be EXT_API, EXT_UI_PROVIDER, EXT_UI_TENANT corresponding to /ext-api, /ext-ui/provider, /ext-ui/tenant/<tenant-name>",
			},
		},
	}
}

func datasourceVcdApiFilterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId(d.Get("api_filter_id").(string))
	return genericVcdApiFilterRead(ctx, d, meta, "datasource")
}
