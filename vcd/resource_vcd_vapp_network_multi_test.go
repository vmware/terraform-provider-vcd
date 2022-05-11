//go:build multinetwork || functional
// +build multinetwork functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	vappNetworkName1        = "TestVappNetwork1"
	vappNetworkName2        = "TestVappNetwork2"
	vappNetworkName3        = "TestVappNetwork3"
	vappNameForNetworkMulti = "TestAccVappForNetworkMulti"
)

// Creates a VM with three vApp networks
// To execute this test, run
// go test -v -timeout 0 -tags multinetwork -run TestAccVcdVappNetworkMulti .
//
func TestAccVcdVappNetworkMulti(t *testing.T) {
	preTestChecks(t)

	const (
		networkBaseIp1     = "192.168.11"
		networkBaseIp2     = "192.168.12"
		networkBaseIp3     = "192.168.13"
		startStaticAddress = ".11"
		endStaticAddress   = ".20"
		startDhcpAddress   = ".21"
		endDhcpAddress     = ".30"
		gatewayMulti1      = networkBaseIp1 + ".1"
		gatewayMulti2      = networkBaseIp2 + ".1"
		gatewayMulti3      = networkBaseIp3 + ".1"
	)
	resourceName1 := "TestVappNetwork1"
	resourceName2 := "TestVappNetwork2"
	resourceName3 := "TestVappNetwork3"

	var params = StringMap{
		"Org":               testConfig.VCD.Org,
		"Vdc":               testConfig.VCD.Vdc,
		"resourceName1":     resourceName1,
		"resourceName2":     resourceName2,
		"resourceName3":     resourceName3,
		"vappNetworkName1":  vappNetworkName1,
		"vappNetworkName2":  vappNetworkName2,
		"vappNetworkName3":  vappNetworkName3,
		"gateway1":          gatewayMulti1,
		"gateway2":          gatewayMulti2,
		"gateway3":          gatewayMulti3,
		"netmask":           netmask,
		"dns1":              dns1,
		"dns2":              dns2,
		"dnsSuffix":         dnsSuffix,
		"guestVlanAllowed":  guestVlanAllowed,
		"startAddress1":     networkBaseIp1 + startStaticAddress,
		"endAddress1":       networkBaseIp1 + endStaticAddress,
		"startAddress2":     networkBaseIp2 + startStaticAddress,
		"endAddress2":       networkBaseIp2 + endStaticAddress,
		"startAddress3":     networkBaseIp3 + startStaticAddress,
		"endAddress3":       networkBaseIp3 + endStaticAddress,
		"vappName":          vappNameForNetworkMulti,
		"maxLeaseTime":      "7200",
		"defaultLeaseTime":  "3600",
		"dhcpStartAddress1": networkBaseIp1 + startDhcpAddress,
		"dhcpEndAddress1":   networkBaseIp1 + endDhcpAddress,
		"dhcpStartAddress2": networkBaseIp2 + startDhcpAddress,
		"dhcpEndAddress2":   networkBaseIp2 + endDhcpAddress,
		"dhcpStartAddress3": networkBaseIp3 + startDhcpAddress,
		"dhcpEndAddress3":   networkBaseIp3 + endDhcpAddress,
		"dhcpEnabled":       "true",
		"Tags":              "multinetwork",
	}

	configText := templateFill(testAccCheckVappNetworkMulti, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testParamsNotEmpty(t, params) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVappNetworkMultiDestroy,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVappNetworkMultiExists("vcd_vapp_network."+resourceName1),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+resourceName1, "gateway", gatewayMulti1),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+resourceName2, "gateway", gatewayMulti2),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+resourceName3, "gateway", gatewayMulti3),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+resourceName1, "netmask", netmask),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+resourceName1, "dns1", dns1),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+resourceName1, "dns2", dns2),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+resourceName1, "dns_suffix", dnsSuffix),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+resourceName1, "guest_vlan_allowed", guestVlanAllowed),
				),
			},
		},
	})
	postTestChecks(t)
}

func testAccCheckVappNetworkMultiExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no vapp network ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		status, err := isVappNetworkMultiFound(conn, rs)
		if err != nil {
			return err
		}

		if status != "installed" {
			return fmt.Errorf("vApp network was not installed. Status: %s", status)
		}

		return nil
	}
}

// TODO: In future this can be improved to check if network delete only,
// when test suite will create vApp which can be reused
func testAccCheckVappNetworkMultiDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_vapp" {
			continue
		}

		_, err := isVappNetworkMultiFound(conn, rs)
		if err == nil {
			return fmt.Errorf("vapp %s still exists", vappNameForNetworkMulti)
		}
	}

	return nil
}

