// +build network vapp ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const vappNameForNetworkTest = "TestAccVappForNetworkTest"
const gateway = "192.168.1.1"
const dns1 = "8.8.8.8"
const dns2 = "1.1.1.1"
const dnsSuffix = "biz.biz"
const newVappNetworkName = "TestAccVcdVappNetwork_Basic"
const netmask = "255.255.255.0"
const guestVlanAllowed = "true"

func TestAccVcdVappNetwork_Basic(t *testing.T) {
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceName := "TestVappNetwork"

	var params = StringMap{
		"Org":              testConfig.VCD.Org,
		"Vdc":              testConfig.VCD.Vdc,
		"resourceName":     resourceName,
		"vappNetworkName":  newVappNetworkName,
		"gateway":          gateway,
		"netmask":          netmask,
		"dns1":             dns1,
		"dns2":             dns2,
		"dnsSuffix":        dnsSuffix,
		"guestVlanAllowed": guestVlanAllowed,
		"startAddress":     "192.168.1.10",
		"endAddress":       "192.168.1.20",
		"vappName":         vappNameForNetworkTest,
		"maxLeaseTime":     "7200",
		"defaultLeaseTime": "3600",
		"dhcpStartAddress": "192.168.1.21",
		"dhcpEndAddress":   "192.168.1.22",
		"dhcpEnabled":      "true",
	}

	configText := templateFill(testAccCheckVappNetwork_basic, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVappNetworkDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVappNetworkExists("vcd_vapp_network."+resourceName),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+resourceName, "gateway", gateway),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+resourceName, "netmask", netmask),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+resourceName, "dns1", dns1),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+resourceName, "dns2", dns2),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+resourceName, "dns_suffix", dnsSuffix),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+resourceName, "guest_vlan_allowed", guestVlanAllowed),
				),
			},
		},
	})
}

func testAccCheckVappNetworkExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no vapp network ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		found, err := isVappNetworkFound(conn, rs, "exist")
		if err != nil {
			return err
		}

		if !found {
			return fmt.Errorf("vApp network was not found")
		}

		return nil
	}
}

// TODO: In future this can be improved to check if network delete only,
// when test suite will create vApp which can be reused
func testAccCheckVappNetworkDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_vapp" {
			continue
		}

		_, err := isVappNetworkFound(conn, rs, "destroy")
		if err == nil {
			return fmt.Errorf("vapp %s still exists", vappNameForNetworkTest)
		}
	}

	return nil
}

func isVappNetworkFound(conn *VCDClient, rs *terraform.ResourceState, origin string) (bool, error) {
	_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
	if err != nil {
		return false, fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vapp, err := vdc.GetVAppByName(vappNameForNetworkTest, false)
	if err != nil {
		return false, fmt.Errorf("error retrieving vApp: %s, %#v", rs.Primary.ID, err)
	}

	// Avoid looking for network when the purpose is only finding whether the vApp exists
	if origin == "destroy" {
		return true, nil
	}
	networkConfig, err := vapp.GetNetworkConfig()
	if err != nil {
		return false, fmt.Errorf("error retrieving network config from vApp: %#v", err)
	}

	var found bool
	for _, vappNetworkConfig := range networkConfig.NetworkConfig {
		if vappNetworkConfig.NetworkName == newVappNetworkName && vappNetworkConfig.Configuration.IPScopes.IPScope[0].DNSSuffix == dnsSuffix {
			found = true
		}
	}

	return found, nil
}

const testAccCheckVappNetwork_basic = `
resource "vcd_vapp" "{{.vappName}}" {
  name = "{{.vappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
}

resource "vcd_vapp_network" "{{.resourceName}}" {
  org                = "{{.Org}}"
  vdc                = "{{.Vdc}}"
  name               = "{{.vappNetworkName}}"
  vapp_name          = "{{.vappName}}"
  gateway            = "{{.gateway}}"
  netmask            = "{{.netmask}}"
  dns1               = "{{.dns1}}"
  dns2               = "{{.dns2}}"
  dns_suffix         = "{{.dnsSuffix}}"
  guest_vlan_allowed = {{.guestVlanAllowed}}

  static_ip_pool {
    start_address = "{{.startAddress}}"
    end_address   = "{{.endAddress}}"
  }

  dhcp_pool {
    max_lease_time     = "{{.maxLeaseTime}}"
    default_lease_time = "{{.defaultLeaseTime}}"
    start_address      = "{{.dhcpStartAddress}}"
    end_address        = "{{.dhcpEndAddress}}"
    enabled            = "{{.dhcpEnabled}}"
  }

  depends_on = ["vcd_vapp.{{.vappName}}"]
}
`
