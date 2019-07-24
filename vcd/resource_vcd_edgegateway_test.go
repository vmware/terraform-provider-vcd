// +build gateway ALL functional

package vcd

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

var (
	edgeGatewayNameBasic   string = "TestEdgeGatewayBasic"
	edgeGatewayNameComplex string = "TestEdgeGatewayComplex"
)

// Since we can't set the "advanced" property to false by default,
// as it would fail on 9.7+, we can run the test with VCD_ADVANCED_FALSE=1
// and see what would happen on different vCD versions. It will succeed on
// 9.1 and 9.5, and fail quickly on 9.7+
// Note: there is another method to handle this issue, and it is by
// checking the vCD version before running the test. However, since the provider is
// not initialized until the test starts, this approach would require an extra
// vCD connection.
func getAdvancedProperty() bool {
	return os.Getenv("VCD_ADVANCED_FALSE") == ""
}

func TestAccVcdEdgeGatewayBasic(t *testing.T) {
	var edgeGatewayVcdName string = "test_edge_gateway_basic"

	// String map to fill the template
	var params = StringMap{
		"Org":             testConfig.VCD.Org,
		"Vdc":             testConfig.VCD.Vdc,
		"EdgeGateway":     edgeGatewayNameBasic,
		"EdgeGatewayVcd":  edgeGatewayVcdName,
		"ExternalNetwork": testConfig.Networking.ExternalNetwork,
		"Advanced":        getAdvancedProperty(),
		"Tags":            "gateway",
	}
	configText := templateFill(testAccEdgeGatewayBasic, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	if !usingSysAdmin() {
		t.Skip("Edge gateway tests requires system admin privileges")
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdEdgeGatewayDestroyBasic,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vcd_edgegateway."+edgeGatewayNameBasic, "default_gateway_network", testConfig.Networking.ExternalNetwork),
				),
			},
		},
	})
}

func TestAccVcdEdgeGatewayComplex(t *testing.T) {
	var (
		edgeGatewayVcdName    string = "test_edge_gateway_complex"
		newExternalNetwork    string = "TestExternalNetwork"
		newExternalNetworkVcd string = "test_external_network"
	)

	// String map to fill the template
	var params = StringMap{
		"Org":                   testConfig.VCD.Org,
		"Vdc":                   testConfig.VCD.Vdc,
		"EdgeGateway":           edgeGatewayNameComplex,
		"EdgeGatewayVcd":        edgeGatewayVcdName,
		"ExternalNetwork":       testConfig.Networking.ExternalNetwork,
		"Tags":                  "gateway",
		"NewExternalNetwork":    newExternalNetwork,
		"NewExternalNetworkVcd": newExternalNetworkVcd,
		"Type":                  testConfig.Networking.ExternalNetworkPortGroupType,
		"PortGroup":             testConfig.Networking.ExternalNetworkPortGroup,
		"Advanced":              getAdvancedProperty(),
		"Vcenter":               testConfig.Networking.Vcenter,
	}
	configText := templateFill(testAccEdgeGatewayComplex, params)

	params["FuncName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccEdgeGatewayComplexWithLb, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	if !usingSysAdmin() {
		t.Skip("Edge gateway tests requires system admin privileges")
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdEdgeGatewayDestroyComplex,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vcd_edgegateway."+edgeGatewayNameComplex, "default_gateway_network", newExternalNetworkVcd),
					// No load balancer fields should appear in statefile if these fields are not used
					resource.TestCheckNoResourceAttr("vcd_edgegateway."+edgeGatewayNameComplex, "lb_enabled"),
					resource.TestCheckNoResourceAttr("vcd_edgegateway."+edgeGatewayNameComplex, "lb_acceleration_enabled"),
					resource.TestCheckNoResourceAttr("vcd_edgegateway."+edgeGatewayNameComplex, "lb_logging_enabled"),
					resource.TestCheckNoResourceAttr("vcd_edgegateway."+edgeGatewayNameComplex, "lb_loglevel"),
				),
			},
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vcd_edgegateway."+edgeGatewayNameComplex, "default_gateway_network", newExternalNetworkVcd),
					// All load balancer fields should appear when load balancing is used
					resource.TestCheckResourceAttr("vcd_edgegateway."+edgeGatewayNameComplex, "lb_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_edgegateway."+edgeGatewayNameComplex, "lb_acceleration_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_edgegateway."+edgeGatewayNameComplex, "lb_logging_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_edgegateway."+edgeGatewayNameComplex, "lb_loglevel", "critical"),
				),
			},
		},
	})
}

