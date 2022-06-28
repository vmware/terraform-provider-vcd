//go:build vm || ALL || functional
// +build vm ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdVappOrgNetworkDS tests a vApp org network data source if a vApp is found in the VDC
func TestAccVcdVappOrgNetworkDS(t *testing.T) {
	preTestChecks(t)
	//var retainIpMacEnabled = true
	var orgNetName = t.Name()

	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.Nsxt.Vdc,
		"vappName":    "TestAccVcdVappOrgNetworkDS",
		"orgNetwork":  orgNetName,
		"EdgeGateway": testConfig.Nsxt.EdgeGateway,
		// "retainIpMacEnabled": retainIpMacEnabled, // only supported by NSX-V
		// "isFenced":           "false", // only supported by NSX-V

		"FuncName": "TestAccVcdVappOrgNetworkDS",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(datasourceTestVappOrgNetwork, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testCheckVappOrgNetworkNonStringOutputs(orgNetName),
				),
			},
		},
	})
	postTestChecks(t)
}

func testCheckVappOrgNetworkNonStringOutputs(orgNetName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		outputs := s.RootModule().Outputs

		if outputs["org_network_name"].Value != fmt.Sprintf("%v", orgNetName) {
			return fmt.Errorf("org_network_name value didn't match")
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

data "vcd_nsxt_edgegateway" "existing" {
  org  = "{{.Org}}"
  name = "{{.EdgeGateway}}"
}

resource "vcd_network_routed_v2" "{{.orgNetwork}}" {
  org          = "{{.Org}}"
  name         = "{{.orgNetwork}}"
  
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  gateway       = "10.10.102.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "10.10.102.2"
    end_address   = "10.10.102.254"
  }
}



resource "vcd_vapp_org_network" "createVappOrgNetwork" {
  org                = "{{.Org}}"
  vdc                = "{{.Vdc}}"
  vapp_name          = vcd_vapp.{{.vappName}}.name
  org_network_name   = vcd_network_routed_v2.{{.orgNetwork}}.name
}

data "vcd_vapp_org_network" "network-ds" {
  vapp_name        = "{{.vappName}}"
  org_network_name = vcd_vapp_org_network.createVappOrgNetwork.org_network_name
  depends_on 	   = [vcd_vapp_org_network.createVappOrgNetwork]
}

output "org_network_name" {
  value = data.vcd_vapp_org_network.network-ds.org_network_name
}  
`
