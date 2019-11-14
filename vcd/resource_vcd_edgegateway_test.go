// +build gateway ALL functional

package vcd

import (
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var (
	edgeGatewayNameBasic   string = "TestEdgeGatewayBasic"
	edgeGatewayNameComplex string = "TestEdgeGatewayComplex"
	// ipV4Regex matches any IP like format x.x.x.x and can be used to check if a returned value
	// resembles an IP address
	ipV4Regex = regexp.MustCompile(`^(?:\d+\.){3}\d+$`)
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
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdEdgeGatewayDestroyBasic,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vcd_edgegateway."+edgeGatewayNameBasic, "default_gateway_network", testConfig.Networking.ExternalNetwork),
					resource.TestMatchResourceAttr("vcd_edgegateway."+edgeGatewayNameBasic, "default_external_network_ip", ipV4Regex),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_edgegateway." + edgeGatewayNameBasic + "-import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgVdcObject(testConfig, edgeGatewayVcdName),
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
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccEdgeGatewayComplexWithFw, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText2)

	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(testAccEdgeGatewayComplexEnableFwLbOnCreate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText4)

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
					// Expect default load balancer settings when the fields are not set
					resource.TestCheckResourceAttr("vcd_edgegateway."+edgeGatewayNameComplex, "lb_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_edgegateway."+edgeGatewayNameComplex, "lb_acceleration_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_edgegateway."+edgeGatewayNameComplex, "lb_logging_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_edgegateway."+edgeGatewayNameComplex, "lb_loglevel", "info"),
					resource.TestCheckResourceAttr("vcd_edgegateway."+edgeGatewayNameComplex, "default_external_network_ip", "192.168.30.51"),

					resourceFieldsEqual("vcd_edgegateway."+edgeGatewayNameComplex, "data.vcd_edgegateway.edge", []string{}),
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
					resource.TestCheckResourceAttr("vcd_edgegateway."+edgeGatewayNameComplex, "default_external_network_ip", "192.168.30.51"),
				),
			},
			resource.TestStep{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vcd_edgegateway."+edgeGatewayNameComplex, "default_gateway_network", newExternalNetworkVcd),
					resource.TestCheckResourceAttr("vcd_edgegateway."+edgeGatewayNameComplex, "fw_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_edgegateway."+edgeGatewayNameComplex, "fw_default_rule_logging_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_edgegateway."+edgeGatewayNameComplex, "fw_default_rule_action", "accept"),
					resource.TestCheckResourceAttr("vcd_edgegateway."+edgeGatewayNameComplex, "default_external_network_ip", "192.168.30.51"),
				),
			},
			resource.TestStep{ // step3
				ResourceName:      "vcd_edgegateway." + edgeGatewayNameComplex + "-import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgVdcObject(testConfig, edgeGatewayVcdName),
			},
			resource.TestStep{
				Config: configText4,
				// Taint the resource to force recreation of edge gateway because this step
				// attempts to test creation problems.
				Taint: []string{"vcd_edgegateway." + edgeGatewayNameComplex},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vcd_edgegateway."+edgeGatewayNameComplex, "default_gateway_network", newExternalNetworkVcd),
					resource.TestCheckResourceAttr("vcd_edgegateway."+edgeGatewayNameComplex, "lb_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_edgegateway."+edgeGatewayNameComplex, "fw_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_edgegateway."+edgeGatewayNameComplex, "default_external_network_ip", "192.168.30.51"),
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

		_, err = vdc.GetEdgeGatewayByName(wantedEgwName, true)
		if err == nil {
			return fmt.Errorf("edge gateway %s was not removed", wantedEgwName)
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

func TestAccVcdEdgeGatewayExternalNetworks(t *testing.T) {
	var (
		edgeGatewayVcdName    string = "test_edge_gateway_networks"
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
	configText := templateFill(testAccEdgeGatewayNetworks, params)
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
		CheckDestroy: stateDumper(),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vcd_edgegateway.egw", "name", "my-egw"),
					stateDumper(),
					sleepTester(),
				),
			},
		},
	})
}

func stateDumper() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		spew.Dump(s)
		return nil
	}
}

func sleepTester() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fmt.Println("sleeping")
		time.Sleep(1 * time.Minute)
		return nil
	}
}

const testAccEdgeGatewayNetworks = `
resource "vcd_edgegateway" "egw" {
  
	name                    = "my-egw"
	description             = "new edge gateway"
	configuration           = "compact"
	# default_gateway_network = "my-ext-net1"
	# external_networks       = [ "my-ext-net1", "my-ext-net2" ]
	advanced                = true
  
	fips_mode_enabled = false
	use_default_route_for_dns_relay = true
  
	external_network {
	  name = "my-ext-net"
	  enable_rate_limit = true
	  incoming_rate_limit = 100
	  outgoing_rate_limit = 100
  
	  subnet {
		ip_address = "192.168.30.52"
		gateway = "192.168.30.49"
		netmask = "255.255.255.240"
		use_for_default_route = true
	  }
  
	  # subnet {
	  #   ip_address = "192.168.31.52"
	  #   gateway = "192.168.31.49"
	  #   netmask = "255.255.255.240"
	  #   use_for_default_route = true
	  # }
	  
	}
  
  }
`

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

data "vcd_edgegateway" "edge" {
  org                     = "{{.Org}}"
  vdc                     = "{{.Vdc}}"

  name = "${vcd_edgegateway.{{.EdgeGateway}}.name}"
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

const testAccEdgeGatewayComplexWithFw = testAccEdgeGatewayComplexNetwork + `
resource "vcd_edgegateway" "{{.EdgeGateway}}" {
  org                     = "{{.Org}}"
  vdc                     = "{{.Vdc}}"
  name                    = "{{.EdgeGatewayVcd}}"
  description             = "Description"
  configuration           = "compact"
  default_gateway_network = "${vcd_external_network.{{.NewExternalNetwork}}.name}"
  advanced                = {{.Advanced}}
  external_networks       = [ "{{.ExternalNetwork}}", "${vcd_external_network.{{.NewExternalNetwork}}.name}" ]

  fw_enabled                      = "true"
  fw_default_rule_logging_enabled = "true"
  fw_default_rule_action          = "accept"
}
`

const testAccEdgeGatewayComplexEnableFwLbOnCreate = testAccEdgeGatewayComplexNetwork + `
resource "vcd_edgegateway" "{{.EdgeGateway}}" {
  org                     = "{{.Org}}"
  vdc                     = "{{.Vdc}}"
  name                    = "{{.EdgeGatewayVcd}}"
  description             = "Description"
  configuration           = "compact"
  default_gateway_network = "${vcd_external_network.{{.NewExternalNetwork}}.name}"
  advanced                = {{.Advanced}}
  external_networks       = [ "{{.ExternalNetwork}}", "${vcd_external_network.{{.NewExternalNetwork}}.name}" ]

  fw_enabled = "true"
  lb_enabled = "true"
}
`
