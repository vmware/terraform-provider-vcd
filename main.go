package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/terraform-providers/terraform-provider-vcd/vcd"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: vcd.Provider})
}
