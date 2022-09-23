package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/util"
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
	util.Logger.Printf("[DEBUG] [VM create] started standalone VM creation")
	if d.Get("vapp_name").(string) != "" {
		return diag.Errorf("vApp name must not be set for a standalone VM (resource `vcd_vm`)")
	}

	util.Logger.Printf("[DEBUG] [VM create] finished standalone VM creation")

	err := genericResourceVmCreate(d, meta, standaloneVmType)
	if err != nil {
		return err
	}

	return genericVcdVmRead(d, meta, "create")
}

func resourceVcdStandaloneVmUpdate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericResourceVcdVmUpdate(d, meta, standaloneVmType)
}

func resourceVcdVStandaloneVmRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdVmRead(d, meta, "resource")
}
