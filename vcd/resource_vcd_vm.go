package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdStandaloneVm() *schema.Resource {

	return &schema.Resource{
		CreateContext: resourceVcdStandaloneVmCreate,
		UpdateContext: resourceVcdStandaloneVmUpdate,
		ReadContext:   resourceVcdVStandaloneVmRead,
		DeleteContext: resourceVcdVAppVmDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdVappVmImport,
		},
		Schema:      vmSchemaFunc(standaloneVmType),
		Description: "Standalone VM",
	}
}

func resourceVcdStandaloneVmCreate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	err := genericResourceVmCreate(d, meta, standaloneVmType)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceVcdStandaloneVmUpdate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	err := genericResourceVcdVmUpdate(d, meta, standaloneVmType)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceVcdVStandaloneVmRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	err := genericVcdVmRead(d, meta, "resource")
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
