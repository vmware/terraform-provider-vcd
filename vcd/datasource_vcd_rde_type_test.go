//go:build rde || ALL || functional
// +build rde ALL functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"regexp"
	"testing"
)

func TestAccVcdRdeTypeDS(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	// This is a RDE Type that comes with VCD out of the box
	var params = StringMap{
		"TypeNamespace":      "tkgcluster",
		"TypeVersion":        "1.0.0",
		"TypeVendor":         "vmware",
		"InterfaceNamespace": "k8s",
		"InterfaceVersion":   "1.0.0",
		"InterfaceVendor":    "vmware",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccVcdRdeTypeDS, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION data source: %s\n", configText)

	typeName := "data.vcd_rde_type.rde-type-ds"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(typeName, "namespace", params["TypeNamespace"].(string)),
					resource.TestCheckResourceAttr(typeName, "version", params["TypeVersion"].(string)),
					resource.TestCheckResourceAttr(typeName, "vendor", params["TypeVendor"].(string)),
					resource.TestCheckResourceAttr(typeName, "name", "TKG Cluster"), // Name is always the same
					resource.TestCheckResourceAttr(typeName, "id", fmt.Sprintf("urn:vcloud:type:%s:%s:%s", params["TypeVendor"].(string), params["TypeNamespace"].(string), params["TypeVersion"].(string))),
					resource.TestCheckResourceAttr(typeName, "readonly", "false"),
					resource.TestMatchResourceAttr(typeName, "schema", regexp.MustCompile("{.*\"title\":\"tkgcluster\".*}")),
					resource.TestCheckResourceAttrPair(typeName, "interface_ids.0", "data.vcd_rde_interface.interface-ds", "id"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdRdeTypeDS = testAccVcdRdeDefinedInterfaceDS + `
data "vcd_rde_type" "rde-type-ds" {
  namespace = "{{.TypeNamespace}}"
  version   = "{{.TypeVersion}}"
  vendor    = "{{.TypeVendor}}"
}
`
