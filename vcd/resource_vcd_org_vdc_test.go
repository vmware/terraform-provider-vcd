// +build vdc ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

var TestAccVcdVdc = "TestAccVcdVdcBasic"

func TestAccVcdOrgVdcReservationPool(t *testing.T) {

	validateConfiguration(t)

	allocationModel := "ReservationPool"

	var params = StringMap{
		"VdcName":                   TestAccVcdVdc,
		"OrgName":                   testConfig.VCD.Org,
		"AllocationModel":           "ReservationPool",
		"ProviderVdc":               testConfig.VCD.ProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.ProviderVdc.NetworkPool,
		"Allocated":                 "2048",
		"Reserved":                  "2048",
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"Tags":                      "vdc",
		"FuncName":                  "TestAccVcdOrgVdcReservationPool",
		// cause vDC ignores empty values and use default
		"MemoryGuaranteed": "1",
		"CpuGuaranteed":    "1",
	}
	runOrgVdcTest(t, params, allocationModel)
}

func TestAccVcdOrgVdcAllocationPool(t *testing.T) {

	validateConfiguration(t)
	allocationModel := "AllocationPool"

	var params = StringMap{
		"VdcName":                   TestAccVcdVdc,
		"OrgName":                   testConfig.VCD.Org,
		"AllocationModel":           "AllocationPool",
		"ProviderVdc":               testConfig.VCD.ProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.ProviderVdc.NetworkPool,
		"Allocated":                 "2048",
		"Reserved":                  "1024",
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"Tags":                      "vdc",
		"FuncName":                  "TestAccVcdOrgVdcAllocationPool",
		"MemoryGuaranteed":          "0.3",
		"CpuGuaranteed":             "0.45",
	}
	runOrgVdcTest(t, params, allocationModel)
}

func TestAccVcdOrgVdcAllocationVApp(t *testing.T) {

	validateConfiguration(t)

	allocationModel := "AllocationVApp"

	var params = StringMap{
		"VdcName":                   TestAccVcdVdc,
		"OrgName":                   testConfig.VCD.Org,
		"AllocationModel":           allocationModel,
		"ProviderVdc":               testConfig.VCD.ProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.ProviderVdc.NetworkPool,
		"Allocated":                 "0",
		"Reserved":                  "0",
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"Tags":                      "vdc",
		"FuncName":                  "TestAccVcdOrgVdcAllocationVapp",
		"MemoryGuaranteed":          "0.5",
		"CpuGuaranteed":             "0.6",
	}
	runOrgVdcTest(t, params, allocationModel)
}

func validateConfiguration(t *testing.T) {
	if !usingSysAdmin() {
		t.Skip("TestAccVcdVdcBasic requires system admin privileges")
	}

	if testConfig.VCD.ProviderVdc.Name == "" {
		t.Skip("Variable providerVdc.Name must be set to run VDC tests")
	}

	if testConfig.VCD.ProviderVdc.NetworkPool == "" {
		t.Skip("Variable providerVdc.NetworkPool must be set to run VDC tests")
	}

	if testConfig.VCD.ProviderVdc.StorageProfile == "" {
		t.Skip("Variable providerVdc.StorageProfile must be set to run VDC tests")
	}

}

