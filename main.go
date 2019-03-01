package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/terraform-providers/terraform-provider-vcd/v2/vcd"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: vcd.Provider})
}
