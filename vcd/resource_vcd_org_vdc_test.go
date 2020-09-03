// +build vdc ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var TestAccVcdVdc = "TestAccVcdVdcBasic"

func TestAccVcdOrgVdcReservationPool(t *testing.T) {
	if !usingSysAdmin() {
		t.Skip("TestAccVcdVdcBasic requires system admin privileges")
	}
	validateConfiguration(t)

	allocationModel := "ReservationPool"

	var params = StringMap{
		"VdcName":                   TestAccVcdVdc,
		"OrgName":                   testConfig.VCD.Org,
		"AllocationModel":           "ReservationPool",
		"ProviderVdc":               testConfig.VCD.ProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.ProviderVdc.NetworkPool,
		"Allocated":                 "1024",
		"Reserved":                  "1024",
		"Limit":                     "1024",
		"LimitIncreased":            "1100",
		"AllocatedIncreased":        "1100",
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"Tags":                      "vdc",
		"FuncName":                  "TestAccVcdOrgVdcReservationPool",
		// cause vDC ignores empty values and use default
		"MemoryGuaranteed": "1",
		"CpuGuaranteed":    "1",
		// The parameters below are for Flex allocation model
		// Part of HCL is created dynamically and with empty values we don't create the Flex part:
		"equalsChar":                         "",
		"FlexElasticKey":                     "",
		"FlexElasticValue":                   "",
		"FlexElasticValueUpdate":             "",
		"FlexMemoryOverheadKey":              "",
		"FlexMemoryOverheadValue":            "",
		"FlexMemoryOverheadValueUpdate":      "",
		"MemoryOverheadValueForAssert":       "true",
		"MemoryOverheadUpdateValueForAssert": "true",
		"ElasticityValueForAssert":           "false",
		"ElasticityUpdateValueForAssert":     "false",
	}

	runOrgVdcTest(t, params, allocationModel)
}

func TestAccVcdOrgVdcAllocationPool(t *testing.T) {
	if !usingSysAdmin() {
		t.Skip("TestAccVcdVdcBasic requires system admin privileges")
	}
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
		"Limit":                     "2048",
		"LimitIncreased":            "2148",
		"AllocatedIncreased":        "2148",
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"Tags":                      "vdc",
		"FuncName":                  "TestAccVcdOrgVdcAllocationPool",
		"MemoryGuaranteed":          "0.3",
		"CpuGuaranteed":             "0.45",
		// The parameters below are for Flex allocation model
		// Part of HCL is created dynamically and with empty values we don't create the Flex part:
		"equalsChar":                         "",
		"FlexElasticKey":                     "",
		"FlexElasticValue":                   "",
		"FlexElasticValueUpdate":             "",
		"ElasticityValueForAssert":           "false",
		"ElasticityUpdateValueForAssert":     "false",
		"FlexMemoryOverheadKey":              "",
		"FlexMemoryOverheadValue":            "",
		"FlexMemoryOverheadValueUpdate":      "",
		"MemoryOverheadValueForAssert":       "true",
		"MemoryOverheadUpdateValueForAssert": "true",
	}

	runOrgVdcTest(t, params, allocationModel)
}

func TestAccVcdOrgVdcAllocationVApp(t *testing.T) {
	if !usingSysAdmin() {
		t.Skip("TestAccVcdVdcBasic requires system admin privileges")
	}
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
		"Limit":                     "2048",
		"LimitIncreased":            "2148",
		"AllocatedIncreased":        "0",
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"Tags":                      "vdc",
		"FuncName":                  "TestAccVcdOrgVdcAllocationVapp",
		"MemoryGuaranteed":          "0.5",
		"CpuGuaranteed":             "0.6",
		// The parameters below are for Flex allocation model
		// Part of HCL is created dynamically and with empty values we don't create the Flex part:
		"equalsChar":                         "",
		"FlexElasticKey":                     "",
		"FlexElasticValue":                   "",
		"FlexElasticValueUpdate":             "",
		"ElasticityValueForAssert":           "true",
		"ElasticityUpdateValueForAssert":     "true",
		"FlexMemoryOverheadKey":              "",
		"FlexMemoryOverheadValue":            "",
		"FlexMemoryOverheadValueUpdate":      "",
		"MemoryOverheadValueForAssert":       "false",
		"MemoryOverheadUpdateValueForAssert": "false",
	}

	runOrgVdcTest(t, params, allocationModel)
}