func runOrgVdcTest(t *testing.T, params StringMap, allocationModel string) {

	if !usingSysAdmin() {
		t.Skip("TestAccVcdVdcBasic requires system admin privileges")
		return
	}

	configText := templateFill(testAccCheckVcdVdc_basic, params)
	params["FuncName"] = t.Name() + "-Update"
	updateText := templateFill(testAccCheckVcdVdc_update, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVdcDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVdcExists("vcd_org_vdc."+TestAccVcdVdc),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "name", TestAccVcdVdc),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "org", testConfig.VCD.Org),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "allocation_model", allocationModel),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "network_pool_name", testConfig.VCD.ProviderVdc.NetworkPool),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "provider_vdc_name", testConfig.VCD.ProviderVdc.Name),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "enabled", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "enable_thin_provisioning", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "enable_fast_provisioning", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "delete_force", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "delete_recursive", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "metadata.vdc_metadata", "VDC Metadata"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "storage_profile.0.name", testConfig.VCD.ProviderVdc.StorageProfile),
					resource.ComposeTestCheckFunc(testAccSourceVcdOrgVdc(TestAccVcdVdc, params["Allocated"].(string), params["Reserved"].(string))),
				),
			},
			resource.TestStep{
				Config: updateText,
				Check: resource.ComposeTestCheckFunc(
					testVcdVdcUpdated("vcd_org_vdc."+TestAccVcdVdc),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "name", TestAccVcdVdc),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "org", testConfig.VCD.Org),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "allocation_model", allocationModel),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "network_pool_name", testConfig.VCD.ProviderVdc.NetworkPool),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "provider_vdc_name", testConfig.VCD.ProviderVdc.Name),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "enabled", "false"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "enable_thin_provisioning", "false"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "enable_fast_provisioning", "false"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "delete_force", "false"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "delete_recursive", "false"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "memory_guaranteed", params["MemoryGuaranteed"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "cpu_guaranteed", params["CpuGuaranteed"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "metadata.vdc_metadata", "VDC Metadata"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "metadata.vdc_metadata2", "VDC Metadata2"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "metadata.vdc_metadata2", "VDC Metadata2"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "storage_profile.0.name", testConfig.VCD.ProviderVdc.StorageProfile),
					resource.ComposeTestCheckFunc(testAccSourceVcdOrgVdc(TestAccVcdVdc, params["Allocated"].(string), params["Reserved"].(string))),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_org_vdc." + TestAccVcdVdc,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdByVdc(TestAccVcdVdc),
				// These fields can't be retrieved
				ImportStateVerifyIgnore: []string{"delete_force", "delete_recursive"},
			},
		},
	})
}

func importStateIdByVdc(objectName string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		importId := testConfig.VCD.Org + "." + objectName
		if testConfig.VCD.Org == "" || objectName == "" {
			return "", fmt.Errorf("missing information to generate import path: %s", importId)
		}
		return importId, nil
	}
}

func testAccSourceVcdOrgVdc(vdcName, allocatedValue, reserveValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vdcResource, ok := s.RootModule().Resources["vcd_org_vdc."+vdcName]
		if !ok {
			return fmt.Errorf("can't find vcd_org_vdc.%s in state", vdcName)
		}

		checkDeepValues := func(expected, parent, parentHash, child, childHash, label string, result *string) bool {

			value := vdcResource.Primary.Attributes[fmt.Sprintf("%s.%s.%s.%s.%s", parent, parentHash, child, childHash, label)]

			if expected != value {
				*result = fmt.Sprintf("%s.%s.%s.%s.%s is %v; want %v", parent, parentHash, child, childHash, label, value, expected)
				return false
			}
			return true
		}

		// Resource uses Set types and to access them in state file when key is `flatten` you have to know hash keys. In this resource we use default hash generation and in
		// test it is quite difficult to recreate real hash, so we regex them from map and use them to access values.
		mainHashValue, cpuHashInternalValue, err := getHashValuesFromKey(vdcResource.Primary.Attributes, "compute_capacity", "cpu")
		if err != nil {
			return err
		}
		_, memoryHashInternalValue, err := getHashValuesFromKey(vdcResource.Primary.Attributes, "compute_capacity", "memory")
		if err != nil {
			return err
		}

		type testInfo struct {
			child         string
			childHash     string
			label         string
			expectedValue string
		}
		var errorMsg string

		var testData = []testInfo{
			{"cpu", cpuHashInternalValue, "allocated", allocatedValue},
			{"cpu", cpuHashInternalValue, "limit", "2048"},
			{"cpu", cpuHashInternalValue, "overhead", "0"},
			{"cpu", cpuHashInternalValue, "reserved", reserveValue},
			{"cpu", cpuHashInternalValue, "used", "0"},
			{"memory", memoryHashInternalValue, "allocated", allocatedValue},
			{"memory", memoryHashInternalValue, "limit", "2048"},
			{"memory", memoryHashInternalValue, "overhead", "0"},
			{"memory", memoryHashInternalValue, "reserved", reserveValue},
			{"memory", memoryHashInternalValue, "used", "0"},
		}
		for _, td := range testData {
			if !checkDeepValues(td.expectedValue, "compute_capacity", mainHashValue, td.child, td.childHash, td.label, &errorMsg) {
				return fmt.Errorf("%s", errorMsg)
			}
		}

		return nil
	}
}

func testAccCheckVcdVdcExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no VDC ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		adminOrg, err := conn.GetAdminOrg(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf(errorRetrievingOrg, testConfig.VCD.Org+" and error: "+err.Error())
		}

		_, err = adminOrg.GetVDCByName(rs.Primary.Attributes["name"], false)
		if err != nil {
			return fmt.Errorf("vdc %s does not exist (%s)", rs.Primary.Attributes["name"], err)
		}

		return nil
	}
}

func testVcdVdcUpdated(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no VDC ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		adminOrg, err := conn.GetAdminOrg(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf(errorRetrievingOrg, testConfig.VCD.Org+" and error: "+err.Error())
		}

		updateVdc, err := adminOrg.GetVDCByName(rs.Primary.Attributes["name"], false)
		if err != nil {
			return fmt.Errorf("vdc %s does not exist (%s)", rs.Primary.Attributes["name"], err)
		}

		if updateVdc.Vdc.IsEnabled != false {
			return fmt.Errorf("VDC update failed - VDC still enabled")
		}
		return nil
	}
}

func testAccCheckVdcDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_org_vdc" && rs.Primary.Attributes["name"] != TestAccVcdVdc {
			continue
		}

		adminOrg, err := conn.GetAdminOrg(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf(errorRetrievingOrg, testConfig.VCD.Org+" and error: "+err.Error())
		}

		_, err = adminOrg.GetVDCByName(rs.Primary.ID, false)

		if err == nil {
			return fmt.Errorf("vdc %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func init() {
	testingTags["vdc"] = "resource_vcd_org_vdc_test.go"
}

const testAccCheckVcdVdc_basic = `
resource "vcd_org_vdc" "{{.VdcName}}" {
  name = "{{.VdcName}}"
  org  = "{{.OrgName}}"

  allocation_model = "{{.AllocationModel}}"
  network_pool_name     = "{{.NetworkPool}}"
  provider_vdc_name     = "{{.ProviderVdc}}"

  compute_capacity {
    cpu {
      allocated = "{{.Allocated}}"
      limit     = 2048
    }

    memory {
      allocated = "{{.Allocated}}"
      limit     = 2048
    }
  }

  storage_profile {
    name = "{{.ProviderVdcStorageProfile}}"
    enabled  = true
    limit    = 10240
    default  = true
  }

  metadata = {
    vdc_metadata = "VDC Metadata"
  }

  enabled                  = true
  enable_thin_provisioning = true
  enable_fast_provisioning = true
  delete_force             = true
  delete_recursive         = true
}
`

const testAccCheckVcdVdc_update = `
# skip-binary-test: only for updates
resource "vcd_org_vdc" "{{.VdcName}}" {
  name = "{{.VdcName}}"
  org  = "{{.OrgName}}"

  allocation_model = "{{.AllocationModel}}"
  network_pool_name     = "{{.NetworkPool}}"
  provider_vdc_name     = "{{.ProviderVdc}}"

  compute_capacity {
    cpu {
      allocated = "{{.Allocated}}"
      limit     = 2048
    }

    memory {
      allocated = "{{.Allocated}}"
      limit     = 2048
    }
  }

  storage_profile {
    name = "{{.ProviderVdcStorageProfile}}"
    enabled  = true
    limit    = 10240
    default  = true
  }

  metadata = {
    vdc_metadata = "VDC Metadata"
    vdc_metadata2 = "VDC Metadata2"
  }

  cpu_guaranteed           = {{.CpuGuaranteed}}
  memory_guaranteed        = {{.MemoryGuaranteed}}
  enabled                  = false
  enable_thin_provisioning = false
  enable_fast_provisioning = false
  delete_force             = false
  delete_recursive         = false
}
`
