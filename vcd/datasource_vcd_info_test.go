package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"testing"
)

func TestAccVcdDatasourceInfo(t *testing.T) {

	if !usingSysAdmin() {
		t.Skip("TestAccVcdDatasourceInfo requires system admin privileges")
		return
	}

	configText := templateFill(testAccCheckVcdDatasourceInfo, StringMap{})

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.vcd_info.list_resources", "name", "list_resources"),
				),
			},
		},
	})
}

const testAccCheckVcdDatasourceInfo = `
data "vcd_info" "list_resources" {
  name          = "list_resources"
  resource_type = "resources"
}

output "resources" {
  value = data.vcd_info.list_resources
}
`
