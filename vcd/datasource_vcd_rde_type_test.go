//go:build rde || ALL || functional
// +build rde ALL functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccVcdRdeTypeDS(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	// This is a RDE Type that comes with VCD out of the box
	var params = StringMap{
		"Namespace": "tkgcluster",
		"Version":   "1.0.0",
		"Vendor":    "vmware",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccVcdRdeTypeDS, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION data source: %s\n", configText)

	interfaceName := "data.vcd_rde_type.rde-type-ds"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(interfaceName, "namespace", params["Namespace"].(string)),
					resource.TestCheckResourceAttr(interfaceName, "version", params["Version"].(string)),
					resource.TestCheckResourceAttr(interfaceName, "vendor", params["Vendor"].(string)),
					resource.TestCheckResourceAttr(interfaceName, "name", "TKG Cluster"), // Name is always the same
					resource.TestCheckResourceAttr(interfaceName, "id", fmt.Sprintf("urn:vcloud:type:%s:%s:%s", params["Vendor"].(string), params["Namespace"].(string), params["Version"].(string))),
					resource.TestCheckResourceAttr(interfaceName, "readonly", "false"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdRdeTypeDS = `
data "vcd_rde_type" "rde-type-ds" {
  namespace = "{{.Namespace}}"
  version   = "{{.Version}}"
  vendor    = "{{.Vendor}}"
}
`
