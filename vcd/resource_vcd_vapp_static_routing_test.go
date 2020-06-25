// +build functional vapp ALL

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccVcdVappStaticRouting(t *testing.T) {
	if testConfig.Networking.EdgeGateway == "" {
		t.Skip("Variable testConfig.Networking.EdgeGateway must be configured")
		return
	}

	var (
		vappName        = t.Name() + "_vapp"
		vappNetworkName = "vapp-routed-net"
	)

	var params = StringMap{
		"Org":           testConfig.VCD.Org,
		"Vdc":           testConfig.VCD.Vdc,
		"EdgeGateway":   testConfig.Networking.EdgeGateway,
		"DefaultAction": "drop",
		"ResourceName":  t.Name(),
		"FuncName":      t.Name(),
		"NetworkName":   "TestAccVcdVAppVmNet",
		"VappName":      vappName,
		"ExternalIp":    testConfig.Networking.ExternalIp,
		"Tags":          "vapp",
	}
	configText := templateFill(testAccVcdVappStaticRouting_routes, params)
	params["FuncName"] = t.Name() + "-step2"
	configTextForUpdate := templateFill(testAccVcdVappStaticRouting_routes_forUpdate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	debugPrintf("#[DEBUG] UPDATE CONFIGURATION: %s", configTextForUpdate)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	resourceName := "vcd_vapp_static_routing." + t.Name()
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVappStaticRoutesExists(resourceName, 2),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),

					resource.TestCheckResourceAttr(resourceName, "rule.0.name", "rule1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.network_cidr", "10.10.0.0/24"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.next_hop_ip", "10.10.102.3"),

					resource.TestCheckResourceAttr(resourceName, "rule.1.name", "rule2"),
					resource.TestCheckResourceAttr(resourceName, "rule.1.network_cidr", "10.10.1.0/24"),
					resource.TestCheckResourceAttr(resourceName, "rule.1.next_hop_ip", "10.10.102.5"),
				),
			},
			// we can reuse importStateVappFirewallRuleObject as import is the same
			resource.TestStep{ // Step 1 - resource import
				ResourceName:            "vcd_vapp_static_routing.imported",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateVappFirewallRuleObject(testConfig, vappName, vappNetworkName),
				ImportStateVerifyIgnore: []string{"network_id", "org", "vdc"},
			},
			resource.TestStep{ // Step 2 - resource import by ID
				ResourceName:            "vcd_vapp_static_routing.imported2",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateVappFirewallRuleById(testConfig, resourceName),
				ImportStateVerifyIgnore: []string{"network_id", "org", "vdc"},
			},
			resource.TestStep{ // Step 3 - update
				Config: configTextForUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVappStaticRoutesExists(resourceName, 2),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),

					resource.TestCheckResourceAttr(resourceName, "rule.0.name", "rule1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.network_cidr", "10.10.1.0/24"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.next_hop_ip", "10.10.102.5"),

					resource.TestCheckResourceAttr(resourceName, "rule.1.name", "rule2"),
					resource.TestCheckResourceAttr(resourceName, "rule.1.network_cidr", "10.10.0.0/24"),
					resource.TestCheckResourceAttr(resourceName, "rule.1.next_hop_ip", "10.10.102.3"),
				),
			},
		},
	})

}

func testAccCheckVcdVappStaticRoutesExists(n string, rulesCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no record ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.VCD.Vdc, testConfig.VCD.Org, err)
		}

		vapp, err := vdc.GetVAppById(rs.Primary.Attributes["vapp_id"], false)
		if err != nil {
			return err
		}

		vapp_network, err := vapp.GetVappNetworkById(rs.Primary.Attributes["network_id"], false)
		if err != nil {
			return err
		}

		if len(vapp_network.Configuration.Features.StaticRoutingService.StaticRoute) == rulesCount {
			return nil
		}
		return fmt.Errorf("no static rule with provided name is found")
	}
}

const testAccVcdVappStaticRouting_vappAndVm = `
resource "vcd_network_routed" "network_routed" {
  name         = "{{.NetworkName}}"
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "10.10.102.1"

  static_ip_pool {
    start_address = "10.10.102.2"
    end_address   = "10.10.102.254"
  }
}

resource "vcd_vapp" "{{.VappName}}" {
  name = "{{.VappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
}

resource "vcd_vapp_network" "vappRoutedNet" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name             = "vapp-routed-net"
  vapp_name        = vcd_vapp.{{.VappName}}.name
  gateway          = "192.168.22.1"
  netmask          = "255.255.255.0"
  org_network_name = vcd_network_routed.network_routed.name

  static_ip_pool {
    start_address = "192.168.22.2"
    end_address   = "192.168.22.254"
  }
}

resource "vcd_vapp_org_network" "vappAttachedNet" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name        = vcd_vapp.{{.VappName}}.name
  org_network_name = vcd_network_routed.network_routed.name
  is_fenced        = true
}
`

const testAccVcdVappStaticRouting_routes = testAccVcdVappStaticRouting_vappAndVm + `
resource "vcd_vapp_static_routing" "{{.ResourceName}}" {
  org        = "{{.Org}}"
  vdc        = "{{.Vdc}}"
  vapp_id    = vcd_vapp.{{.VappName}}.id
  network_id = vcd_vapp_network.vappRoutedNet.id
  enabled     = true

  rule {
    name         = "rule1"
    network_cidr = "10.10.0.0/24"
    next_hop_ip  = "10.10.102.3"
  }

  rule {
    name         = "rule2"
    network_cidr = "10.10.1.0/24"
    next_hop_ip  = "10.10.102.5"
  }
}
`

const testAccVcdVappStaticRouting_routes_forUpdate = testAccVcdVappStaticRouting_vappAndVm + `
resource "vcd_vapp_static_routing" "{{.ResourceName}}" {
  org        = "{{.Org}}"
  vdc        = "{{.Vdc}}"
  vapp_id    = vcd_vapp.{{.VappName}}.id
  network_id = vcd_vapp_network.vappRoutedNet.id
  enabled     = false

  rule {
    name         = "rule1"
    network_cidr = "10.10.1.0/24"
    next_hop_ip  = "10.10.102.5"
  }

  rule {
    name         = "rule2"
    network_cidr = "10.10.0.0/24"
    next_hop_ip  = "10.10.102.3"
  }
}
`
