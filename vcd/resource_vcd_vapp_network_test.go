//go:build network || vapp || ALL || functional

package vcd

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const vappNameForNetworkTest = "TestAccVappForNetworkTest"
const gateway = "192.168.1.1"
const dns1 = "8.8.8.8"
const dns2 = "1.1.1.1"
const dnsSuffix = "biz.biz"
const netmask = "255.255.255.0"
const guestVlanAllowed = "true"

func TestAccVcdVappNetwork_Isolated(t *testing.T) {
	preTestChecks(t)
	vappNetworkResourceName := "TestAccVcdVappNetwork_Isolated"

	var params = StringMap{
		"Org":          testConfig.VCD.Org,
		"Vdc":          testConfig.VCD.Vdc,
		"resourceName": vappNetworkResourceName,
		// we can't change network name as this results in ID (HREF) change
		"vappNetworkName":           vappNetworkResourceName,
		"description":               "network description",
		"descriptionForUpdate":      "update",
		"gateway":                   gateway,
		"netmask":                   netmask,
		"dns1":                      dns1,
		"dns1ForUpdate":             "8.8.8.7",
		"dns2":                      dns2,
		"dns2ForUpdate":             "1.1.1.2",
		"dnsSuffix":                 dnsSuffix,
		"dnsSuffixForUpdate":        "updated",
		"guestVlanAllowed":          guestVlanAllowed,
		"guestVlanAllowedForUpdate": "false",
		"startAddress":              "192.168.1.10",
		"startAddressForUpdate":     "192.168.1.11",
		"endAddress":                "192.168.1.20",
		"endAddressForUpdate":       "192.168.1.21",
		"vappName":                  vappNameForNetworkTest,
		"maxLeaseTime":              "7200",
		"maxLeaseTimeForUpdate":     "7300",
		"defaultLeaseTime":          "3600",
		"defaultLeaseTimeForUpdate": "3500",
		"dhcpStartAddress":          "192.168.1.21",
		"dhcpStartAddressForUpdate": "192.168.1.22",
		"dhcpEndAddress":            "192.168.1.22",
		"dhcpEndAddressForUpdate":   "192.168.1.23",
		"dhcpEnabled":               "true",
		"dhcpEnabledForUpdate":      "false",
		"EdgeGateway":               testConfig.Networking.EdgeGateway,
		"NetworkName":               "TestAccVcdVAppNet",
		"NetworkName2":              "TestAccVcdVAppNet2",
		// adding space to allow pass validation in testParamsNotEmpty which skips the test if param value is empty
		// to avoid running test when test data is missing
		"OrgNetworkKey":               " ",
		"equalsChar":                  " ",
		"quotationChar":               " ",
		"orgNetwork":                  " ",
		"orgNetworkForUpdate":         " ",
		"retainIpMacEnabled":          "false",
		"retainIpMacEnabledForUpdate": "false",
	}
	testParamsNotEmpty(t, params)

	runVappNetworkTestNetmask(t, params)
	postTestChecks(t)
}

func TestAccVcdVappNetwork_Isolated_ipv6(t *testing.T) {
	preTestChecks(t)

	var params = StringMap{
		"Org":          testConfig.VCD.Org,
		"Vdc":          testConfig.Nsxt.Vdc,
		"resourceName": t.Name(),
		// we can't change network name as this results in ID (HREF) change
		"vappNetworkName":           t.Name(),
		"description":               "network description",
		"descriptionForUpdate":      "update",
		"gateway":                   "fe80:0:0:0:0:0:0:aaaa",
		"prefix_length":             "100",
		"dns1":                      "ab:ab:ab:ab:ab:ab:ab:ab",
		"dns1ForUpdate":             "ab:ab:ab:ab:ab:ab:ab:ac",
		"dns2":                      "bb:bb:bb:bb:bb:bb:bb:bb",
		"dns2ForUpdate":             "bb:bb:bb:bb:bb:bb:bb:bc",
		"dnsSuffix":                 dnsSuffix,
		"dnsSuffixForUpdate":        "updated",
		"guestVlanAllowed":          guestVlanAllowed,
		"guestVlanAllowedForUpdate": "false",
		"startAddress":              "fe80:0:0:0:0:0:0:aa",
		"startAddressForUpdate":     "fe80:0:0:0:0:0:0:bb",
		"endAddress":                "fe80:0:0:0:0:0:0:ab",
		"endAddressForUpdate":       "fe80:0:0:0:0:0:0:bc",
		"vappName":                  vappNameForNetworkTest,
		"vappVmName":                t.Name(),
		"NetworkName":               "TestAccVcdVAppNet",
		// adding space to allow pass validation in testParamsNotEmpty which skips the test if param value is empty
		// to avoid running test when test data is missing
		"OrgNetworkKey":               " ",
		"equalsChar":                  " ",
		"quotationChar":               " ",
		"orgNetwork":                  " ",
		"orgNetworkForUpdate":         " ",
		"retainIpMacEnabled":          "false",
		"retainIpMacEnabledForUpdate": "false",
		"RebootVappOnRemoval":         "true",
	}
	testParamsNotEmpty(t, params)

	runVappNetworkTestPrefixLength(t, params)
	postTestChecks(t)
}

