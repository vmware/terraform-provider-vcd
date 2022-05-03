package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

// Test
func TestAccVcdSecurityTag(t *testing.T) {
	tagName1 := t.Name() + "-tag1"
	tagName2 := t.Name() + "-tag2"

	var params = StringMap{
		"Org":          testConfig.VCD.Org,
		"Vdc":          testConfig.VCD.Vdc,
		"VappName":     t.Name() + "-vapp",
		"VmName":       t.Name() + "-vm",
		"ComputerName": t.Name() + "-vm",
		"Catalog":      testConfig.VCD.Catalog.Name,
		"CatalogItem":  testConfig.VCD.Catalog.CatalogItem,
		"SecurityTag1": tagName1,
		"SecurityTag2": tagName2,
	}

	configText := templateFill(testAccSecurityTag, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSecurityTagDestroy(),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityTagCreated(),
				),
			},
		},
	})
	postTestChecks(t)
}

// Templates to use on different stages
const testAccSecurityTag = `
resource "vcd_vapp" "{{.VappName}}" {
  name = "{{.VappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
}

resource "vcd_vapp_vm" "{{.VmName}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  vapp_name     = vcd_vapp.{{.VappName}}.name
  name          = "{{.VmName}}"
  computer_name = "{{.ComputerName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 1024
  cpus          = 2
  cpu_cores     = 1
}

resource "vcd_security_tag" "{{.tagName1}}" {
  name = "{{.tagName1}}"
  vm_ids = [vcd_vapp_vm.{{.VmName}}.id]
}
`

func testAccCheckSecurityTagDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// TBD
		return nil
	}
}

func testAccCheckSecurityTagCreated() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// TBD
		return nil
	}
}
