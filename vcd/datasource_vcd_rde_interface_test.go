//go:build rde || ALL || functional
// +build rde ALL functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccVcdRdeDefinedInterfaceDS(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	// This is a Defined Interface that comes with VCD out of the box
	var params = StringMap{
		"InterfaceNamespace": "k8s",
		"InterfaceVersion":   "1.0.0",
		"InterfaceVendor":    "vmware",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccVcdRdeDefinedInterfaceDS, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION data source: %s\n", configText)

	interfaceName := "data.vcd_rde_interface.interface-ds"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(interfaceName, "namespace", params["InterfaceNamespace"].(string)),
					resource.TestCheckResourceAttr(interfaceName, "version", params["InterfaceVersion"].(string)),
					resource.TestCheckResourceAttr(interfaceName, "vendor", params["InterfaceVendor"].(string)),
					resource.TestCheckResourceAttr(interfaceName, "name", "Kubernetes"), // Name is always the same
					resource.TestCheckResourceAttr(interfaceName, "id", fmt.Sprintf("urn:vcloud:interface:%s:%s:%s", params["InterfaceVendor"].(string), params["InterfaceNamespace"].(string), params["InterfaceVersion"].(string))),
					resource.TestCheckResourceAttr(interfaceName, "readonly", "false"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdRdeDefinedInterfaceDS = `
data "vcd_rde_interface" "interface-ds" {
  namespace = "{{.InterfaceNamespace}}"
  version   = "{{.InterfaceVersion}}"
  vendor    = "{{.InterfaceVendor}}"
}
`