func TestAccVcdVappNetwork_Nat(t *testing.T) {
	preTestChecks(t)
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
		"OrgNetworkKey":               "org_network_name",
		"equalsChar":                  "=",
		"quotationChar":               "\"",
		"orgNetwork":                  "TestAccVcdVAppNet",
		"orgNetworkForUpdate":         "TestAccVcdVAppNet2",
		"retainIpMacEnabled":          "false",
		"retainIpMacEnabledForUpdate": "true",
		"FuncName":                    "TestAccVcdVappNetwork_Nat",
	}
	testParamsNotEmpty(t, params)

	runVappNetworkTestNetmask(t, params)
	postTestChecks(t)
}

// TODO leave only one test, runVappNetworkTestPrefixLength after Netmask is fully deprecated
func runVappNetworkTestNetmask(t *testing.T, params StringMap) {
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
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVappNetworkDestroyNsxv,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVappNetworkExists(resourceName, testConfig.VCD.Vdc),
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
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "static_ip_pool.*", map[string]string{
						"start_address": params["startAddress"].(string),
						"end_address":   params["endAddress"].(string),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "dhcp_pool.*", map[string]string{
						"start_address":      params["dhcpStartAddress"].(string),
						"end_address":        params["dhcpEndAddress"].(string),
						"enabled":            params["dhcpEnabled"].(string),
						"default_lease_time": params["defaultLeaseTime"].(string),
						"max_lease_time":     params["maxLeaseTime"].(string),
					}),
					resource.TestCheckResourceAttr(
						resourceName, "org_network_name", strings.TrimSpace(params["orgNetwork"].(string))),
					resource.TestCheckResourceAttr(
						resourceName, "retain_ip_mac_enabled", params["retainIpMacEnabled"].(string)),
				),
			},
			{
				Config: updateConfigText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVappNetworkExists(resourceName, testConfig.VCD.Vdc),
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
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "static_ip_pool.*", map[string]string{
						"start_address": params["startAddressForUpdate"].(string),
						"end_address":   params["endAddressForUpdate"].(string),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "dhcp_pool.*", map[string]string{
						"start_address":      params["dhcpStartAddressForUpdate"].(string),
						"end_address":        params["dhcpEndAddressForUpdate"].(string),
						"enabled":            params["dhcpEnabledForUpdate"].(string),
						"default_lease_time": params["defaultLeaseTimeForUpdate"].(string),
						"max_lease_time":     params["maxLeaseTimeForUpdate"].(string),
					}),
					resource.TestCheckResourceAttr(
						resourceName, "org_network_name", strings.TrimSpace(params["orgNetworkForUpdate"].(string))),
					resource.TestCheckResourceAttr(
						resourceName, "retain_ip_mac_enabled", params["retainIpMacEnabledForUpdate"].(string)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdVappObject(params["vappName"].(string), params["vappNetworkName"].(string), testConfig.VCD.Vdc),
				// These fields can't be retrieved from user data.
				ImportStateVerifyIgnore: []string{"org", "vdc", "reboot_vapp_on_removal"},
			},
		},
	})
}

