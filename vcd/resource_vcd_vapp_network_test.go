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
const netmask = "255.255.255.0"
const guestVlanAllowed = "true"

func TestAccVcdVappNetwork_Isolated(t *testing.T) {
	vappNetworkResourceName := "TestAccVcdVappNetwork_Isolated"

	var params = StringMap{
		"Org":          testConfig.VCD.Org,
		"Vdc":          testConfig.VCD.Vdc,
		"resourceName": vappNetworkResourceName,
		// we can't change network name as this results in ID (HREF) change
		"vappNetworkName":             vappNetworkResourceName,
		"description":                 "network description",
		"descriptionForUpdate":        "update",
		"gateway":                     gateway,
		"netmask":                     netmask,
		"dns1":                        dns1,
		"dns1ForUpdate":               "8.8.8.7",
		"dns2":                        dns2,
		"dns2ForUpdate":               "1.1.1.2",
		"dnsSuffix":                   dnsSuffix,
		"dnsSuffixForUpdate":          "updated",
		"guestVlanAllowed":            guestVlanAllowed,
		"guestVlanAllowedForUpdate":   "false",
		"startAddress":                "192.168.1.10",
		"startAddressForUpdate":       "192.168.1.11",
		"endAddress":                  "192.168.1.20",
		"endAddressForUpdate":         "192.168.1.21",
		"vappName":                    vappNameForNetworkTest,
		"maxLeaseTime":                "7200",
		"maxLeaseTimeForUpdate":       "7300",
		"defaultLeaseTime":            "3600",
		"defaultLeaseTimeForUpdate":   "3500",
		"dhcpStartAddress":            "192.168.1.21",
		"dhcpStartAddressForUpdate":   "192.168.1.22",
		"dhcpEndAddress":              "192.168.1.22",
		"dhcpEndAddressForUpdate":     "192.168.1.23",
		"dhcpEnabled":                 "true",
		"dhcpEnabledForUpdate":        "false",
		"EdgeGateway":                 testConfig.Networking.EdgeGateway,
		"NetworkName":                 "TestAccVcdVAppNet",
		"NetworkName2":                "TestAccVcdVAppNet2",
		"orgNetwork":                  "",
		"orgNetworkForUpdate":         "",
		"firewallEnabled":             "false",
		"firewallEnabledForUpdate":    "false",
		"natEnabled":                  "true",
		"natEnabledForUpdate":         "true",
		"retainIpMacEnabled":          "false",
		"retainIpMacEnabledForUpdate": "false",
	}

	runVappNetworkTest(t, params)
}

func TestAccVcdVappNetwork_Nat(t *testing.T) {
	vappNetworkResourceName := "TestAccVcdVappNetwork_Nat"

	var params = StringMap{
		"Org":          testConfig.VCD.Org,
		"Vdc":          testConfig.VCD.Vdc,
		"resourceName": vappNetworkResourceName,
		// we can't change network name as this results in ID (HREF) change
		"vappNetworkName":             vappNetworkResourceName,
		"description":                 "network description",
		"descriptionForUpdate":        "update",
		"gateway":                     gateway,
		"netmask":                     netmask,
		"dns1":                        dns1,
		"dns1ForUpdate":               "8.8.8.7",
		"dns2":                        dns2,
		"dns2ForUpdate":               "1.1.1.2",
		"dnsSuffix":                   dnsSuffix,
		"dnsSuffixForUpdate":          "updated",
		"guestVlanAllowed":            guestVlanAllowed,
		"guestVlanAllowedForUpdate":   "false",
		"startAddress":                "192.168.1.10",
		"startAddressForUpdate":       "192.168.1.11",
		"endAddress":                  "192.168.1.20",
		"endAddressForUpdate":         "192.168.1.21",
		"vappName":                    vappNameForNetworkTest,
		"maxLeaseTime":                "7200",
		"maxLeaseTimeForUpdate":       "7300",
		"defaultLeaseTime":            "3600",
		"defaultLeaseTimeForUpdate":   "3500",
		"dhcpStartAddress":            "192.168.1.21",
		"dhcpStartAddressForUpdate":   "192.168.1.22",
		"dhcpEndAddress":              "192.168.1.22",
		"dhcpEndAddressForUpdate":     "192.168.1.23",
		"dhcpEnabled":                 "true",
		"dhcpEnabledForUpdate":        "false",
		"EdgeGateway":                 testConfig.Networking.EdgeGateway,
		"NetworkName":                 "TestAccVcdVAppNet",
		"NetworkName2":                "TestAccVcdVAppNet2",
		"orgNetwork":                  "TestAccVcdVAppNet",
		"orgNetworkForUpdate":         "TestAccVcdVAppNet2",
		"firewallEnabled":             "false",
		"firewallEnabledForUpdate":    "true",
		"natEnabled":                  "false",
		"natEnabledForUpdate":         "true",
		"retainIpMacEnabled":          "false",
		"retainIpMacEnabledForUpdate": "true",
		"FuncName":                    "TestAccVcdVappNetwork_Nat",
	}

	runVappNetworkTest(t, params)
}

