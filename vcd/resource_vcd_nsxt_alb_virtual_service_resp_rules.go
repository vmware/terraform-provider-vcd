package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdAlbVirtualServiceRespRules() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdAlbVirtualServiceRespRulesCreate,
		ReadContext:   resourceVcdAlbVirtualServiceRespRulesRead,
		UpdateContext: resourceVcdAlbVirtualServiceRespRulesUpdate,
		DeleteContext: resourceVcdAlbVirtualServiceRespRulesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdAlbVirtualServiceRespRulesImport,
		},

		Schema: map[string]*schema.Schema{
			"virtual_service_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "NSX-T ALB Virtual Service ID",
			},
		},
	}
}

func resourceVcdAlbVirtualServiceRespRulesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdAlbVirtualServiceRespRulesRead(ctx, d, meta)
}

func resourceVcdAlbVirtualServiceRespRulesUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdAlbVirtualServiceRespRulesRead(ctx, d, meta)
}

func resourceVcdAlbVirtualServiceRespRulesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceVcdAlbVirtualServiceRespRulesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceVcdAlbVirtualServiceRespRulesImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}