func runVappNetworkTestPrefixLength(t *testing.T, params StringMap) {
	configText := templateFill(testAccCheckVappNetwork_basic_ipv6, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	params["FuncName"] = t.Name() + "-Update"
	updateConfigText := templateFill(testAccCheckVappNetwork_update_ipv6, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", updateConfigText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceName := "vcd_vapp_network." + params["resourceName"].(string)
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVappNetworkDestroyNsxt,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVappNetworkExists(resourceName, testConfig.Nsxt.Vdc),
					resource.TestCheckResourceAttr(
						resourceName, "name", params["vappNetworkName"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "description", params["description"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "gateway", params["gateway"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "prefix_length", params["prefix_length"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "dns1", params["dns1"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "dns2", params["dns2"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "dns_suffix", dnsSuffix),
					resource.TestCheckResourceAttr(
						resourceName, "guest_vlan_allowed", guestVlanAllowed),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "static_ip_pool.*", map[string]string{
						"start_address": params["startAddress"].(string),
						"end_address":   params["endAddress"].(string),
					}),
					resource.TestCheckResourceAttr(
						resourceName, "org_network_name", strings.TrimSpace(params["orgNetwork"].(string))),
					resource.TestCheckResourceAttr(
						resourceName, "retain_ip_mac_enabled", params["retainIpMacEnabled"].(string)),
				),
			},
			{
				Config: updateConfigText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVappNetworkExists(resourceName, testConfig.Nsxt.Vdc),
					resource.TestCheckResourceAttr(
						resourceName, "name", params["vappNetworkName"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "description", params["descriptionForUpdate"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "gateway", params["gateway"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "prefix_length", params["prefix_length"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "dns1", params["dns1ForUpdate"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "dns2", params["dns2ForUpdate"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "dns_suffix", params["dnsSuffixForUpdate"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "guest_vlan_allowed", params["guestVlanAllowedForUpdate"].(string)),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "static_ip_pool.*", map[string]string{
						"start_address": params["startAddressForUpdate"].(string),
						"end_address":   params["endAddressForUpdate"].(string),
					}),
					resource.TestCheckResourceAttr(
						resourceName, "org_network_name", strings.TrimSpace(params["orgNetworkForUpdate"].(string))),
					resource.TestCheckResourceAttr(
						resourceName, "retain_ip_mac_enabled", params["retainIpMacEnabledForUpdate"].(string)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdVappObject(params["vappName"].(string), params["vappNetworkName"].(string), testConfig.Nsxt.Vdc),
				// These fields can't be retrieved from user data.
				ImportStateVerifyIgnore: []string{"org", "vdc", "reboot_vapp_on_removal"},
			},
		},
	})
}

func testAccCheckVappNetworkExists(n, vdc string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no vapp network ID is set")
		}

		found, err := doesVappNetworkExist(rs, vdc)
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
func testAccCheckVappNetworkDestroyNsxv(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_vapp" {
			continue
		}

		_, err := doesVappNetworkExist(rs, testConfig.VCD.Vdc)
		if err == nil {
			return fmt.Errorf("vapp %s still exists", vappNameForNetworkTest)
		}
	}

	return nil
}

func testAccCheckVappNetworkDestroyNsxt(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_vapp" {
			continue
		}

		_, err := doesVappNetworkExist(rs, testConfig.Nsxt.Vdc)
		if err == nil {
			return fmt.Errorf("vapp %s still exists", vappNameForNetworkTest)
		}
	}

	return nil
}

func doesVappNetworkExist(rs *terraform.ResourceState, testVdc string) (bool, error) {
	conn := testAccProvider.Meta().(*VCDClient)
	_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testVdc)
	if err != nil {
		return false, fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vapp, err := vdc.GetVAppByName(rs.Primary.Attributes["vapp_name"], false)
	if err != nil {
		return false, fmt.Errorf("error retrieving vApp: %s, %#v", rs.Primary.ID, err)
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
		if normalizeId("urn:vcloud:network:", networkId) == rs.Primary.ID {
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

  {{.OrgNetworkKey}} {{.equalsChar}} {{.quotationChar}}{{.orgNetwork}}{{.quotationChar}}

  retain_ip_mac_enabled = "{{.retainIpMacEnabled}}"

  depends_on = ["vcd_vapp.{{.vappName}}", "vcd_network_routed.{{.NetworkName}}"]
}
`

const testAccCheckVappNetwork_basic_ipv6 = `
resource "vcd_vapp" "{{.vappName}}" {
  name = "{{.vappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
}

resource "vcd_vapp_vm" "{{.vappVmName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name     = "{{.vappName}}"
  name          = "{{.vappVmName}}"
  computer_name = "emptyVM"
  memory        = 2048
  cpus          = 2
  cpu_cores     = 1

  network {
    type               = "vapp"
    name               = vcd_vapp_network.{{.resourceName}}.name
    ip_allocation_mode = "POOL"
  }

  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"

  depends_on = [vcd_vapp.{{.vappName}}, vcd_vapp_network.{{.resourceName}}]
}

resource "vcd_vapp_network" "{{.resourceName}}" {
  org                = "{{.Org}}"
  vdc                = "{{.Vdc}}"
  name               = "{{.vappNetworkName}}"
  description        = "{{.description}}"
  vapp_name          = "{{.vappName}}"
  gateway            = "{{.gateway}}"
  prefix_length      = "{{.prefix_length}}"
  dns1               = "{{.dns1}}"
  dns2               = "{{.dns2}}"
  dns_suffix         = "{{.dnsSuffix}}"
  guest_vlan_allowed = {{.guestVlanAllowed}}

  static_ip_pool {
    start_address = "{{.startAddress}}"
    end_address   = "{{.endAddress}}"
  }

  {{.OrgNetworkKey}} {{.equalsChar}} {{.quotationChar}}{{.orgNetwork}}{{.quotationChar}}
  
  retain_ip_mac_enabled = "{{.retainIpMacEnabled}}"
  reboot_vapp_on_removal = {{.RebootVappOnRemoval}}

  depends_on = [vcd_vapp.{{.vappName}}]
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

  {{.OrgNetworkKey}} {{.equalsChar}} {{.quotationChar}}{{.orgNetworkForUpdate}}{{.quotationChar}}

  retain_ip_mac_enabled = "{{.retainIpMacEnabledForUpdate}}"

  depends_on = ["vcd_vapp.{{.vappName}}", "vcd_network_routed.{{.NetworkName}}", "vcd_network_routed.{{.NetworkName2}}"]
}
`

const testAccCheckVappNetwork_update_ipv6 = `
resource "vcd_vapp" "{{.vappName}}" {
  name = "{{.vappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
}

resource "vcd_vapp_vm" "{{.vappVmName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name     = "{{.vappName}}"
  name          = "{{.vappVmName}}"
  computer_name = "emptyVM"
  memory        = 2048
  cpus          = 2
  cpu_cores     = 1

  network {
    type               = "vapp"
    name               = vcd_vapp_network.{{.resourceName}}.name
    ip_allocation_mode = "POOL"
  }

  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"

  depends_on = ["vcd_vapp.{{.vappName}}", "vcd_vapp_network.{{.resourceName}}"]
}

resource "vcd_vapp_network" "{{.resourceName}}" {
  org                = "{{.Org}}"
  vdc                = "{{.Vdc}}"
  name               = "{{.vappNetworkName}}"
  description        = "{{.descriptionForUpdate}}"
  vapp_name          = "{{.vappName}}"
  gateway            = "{{.gateway}}"
  prefix_length      = "{{.prefix_length}}"
  dns1               = "{{.dns1ForUpdate}}"
  dns2               = "{{.dns2ForUpdate}}"
  dns_suffix         = "{{.dnsSuffixForUpdate}}"
  guest_vlan_allowed = {{.guestVlanAllowedForUpdate}}
  static_ip_pool {
    start_address = "{{.startAddressForUpdate}}"
    end_address   = "{{.endAddressForUpdate}}"
  }

  {{.OrgNetworkKey}} {{.equalsChar}} {{.quotationChar}}{{.orgNetworkForUpdate}}{{.quotationChar}}

  retain_ip_mac_enabled  = "{{.retainIpMacEnabledForUpdate}}"
  reboot_vapp_on_removal = {{.RebootVappOnRemoval}}

  depends_on = ["vcd_vapp.{{.vappName}}"]
}
`

// TestAccVcdNsxtVappNetworks checks that NSX-T Org networks can be attached to vApp, given that
// NSX-T Edge Cluster is specified in NSX-T VDC
func TestAccVcdNsxtVappNetworks(t *testing.T) {
	preTestChecks(t)
	// String map to fill the template
	var params = StringMap{
		"Org":                       testConfig.VCD.Org,
		"VdcName":                   t.Name(),
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"ProviderVdcStorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"EdgeCluster":               testConfig.Nsxt.NsxtEdgeCluster,
		"ExternalNetwork":           testConfig.Nsxt.ExternalNetwork,
		"TestName":                  t.Name(),
		"NetworkName":               t.Name(),

		"Tags": "network vapp",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccVcdNsxtVappNetwork, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText)

	params["FuncName"] = t.Name() + "step-2"
	configTextDS := templateFill(testAccVcdNsxtVappNetworkDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configTextDS)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_vapp_org_network.with-isolated", "id"),
					resource.TestCheckResourceAttrSet("vcd_vapp_org_network.with-routed", "id"),
					resource.TestCheckResourceAttrSet("vcd_vapp_network.isolated-attached", "id"),
					resource.TestCheckResourceAttrSet("vcd_vapp_network.routed-attached", "id"),
					resource.TestCheckResourceAttrSet("vcd_vapp.test", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.nsxt-edge", "id"),
					resource.TestCheckResourceAttrSet("vcd_org_vdc.with-edge-cluster", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_isolated_v2.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_isolated_v2.net2", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.nsxt-backed", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.nsxt-backed2", "id"),
				),
			},
			{
				Config: configTextDS,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("vcd_vapp_network.routed-attached", "id", "data.vcd_vapp_network.routed-attached", "id"),
					resource.TestCheckResourceAttrPair("vcd_vapp_network.isolated-attached", "id", "data.vcd_vapp_network.isolated-attached", "id"),

					resource.TestCheckResourceAttrPair("vcd_vapp_org_network.with-isolated", "id", "data.vcd_vapp_org_network.with-isolated", "id"),
					resource.TestCheckResourceAttrPair("vcd_vapp_org_network.with-routed", "id", "data.vcd_vapp_org_network.with-routed", "id"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtVappNetwork = `
data "vcd_provider_vdc" "pvdc" {
  name = "{{.ProviderVdc}}"
}

data "vcd_nsxt_edge_cluster" "ec" {
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
  name            = "{{.EdgeCluster}}"
}

resource "vcd_org_vdc" "with-edge-cluster" {
  name = "{{.VdcName}}"
  org  = "{{.Org}}"

  allocation_model  = "ReservationPool"
  network_pool_name = "{{.NetworkPool}}"
  provider_vdc_name = data.vcd_provider_vdc.pvdc.name
  network_quota     = 5

  edge_cluster_id = data.vcd_nsxt_edge_cluster.ec.id

  compute_capacity {
    cpu {
      allocated = 1024
      limit     = 1024
    }

    memory {
      allocated = 1024
      limit     = 1024
    }
  }

  storage_profile {
    name     = "{{.ProviderVdcStorageProfile}}"
    enabled  = true
    limit    = 10240
    default  = true
  }

  enabled                    = true
  enable_thin_provisioning   = true
  enable_fast_provisioning   = true
  delete_force               = true
  delete_recursive           = true
}

resource "vcd_network_isolated_v2" "net1" {
  org      = "{{.Org}}"
  owner_id = vcd_org_vdc.with-edge-cluster.id
  name     = "{{.NetworkName}}-isolated"

  gateway       = "7.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "7.1.1.10"
    end_address   = "7.1.1.20"
  }
}

resource "vcd_vapp" "test" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.with-edge-cluster.name

  name     = "{{.TestName}}"
  power_on = false

  depends_on = [vcd_org_vdc.with-edge-cluster]
}

resource "vcd_vapp_network" "isolated-attached" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.with-edge-cluster.name

  name               = "{{.TestName}}-isolated-attached"
  vapp_name          = vcd_vapp.test.name
  gateway            = "10.10.66.1"
  netmask            = "255.255.255.0"
  guest_vlan_allowed = false
  
  org_network_name   = vcd_network_isolated_v2.net1.name

  static_ip_pool {
    start_address = "10.10.66.10"
    end_address   = "10.10.66.20"
  }

  depends_on = [
	vcd_vapp.test,
	vcd_network_isolated_v2.net1
  ]
}

resource "vcd_network_isolated_v2" "net2" {
	org      = "{{.Org}}"
	owner_id = vcd_org_vdc.with-edge-cluster.id
	name     = "{{.NetworkName}}-isolated-2"
  
	gateway       = "8.1.1.1"
	prefix_length = 24
  
	static_ip_pool {
	  start_address = "8.1.1.10"
	  end_address   = "8.1.1.20"
	}
  }

resource "vcd_vapp_org_network" "with-isolated" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.with-edge-cluster.name

  vapp_name = vcd_vapp.test.name
  
  org_network_name = vcd_network_isolated_v2.net2.name

  depends_on = [
	vcd_vapp.test,
	vcd_network_isolated_v2.net1
  ]
}

data "vcd_external_network_v2" "existing-extnet" {
	name = "{{.ExternalNetwork}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org         = "{{.Org}}"
  owner_id    = vcd_org_vdc.with-edge-cluster.id
  name        = "{{.TestName}}"

  external_network_id = data.vcd_external_network_v2.existing-extnet.id

  subnet {
     gateway       = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].gateway
     prefix_length = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].prefix_length

     primary_ip = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }
}

