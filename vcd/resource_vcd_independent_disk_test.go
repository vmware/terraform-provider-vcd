// +build disk ALL functional

package vcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

var resourceName = "TestAccVcdIndependentDiskBasic_1"
var name = "TestAccVcdIndependentDiskBasic"

func TestAccVcdIndependentDiskBasic(t *testing.T) {

	var params = StringMap{
		"Org":                testConfig.VCD.Org,
		"Vdc":                testConfig.VCD.Vdc,
		"name":               name,
		"size":               "5000",
		"busType":            "SCSI",
		"busSubType":         "lsilogicsas",
		"storageProfileName": "*",
		"ResourceName":       resourceName,
		"Tags":               "disk",
	}

	configText := templateFill(testAccCheckVcdIndependentDiskBasic, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testDiskResourcesDestroyed,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskCreated("vcd_independent_disk."+resourceName, name),
				),
			},
			resource.TestStep{
				ResourceName:            "vcd_independent_disk." + resourceName + "-import",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdByDisk(name),
				ImportStateVerifyIgnore: []string{"org", "vdc", "size"},
			},
		},
	})
}

func testAccCheckDiskCreated(itemName, diskName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		injectItemRs, ok := s.RootModule().Resources[itemName]
		if !ok {
			return fmt.Errorf("not found: %s", itemName)
		}

		if injectItemRs.Primary.ID == "" {
			return fmt.Errorf("no disk insert ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.VCD.Vdc, testConfig.VCD.Org, err)
		}

		_, err = vdc.GetDiskByName(diskName, true)
		if err != nil {
			return fmt.Errorf("independent disk %s isn't exist and error: %#v", itemName, err)
		}

		return nil
	}
}

func testDiskResourcesDestroyed(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)
	for _, rs := range s.RootModule().Resources {
		itemName := rs.Primary.Attributes["name"]
		if rs.Type != "vcd_independent_disk" && itemName != name {
			continue
		}

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.VCD.Vdc, testConfig.VCD.Org, err)
		}

		_, err = vdc.GetDiskByName(name, true)
		if !govcd.IsNotFound(err) {
			return fmt.Errorf("independent disk %s still exist and error: %#v", itemName, err)
		}

	}
	return nil
}

func importStateIdByDisk(objectName string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		importId := testConfig.VCD.Vdc + "." + objectName
		if testConfig.VCD.Vdc == "" || objectName == "" {
			return "", fmt.Errorf("missing information to generate import path: %s", importId)
		}
		return importId, nil
	}
}

func init() {
	testingTags["disk"] = "resource_vcd_independent_disk_test.go"
}

const testAccCheckVcdIndependentDiskBasic = `
resource "vcd_independent_disk" "{{.ResourceName}}" {
  org             = "{{.Org}}"
  vdc             = "{{.Vdc}}"
  name            = "{{.name}}"
  size            = "{{.size}}"
  bus_type        = "{{.busType}}"
  bus_sub_type    = "{{.busSubType}}"
  storage_profile = "{{.storageProfileName}}"
}
`
