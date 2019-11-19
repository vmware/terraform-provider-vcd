// +build gateway ALL functional

package vcd

import (
	"fmt"
	"os"
	"regexp"
	"testing"

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
				ResourceName:            "vcd_edgegateway." + edgeGatewayNameBasic + "-import",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdOrgVdcObject(testConfig, edgeGatewayVcdName),
				ImportStateVerifyIgnore: []string{"external_network", "external_networks"},
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

					resourceFieldsEqual("vcd_edgegateway."+edgeGatewayNameComplex, "data.vcd_edgegateway.edge", []string{"external_network.#"}),
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
				// import only imports values in 'external_network' block, while this test uses
				// older 'external_networks therefore we need to skip validation of these fields
				ImportStateVerifyIgnore: []string{"external_networks", "external_network"},
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

	// params["FuncName"] = t.Name() + "-step2"
	// configText2 := templateFill(testAccEdgeGatewayNetworks2, params)

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
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		// CheckDestroy: stateDumper(),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "name", "edge-with-complex-networks"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "advanced", "true"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "description", "new edge gateway"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "configuration", "compact"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "advanced", "true"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "distributed_routing", "true"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "fips_mode_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "use_default_route_for_dns_relay", "true"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "default_gateway_network", newExternalNetworkVcd),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "default_external_network_ip", "192.168.30.51"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "external_network.#", "2"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "external_network.2798412847.enable_rate_limit", "true"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "external_network.2798412847.incoming_rate_limit", "100"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "external_network.2798412847.name", "test_external_network"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "external_network.2798412847.outgoing_rate_limit", "100"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "external_network.2798412847.subnet.#", "1"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "external_network.2798412847.subnet.3598571839.gateway", "192.168.30.49"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "external_network.2798412847.subnet.3598571839.ip_address", "192.168.30.51"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "external_network.2798412847.subnet.3598571839.netmask", "255.255.255.240"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "external_network.2798412847.subnet.3598571839.use_for_default_route", "true"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "external_network.2798412847.subnet.3598571839.suballocate_pool.#", "2"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "external_network.2798412847.subnet.3598571839.suballocate_pool.3548736268.end_address", "192.168.30.55"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "external_network.2798412847.subnet.3598571839.suballocate_pool.3548736268.start_address", "192.168.30.53"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "external_network.2798412847.subnet.3598571839.suballocate_pool.4005225628.end_address", "192.168.30.60"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "external_network.2798412847.subnet.3598571839.suballocate_pool.4005225628.start_address", "192.168.30.58"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "external_network_ips.#", "2"),
					resource.TestMatchResourceAttr("vcd_edgegateway.egw", "external_network_ips.0", ipV4Regex),
					resource.TestMatchResourceAttr("vcd_edgegateway.egw", "external_network_ips.1", ipV4Regex),

					// TODO after
					// https://github.com/terraform-providers/terraform-provider-aws/issues/7198
					// Data source checks. There is a bug in Terraform where a data source cannot
					// have two computed TypeSet variables because they get overwriten The test
					// below is left such, that it triggers an error as soon as the bug is fixed.
					// (probably when we pull in newer SDK)
					resource.TestCheckResourceAttr("data.vcd_edgegateway.egw", "external_network.#", "1"),

					// Working data source tests
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "name", "data.vcd_edgegateway.egw", "name"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "external_network_ips.#", "data.vcd_edgegateway.egw", "external_network_ips.#"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "external_network_ips.0", "data.vcd_edgegateway.egw", "external_network_ips.0"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "external_network_ips.1", "data.vcd_edgegateway.egw", "external_network_ips.1"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "fips_mode_enabled", "data.vcd_edgegateway.egw", "fips_mode_enabled"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "use_default_route_for_dns_relay", "data.vcd_edgegateway.egw", "use_default_route_for_dns_relay"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "advanced", "data.vcd_edgegateway.egw", "advanced"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "distributed_routing", "data.vcd_edgegateway.egw", "distributed_routing"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "configuration", "data.vcd_edgegateway.egw", "configuration"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "default_external_network_ip", "data.vcd_edgegateway.egw", "default_external_network_ip"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "default_external_network_ip", "data.vcd_edgegateway.egw", "default_external_network_ip"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "default_gateway_network", "data.vcd_edgegateway.egw", "default_gateway_network"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "description", "data.vcd_edgegateway.egw", "description"),
					resource.TestCheckResourceAttr("data.vcd_edgegateway.egw", "external_networks.#", "2"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "fw_default_rule_action", "data.vcd_edgegateway.egw", "fw_default_rule_action"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "fw_default_rule_logging_enabled", "data.vcd_edgegateway.egw", "fw_default_rule_logging_enabled"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "fw_enabled", "data.vcd_edgegateway.egw", "fw_enabled"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "ha_enabled", "data.vcd_edgegateway.egw", "ha_enabled"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "lb_acceleration_enabled", "data.vcd_edgegateway.egw", "lb_acceleration_enabled"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "lb_enabled", "data.vcd_edgegateway.egw", "lb_enabled"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "lb_logging_enabled", "data.vcd_edgegateway.egw", "lb_logging_enabled"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "lb_loglevel", "data.vcd_edgegateway.egw", "lb_loglevel"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "use_default_route_for_dns_relay", "data.vcd_edgegateway.egw", "use_default_route_for_dns_relay"),
				),
			},
		},
	})
}

