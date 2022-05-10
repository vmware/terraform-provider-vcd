package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdOrgVdcAccessControl() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVdcVdcAccessControlCreate,
		ReadContext:   resourceVdcVdcAccessControlRead,
		UpdateContext: resourceVdcVdcAccessControlUpdate,
		DeleteContext: resourceVdcVdcAccessControlDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVdcVdcAccessControlImport,
		},
		Schema: map[string]*schema.Schema{},
	}
}

func resourceVdcVdcAccessControlCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceVdcVdcAccessControlRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceVdcVdcAccessControlUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceVdcVdcAccessControlDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceVdcVdcAccessControlImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return nil, nil
}