resource "vcd_network_routed_v2" "nsxt-backed" {
  org         = "{{.Org}}"
  name        = "{{.TestName}}-routed"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  gateway       = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "1.1.1.10"
    end_address   = "1.1.1.20"
  }
}

resource "vcd_network_routed_v2" "nsxt-backed2" {
  org  = "{{.Org}}"
  name = "{{.TestName}}-routed-2"
  
  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id
  
  gateway       = "2.1.1.1"
  prefix_length = 24
  
  static_ip_pool {
    start_address = "2.1.1.10"
    end_address   = "2.1.1.20"
  }
}

resource "vcd_vapp_network" "routed-attached" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.with-edge-cluster.name

  name               = "{{.TestName}}-routed-attached"
  vapp_name          = vcd_vapp.test.name
  gateway            = "20.10.66.1"
  netmask            = "255.255.255.0"
  guest_vlan_allowed = false
  
  org_network_name   = vcd_network_routed_v2.nsxt-backed2.name

  static_ip_pool {
    start_address = "20.10.66.10"
    end_address   = "20.10.66.20"
  }

  depends_on = [
	vcd_vapp.test,
	vcd_network_routed_v2.nsxt-backed2
  ]
}

resource "vcd_vapp_org_network" "with-routed" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.with-edge-cluster.name

  vapp_name = vcd_vapp.test.name
  
  org_network_name = vcd_network_routed_v2.nsxt-backed2.name

  depends_on = [
	vcd_vapp.test,
	vcd_network_routed_v2.nsxt-backed2
  ]
}
`

const testAccVcdNsxtVappNetworkDS = testAccVcdNsxtVappNetwork + `
# skip-binary-test: Data Source test
data "vcd_vapp_network" "routed-attached" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.with-edge-cluster.name

  vapp_name = vcd_vapp.test.name
  name      = vcd_vapp_network.routed-attached.name
}

