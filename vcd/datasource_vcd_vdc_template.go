package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdOrgVdcTemplate() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdVdcTemplateRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the VDC Template",
			},
		},
	}
}

func datasourceVcdVdcTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdVdcTemplateRead(ctx, d, meta)
}
