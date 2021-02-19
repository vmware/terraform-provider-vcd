package vcd

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func datasourceVcdStandaloneVm() *schema.Resource {
	return &schema.Resource{
		Read:        datasourceVcdStandaloneVmRead,
		Schema:      vcdVmDS(standaloneVmType),
		Description: "Standalone VM",
	}
}

func datasourceVcdStandaloneVmRead(d *schema.ResourceData, meta interface{}) error {
	return genericVcdVmRead(d, meta, "datasource", standaloneVmType)
}