data "vcd_vapp_network" "isolated-attached" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.with-edge-cluster.name

  vapp_name = vcd_vapp.test.name
  name      = vcd_vapp_network.isolated-attached.name
}

data "vcd_vapp_org_network" "with-isolated" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.with-edge-cluster.name

  vapp_name        = vcd_vapp.test.name
  org_network_name = vcd_vapp_org_network.with-isolated.org_network_name
}

data "vcd_vapp_org_network" "with-routed" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.with-edge-cluster.name

  vapp_name        = vcd_vapp.test.name
  org_network_name = vcd_vapp_org_network.with-routed.org_network_name
}
`

// TestAccVcdNsxtVappNetworkRemoval checks the following:
// * Creates a vApp with two networks (vcd_vapp_network and vcd_vapp_org_network), both having
// reboot_vapp_on_removal = true
// * Removes everything, except vApp to check that its power state remains POWERED_ON (int status 4)
// after network removal
func TestAccVcdNsxtVappNetworkRemoval(t *testing.T) {
	preTestChecks(t)
	// String map to fill the template
	var params = StringMap{
		"Org":                       testConfig.VCD.Org,
		"VdcName":                   testConfig.Nsxt.Vdc,
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"ProviderVdcStorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"EdgeCluster":               testConfig.Nsxt.NsxtEdgeCluster,
		"ExternalNetwork":           testConfig.Nsxt.ExternalNetwork,
		"TestName":                  t.Name(),
		"ExistingRoutedNetwork":     testConfig.Nsxt.RoutedNetwork,
		"RebootVappOnRemoval":       "true",

		"Tags": "network vapp",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdNsxtVappNetworkRemoval, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdNsxtVappNetworkRemovalVApp, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{ // Create setup
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_vapp.test", "id"),
					resource.TestCheckResourceAttrSet("vcd_vapp_network.test", "id"),
					resource.TestCheckResourceAttrSet("vcd_vapp_org_network.test", "id"),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm.test", "id"),
				),
			},
			{ // Delete everything, except vApp to check that its power state is still powered on
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_vapp.test", "id"),
					resource.TestCheckResourceAttr("vcd_vapp.test", "status", "4"), // 4 is POWERED_ON
					resource.TestCheckResourceAttr("vcd_vapp.test", "status_text", "POWERED_ON"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtVappNetworkRemovalVApp = `
