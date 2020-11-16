// +build ALL functional

package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccVcdDatasourceResourceSchema(t *testing.T) {

	configText := templateFill(testAccCheckVcdDatasourceStructure, StringMap{})

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.vcd_resource_schema.struct_org", "name", "struct_org"),
				),
			},
		},
	})
}

const testAccCheckVcdDatasourceStructure = `
data "vcd_resource_schema" "struct_org" {
  name          = "struct_org"
  resource_type = "vcd_org"
}

output "resources" {
  value = data.vcd_resource_schema.struct_org
}
`
