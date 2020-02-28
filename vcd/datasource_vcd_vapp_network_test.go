// +build vm ALL functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// TestAccVcdVappNetworkDS tests a vApp network data source if a vApp is found in the VDC
func TestAccVcdVappNetworkDS(t *testing.T) {
	// This test requires access to the vCD before filling templates
	// Thus it won't run in the short test
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	vapp, err := getAvailableVapp()
	if err != nil {
		t.Skip("No suitable vApp found for this test")
		return
	}

	err = getAvailableNetworks()

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

	vappNetworkSettings := &govcd.VappNetworkSettings{
		Name:               networkName,
		Gateway:            gateway,
		NetMask:            netmask,
		DNS1:               dns1,
		DNS2:               dns2,
		DNSSuffix:          dnsSuffix,
		StaticIPRanges:     []*types.IPRange{{StartAddress: startAddress, EndAddress: endAddress}},
		DhcpSettings:       &govcd.DhcpSettings{IsEnabled: true, MaxLeaseTime: maxLeaseTime, DefaultLeaseTime: defaultLeaseTime, IPRange: &types.IPRange{StartAddress: dhcpStartAddress, EndAddress: dhcpEndAddress}},
		GuestVLANAllowed:   &guestVlanAllowed,
		Description:        description,
		FirewallEnabled:    &fwEnabled,
		NatEnabled:         &natEnabled,
		RetainIpMacEnabled: &retainIpMacEnabled,
	}

	_, err = vapp.CreateVappNetwork(vappNetworkSettings, data.network)
	if err != nil {
		fmt.Printf("%s\n", err)
		t.Skip("error adding vApp network")
		return
	}

	var params = StringMap{
		"Org":             testConfig.VCD.Org,
		"VDC":             testConfig.VCD.Vdc,
		"VappName":        vapp.VApp.Name,
		"FuncName":        "TestVappVmDS",
		"vappNetworkName": networkName,
	}
	configText := templateFill(datasourceTestVappNetwork, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.ParallelTest(t, resource.TestCase{
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

data "vcd_vapp_network" "network-ds" {
  name       = "{{.vappNetworkName}}"
  vapp_name  = "{{.VappName}}"
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