resource "vcd_vapp" "test" {
  org  = "{{.Org}}"
  vdc  = "{{.VdcName}}"
  name = "{{.TestName}}"
}
`

const testAccVcdNsxtVappNetworkRemoval = testAccVcdNsxtVappNetworkRemovalVApp + `
resource "vcd_vapp_network" "test" {
  org = "{{.Org}}"
  vdc = "{{.VdcName}}"

  name      = "{{.TestName}}"
  vapp_name = vcd_vapp.test.name
  gateway   = "192.168.2.1"
  netmask   = "255.255.255.0"

  static_ip_pool {
    start_address = "192.168.2.51"
    end_address   = "192.168.2.100"
  }

  reboot_vapp_on_removal = {{.RebootVappOnRemoval}}
}

resource "vcd_vapp_org_network" "test" {
  org = "{{.Org}}"
  vdc = "{{.VdcName}}"

  vapp_name        = vcd_vapp.test.name
  org_network_name = "{{.ExistingRoutedNetwork}}"

  reboot_vapp_on_removal = {{.RebootVappOnRemoval}}
}

resource "vcd_vapp_vm" "test" {
  vapp_name     = vcd_vapp.test.name
  name          = "{{.TestName}}"
  computer_name = "emptyVM"
  memory        = 1048
  cpus          = 2
  cpu_cores     = 1

  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"
  
  power_on = true
  
  depends_on = [vcd_vapp_network.test, vcd_vapp_org_network.test]
}
`

// TestAccVcdNsxtVappNetworkRemovalFails does the following:
// * Attempts to replicate a user process of removing vApp networks without reboot_vapp_on_removal
// * Checks that user gets a clear error message with a hint to use reboot_vapp_on_removal
// * Checks that one can update the value and destroy whole environment afterwards
func TestAccVcdNsxtVappNetworkRemovalFails(t *testing.T) {
	preTestChecks(t)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	vcdClient := createTemporaryVCDConnection(false)
	if vcdClient.Client.APIVCDMaxVersionIs("< 37.1") {
		t.Skipf("This test tests VCD 10.4.1+ (API V37.1+) features. Skipping.")
	}

	// String map to fill the template
	var params = StringMap{
		"Org":                       testConfig.VCD.Org,
		"VdcName":                   testConfig.Nsxt.Vdc,
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"ProviderVdcStorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"EdgeCluster":               testConfig.Nsxt.NsxtEdgeCluster,
		"ExternalNetwork":           testConfig.Nsxt.ExternalNetwork,
		"TestName":                  t.Name(),
		"ExistingRoutedNetwork":     testConfig.Nsxt.RoutedNetwork,
		"RebootVappOnRemoval":       "false",

		"Tags": "network vapp",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdNsxtVappNetworkRemoval, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["RebootVappOnRemoval"] = "true"
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdNsxtVappNetworkRemoval, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVappNetworkDestroyNsxv,
		),
		Steps: []resource.TestStep{
			{ // Create setup
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_vapp.test", "id"),
					resource.TestCheckResourceAttrSet("vcd_vapp_network.test", "id"),
					resource.TestCheckResourceAttrSet("vcd_vapp_org_network.test", "id"),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm.test", "id"),
				),
			},
			{ // Explicitly attempt to destroy the vApp with reboot_vapp_on_removal = false (default value)
				Config:  configText1,
				Destroy: true,
				// Test that enriched error message with hint for 'reboot_vapp_on_removal' is
				// returned. This is to validate that error catching mechanism is working as VCD API
				// evolves.
				ExpectError: regexp.MustCompile("reboot_vapp_on_removal"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_vapp.test", "id"),
					resource.TestCheckResourceAttrSet("vcd_vapp_network.test", "id"),
					resource.TestCheckResourceAttrSet("vcd_vapp_org_network.test", "id"),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm.test", "id"),
				),
			},
			{ // Set the flag reboot_vapp_on_removal=true so that vApp can be destroyed
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_vapp.test", "id"),
					resource.TestCheckResourceAttrSet("vcd_vapp_network.test", "id"),
					resource.TestCheckResourceAttrSet("vcd_vapp_org_network.test", "id"),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm.test", "id"),
				),
			},
		},
	})
	postTestChecks(t)
}
