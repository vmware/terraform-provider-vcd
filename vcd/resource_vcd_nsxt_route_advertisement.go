package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdNsxtRouteAdvertisement() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdNsxtRouteAdvertisementCreateUpdate,
		ReadContext:   resourceVcdNsxtRouteAdvertisementRead,
		UpdateContext: resourceVcdNsxtRouteAdvertisementCreateUpdate,
		DeleteContext: resourceVcdNsxtRouteAdvertisementDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdNsxtRouteAdvertisementImport,
		},
	}
}

func resourceVcdNsxtRouteAdvertisementCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceVcdNsxtRouteAdvertisementRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceVcdNsxtRouteAdvertisementDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceVcdNsxtRouteAdvertisementImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return nil, nil
}