func runVappNetworkTest(t *testing.T, params StringMap) {
	configText := templateFill(testAccCheckVappNetwork_basic, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	params["FuncName"] = t.Name() + "-Update"
	updateConfigText := templateFill(testAccCheckVappNetwork_update, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", updateConfigText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceName := "vcd_vapp_network." + params["resourceName"].(string)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVappNetworkDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVappNetworkExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "name", params["vappNetworkName"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "description", params["description"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "gateway", gateway),
					resource.TestCheckResourceAttr(
						resourceName, "netmask", netmask),
					resource.TestCheckResourceAttr(
						resourceName, "dns1", dns1),
					resource.TestCheckResourceAttr(
						resourceName, "dns2", dns2),
					resource.TestCheckResourceAttr(
						resourceName, "dns_suffix", dnsSuffix),
					resource.TestCheckResourceAttr(
						resourceName, "guest_vlan_allowed", guestVlanAllowed),
					resource.TestCheckResourceAttr(
						resourceName, "static_ip_pool.2802459930.start_address", params["startAddress"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "static_ip_pool.2802459930.end_address", params["endAddress"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "dhcp_pool.3992831158.start_address", params["dhcpStartAddress"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "dhcp_pool.3992831158.end_address", params["dhcpEndAddress"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "dhcp_pool.3992831158.enabled", params["dhcpEnabled"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "dhcp_pool.3992831158.default_lease_time", params["defaultLeaseTime"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "dhcp_pool.3992831158.max_lease_time", params["maxLeaseTime"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "org_network_name", params["orgNetwork"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "retain_ip_mac_enabled", params["retainIpMacEnabled"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "firewall_enabled", params["firewallEnabled"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "nat_enabled", params["natEnabled"].(string)),
				),
			},
			resource.TestStep{
				Config: updateConfigText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVappNetworkExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "name", params["vappNetworkName"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "description", params["descriptionForUpdate"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "gateway", gateway),
					resource.TestCheckResourceAttr(
						resourceName, "netmask", netmask),
					resource.TestCheckResourceAttr(
						resourceName, "dns1", params["dns1ForUpdate"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "dns2", params["dns2ForUpdate"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "dns_suffix", params["dnsSuffixForUpdate"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "guest_vlan_allowed", params["guestVlanAllowedForUpdate"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "static_ip_pool.132829107.start_address", params["startAddressForUpdate"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "static_ip_pool.132829107.end_address", params["endAddressForUpdate"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "dhcp_pool.4252262036.start_address", params["dhcpStartAddressForUpdate"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "dhcp_pool.4252262036.end_address", params["dhcpEndAddressForUpdate"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "dhcp_pool.4252262036.enabled", params["dhcpEnabledForUpdate"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "dhcp_pool.4252262036.default_lease_time", params["defaultLeaseTimeForUpdate"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "dhcp_pool.4252262036.max_lease_time", params["maxLeaseTimeForUpdate"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "org_network_name", params["orgNetworkForUpdate"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "retain_ip_mac_enabled", params["retainIpMacEnabledForUpdate"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "firewall_enabled", params["firewallEnabledForUpdate"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "nat_enabled", params["natEnabledForUpdate"].(string)),
				),
			},
			resource.TestStep{
				ResourceName:      resourceName + "-import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdVappObject(testConfig, params["vappName"].(string), params["vappNetworkName"].(string)),
				// These fields can't be retrieved from user data. firewall_enabled is null and as so we need ignore in some cases
				ImportStateVerifyIgnore: []string{"org", "vdc", "firewall_enabled"},
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

	vapp, err := vdc.GetVAppByName(rs.Primary.Attributes["vapp_name"], false)
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
		networkId, err := govcd.GetUuidFromHref(vappNetworkConfig.Link.HREF, false)
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

  org_network_name      = "{{.orgNetwork}}"
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

resource "vcd_network_routed" "{{.NetworkName2}}" {
  name         = "{{.NetworkName2}}"
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "10.10.103.1"

  static_ip_pool {
    start_address = "10.10.103.2"
    end_address   = "10.10.103.254"
  }
}

resource "vcd_vapp_network" "{{.resourceName}}" {
  org                = "{{.Org}}"
  vdc                = "{{.Vdc}}"
  name               = "{{.vappNetworkName}}"
  description        = "{{.descriptionForUpdate}}"
  vapp_name          = "{{.vappName}}"
  gateway            = "{{.gateway}}"
  netmask            = "{{.netmask}}"
  dns1               = "{{.dns1ForUpdate}}"
  dns2               = "{{.dns2ForUpdate}}"
  dns_suffix         = "{{.dnsSuffixForUpdate}}"
  guest_vlan_allowed = {{.guestVlanAllowedForUpdate}}
  static_ip_pool {
    start_address = "{{.startAddressForUpdate}}"
    end_address   = "{{.endAddressForUpdate}}"
  }

  dhcp_pool {
    max_lease_time     = "{{.maxLeaseTimeForUpdate}}"
    default_lease_time = "{{.defaultLeaseTimeForUpdate}}"
    start_address      = "{{.dhcpStartAddressForUpdate}}"
    end_address        = "{{.dhcpEndAddressForUpdate}}"
    enabled            = "{{.dhcpEnabledForUpdate}}"
  }

  org_network_name      = "{{.orgNetworkForUpdate}}"
  firewall_enabled      = "{{.firewallEnabledForUpdate}}"
  nat_enabled           = "{{.natEnabledForUpdate}}"
  retain_ip_mac_enabled = "{{.retainIpMacEnabledForUpdate}}"

  depends_on = ["vcd_vapp.{{.vappName}}", "vcd_network_routed.{{.NetworkName}}", "vcd_network_routed.{{.NetworkName2}}"]
}
`
