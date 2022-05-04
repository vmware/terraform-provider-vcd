package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"strings"
	"testing"
)

func TestAccVcdSecurityTag(t *testing.T) {
	tag1 := strings.ToLower(t.Name() + "-tag1") // security tags are always lowercase in serverside
	tag2 := strings.ToLower(t.Name() + "-tag2")

	var params = StringMap{
		"Org":          testConfig.VCD.Org,
		"Vdc":          testConfig.VCD.Vdc,
		"VappName":     t.Name() + "-vapp",
		"VmName":       t.Name() + "-vm",
		"ComputerName": t.Name() + "-vm",
		"Catalog":      testConfig.VCD.Catalog.Name,
		"CatalogItem":  testConfig.VCD.Catalog.CatalogItem,
		"SecurityTag1": tag1,
		"SecurityTag2": tag2,
		"FuncName":     t.Name(),
	}

	configText := templateFill(testAccSecurityTag, params)

	params["FuncName"] = t.Name() + "-update"
	configTextUpdate := templateFill(testAccSecurityTagUpdate, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSecurityTagDestroy(tag1),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityTagCreated(tag1),
				),
			},
			{
				Config: configTextUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityTagCreated(tag1, tag2),
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

resource "vcd_security_tag" "{{.SecurityTag1}}" {
  name = "{{.SecurityTag1}}"
  vm_ids = [vcd_vapp_vm.{{.VmName}}.id]
}
`

const testAccSecurityTagUpdate = `
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

resource "vcd_security_tag" "{{.SecurityTag1}}" {
  name = "{{.SecurityTag1}}"
  vm_ids = [vcd_vapp_vm.{{.VmName}}.id]
}

resource "vcd_security_tag" "{{.SecurityTag2}}" {
  name = "{{.SecurityTag2}}"
  vm_ids = [vcd_vapp_vm.{{.VmName}}.id]
}
`

func testAccCheckSecurityTagDestroy(securityTags ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		org, err := conn.GetOrgByName(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf("error retrieving the Org - %s", testConfig.VCD.Org)
		}

		securityTagValues, err := org.GetSecurityTagValues("")
		if err != nil {
			return fmt.Errorf("error retrieving Org %s security tags", testConfig.VCD.Org)
		}

		for _, tag := range securityTags {
			for _, securityTagValue := range securityTagValues {
				if securityTagValue.Tag == tag {
					return fmt.Errorf("found tag %s after destroying", tag)
				}
			}
		}
		return nil
	}
}

func testAccCheckSecurityTagCreated(securityTags ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		org, err := conn.GetOrgByName(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf("error retrieving the Org - %s", testConfig.VCD.Org)
		}

		securityTagValues, err := org.GetSecurityTagValues("")
		if err != nil {
			return fmt.Errorf("error retrieving Org %s security tags", testConfig.VCD.Org)
		}

		for _, tag := range securityTags {
			for i, securityTagValue := range securityTagValues {
				if securityTagValue.Tag == tag {
					break
				}
				if i == len(securityTagValues)-1 {
					return fmt.Errorf("tag %s wasn't found", tag)
				}
			}
		}
		return nil
	}
}
