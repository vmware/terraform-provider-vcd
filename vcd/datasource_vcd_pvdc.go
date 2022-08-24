package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdPvdc() *schema.Resource {

	return &schema.Resource{
		ReadContext: datasourceVcdVmSizingPolicyRead,
		Schema: map[string]*schema.Schema{

		},
	}
}

func datasourceVcdPvdcRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
