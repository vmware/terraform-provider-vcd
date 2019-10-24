// +build disk ALL functional

package vcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"regexp"
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
		"secondName":         name + "second",
		"size":               "5000",
		"sizeInBytes":        "5242880000",
		"busType":            "SCSI",
		"busSubType":         "lsilogicsas",
		"storageProfileName": "*",
		"ResourceName":       resourceName,
		"secondResourceName": resourceName + "second",
		"Tags":               "disk",
	}

	params["FuncName"] = t.Name() + "-Compatibility"
	configTextForCompatibility := templateFill(testAccCheckVcdIndependentDiskForCompatibility, params)
	/*	params["FuncName"] = t.Name()
		configText := templateFill(testAccCheckVcdIndependentDiskBasic, params)
	*/if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s", configTextForCompatibility)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testDiskResourcesDestroyed,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configTextForCompatibility,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskCreated("vcd_independent_disk."+resourceName),
					//resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "size_in_bytes", "5242880000"),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "owner_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "datastore_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "iops", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "is_attached", "false"),
				),
			},
			/*			resource.TestStep{
						Config: configText,
						Check: resource.ComposeTestCheckFunc(
							testAccCheckDiskCreated("vcd_independent_disk."+resourceName+"second"),
							resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName+"second", "size_in_bytes", "5242880000"),
							resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName+"second", "owner_name", regexp.MustCompile(`^\S+`)),
							resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName+"second", "datastore_name", regexp.MustCompile(`^\S+`)),
							resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName+"second", "iops", regexp.MustCompile(`^\d+$`)),
							resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName+"second", "is_attached", "false"),
						),
					},*/
			resource.TestStep{
				ResourceName:            "vcd_independent_disk." + resourceName + "-import",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdByDisk("vcd_independent_disk." + resourceName),
				ImportStateVerifyIgnore: []string{"org", "vdc", "size"},
			},
		},
	})
}

func testAccCheckDiskCreated(itemName string) resource.TestCheckFunc {
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

		_, err = vdc.GetDiskById(injectItemRs.Primary.ID, true)
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

		_, err = vdc.GetDisksByName(name, true)
		if !govcd.IsNotFound(err) {
			return fmt.Errorf("independent disk %s still exist and error: %#v", itemName, err)
		}

	}
	return nil
}

func importStateIdByDisk(resource string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return "", fmt.Errorf("not found resource: %s", resource)
		}

		if rs.Primary.ID == "" {
			return "", fmt.Errorf("no ID is set for %s resource", resource)
		}

		importId := testConfig.VCD.Org + "." + testConfig.VCD.Vdc + "." + rs.Primary.ID
		if testConfig.VCD.Org == "" || testConfig.VCD.Vdc == "" || rs.Primary.ID == "" {
			return "", fmt.Errorf("missing information to generate import path: %s", importId)
		}
		return importId, nil
	}
}

func init() {
	testingTags["disk"] = "resource_vcd_independent_disk_test.go"
}

const testAccCheckVcdIndependentDiskForCompatibility = `
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

/*const testAccCheckVcdIndependentDiskBasic = `
resource "vcd_independent_disk" "{{.secondResourceName}}" {
  org             = "{{.Org}}"
  vdc             = "{{.Vdc}}"
  name            = "{{.secondName}}"
  size_in_bytes   = "{{.sizeInBytes}}"
  bus_type        = "{{.busType}}"
  bus_sub_type    = "{{.busSubType}}"
  storage_profile = "{{.storageProfileName}}"
}
`
*/
