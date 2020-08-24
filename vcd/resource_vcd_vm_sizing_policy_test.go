// +build vdc ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var TestVmPolicy = "TestVmPolicyBasic"

func TestAccVcdVmSizingPolicy(t *testing.T) {
	if !usingSysAdmin() {
		t.Skip("TestAccVcdVmSizingPolicy requires system admin privileges")
	}

	if testConfig.VCD.ProviderVdc.Name == "" {
		t.Skip("Variable providerVdc.Name must be set to run VDC tests")
	}

	var params = StringMap{
		"OrgName":     testConfig.VCD.Org,
		"PolicyName":  "TestAccVcdVmSizingPolicy",
		"Description": "TestAccVcdVmSizingPolicyDescription",

		"CpuShare":       "886",
		"CpuLimit":       "2400",
		"CpuCount":       "9",
		"CpuSpeed":       "2500",
		"CoresPerSocket": "3",
		"CpuReservation": "0.55",

		"MemoryShare":       "1580",
		"MemorySize":        "3200",
		"MemoryLimit":       "2800",
		"MemoryReservation": "0.3",
	}

	configText := templateFill(testAccCheckVmSizingPolicy_basic, params)
	params["FuncName"] = t.Name() + "-Update"
	updateText := templateFill(testAccCheckVmSizingPolicy_update, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", updateText)

	resource1 := "vcd_vm_sizing_policy." + params["PolicyName"].(string) + "_1"
	resource2 := "vcd_vm_sizing_policy." + params["PolicyName"].(string) + "_2"
	resource3 := "vcd_vm_sizing_policy." + params["PolicyName"].(string) + "_3"
	resource4 := "vcd_vm_sizing_policy." + params["PolicyName"].(string) + "_4"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVmSizingPolicyDestroyed,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVmSizingPolicyExists(resource1),
					resource.TestCheckResourceAttr(resource1, "org", testConfig.VCD.Org),
					resource.TestCheckResourceAttr(resource1, "name", params["PolicyName"].(string)+"_1"),
					resource.TestCheckResourceAttr(resource1, "description", params["Description"].(string)+"_1"),

					testAccCheckVmSizingPolicyExists(resource2),
					resource.TestCheckResourceAttr(resource2, "org", testConfig.VCD.Org),
					resource.TestCheckResourceAttr(resource2, "name", params["PolicyName"].(string)+"_2"),
					resource.TestCheckResourceAttr(resource2, "description", params["Description"].(string)+"_2"),

					resource.TestCheckResourceAttr(resource2, "cpu.0.shares", params["CpuShare"].(string)),
					resource.TestCheckResourceAttr(resource2, "cpu.0.limit_in_mhz", params["CpuLimit"].(string)),
					resource.TestCheckResourceAttr(resource2, "cpu.0.count", params["CpuCount"].(string)),
					resource.TestCheckResourceAttr(resource2, "cpu.0.speed_in_mhz", params["CpuSpeed"].(string)),
					resource.TestCheckResourceAttr(resource2, "cpu.0.cores_per_socket", params["CoresPerSocket"].(string)),
					resource.TestCheckResourceAttr(resource2, "cpu.0.reservation_guarantee", params["CpuReservation"].(string)),

					testAccCheckVmSizingPolicyExists(resource3),
					resource.TestCheckResourceAttr(resource3, "org", testConfig.VCD.Org),
					resource.TestCheckResourceAttr(resource3, "name", params["PolicyName"].(string)+"_3"),
					resource.TestCheckResourceAttr(resource3, "description", params["Description"].(string)+"_3"),

					resource.TestCheckResourceAttr(resource3, "memory.0.shares", params["MemoryShare"].(string)),
					resource.TestCheckResourceAttr(resource3, "memory.0.size_in_mb", params["MemorySize"].(string)),
					resource.TestCheckResourceAttr(resource3, "memory.0.limit_in_mb", params["MemoryLimit"].(string)),
					resource.TestCheckResourceAttr(resource3, "memory.0.reservation_guarantee", params["MemoryReservation"].(string)),

					testAccCheckVmSizingPolicyExists(resource4),
					resource.TestCheckResourceAttr(resource4, "org", testConfig.VCD.Org),
					resource.TestCheckResourceAttr(resource4, "name", params["PolicyName"].(string)+"_4"),
					resource.TestCheckResourceAttr(resource4, "description", params["Description"].(string)+"_4"),

					resource.TestCheckResourceAttr(resource4, "cpu.0.shares", params["CpuShare"].(string)),
					resource.TestCheckResourceAttr(resource4, "cpu.0.limit_in_mhz", params["CpuLimit"].(string)),
					resource.TestCheckResourceAttr(resource4, "cpu.0.count", params["CpuCount"].(string)),
					resource.TestCheckResourceAttr(resource4, "cpu.0.speed_in_mhz", params["CpuSpeed"].(string)),
					resource.TestCheckResourceAttr(resource4, "cpu.0.cores_per_socket", params["CoresPerSocket"].(string)),
					resource.TestCheckResourceAttr(resource4, "cpu.0.reservation_guarantee", params["CpuReservation"].(string)),

					resource.TestCheckResourceAttr(resource4, "memory.0.shares", params["MemoryShare"].(string)),
					resource.TestCheckResourceAttr(resource4, "memory.0.size_in_mb", params["MemorySize"].(string)),
					resource.TestCheckResourceAttr(resource4, "memory.0.limit_in_mb", params["MemoryLimit"].(string)),
					resource.TestCheckResourceAttr(resource4, "memory.0.reservation_guarantee", params["MemoryReservation"].(string)),
				),
			},
			resource.TestStep{
				Config: updateText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVmSizingPolicyExists(resource1),
					resource.TestCheckResourceAttr(resource1, "org", testConfig.VCD.Org),
					resource.TestCheckResourceAttr(resource1, "name", params["PolicyName"].(string)+"_1"),
					resource.TestCheckResourceAttr(resource1, "description", params["Description"].(string)+"_1"),

					testAccCheckVmSizingPolicyExists(resource2),
					resource.TestCheckResourceAttr(resource2, "org", testConfig.VCD.Org),
					resource.TestCheckResourceAttr(resource2, "name", params["PolicyName"].(string)+"_2"),
					resource.TestCheckResourceAttr(resource2, "description", params["Description"].(string)+"_2"),

					resource.TestCheckResourceAttr(resource2, "cpu.0.shares", params["CpuShare"].(string)),
					resource.TestCheckResourceAttr(resource2, "cpu.0.limit_in_mhz", params["CpuLimit"].(string)),
					resource.TestCheckResourceAttr(resource2, "cpu.0.count", params["CpuCount"].(string)),
					resource.TestCheckResourceAttr(resource2, "cpu.0.speed_in_mhz", params["CpuSpeed"].(string)),
					resource.TestCheckResourceAttr(resource2, "cpu.0.cores_per_socket", params["CoresPerSocket"].(string)),
					resource.TestCheckResourceAttr(resource2, "cpu.0.reservation_guarantee", params["CpuReservation"].(string)),

					testAccCheckVmSizingPolicyExists(resource3),
					resource.TestCheckResourceAttr(resource3, "org", testConfig.VCD.Org),
					resource.TestCheckResourceAttr(resource3, "name", params["PolicyName"].(string)+"_3"),
					resource.TestCheckResourceAttr(resource3, "description", params["Description"].(string)+"_3"),

					resource.TestCheckResourceAttr(resource3, "memory.0.shares", params["MemoryShare"].(string)),
					resource.TestCheckResourceAttr(resource3, "memory.0.size_in_mb", params["MemorySize"].(string)),
					resource.TestCheckResourceAttr(resource3, "memory.0.limit_in_mb", params["MemoryLimit"].(string)),
					resource.TestCheckResourceAttr(resource3, "memory.0.reservation_guarantee", params["MemoryReservation"].(string)),

					testAccCheckVmSizingPolicyExists(resource4),
					resource.TestCheckResourceAttr(resource4, "org", testConfig.VCD.Org),
					resource.TestCheckResourceAttr(resource4, "name", params["PolicyName"].(string)+"_updated"),
					resource.TestCheckResourceAttr(resource4, "description", params["Description"].(string)+"_updated"),

					resource.TestCheckResourceAttr(resource4, "cpu.0.shares", params["CpuShare"].(string)),
					resource.TestCheckResourceAttr(resource4, "cpu.0.limit_in_mhz", params["CpuLimit"].(string)),
					resource.TestCheckResourceAttr(resource4, "cpu.0.count", params["CpuCount"].(string)),
					resource.TestCheckResourceAttr(resource4, "cpu.0.speed_in_mhz", params["CpuSpeed"].(string)),
					resource.TestCheckResourceAttr(resource4, "cpu.0.cores_per_socket", params["CoresPerSocket"].(string)),
					resource.TestCheckResourceAttr(resource4, "cpu.0.reservation_guarantee", params["CpuReservation"].(string)),

					resource.TestCheckResourceAttr(resource4, "memory.0.shares", params["MemoryShare"].(string)),
					resource.TestCheckResourceAttr(resource4, "memory.0.size_in_mb", params["MemorySize"].(string)),
					resource.TestCheckResourceAttr(resource4, "memory.0.limit_in_mb", params["MemoryLimit"].(string)),
					resource.TestCheckResourceAttr(resource4, "memory.0.reservation_guarantee", params["MemoryReservation"].(string)),
				),
			},
			resource.TestStep{
				ResourceName:            resource4,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateVmSizingPolicyById(testConfig, resource4),
				ImportStateVerifyIgnore: []string{"org"},
			},
		},
	})
}

