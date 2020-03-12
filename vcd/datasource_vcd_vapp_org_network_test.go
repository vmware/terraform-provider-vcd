// +build vm ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// TestAccVcdVappOrgNetworkDS tests a vApp org network data source if a vApp is found in the VDC
func TestAccVcdVappOrgNetworkDS(t *testing.T) {
	var fwEnabled = false
	var natEnabled = false
	var retainIpMacEnabled = true

	var params = StringMap{
		"Org":                testConfig.VCD.Org,
		"Vdc":                testConfig.VCD.Vdc,
		"vappName":           "TestAccVcdVappOrgNetworkDS",
		"orgNetwork":         "TestAccVcdVappOrgNetworkDSOrgNetwork",
		"EdgeGateway":        testConfig.Networking.EdgeGateway,
		"firewallEnabled":    fwEnabled,
		"natEnabled":         natEnabled,
		"retainIpMacEnabled": retainIpMacEnabled,
		"isFenced":           "true",

		"FuncName": "TestAccVcdVappOrgNetworkDS",
	}
	configText := templateFill(datasourceTestVappOrgNetwork, params)
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
					testCheckVappOrgNetworkNonStringOutputs(fwEnabled, natEnabled, retainIpMacEnabled),
				),
			},
		},
	})
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
resource "vcd_vapp" "{{.vappName}}" {
  name = "{{.vappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
}

resource "vcd_network_routed" "{{.orgNetwork}}" {
  name         = "{{.orgNetwork}}"
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "10.10.102.1"

  static_ip_pool {
    start_address = "10.10.102.2"
    end_address   = "10.10.102.254"
  }
}

resource "vcd_vapp_org_network" "createVappOrgNetwork" {
  org                = "{{.Org}}"
  vdc                = "{{.Vdc}}"
  vapp_name          = vcd_vapp.{{.vappName}}.name
  org_network_name   = vcd_network_routed.{{.orgNetwork}}.name
  
  is_fenced = "{{.isFenced}}"

  firewall_enabled      = "{{.firewallEnabled}}"
  nat_enabled           = "{{.natEnabled}}"
  retain_ip_mac_enabled = "{{.retainIpMacEnabled}}"
}

data "vcd_vapp_org_network" "network-ds" {
  vapp_name        = "{{.vappName}}"
  org_network_name = vcd_vapp_org_network.createVappOrgNetwork.org_network_name
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
