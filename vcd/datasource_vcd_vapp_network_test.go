// +build vm ALL functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// TestAccVcdVappNetworkDS tests a vApp network data source if a vApp is found in the VDC
func TestAccVcdVappNetworkDS(t *testing.T) {
	err := getAvailableNetworks()

	if err != nil {
		fmt.Printf("%s\n", err)
		t.Skip("error getting available networks")
		return
	}
	if len(availableNetworks) == 0 {
		t.Skip("No networks found - data source test skipped")
		return
	}

	networkType := "vcd_network_routed"
	data, ok := availableNetworks[networkType]
	if !ok {
		t.Skip("no routed network found ")
		return
	}

	networkName := "TestAccVcdVappNetworkDS"
	description := "Created in test"
	const gateway = "192.168.0.1"
	const netmask = "255.255.255.0"
	const dns1 = "8.8.8.8"
	const dns2 = "1.1.1.1"
	const dnsSuffix = "biz.biz"
	const startAddress = "192.168.0.10"
	const endAddress = "192.168.0.20"
	const dhcpStartAddress = "192.168.0.30"
	const dhcpEndAddress = "192.168.0.40"
	const maxLeaseTime = 3500
	const defaultLeaseTime = 2400
	var guestVlanAllowed = true
	var fwEnabled = false
	var natEnabled = false
	var retainIpMacEnabled = true

	var params = StringMap{
		"Org":                testConfig.VCD.Org,
		"VDC":                testConfig.VCD.Vdc,
		"vappName":           "TestAccVcdVappNetworkDS",
		"FuncName":           "TestVappNetworkDS",
		"vappNetworkName":    networkName,
		"description":        description,
		"gateway":            gateway,
		"netmask":            netmask,
		"dns1":               dns1,
		"dns2":               dns2,
		"dnsSuffix":          dnsSuffix,
		"guestVlanAllowed":   guestVlanAllowed,
		"startAddress":       startAddress,
		"endAddress":         endAddress,
		"maxLeaseTime":       maxLeaseTime,
		"defaultLeaseTime":   defaultLeaseTime,
		"dhcpStartAddress":   dhcpStartAddress,
		"dhcpEndAddress":     dhcpEndAddress,
		"dhcpEnabled":        "true",
		"orgNetwork":         data.network.Name,
		"firewallEnabled":    fwEnabled,
		"natEnabled":         natEnabled,
		"retainIpMacEnabled": retainIpMacEnabled,
	}
	configText := templateFill(datasourceTestVappNetwork, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("netmask", netmask),
					resource.TestCheckOutput("description", description),
					resource.TestCheckOutput("gateway", gateway),
					resource.TestCheckOutput("dns1", dns1),
					resource.TestCheckOutput("dns2", dns2),
					resource.TestCheckOutput("dnsSuffix", dnsSuffix),
					resource.TestCheckOutput("dhcpStartAddress", dhcpStartAddress),
					resource.TestCheckOutput("dhcpEndAddress", dhcpEndAddress),
					resource.TestCheckOutput("staticIpPoolStartAddress", startAddress),
					resource.TestCheckOutput("staticIpPoolEndAddress", endAddress),
					resource.TestCheckOutput("orgNetwork", data.network.Name),
					testCheckVappNetworkNonStringOutputs(guestVlanAllowed, fwEnabled, natEnabled, retainIpMacEnabled),
				),
			},
		},
	})
}

func testCheckVappNetworkNonStringOutputs(guestVlanAllowed, firewallEnabled, natEnabled, retainIpMacEnabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		outputs := s.RootModule().Outputs

		if outputs["guestVlanAllowed"].Value != guestVlanAllowed {
			return fmt.Errorf("guestVlanAllowed value didn't match")
		}

		if outputs["retain_ip_mac_enabled"].Value != retainIpMacEnabled {
			return fmt.Errorf("retain_ip_mac_enabled value didn't match")
		}

		if outputs["firewall_enabled"].Value != firewallEnabled {
			return fmt.Errorf("retain_ip_mac_enabled value didn't match")
		}

		if outputs["nat_enabled"].Value != natEnabled {
			return fmt.Errorf("retain_ip_mac_enabled value didn't match")
		}
		return nil
	}
}

const datasourceTestVappNetwork = `
resource "vcd_vapp" "{{.vappName}}" {
  name = "{{.vappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.VDC}}"
}

resource "vcd_vapp_network" "createdVappNetwork" {
  org                = "{{.Org}}"
  vdc                = "{{.VDC}}"
  name               = "{{.vappNetworkName}}"
  description        = "{{.description}}"
  vapp_name          = vcd_vapp.{{.vappName}}.name
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
}
 

data "vcd_vapp_network" "network-ds" {
  name       =  vcd_vapp_network.createdVappNetwork.name
  vapp_name  = "{{.vappName}}"
}

output "netmask" {
  value = data.vcd_vapp_network.network-ds.netmask 
} 
output "description" {
  value = data.vcd_vapp_network.network-ds.description 
} 
output "gateway" {
  value = data.vcd_vapp_network.network-ds.gateway 
} 
output "dns1" {
  value = data.vcd_vapp_network.network-ds.dns1 
} 
output "dns2" {
  value = data.vcd_vapp_network.network-ds.dns2 
} 
output "dnsSuffix" {
  value = data.vcd_vapp_network.network-ds.dns_suffix 
} 
output "guestVlanAllowed" {
  value = data.vcd_vapp_network.network-ds.guest_vlan_allowed
} 
output "dhcpStartAddress" {
  value  = tolist(data.vcd_vapp_network.network-ds.dhcp_pool)[0].start_address
}
output "dhcpEndAddress" {
  value  = tolist(data.vcd_vapp_network.network-ds.dhcp_pool)[0].end_address
}
output "staticIpPoolStartAddress" {
  value  = tolist(data.vcd_vapp_network.network-ds.static_ip_pool)[0].start_address
}
output "staticIpPoolEndAddress" {
  value  = tolist(data.vcd_vapp_network.network-ds.static_ip_pool)[0].end_address
}
output "orgNetwork" {
  value = data.vcd_vapp_network.network-ds.org_network_name
} 
output "retain_ip_mac_enabled" {
  value = data.vcd_vapp_network.network-ds.retain_ip_mac_enabled
} 
output "firewall_enabled" {
  value = data.vcd_vapp_network.network-ds.firewall_enabled
} 
output "nat_enabled" {
  value = data.vcd_vapp_network.network-ds.nat_enabled
} 
`