func importStateVmSizingPolicyById(testConfig TestConfig, resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}
		if testConfig.VCD.Org == "" || rs.Primary.ID == "" {
			return "", fmt.Errorf("missing information to generate import path %s, %s", testConfig.VCD.Org, rs.Primary.ID)
		}
		return testConfig.VCD.Org +
			ImportSeparator +
			rs.Primary.ID, nil
	}
}

func testAccCheckVmSizingPolicyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no VM sizing policy ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		adminOrg, err := conn.GetAdminOrg(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf(errorRetrievingOrg, testConfig.VCD.Org+" and error: "+err.Error())
		}

		_, err = adminOrg.GetVdcComputePolicyById(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("VM sizing policy %s does not exist (%s)", rs.Primary.Attributes["name"], err)
		}

		return nil
	}
}

func testAccCheckVmSizingPolicyDestroyed(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_org_vdc" && rs.Primary.Attributes["name"] != TestVmPolicy {
			continue
		}

		adminOrg, err := conn.GetAdminOrg(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf(errorRetrievingOrg, testConfig.VCD.Org+" and error: "+err.Error())
		}

		_, err = adminOrg.GetVdcComputePolicyById(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("VM sizing policy %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func init() {
	testingTags["vdc"] = "resource_vcd_org_vdc_test.go"
}

const testAccCheckVmSizingPolicy_basic = `
resource "vcd_vm_sizing_policy" "{{.PolicyName}}_1" {
  org         = "{{.OrgName}}"
  name        = "{{.PolicyName}}_1"
  description = "{{.Description}}_1"
}

resource "vcd_vm_sizing_policy" "{{.PolicyName}}_2" {
  org         = "{{.OrgName}}"
  name        = "{{.PolicyName}}_2"
  description = "{{.Description}}_2"

  cpu {
    shares                = "{{.CpuShare}}"
    limit_in_mhz          = "{{.CpuLimit}}"
    count                 = "{{.CpuCount}}"
    speed_in_mhz          = "{{.CpuSpeed}}"
    cores_per_socket      = "{{.CoresPerSocket}}"
    reservation_guarantee = "{{.CpuReservation}}"
  }
}

resource "vcd_vm_sizing_policy" "{{.PolicyName}}_3" {
  org         = "{{.OrgName}}"
  name        = "{{.PolicyName}}_3"
  description = "{{.Description}}_3"

  memory {
    shares                = "{{.MemoryShare}}"
    size_in_mb            = "{{.MemorySize}}"
    limit_in_mb           = "{{.MemoryLimit}}"
    reservation_guarantee = "{{.MemoryReservation}}"
  }
}

resource "vcd_vm_sizing_policy" "{{.PolicyName}}_4" {
  org         = "{{.OrgName}}"
  name        = "{{.PolicyName}}_4"
  description = "{{.Description}}_4"

  cpu {
    shares                = "{{.CpuShare}}"
    limit_in_mhz          = "{{.CpuLimit}}"
    count                 = "{{.CpuCount}}"
    speed_in_mhz          = "{{.CpuSpeed}}"
    cores_per_socket      = "{{.CoresPerSocket}}"
    reservation_guarantee = "{{.CpuReservation}}"
  }

  memory {
    shares                = "{{.MemoryShare}}"
    size_in_mb            = "{{.MemorySize}}"
    limit_in_mb           = "{{.MemoryLimit}}"
    reservation_guarantee = "{{.MemoryReservation}}"
  }
}
`

const testAccCheckVmSizingPolicy_update = `
# skip-binary-test: only for updates
resource "vcd_vm_sizing_policy" "{{.PolicyName}}_1" {
  org         = "{{.OrgName}}"
  name        = "{{.PolicyName}}_1"
  description = "{{.Description}}_1"
}

resource "vcd_vm_sizing_policy" "{{.PolicyName}}_2" {
  org         = "{{.OrgName}}"
  name        = "{{.PolicyName}}_2"
  description = "{{.Description}}_2"

  cpu {
    shares                = "{{.CpuShare}}"
    limit_in_mhz          = "{{.CpuLimit}}"
    count                 = "{{.CpuCount}}"
    speed_in_mhz          = "{{.CpuSpeed}}"
    cores_per_socket      = "{{.CoresPerSocket}}"
    reservation_guarantee = "{{.CpuReservation}}"
  }
}

resource "vcd_vm_sizing_policy" "{{.PolicyName}}_3" {
  org         = "{{.OrgName}}"
  name        = "{{.PolicyName}}_3"
  description = "{{.Description}}_3"

  memory {
    shares                = "{{.MemoryShare}}"
    size_in_mb            = "{{.MemorySize}}"
    limit_in_mb           = "{{.MemoryLimit}}"
    reservation_guarantee = "{{.MemoryReservation}}"
  }
}

resource "vcd_vm_sizing_policy" "{{.PolicyName}}_4" {
  org         = "{{.OrgName}}"
  name        = "{{.PolicyName}}_updated"
  description = "{{.Description}}_updated"

  cpu {
    shares                = "{{.CpuShare}}"
    limit_in_mhz          = "{{.CpuLimit}}"
    count                 = "{{.CpuCount}}"
    speed_in_mhz          = "{{.CpuSpeed}}"
    cores_per_socket      = "{{.CoresPerSocket}}"
    reservation_guarantee = "{{.CpuReservation}}"
  }

  memory {
    shares                = "{{.MemoryShare}}"
    size_in_mb            = "{{.MemorySize}}"
    limit_in_mb           = "{{.MemoryLimit}}"
    reservation_guarantee = "{{.MemoryReservation}}"
  }
}
`
