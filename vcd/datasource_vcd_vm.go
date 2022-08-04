package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdStandaloneVm() *schema.Resource {
	return &schema.Resource{
		ReadContext:        datasourceVcdStandaloneVmRead,
		Schema:      vcdVmDS(standaloneVmType),
		Description: "Standalone VM",
	}
}

func datasourceVcdStandaloneVmRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdVmRead(d, meta, "datasource", standaloneVmType)
}
