//go:build rde || ALL || functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdRdeTypeDS(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	// This is a RDE Type that comes with VCD out of the box
	var params = StringMap{
		"TypeNss":          "tkgcluster",
		"TypeVersion":      "1.0.0",
		"TypeVendor":       "vmware",
		"InterfaceNss":     "k8s",
		"InterfaceVersion": "1.0.0",
		"InterfaceVendor":  "vmware",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccVcdRdeTypeDS, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION data source: %s\n", configText)

	typeName := "data.vcd_rde_type.rde_type_ds"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(typeName, "nss", params["TypeNss"].(string)),
					resource.TestCheckResourceAttr(typeName, "version", params["TypeVersion"].(string)),
					resource.TestCheckResourceAttr(typeName, "vendor", params["TypeVendor"].(string)),
					resource.TestCheckResourceAttr(typeName, "name", "TKG Cluster"), // Name is always the same
					resource.TestCheckResourceAttr(typeName, "id", fmt.Sprintf("urn:vcloud:type:%s:%s:%s", params["TypeVendor"].(string), params["TypeNss"].(string), params["TypeVersion"].(string))),
					resource.TestCheckResourceAttr(typeName, "readonly", "false"),
					resource.TestMatchResourceAttr(typeName, "schema", regexp.MustCompile("{.*\"title\":\"tkgcluster\".*}")),
					resource.TestCheckResourceAttrPair(typeName, "interface_ids.0", "data.vcd_rde_interface.interface_ds", "id"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdRdeTypeDS = `
data "vcd_rde_interface" "interface_ds" {
  nss     = "{{.InterfaceNss}}"
  version = "{{.InterfaceVersion}}"
  vendor  = "{{.InterfaceVendor}}"
}

data "vcd_rde_type" "rde_type_ds" {
  nss     = "{{.TypeNss}}"
  version = "{{.TypeVersion}}"
  vendor  = "{{.TypeVendor}}"
}
`
