package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdStandaloneVm() *schema.Resource {

	return &schema.Resource{
		Create: resourceVcdStandaloneVmCreate,
		Update: resourceVcdStandaloneVmUpdate,
		Read:   resourceVcdVStandaloneVmRead,
		Delete: resourceVcdVAppVmDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdVappVmImport,
		},
		Schema:      vmSchemaFunc("standalone"),
		Description: "Standalone VM",
	}
}

func resourceVcdStandaloneVmCreate(d *schema.ResourceData, meta interface{}) error {
	return genericResourceVmCreate(d, meta, "standalone")
}

func resourceVcdStandaloneVmUpdate(d *schema.ResourceData, meta interface{}) error {
	return genericResourceVcdVmUpdate(d, meta, "standalone")
}

func resourceVcdVStandaloneVmRead(d *schema.ResourceData, meta interface{}) error {
	return genericVcdVmRead(d, meta, "resource", "standalone")
}
