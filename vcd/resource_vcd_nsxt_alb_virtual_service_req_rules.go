package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdAlbVirtualServiceReqRules() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdAlbVirtualServiceReqRulesCreate,
		ReadContext:   resourceVcdAlbVirtualServiceReqRulesRead,
		UpdateContext: resourceVcdAlbVirtualServiceReqRulesUpdate,
		DeleteContext: resourceVcdAlbVirtualServiceReqRulesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdAlbVirtualServiceReqRulesImport,
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

func resourceVcdAlbVirtualServiceReqRulesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdAlbVirtualServiceReqRulesRead(ctx, d, meta)
}

func resourceVcdAlbVirtualServiceReqRulesUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdAlbVirtualServiceReqRulesRead(ctx, d, meta)
}

func resourceVcdAlbVirtualServiceReqRulesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceVcdAlbVirtualServiceReqRulesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceVcdAlbVirtualServiceReqRulesImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}