// TestAccVcdEdgeGatewayParallelCreation attaches multiple edge gateways to the same external
// network as it was reported that edge gateways step on each other while trying to attach to the
// same external network. If this test ever fails then it means locks have to be used on external
// networks.
func TestAccVcdEdgeGatewayParallelCreation(t *testing.T) {
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
	configText := templateFill(testAccEdgeGatewayParallel, params)

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
					resource.TestCheckResourceAttr("vcd_edgegateway.egw.0", "name", "parallel-0"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw.1", "name", "parallel-1"),
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

// func sleepTester() resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		fmt.Println("sleeping")
// 		time.Sleep(1 * time.Minute)
// 		return nil
// 	}
// }

const testAccEdgeGatewayNetworks = testAccEdgeGatewayComplexNetwork + `
resource "vcd_edgegateway" "egw" {
	org                     = "{{.Org}}"
	vdc                     = "{{.Vdc}}"

	name                    = "edge-with-complex-networks"
	description             = "new edge gateway"
	configuration           = "compact"
	advanced                = true
  
	fips_mode_enabled               = false
	use_default_route_for_dns_relay = true
	distributed_routing             = true
  
	external_network {
	  name = "${vcd_external_network.{{.NewExternalNetwork}}.name}"
	  enable_rate_limit = true
	  incoming_rate_limit = 100
	  outgoing_rate_limit = 100
  
	  subnet {
		ip_address = "192.168.30.51"
		gateway = "192.168.30.49"
		netmask = "255.255.255.240"
		use_for_default_route = true

		suballocate_pool {
			start_address = "192.168.30.53"
			end_address   = "192.168.30.55"
		}

		suballocate_pool {
			start_address = "192.168.30.58"
			end_address   = "192.168.30.60"
		}
	  }
	}

	# Attach to existing external network
	external_network {
	  name = "${data.vcd_external_network.ds-network.name}"

		subnet {
			# ip_address is skipped here on purpose to get dynamic IP
			use_for_default_route = false
			gateway = "${data.vcd_external_network.ds-network.ip_scope[0].gateway}"
			netmask = "${data.vcd_external_network.ds-network.ip_scope[0].netmask}"
	}
  }
}

data "vcd_edgegateway" "egw" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"
	
  name = vcd_edgegateway.egw.name
}

# Use data source of existing external network to get needed gateway and netmask
# for subnet participation details
data "vcd_external_network" "ds-network" {
	name = "{{.ExternalNetwork}}"
}
`

const testAccEdgeGatewayParallel = testAccEdgeGatewayComplexNetwork + `
resource "vcd_edgegateway" "egw" {
	count = 2

	org                     = "{{.Org}}"
	vdc                     = "{{.Vdc}}"

	name                    = "parallel-${count.index}"
	configuration           = "compact"
	advanced                = true

	external_network {
	  name = "${vcd_external_network.{{.NewExternalNetwork}}.name}"
	  subnet {
		gateway = "192.168.30.49"
		netmask = "255.255.255.240"
	  }
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

// TODO external network has a bug that it uses a TypeList for `ip_scope` field. If the below two
// network has second ip_scope defined - then vCD API orders them differently and a replacement is
// suggested.
// GitHUB issue - https://github.com/terraform-providers/terraform-provider-vcd/issues/371
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
  
#  ip_scope {
# 	gateway      = "192.168.40.149"
# 	netmask      = "255.255.255.0"
# 	dns1         = "192.168.0.164"
# 	dns2         = "192.168.0.196"
# 	dns_suffix   = "company.biz"

# 	static_ip_pool {
# 	  start_address = "192.168.40.151"
# 	  end_address   = "192.168.40.162"
# 	}
#   }

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