func TestAccVcdOrgVdcAllocationFlex(t *testing.T) {
	if !usingSysAdmin() {
		t.Skip("TestAccVcdVdcBasic requires system admin privileges")
	}

	validateConfiguration(t)

	allocationModel := "Flex"

	var params = StringMap{
		"VdcName":                   TestAccVcdVdc,
		"OrgName":                   testConfig.VCD.Org,
		"AllocationModel":           allocationModel,
		"ProviderVdc":               testConfig.VCD.ProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.ProviderVdc.NetworkPool,
		"Allocated":                 "1024",
		"Reserved":                  "0",
		"Limit":                     "1024",
		"LimitIncreased":            "1124",
		"AllocatedIncreased":        "1124",
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"Tags":                      "vdc",
		"FuncName":                  t.Name(),
		"MemoryGuaranteed":          "0.5",
		"CpuGuaranteed":             "0.6",
		// The parameters below are for Flex allocation model
		// Part of HCL is created dynamically and these parameters with values result in the Flex part of the template being filled:
		"equalsChar":                         "=",
		"FlexElasticKey":                     "elasticity",
		"FlexElasticValue":                   "false",
		"FlexElasticValueUpdate":             "true",
		"ElasticityValueForAssert":           "false",
		"ElasticityUpdateValueForAssert":     "true",
		"FlexMemoryOverheadKey":              "include_vm_memory_overhead",
		"FlexMemoryOverheadValue":            "false",
		"FlexMemoryOverheadValueUpdate":      "true",
		"MemoryOverheadValueForAssert":       "false",
		"MemoryOverheadUpdateValueForAssert": "true",
	}
	runOrgVdcTest(t, params, allocationModel)
}

func validateConfiguration(t *testing.T) {
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
	debugPrintf("#[DEBUG] CONFIGURATION: %s", updateText)

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
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "storage_profile.0.enabled", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "storage_profile.0.limit", "10240"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "storage_profile.0.default", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.cpu.0.allocated", params["Allocated"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.cpu.0.limit", params["Limit"].(string)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.cpu.0.reserved", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.cpu.0.used", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.memory.0.allocated", params["Allocated"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.memory.0.limit", params["Limit"].(string)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.memory.0.reserved", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.memory.0.used", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "elasticity", regexp.MustCompile(`^`+params["ElasticityValueForAssert"].(string)+`$`)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "include_vm_memory_overhead", regexp.MustCompile(`^`+params["MemoryOverheadValueForAssert"].(string)+`$`)),
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
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "storage_profile.0.enabled", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "storage_profile.0.limit", "10240"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "storage_profile.0.default", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.cpu.0.allocated", params["AllocatedIncreased"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.cpu.0.limit", params["LimitIncreased"].(string)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.cpu.0.reserved", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.cpu.0.used", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.memory.0.allocated", params["AllocatedIncreased"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.memory.0.limit", params["LimitIncreased"].(string)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.memory.0.reserved", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.memory.0.used", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "elasticity", regexp.MustCompile(`^`+params["ElasticityUpdateValueForAssert"].(string)+`$`)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "include_vm_memory_overhead", regexp.MustCompile(`^`+params["MemoryOverheadUpdateValueForAssert"].(string)+`$`)),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_org_vdc." + TestAccVcdVdc,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgObject(testConfig, TestAccVcdVdc),
				// These fields can't be retrieved
				ImportStateVerifyIgnore: []string{"delete_force", "delete_recursive"},
			},
		},
	})
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

  allocation_model  = "{{.AllocationModel}}"
  network_pool_name = "{{.NetworkPool}}"
  provider_vdc_name = "{{.ProviderVdc}}"

  compute_capacity {
    cpu {
      allocated = "{{.Allocated}}"
      limit     = "{{.Limit}}"
    }

    memory {
      allocated = "{{.Allocated}}"
      limit     = "{{.Limit}}"
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

  enabled                    = true
  enable_thin_provisioning   = true
  enable_fast_provisioning   = true
  delete_force               = true
  delete_recursive           = true
  {{.FlexElasticKey}}                 {{.equalsChar}} {{.FlexElasticValue}}
  {{.FlexMemoryOverheadKey}} {{.equalsChar}} {{.FlexMemoryOverheadValue}}
}
`

const testAccCheckVcdVdc_update = `
# skip-binary-test: only for updates
resource "vcd_org_vdc" "{{.VdcName}}" {
  name = "{{.VdcName}}"
  org  = "{{.OrgName}}"

  allocation_model  = "{{.AllocationModel}}"
  network_pool_name = "{{.NetworkPool}}"
  provider_vdc_name = "{{.ProviderVdc}}"

  compute_capacity {
    cpu {
      allocated = "{{.AllocatedIncreased}}"
      limit     = "{{.LimitIncreased}}"
    }

    memory {
      allocated = "{{.AllocatedIncreased}}"
      limit     = "{{.LimitIncreased}}"
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

  cpu_guaranteed             = {{.CpuGuaranteed}}
  memory_guaranteed          = {{.MemoryGuaranteed}}
  enabled                    = false
  enable_thin_provisioning   = false
  enable_fast_provisioning   = false
  delete_force               = false
  delete_recursive           = false
  {{.FlexElasticKey}}                 {{.equalsChar}} {{.FlexElasticValueUpdate}}
  {{.FlexMemoryOverheadKey}} {{.equalsChar}} {{.FlexMemoryOverheadValueUpdate}}
}
`
