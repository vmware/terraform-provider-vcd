// +build network vapp ALL functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccVcdVappOrgNetwork_NotFenced(t *testing.T) {
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceName = "TestAccVcdVappOrgNetwork_NotFenced"

	var params = StringMap{
		"Org":                testConfig.VCD.Org,
		"Vdc":                testConfig.VCD.Vdc,
		"resourceName":       resourceName,
		"vappNetworkName":    newVappNetworkName,
		"gateway":            gateway,
		"netmask":            netmask,
		"dns1":               dns1,
		"dns2":               dns2,
		"dnsSuffix":          dnsSuffix,
		"guestVlanAllowed":   guestVlanAllowed,
		"startAddress":       "192.168.1.10",
		"endAddress":         "192.168.1.20",
		"vappName":           vappNameForNetworkTest,
		"maxLeaseTime":       "7200",
		"defaultLeaseTime":   "3600",
		"dhcpStartAddress":   "192.168.1.21",
		"dhcpEndAddress":     "192.168.1.22",
		"dhcpEnabled":        "true",
		"EdgeGateway":        testConfig.Networking.EdgeGateway,
		"NetworkName":        "TestAccVcdVAppNet",
		"orgNetwork":         "",
		"firewallEnabled":    "false",
		"natEnabled":         "false",
		"retainIpMacEnabled": "false",
	}

	rungVappOrgNetworkTest(t, params)
}

func TestAccVcdVappOrgNetwork_Fenced(t *testing.T) {
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceName = "TestAccVcdVappOrgNetwork_Fenced"

	var params = StringMap{
		"Org":                testConfig.VCD.Org,
		"Vdc":                testConfig.VCD.Vdc,
		"resourceName":       resourceName,
		"vappNetworkName":    newVappNetworkName,
		"gateway":            gateway,
		"netmask":            netmask,
		"dns1":               dns1,
		"dns2":               dns2,
		"dnsSuffix":          dnsSuffix,
		"guestVlanAllowed":   guestVlanAllowed,
		"startAddress":       "192.168.1.10",
		"endAddress":         "192.168.1.20",
		"vappName":           vappNameForNetworkTest,
		"maxLeaseTime":       "7200",
		"defaultLeaseTime":   "3600",
		"dhcpStartAddress":   "192.168.1.21",
		"dhcpEndAddress":     "192.168.1.22",
		"dhcpEnabled":        "true",
		"EdgeGateway":        testConfig.Networking.EdgeGateway,
		"NetworkName":        "TestAccVcdVAppNet",
		"orgNetwork":         "TestAccVcdVAppNet",
		"firewallEnabled":    "false",
		"natEnabled":         "false",
		"retainIpMacEnabled": "true",
		"FuncName":           "TestAccVcdVappNetwork_Nat",
	}

	rungVappOrgNetworkTest(t, params)
}

func rungVappOrgNetworkTest(t *testing.T, params StringMap) {
	configText := templateFill(testAccCheckOrgVappNetwork_basic, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

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
				),
			},
		},
	})
}

const testAccCheckOrgVappNetwork_basic = `
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

resource "vcd_vapp_org_network" "{{.resourceName}}" {
  org                = "{{.Org}}"
  vdc                = "{{.Vdc}}"
  vapp_name          = "{{.vappName}}"
  org_network        = "{{.orgNetwork}}"
  
  is_fenced = true

  firewall_enabled      = "{{.firewallEnabled}}"
  nat_enabled           = "{{.natEnabled}}"
  retain_ip_mac_enabled = "{{.retainIpMacEnabled}}"

  depends_on = ["vcd_vapp.{{.vappName}}", "vcd_network_routed.{{.NetworkName}}"]
}
`