func testEdgeGatewayDestroy(s *terraform.State, wantedEgwName string) error {

	for _, rs := range s.RootModule().Resources {
		edgeGatewayName := rs.Primary.Attributes["name"]
		if rs.Type != "vcd_edgegateway" {
			continue
		}
		if edgeGatewayName != wantedEgwName {
			continue
		}
		conn := testAccProvider.Meta().(*VCDClient)
		orgName := rs.Primary.Attributes["org"]
		vdcName := rs.Primary.Attributes["vdc"]

		_, vdc, err := conn.GetOrgAndVdc(orgName, vdcName)
		if err != nil {
			return fmt.Errorf("error retrieving org %s and vdc %s : %s ", orgName, vdcName, err)
		}

		_, err = vdc.FindEdgeGateway(edgeGatewayNameBasic)
		if err == nil {
			return fmt.Errorf("edge gateway %s was not removed", edgeGatewayNameBasic)
		}
	}

	return nil
}

func testAccCheckVcdEdgeGatewayDestroyBasic(s *terraform.State) error {
	return testEdgeGatewayDestroy(s, edgeGatewayNameBasic)
}

func testAccCheckVcdEdgeGatewayDestroyComplex(s *terraform.State) error {
	return testEdgeGatewayDestroy(s, edgeGatewayNameComplex)
}

const testAccEdgeGatewayBasic = `
resource "vcd_edgegateway" "{{.EdgeGateway}}" {
  org                     = "{{.Org}}"
  vdc                     = "{{.Vdc}}"
  name                    = "{{.EdgeGatewayVcd}}"
  description             = "Description"
  configuration           = "compact"
  default_gateway_network = "{{.ExternalNetwork}}"
  advanced                = {{.Advanced}}
  external_networks       = [ "{{.ExternalNetwork}}" ]
}
`

const testAccEdgeGatewayComplexNetwork = `
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
`

const testAccEdgeGatewayComplex = testAccEdgeGatewayComplexNetwork + `
resource "vcd_edgegateway" "{{.EdgeGateway}}" {
  org                     = "{{.Org}}"
  vdc                     = "{{.Vdc}}"
  name                    = "{{.EdgeGatewayVcd}}"
  description             = "Description"
  configuration           = "compact"
  default_gateway_network = "${vcd_external_network.{{.NewExternalNetwork}}.name}"
  advanced                = {{.Advanced}}
  external_networks       = [ "{{.ExternalNetwork}}", "${vcd_external_network.{{.NewExternalNetwork}}.name}" ]
}
`

const testAccEdgeGatewayComplexWithLb = testAccEdgeGatewayComplexNetwork + `
resource "vcd_edgegateway" "{{.EdgeGateway}}" {
  org                     = "{{.Org}}"
  vdc                     = "{{.Vdc}}"
  name                    = "{{.EdgeGatewayVcd}}"
  description             = "Description"
  configuration           = "compact"
  default_gateway_network = "${vcd_external_network.{{.NewExternalNetwork}}.name}"
  advanced                = {{.Advanced}}
  external_networks       = [ "{{.ExternalNetwork}}", "${vcd_external_network.{{.NewExternalNetwork}}.name}" ]

  lb_enabled              = "true"
  lb_acceleration_enabled = "true"
  lb_logging_enabled      = "true"
  lb_loglevel             = "critical"
}
`
