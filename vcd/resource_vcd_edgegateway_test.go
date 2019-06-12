// +build gateway ALL functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccVcdEdgeGatewayBasic(t *testing.T) {
	var edgeGatewayName string = "TestEdgeGatewayBasic"
	var edgeGatewayVcdName string = "test_edge_gateway_basic"

	// String map to fill the template
	var params = StringMap{
		"Org":             testConfig.VCD.Org,
		"Vdc":             testConfig.VCD.Vdc,
		"EdgeGateway":     edgeGatewayName,
		"EdgeGatewayVcd":  edgeGatewayVcdName,
		"ExternalNetwork": testConfig.Networking.ExternalNetwork,
		"Tags":            "gateway",
	}
	configText := templateFill(testAccEdgeGatewayBasic, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdEdgeGatewayDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vcd_edgegateway."+edgeGatewayName, "default_gateway", testConfig.Networking.ExternalNetwork),
				),
			},
		},
	})
}

func TestAccVcdEdgeGatewayComplex(t *testing.T) {
	var (
		edgeGatewayName       string = "TestEdgeGatewayComplex"
		edgeGatewayVcdName    string = "test_edge_gateway_complex"
		newExternalNetwork    string = "TestExternalNetwork"
		newExternalNetworkVcd string = "test_external_network"
	)

	// String map to fill the template
	var params = StringMap{
		"Org":                   testConfig.VCD.Org,
		"Vdc":                   testConfig.VCD.Vdc,
		"EdgeGateway":           edgeGatewayName,
		"EdgeGatewayVcd":        edgeGatewayVcdName,
		"ExternalNetwork":       testConfig.Networking.ExternalNetwork,
		"Tags":                  "gateway",
		"NewExternalNetwork":    newExternalNetwork,
		"NewExternalNetworkVcd": newExternalNetworkVcd,
		"Type":                  testConfig.Networking.ExternalNetworkPortGroupType,
		"PortGroup":             testConfig.Networking.ExternalNetworkPortGroup,
		"Vcenter":               testConfig.Networking.Vcenter,
	}
	configText := templateFill(testAccEdgeGatewayComplex, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdEdgeGatewayDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vcd_edgegateway."+edgeGatewayName, "default_gateway", newExternalNetworkVcd),
				),
			},
		},
	})
}

func testAccCheckVcdEdgeGatewayDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_edgegateway" {
			continue
		}

		return nil
	}

	return nil
}

const testAccEdgeGatewayBasic = `
resource "vcd_edgegateway" "{{.EdgeGateway}}" {
  org                 = "{{.Org}}"
  vdc                 = "{{.Vdc}}"
  name                = "{{.EdgeGatewayVcd}}"
  description         = "Description"
  backing_config      = "compact"
  default_gateway     = "{{.ExternalNetwork}}"

  external_networks   = [ "{{.ExternalNetwork}}" ]
}
`

const testAccEdgeGatewayComplex = `

resource "vcd_external_network" "{{.NewExternalNetwork}}" {
  name        = "{{.NewExternalNetworkVcd}}"
  description = "Test External Network"

  vsphere_network {
    vcenter = "{{.Vcenter}}"
    name    = "{{.PortGroup}}"
    type    = "{{.Type}}"
  }

  ip_scope {
    gateway      = "192.168.30.49"
    netmask      = "255.255.255.240"
    dns1         = "192.168.0.164"
    dns2         = "192.168.0.196"
    dns_suffix   = "company.biz"

    static_ip_pool {
      start_address = "192.168.30.51"
      end_address   = "192.168.30.62"
    }
  }

  retain_net_info_across_deployments = "false"
}

resource "vcd_edgegateway" "{{.EdgeGateway}}" {
  org                 = "{{.Org}}"
  vdc                 = "{{.Vdc}}"
  name                = "{{.EdgeGatewayVcd}}"
  description         = "Description"
  backing_config      = "compact"
  default_gateway     = "${vcd_external_network.{{.NewExternalNetwork}}.name}"

  external_networks   = [ "{{.ExternalNetwork}}", "${vcd_external_network.{{.NewExternalNetwork}}.name}" ]
}
`