func isVappNetworkMultiFound(conn *VCDClient, rs *terraform.ResourceState) (string, error) {
	_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
	if err != nil {
		return "uninstalled", fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vapp, err := vdc.GetVAppByName(vappNameForNetworkMulti, false)
	if err != nil {
		return "uninstalled", fmt.Errorf("error retrieving vApp: %s, %#v", rs.Primary.ID, err)
	}

	networkConfig, err := vapp.GetNetworkConfig()
	if err != nil {
		return "uninstalled", fmt.Errorf("error retrieving network config from vApp: %#v", err)
	}

	var foundNetworks int
	for _, vappNetworkConfig := range networkConfig.NetworkConfig {
		if vappNetworkConfig.Configuration.IPScopes.IPScope[0].DNSSuffix == dnsSuffix {
			switch vappNetworkConfig.NetworkName {
			case vappNetworkName1:
				foundNetworks += 10
			case vappNetworkName2:
				foundNetworks += 100
			case vappNetworkName3:
				foundNetworks += 1000
			}
		}
	}

	if foundNetworks == 1110 {
		return "installed", nil
	} else {
		if foundNetworks > 0 {
			return "partial", nil
		}
	}
	return "uninstalled", nil
}

const testAccCheckVappNetworkMulti = `
resource "vcd_vapp" "{{.vappName}}" {
  name = "{{.vappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
}

resource "vcd_vapp_network" "{{.resourceName1}}" {
  org                = "{{.Org}}"
  vdc                = "{{.Vdc}}"
  name               = "{{.vappNetworkName1}}"
  vapp_name          = "{{.vappName}}"
  gateway            = "{{.gateway1}}"
  netmask            = "{{.netmask}}"
  dns1               = "{{.dns1}}"
  dns2               = "{{.dns2}}"
  dns_suffix         = "{{.dnsSuffix}}"
  guest_vlan_allowed = "{{.guestVlanAllowed}}"

  static_ip_pool {
    start_address = "{{.startAddress1}}"
    end_address   = "{{.endAddress1}}"
  }

  dhcp_pool {
    max_lease_time     = "{{.maxLeaseTime}}"
    default_lease_time = "{{.defaultLeaseTime}}"
    start_address      = "{{.dhcpStartAddress1}}"
    end_address        = "{{.dhcpEndAddress1}}"
    enabled            = "{{.dhcpEnabled}}"
  }

  depends_on = [vcd_vapp.{{.vappName}}]
}

resource "vcd_vapp_network" "{{.resourceName2}}" {
  org                = "{{.Org}}"
  vdc                = "{{.Vdc}}"
  name               = "{{.vappNetworkName2}}"
  vapp_name          = "{{.vappName}}"
  gateway            = "{{.gateway2}}"
  netmask            = "{{.netmask}}"
  dns1               = "{{.dns1}}"
  dns2               = "{{.dns2}}"
  dns_suffix         = "{{.dnsSuffix}}"
  guest_vlan_allowed = "{{.guestVlanAllowed}}"

  static_ip_pool {
    start_address = "{{.startAddress2}}"
    end_address   = "{{.endAddress2}}"
  }

  dhcp_pool {
    max_lease_time     = "{{.maxLeaseTime}}"
    default_lease_time = "{{.defaultLeaseTime}}"
    start_address      = "{{.dhcpStartAddress2}}"
    end_address        = "{{.dhcpEndAddress2}}"
    enabled            = "{{.dhcpEnabled}}"
  }

  depends_on = [vcd_vapp.{{.vappName}}]
}

resource "vcd_vapp_network" "{{.resourceName3}}" {
  org                = "{{.Org}}"
  vdc                = "{{.Vdc}}"
  name               = "{{.vappNetworkName3}}"
  vapp_name          = "{{.vappName}}"
  gateway            = "{{.gateway3}}"
  netmask            = "{{.netmask}}"
  dns1               = "{{.dns1}}"
  dns2               = "{{.dns2}}"
  dns_suffix         = "{{.dnsSuffix}}"
  guest_vlan_allowed = "{{.guestVlanAllowed}}"

  static_ip_pool {
    start_address = "{{.startAddress3}}"
    end_address   = "{{.endAddress3}}"
  }

  dhcp_pool {
    max_lease_time     = "{{.maxLeaseTime}}"
    default_lease_time = "{{.defaultLeaseTime}}"
    start_address      = "{{.dhcpStartAddress3}}"
    end_address        = "{{.dhcpEndAddress3}}"
    enabled            = "{{.dhcpEnabled}}"
  }

  depends_on = [vcd_vapp.{{.vappName}}]
}
`
