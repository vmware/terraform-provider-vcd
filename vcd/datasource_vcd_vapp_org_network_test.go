// +build vm ALL functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// TestAccVcdVappOrgNetworkDS tests a vApp org network data source if a vApp is found in the VDC
func TestAccVcdVappOrgNetworkDS(t *testing.T) {
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

	var fwEnabled = false
	var natEnabled = false
	var retainIpMacEnabled = true

	vappNetworkSettings := &govcd.VappNetworkSettings{
		FirewallEnabled:    &fwEnabled,
		NatEnabled:         &natEnabled,
		RetainIpMacEnabled: &retainIpMacEnabled,
	}

	_, err = vapp.AddOrgNetwork(vappNetworkSettings, data.network, true)
	if err != nil {
		fmt.Printf("%s\n", err)
		t.Skip("error adding vApp network")
		return
	}

	var params = StringMap{
		"Org":        testConfig.VCD.Org,
		"VDC":        testConfig.VCD.Vdc,
		"VappName":   vapp.VApp.Name,
		"orgNetwork": data.network.Name,
		"FuncName":   "TestVappOrgNetworkDS",
	}
	configText := templateFill(datasourceTestVappOrgNetwork, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testCheckVappOrgNetworkNonStringOutputs(fwEnabled, natEnabled, retainIpMacEnabled),
				),
			},
		},
	})

	_, err = vapp.RemoveNetwork(data.network.Name)
	if err != nil {
		fmt.Printf("%s\n", err)
		t.Skip("error removing vApp network")
		return
	}
}

func testCheckVappOrgNetworkNonStringOutputs(firewallEnabled, natEnabled, retainIpMacEnabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		outputs := s.RootModule().Outputs

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

const datasourceTestVappOrgNetwork = `

data "vcd_vapp_org_network" "network-ds" {
  vapp_name        = "{{.VappName}}"
  org_network_name = "{{.orgNetwork}}" 
}

output "retain_ip_mac_enabled" {
  value = data.vcd_vapp_org_network.network-ds.retain_ip_mac_enabled
} 
output "firewall_enabled" {
  value = data.vcd_vapp_org_network.network-ds.firewall_enabled
} 
output "nat_enabled" {
  value = data.vcd_vapp_org_network.network-ds.nat_enabled
} 
`
