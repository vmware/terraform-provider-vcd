package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/vmware/terraform-provider-vcd/v3/vcd"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: vcd.Provider})
}
