// +build vdc ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
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
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"Tags":                      "vdc",
		"FuncName":                  "TestAccVcdOrgVdcReservationPool",
		// cause vDC ignores values
		"MemoryGuaranteed": "1.00",
		"CpuGuaranteed":    "1.00",
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
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"Tags":                      "vdc",
		"FuncName":                  "TestAccVcdOrgVdcAllocationPool",
		"MemoryGuaranteed":          "0.30",
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
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"Tags":                      "vdc",
		"FuncName":                  "TestAccVcdOrgVdcAllocationVapp",
		"MemoryGuaranteed":          "0.50",
		"CpuGuaranteed":             "0.65",
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

	var vdc govcd.Vdc

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
					testAccCheckVcdVdcExists("vcd_org_vdc."+TestAccVcdVdc, &vdc),
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
				),
			},
			resource.TestStep{
				Config: updateText,
				Check: resource.ComposeTestCheckFunc(
					testVcdVdcUpdated("vcd_org_vdc."+TestAccVcdVdc, &vdc),
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
				),
			},
		},
	})
}

func testAccCheckVcdVdcExists(name string, vdc *govcd.Vdc) resource.TestCheckFunc {
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

		newVdc, err := adminOrg.GetVdcByName(rs.Primary.Attributes["name"])
		if err != nil {
			return fmt.Errorf("vdc %s does not exist (%#v)", rs.Primary.Attributes["name"], newVdc)
		}

		vdc = &newVdc
		return nil
	}
}

func testVcdVdcUpdated(name string, vdc *govcd.Vdc) resource.TestCheckFunc {
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

		updateVdc, err := adminOrg.GetVdcByName(rs.Primary.Attributes["name"])
		if err != nil {
			return fmt.Errorf("vdc %s does not exist (%#v)", rs.Primary.Attributes["name"], updateVdc)
		}

		if updateVdc.Vdc.IsEnabled != false {
			return fmt.Errorf("VDC update failed - VDC still enabled")
		}

		vdc = &updateVdc
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

		vdc, err := adminOrg.GetVdcByName(rs.Primary.ID)

		if vdc != (govcd.Vdc{}) {
			return fmt.Errorf("vdc %s still exists", rs.Primary.ID)
		}
		if err != nil {
			return fmt.Errorf("vdc %s still exists or other error: %#v", rs.Primary.ID, err)
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
      allocated = 2048
      limit     = 2048
    }

    memory {
      allocated = 2048
      limit     = 2048
    }
  }

  storage_profile {
    name = "{{.ProviderVdcStorageProfile}}"
    enabled  = true
    limit    = 10240
    default  = true
  }

  enabled                  = true
  enable_thin_provisioning = true
  enable_fast_provisioning = true
  delete_force             = true
  delete_recursive         = true
}
`

const testAccCheckVcdVdc_update = `
resource "vcd_org_vdc" "{{.VdcName}}" {
  name = "{{.VdcName}}"
  org  = "{{.OrgName}}"

  allocation_model = "{{.AllocationModel}}"
  network_pool_name     = "{{.NetworkPool}}"
  provider_vdc_name     = "{{.ProviderVdc}}"

  compute_capacity {
    cpu {
      allocated = 2048
      limit     = 2048
    }

    memory {
      allocated = 2048
      limit     = 2048
    }
  }

  storage_profile {
    name = "{{.ProviderVdcStorageProfile}}"
    enabled  = true
    limit    = 10240
    default  = true
  }

  cpu_guaranteed           = "{{.CpuGuaranteed}}"
  memory_guaranteed        = "{{.MemoryGuaranteed}}"
  enabled                  = false
  enable_thin_provisioning = false
  enable_fast_provisioning = false
  delete_force             = false
  delete_recursive         = false
}
`
