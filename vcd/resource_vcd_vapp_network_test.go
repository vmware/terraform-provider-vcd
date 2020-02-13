// +build network vapp ALL functional

package vcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const vappNameForNetworkTest = "TestAccVappForNetworkTest"
const gateway = "192.168.1.1"
const dns1 = "8.8.8.8"
const dns2 = "1.1.1.1"
const dnsSuffix = "biz.biz"
const newVappNetworkName = "TestAccVcdVappNetwork_Basic"
const netmask = "255.255.255.0"
const guestVlanAllowed = "true"

func TestAccVcdVappNetwork_Isolated(t *testing.T) {
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceName = "TestAccVcdVappNetwork_Isolated"

	var params = StringMap{
		"Org":             testConfig.VCD.Org,
		"Vdc":             testConfig.VCD.Vdc,
		"resourceName":    resourceName,
		"vappNetworkName": newVappNetworkName,
		// change id with name?
		"vappNetworkNameForUpdate": newVappNetworkName + "updated",
		"description":              "network description",
		"gateway":                  gateway,
		"netmask":                  netmask,
		"dns1":                     dns1,
		"dns2":                     dns2,
		"dnsSuffix":                dnsSuffix,
		"guestVlanAllowed":         guestVlanAllowed,
		"startAddress":             "192.168.1.10",
		"endAddress":               "192.168.1.20",
		"vappName":                 vappNameForNetworkTest,
		"maxLeaseTime":             "7200",
		"defaultLeaseTime":         "3600",
		"dhcpStartAddress":         "192.168.1.21",
		"dhcpEndAddress":           "192.168.1.22",
		"dhcpEnabled":              "true",
		"EdgeGateway":              testConfig.Networking.EdgeGateway,
		"NetworkName":              "TestAccVcdVAppNet",
		"orgNetwork":               "",
		"firewallEnabled":          "false",
		"natEnabled":               "false",
		"retainIpMacEnabled":       "false",
	}

	rungVappNetworkTest(t, params)
}

func TestAccVcdVappNetwork_Nat(t *testing.T) {
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceName = "TestAccVcdVappNetwork_Nat"

	var params = StringMap{
		"Org":             testConfig.VCD.Org,
		"Vdc":             testConfig.VCD.Vdc,
		"resourceName":    resourceName,
		"vappNetworkName": newVappNetworkName,
		// change id with name?
		"vappNetworkNameForUpdate": newVappNetworkName + "updated",
		"description":              "network description",
		"gateway":                  gateway,
		"netmask":                  netmask,
		"dns1":                     dns1,
		"dns2":                     dns2,
		"dnsSuffix":                dnsSuffix,
		"guestVlanAllowed":         guestVlanAllowed,
		"startAddress":             "192.168.1.10",
		"endAddress":               "192.168.1.20",
		"vappName":                 vappNameForNetworkTest,
		"maxLeaseTime":             "7200",
		"defaultLeaseTime":         "3600",
		"dhcpStartAddress":         "192.168.1.21",
		"dhcpEndAddress":           "192.168.1.22",
		"dhcpEnabled":              "true",
		"EdgeGateway":              testConfig.Networking.EdgeGateway,
		"NetworkName":              "TestAccVcdVAppNet",
		"orgNetwork":               "TestAccVcdVAppNet",
		"firewallEnabled":          "false",
		"natEnabled":               "false",
		"retainIpMacEnabled":       "true",
		"FuncName":                 "TestAccVcdVappNetwork_Nat",
	}

	rungVappNetworkTest(t, params)
}

func rungVappNetworkTest(t *testing.T, params StringMap) {
	configText := templateFill(testAccCheckVappNetwork_basic, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	params["FuncName"] = t.Name() + "-Update"
	updateConfigText := templateFill(testAccCheckVappNetwork_update, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", testAccCheckVappNetwork_update)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVappNetworkDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVappNetworkExists("vcd_vapp_network."+params["resourceName"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "description", params["description"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "gateway", gateway),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "netmask", netmask),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "dns1", dns1),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "dns2", dns2),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "dns_suffix", dnsSuffix),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "guest_vlan_allowed", guestVlanAllowed),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "static_ip_pool.2802459930.start_address", params["startAddress"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "static_ip_pool.2802459930.end_address", params["endAddress"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "dhcp_pool.84879490.start_address", params["dhcpStartAddress"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "dhcp_pool.84879490.end_address", params["dhcpEndAddress"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "dhcp_pool.84879490.enabled", params["dhcpEnabled"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "org_network", params["orgNetwork"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "retain_ip_mac_enabled", params["retainIpMacEnabled"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "firewall_enabled", params["firewallEnabled"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "nat_enabled", params["natEnabled"].(string)),
				),
			},
			resource.TestStep{
				Config: updateConfigText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVappNetworkExists("vcd_vapp_network."+params["resourceName"].(string)),
					//resource.TestCheckResourceAttr(
					//	"vcd_vapp_network."+params["resourceName"].(string), "name", params["vappNetworkNameForUpdate"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "description", params["description"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "gateway", gateway),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "netmask", netmask),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "dns1", dns1),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "dns2", dns2),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "dns_suffix", dnsSuffix),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "guest_vlan_allowed", guestVlanAllowed),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "static_ip_pool.2802459930.start_address", params["startAddress"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "static_ip_pool.2802459930.end_address", params["endAddress"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "dhcp_pool.84879490.start_address", params["dhcpStartAddress"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "dhcp_pool.84879490.end_address", params["dhcpEndAddress"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "dhcp_pool.84879490.enabled", params["dhcpEnabled"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "org_network", params["orgNetwork"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "retain_ip_mac_enabled", params["retainIpMacEnabled"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "firewall_enabled", params["firewallEnabled"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_network."+params["resourceName"].(string), "nat_enabled", params["natEnabled"].(string)),
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
		networkId, err := govcd.GetUuidFromHref(vappNetworkConfig.Link.HREF)
		if err != nil {
			return false, fmt.Errorf("unable to get network ID from HREF: %s", err)
		}
		if networkId == rs.Primary.ID {
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

resource "vcd_network_routed" "{{.NetworkName}}" {
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

resource "vcd_vapp_network" "{{.resourceName}}" {
  org                = "{{.Org}}"
  vdc                = "{{.Vdc}}"
  name               = "{{.vappNetworkName}}"
  description        = "{{.description}}"
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

  org_network           = "{{.orgNetwork}}"
  firewall_enabled      = "{{.firewallEnabled}}"
  nat_enabled           = "{{.natEnabled}}"
  retain_ip_mac_enabled = "{{.retainIpMacEnabled}}"

  depends_on = ["vcd_vapp.{{.vappName}}", "vcd_network_routed.{{.NetworkName}}"]
}
`

const testAccCheckVappNetwork_update = `
resource "vcd_vapp" "{{.vappName}}" {
  name = "{{.vappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
}

resource "vcd_network_routed" "{{.NetworkName}}" {
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

resource "vcd_vapp_network" "{{.resourceName}}" {
  org                = "{{.Org}}"
  vdc                = "{{.Vdc}}"
  name               = "{{.vappNetworkName}}"
  description        = "{{.description}}"
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

  org_network           = "{{.orgNetwork}}"
  firewall_enabled      = "{{.firewallEnabled}}"
  nat_enabled           = "{{.natEnabled}}"
  retain_ip_mac_enabled = "{{.retainIpMacEnabled}}"

  depends_on = ["vcd_vapp.{{.vappName}}", "vcd_network_routed.{{.NetworkName}}"]
}
`
