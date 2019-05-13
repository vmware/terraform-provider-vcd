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

func TestAccVcdVdcBasic(t *testing.T) {

	if !usingSysAdmin() {
		t.Skip("TestAccVcdVdcBasic requires system admin privileges")
		return
	}

	if testConfig.VCD.ProviderVdc.Name == "" {
		t.Skip("Variable networking.Name must be set to run VDC tests")
		return
	}

	if testConfig.VCD.ProviderVdc.NetworkPool == "" {
		t.Skip("Variable networking.NetworkPool must be set to run VDC tests")
		return
	}

	if testConfig.VCD.ProviderVdc.StorageProfile == "" {
		t.Skip("Variable networking.StorageProfile must be set to run VDC tests")
		return
	}

	var vdc govcd.Vdc
	var params = StringMap{
		"VdcName":                   TestAccVcdVdc,
		"OrgName":                   testConfig.VCD.Org,
		"AllocationModel":           "ReservationPool",
		"ProviderVdc":               testConfig.VCD.ProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.ProviderVdc.NetworkPool,
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
	}

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	if !usingSysAdmin() {
		t.Skip("TestAccVcdVdcBasic requires system admin privileges")
		return
	}

	configText := templateFill(testAccCheckVcdVdc_basic, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVdcDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVdcExists("vcd_vdc."+TestAccVcdVdc, &vdc),
					resource.TestCheckResourceAttr(
						"vcd_vdc."+TestAccVcdVdc, "name", TestAccVcdVdc),
					resource.TestCheckResourceAttr(
						"vcd_vdc."+TestAccVcdVdc, "org", testConfig.VCD.Org),
					resource.TestCheckResourceAttr(
						"vcd_vdc."+TestAccVcdVdc, "allocation_model", "ReservationPool"),
					resource.TestCheckResourceAttr(
						"vcd_vdc."+TestAccVcdVdc, "network_pool_name", testConfig.VCD.ProviderVdc.NetworkPool),
					resource.TestCheckResourceAttr(
						"vcd_vdc."+TestAccVcdVdc, "provider_vdc_name", testConfig.VCD.ProviderVdc.Name),
					resource.TestCheckResourceAttr(
						"vcd_vdc."+TestAccVcdVdc, "is_enabled", "true"),
					resource.TestCheckResourceAttr(
						"vcd_vdc."+TestAccVcdVdc, "enable_thin_provisioning", "true"),
					resource.TestCheckResourceAttr(
						"vcd_vdc."+TestAccVcdVdc, "uses_fast_provisioning", "true"),
					resource.TestCheckResourceAttr(
						"vcd_vdc."+TestAccVcdVdc, "delete_force", "true"),
					resource.TestCheckResourceAttr(
						"vcd_vdc."+TestAccVcdVdc, "delete_recursive", "true"),
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

		newVdc, err := adminOrg.GetVdcByName(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("vdc %s does not exist (%#v)", rs.Primary.ID, newVdc)
		}

		vdc = &newVdc
		return nil
	}
}

func testAccCheckVdcDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_vdc" && rs.Primary.Attributes["name"] != TestAccVcdVdc {
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
	testingTags["vdc"] = "resource_vcd_vdc_test.go"
}

const testAccCheckVcdVdc_basic = `
resource "vcd_vdc" "{{.VdcName}}" {
  name = "{{.VdcName}}"
  org  = "{{.OrgName}}"

  allocation_model = "{{.AllocationModel}}"
  network_pool_name     = "{{.NetworkPool}}"
  provider_vdc_name     = "{{.ProviderVdc}}"

  compute_capacity {
    cpu {
      units     = "MHz"
      allocated = 2048
      limit     = 2048
      reserved  = 2048
      used      = 0
      overhead  = 0
    }

    memory {
      units     = "MB"
      allocated = 2048
      limit     = 2048
      reserved  = 2048
      used      = 0
      overhead  = 0
    }
  }

  storage_profile {
    name = "{{.ProviderVdcStorageProfile}}"
    enabled  = true
    limit    = 10240
    default  = true
  }

  is_enabled               = true
  enable_thin_provisioning = true
  uses_fast_provisioning   = true
  delete_force             = true
  delete_recursive         = true
}
`
