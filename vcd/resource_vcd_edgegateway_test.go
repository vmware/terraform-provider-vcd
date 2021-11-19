//go:build gateway || ALL || functional
// +build gateway ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

var (
	edgeGatewayNameBasic   string = "TestEdgeGatewayBasic"
	edgeGatewayNameComplex string = "TestEdgeGatewayComplex"
	// ipV4Regex matches any IP like format x.x.x.x and can be used to check if a returned value
	// resembles an IP address
	ipV4Regex = regexp.MustCompile(`^(?:\d+\.){3}\d+$`)
)

func TestAccVcdEdgeGatewayBasic(t *testing.T) {
	preTestChecks(t)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	var (
		edgeGatewayVcdName    string = "test_edge_gateway_basic"
		newExternalNetwork    string = "TestExternalNetwork"
		newExternalNetworkVcd string = "test_external_network"
	)

	// String map to fill the template
	var params = StringMap{
		"Org":                   testConfig.VCD.Org,
		"Vdc":                   testConfig.VCD.Vdc,
		"EdgeGateway":           edgeGatewayNameBasic,
		"EdgeGatewayVcd":        edgeGatewayVcdName,
		"ExternalNetwork":       testConfig.Networking.ExternalNetwork,
		"Advanced":              "true",
		"Tags":                  "gateway",
		"NewExternalNetwork":    newExternalNetwork,
		"NewExternalNetworkVcd": newExternalNetworkVcd,
		"Type":                  testConfig.Networking.ExternalNetworkPortGroupType,
		"PortGroup":             testConfig.Networking.ExternalNetworkPortGroup,
		"Vcenter":               testConfig.Networking.Vcenter,
	}
	configText := templateFill(testAccEdgeGatewayBasic, params)

	if !usingSysAdmin() {
		t.Skip("Edge Gateway tests require system admin privileges")
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdEdgeGatewayDestroy(edgeGatewayNameBasic),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_edgegateway."+edgeGatewayNameBasic, "default_external_network_ip", ipV4Regex),
				),
			},
			resource.TestStep{
				ResourceName:            "vcd_edgegateway." + edgeGatewayNameBasic,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdOrgVdcObject(testConfig, edgeGatewayVcdName),
				ImportStateVerifyIgnore: []string{"external_network", "external_networks"},
			},
		},
	})
	postTestChecks(t)
}

func TestAccVcdEdgeGatewayComplex(t *testing.T) {
	preTestChecks(t)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	var (
		edgeGatewayVcdName    string = "test_edge_gateway_basic"
		newExternalNetwork    string = "TestExternalNetwork"
		newExternalNetworkVcd string = "test_external_network"
	)

	// String map to fill the template
	var params = StringMap{
		"Org":                   testConfig.VCD.Org,
		"Vdc":                   testConfig.VCD.Vdc,
		"EdgeGateway":           edgeGatewayNameBasic,
		"EdgeGatewayVcd":        edgeGatewayVcdName,
		"ExternalNetwork":       testConfig.Networking.ExternalNetwork,
		"Tags":                  "gateway",
		"NewExternalNetwork":    newExternalNetwork,
		"NewExternalNetworkVcd": newExternalNetworkVcd,
		"Type":                  testConfig.Networking.ExternalNetworkPortGroupType,
		"PortGroup":             testConfig.Networking.ExternalNetworkPortGroup,
		"Vcenter":               testConfig.Networking.Vcenter,
	}
	configText := templateFill(testAccEdgeGatewayBasic, params)

	if !usingSysAdmin() {
		t.Skip("Edge gateway tests requires system admin privileges")
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdEdgeGatewayDestroy(edgeGatewayNameBasic),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_edgegateway."+edgeGatewayNameBasic, "default_external_network_ip", ipV4Regex),
				),
			},
			resource.TestStep{
				ResourceName:            "vcd_edgegateway." + edgeGatewayNameBasic,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdOrgVdcObject(testConfig, edgeGatewayVcdName),
				ImportStateVerifyIgnore: []string{"external_network"},
			},
		},
	})
	postTestChecks(t)
}

func testAccCheckVcdEdgeGatewayDestroy(edgeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		for _, rs := range s.RootModule().Resources {
			edgeGatewayName := rs.Primary.Attributes["name"]
			if rs.Type != "vcd_edgegateway" {
				continue
			}
			if edgeGatewayName != edgeName {
				continue
			}
			conn := testAccProvider.Meta().(*VCDClient)
			orgName := rs.Primary.Attributes["org"]
			vdcName := rs.Primary.Attributes["vdc"]

			_, vdc, err := conn.GetOrgAndVdc(orgName, vdcName)
			if err != nil {
				return fmt.Errorf("error retrieving org %s and vdc %s : %s ", orgName, vdcName, err)
			}

			_, err = vdc.GetEdgeGatewayByName(edgeName, true)
			if err == nil {
				return fmt.Errorf("edge gateway %s was not removed", edgeName)
			}
		}

		return nil
	}
}

func TestAccVcdEdgeGatewayExternalNetworks(t *testing.T) {
	preTestChecks(t)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
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
		"Vcenter":               testConfig.Networking.Vcenter,
	}
	configText := templateFill(testAccEdgeGatewayNetworks, params)

	params["FuncName"] = t.Name() + "-step2"
	configText1 := templateFill(testAccEdgeGatewayNetworks2, params)

	if !usingSysAdmin() {
		t.Skip("Edge gateway tests requires system admin privileges")
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdEdgeGatewayDestroy("edge-with-complex-networks"),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "name", "edge-with-complex-networks"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "description", "new edge gateway"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "configuration", "compact"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "distributed_routing", "false"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "fips_mode_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "use_default_route_for_dns_relay", "true"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "default_external_network_ip", "192.168.30.51"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "external_network.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_edgegateway.egw", "external_network.*", map[string]string{
						"enable_rate_limit":   "false",
						"incoming_rate_limit": "0",
						"name":                testConfig.Networking.ExternalNetwork,
						"outgoing_rate_limit": "0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_edgegateway.egw", "external_network.*", map[string]string{
						"enable_rate_limit":   "false",
						"incoming_rate_limit": "0",
						"name":                "test_external_network",
						"outgoing_rate_limit": "0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_edgegateway.egw", "external_network.*.subnet.*", map[string]string{
						"gateway":               "10.150.191.253",
						"ip_address":            "",
						"netmask":               "255.255.255.240",
						"use_for_default_route": "false",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_edgegateway.egw", "external_network.*.subnet.*", map[string]string{
						"gateway":               "192.168.30.49",
						"ip_address":            "192.168.30.51",
						"netmask":               "255.255.255.240",
						"use_for_default_route": "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_edgegateway.egw", "external_network.1.subnet.0.suballocate_pool.*", map[string]string{
						"start_address": "192.168.30.53",
						"end_address":   "192.168.30.55",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_edgegateway.egw", "external_network.1.subnet.0.suballocate_pool.*", map[string]string{
						"start_address": "192.168.30.58",
						"end_address":   "192.168.30.60",
					}),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "external_network_ips.#", "2"),
					resource.TestMatchResourceAttr("vcd_edgegateway.egw", "external_network_ips.0", ipV4Regex),
					resource.TestMatchResourceAttr("vcd_edgegateway.egw", "external_network_ips.1", ipV4Regex),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "lb_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "lb_acceleration_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "lb_logging_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "lb_loglevel", "critical"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "fw_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "fw_default_rule_logging_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "fw_default_rule_action", "accept"),

					resource.TestCheckResourceAttr("data.vcd_edgegateway.egw", "external_network.#", "2"),
					// Working data source tests
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "name", "data.vcd_edgegateway.egw", "name"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "external_network_ips.#", "data.vcd_edgegateway.egw", "external_network_ips.#"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "external_network_ips.0", "data.vcd_edgegateway.egw", "external_network_ips.0"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "external_network_ips.1", "data.vcd_edgegateway.egw", "external_network_ips.1"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "fips_mode_enabled", "data.vcd_edgegateway.egw", "fips_mode_enabled"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "use_default_route_for_dns_relay", "data.vcd_edgegateway.egw", "use_default_route_for_dns_relay"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "distributed_routing", "data.vcd_edgegateway.egw", "distributed_routing"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "configuration", "data.vcd_edgegateway.egw", "configuration"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "default_external_network_ip", "data.vcd_edgegateway.egw", "default_external_network_ip"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "default_external_network_ip", "data.vcd_edgegateway.egw", "default_external_network_ip"),
					resource.TestCheckResourceAttrPair("vcd_edgegateway.egw", "description", "data.vcd_edgegateway.egw", "description"),
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
			resource.TestStep{
				Taint:  []string{"vcd_edgegateway.egw"},
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "name", "simple-edge-with-complex-networks"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "configuration", "compact"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "distributed_routing", "false"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "fips_mode_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "ha_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "description", ""),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "external_network.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_edgegateway.egw", "external_network.*", map[string]string{
						"enable_rate_limit":   "false",
						"incoming_rate_limit": "0",
						"name":                "test_external_network",
						"outgoing_rate_limit": "0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_edgegateway.egw", "external_network.*.subnet.*", map[string]string{
						"gateway":               "192.168.30.49",
						"ip_address":            "",
						"netmask":               "255.255.255.240",
						"use_for_default_route": "true",
					}),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "external_network_ips.#", "1"),
					resource.TestMatchResourceAttr("vcd_edgegateway.egw", "external_network_ips.0", ipV4Regex),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "fw_default_rule_action", "deny"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "fw_default_rule_logging_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "fw_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "lb_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "lb_acceleration_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "lb_logging_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "lb_loglevel", "info"),
				),
			},
			resource.TestStep{ // step2 - import
				ResourceName:            "vcd_edgegateway.egw",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdOrgVdcObject(testConfig, "simple-edge-with-complex-networks"),
				ImportStateVerifyIgnore: []string{"external_network"},
			},
		},
	})
	postTestChecks(t)
}

// isPortGroupDistributed  checks if portgroup is defined in Distributed or Standard vSwitch
func isPortGroupDistributed(portGroupName string) (bool, error) {
	if portGroupName == "" {
		return false, nil
	}

	// Get the data from configuration file. This client is still inactive at this point
	vcdClient, err := getTestVCDFromJson(testConfig)
	if err != nil {
		return false, fmt.Errorf("error getting client configuration: %s", err)
	}
	err = ProviderAuthenticate(vcdClient, testConfig.Provider.User, testConfig.Provider.Password, testConfig.Provider.Token, testConfig.Provider.SysOrg)
	if err != nil {
		return false, fmt.Errorf("authentication error: %s", err)
	}

	portGroupRecord, err := govcd.QueryDistributedPortGroup(vcdClient, portGroupName)
	if err != nil {
		return false, fmt.Errorf("got error while querying portgroup type: %s", err)
	}
	if len(portGroupRecord) == 1 {
		return true, nil
	}

	return false, nil
}

// TestAccVcdEdgeGatewayRateLimits focuses on testing how the `external_network` block handles
// network interface limits. It escapes quickly when the ExternalNetworkPortGroup of external
// network is of type "NETWORK" (standard switch portgroup) and only proceeds when it is of type
// "DV_PORTGROUP" (Backed by distributed switch). Only "DV_PORTGROUP" support rate limiting.
func TestAccVcdEdgeGatewayRateLimits(t *testing.T) {
	preTestChecks(t)
	isPgDistributed, err := isPortGroupDistributed(testConfig.Networking.ExternalNetworkPortGroup)
	if err != nil {
		t.Skipf("Skipping test because port group type could not be validated")
	}

	if !isPgDistributed {
		t.Skipf("Skipping test because specified portgroup is from standard vSwitch and " +
			"does not support rate limiting")
	}

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
		"Vcenter":               testConfig.Networking.Vcenter,
		"EnableRateLimit":       "true",
		"IncomingRateLimit":     "88.888",
		"OutgoingRateLimit":     "55.335",
		"Type":                  testConfig.Networking.ExternalNetworkPortGroupType,
		"PortGroup":             testConfig.Networking.ExternalNetworkPortGroup,
	}
	configText := templateFill(testAccEdgeGatewayRateLimits, params)

	params["EnableRateLimit"] = "false"
	params["IncomingRateLimit"] = "0"
	params["OutgoingRateLimit"] = "0"
	params["FuncName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccEdgeGatewayRateLimits, params)

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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdEdgeGatewayDestroy("edge-with-complex-networks"),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "name", "edge-with-rate-limits"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_edgegateway.egw", "external_network.*", map[string]string{
						"name":                "test_external_network",
						"enable_rate_limit":   "true",
						"incoming_rate_limit": "88.888",
						"outgoing_rate_limit": "55.335",
					}),
				),
			},
			resource.TestStep{
				Taint:  []string{"vcd_edgegateway.egw"},
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_edgegateway.egw", "name", "edge-with-rate-limits"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_edgegateway.egw", "external_network.*", map[string]string{
						"name":                "test_external_network",
						"enable_rate_limit":   "false",
						"incoming_rate_limit": "0",
						"outgoing_rate_limit": "0",
					}),
				),
			},
			resource.TestStep{ // step2 - import
				ResourceName:            "vcd_edgegateway.egw",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdOrgVdcObject(testConfig, "edge-with-rate-limits"),
				ImportStateVerifyIgnore: []string{"external_network"},
			},
		},
	})
	postTestChecks(t)
}

const testAccEdgeGatewayRateLimits = testAccEdgeGatewayComplexNetwork + `
resource "vcd_edgegateway" "egw" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name          = "edge-with-rate-limits"
  configuration = "compact"

  external_network {
    name                = vcd_external_network.{{.NewExternalNetwork}}.name
    enable_rate_limit   = {{.EnableRateLimit}}
    incoming_rate_limit = {{.IncomingRateLimit}}
    outgoing_rate_limit = {{.OutgoingRateLimit}}

    subnet {
      gateway               = "192.168.30.49"
      netmask               = "255.255.255.240"
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
}
`

// TestAccVcdEdgeGatewayParallelCreation attaches multiple edge gateways to the same external
// network as it was reported that edge gateways step on each other while trying to attach to the
// same external network. If this test ever fails then it means locks have to be used on external
// networks.
func TestAccVcdEdgeGatewayParallelCreation(t *testing.T) {
	preTestChecks(t)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
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
		"Vcenter":               testConfig.Networking.Vcenter,
	}
	configText := templateFill(testAccEdgeGatewayParallel, params)

	if !usingSysAdmin() {
		t.Skip("Edge gateway tests requires system admin privileges")
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVcdEdgeGatewayDestroy("parallel-0"),
			testAccCheckVcdEdgeGatewayDestroy("parallel-1"),
		),
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
	postTestChecks(t)
}

// TODO external network has a bug that it uses a TypeList for `ip_scope` field. If the below two
// network has second ip_scope defined - then vCD API orders them differently and a replacement is
// suggested.
// GitHUB issue - https://github.com/vmware/terraform-provider-vcd/issues/395
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
    gateway    = "192.168.30.49"
    netmask    = "255.255.255.240"
    dns1       = "192.168.0.164"
    dns2       = "192.168.0.196"
    dns_suffix = "company.biz"

    static_ip_pool {
      start_address = "192.168.30.51"
      end_address   = "192.168.30.62"
    }
  }

  #  ip_scope {
  # 	gateway    = "192.168.40.149"
  # 	netmask    = "255.255.255.0"
  # 	dns1       = "192.168.0.164"
  # 	dns2       = "192.168.0.196"
  # 	dns_suffix = "company.biz"

  # 	static_ip_pool {
  # 	  start_address = "192.168.40.151"
  # 	  end_address   = "192.168.40.162"
  # 	}
  #   }

  retain_net_info_across_deployments = "false"
}
`

const testAccEdgeGatewayBasic = testAccEdgeGatewayComplexNetwork + `
resource "vcd_edgegateway" "{{.EdgeGateway}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  name          = "{{.EdgeGatewayVcd}}"
  description   = "Description"
  configuration = "compact"

  external_network {
    name = vcd_external_network.{{.NewExternalNetwork}}.name

    subnet {
      ip_address            = "192.168.30.51"
      gateway               = "192.168.30.49"
      netmask               = "255.255.255.240"
      use_for_default_route = true
    }
  }
}
`

const testAccEdgeGatewayNetworks = testAccEdgeGatewayComplexNetwork + `
resource "vcd_edgegateway" "egw" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name          = "edge-with-complex-networks"
  description   = "new edge gateway"
  configuration = "compact"

  # can be only true when system setting Allow FIPS Mode is enabled
  fips_mode_enabled               = false
  use_default_route_for_dns_relay = true
  distributed_routing             = false

  lb_enabled              = "true"
  lb_acceleration_enabled = "true"
  lb_logging_enabled      = "true"
  lb_loglevel             = "critical"

  fw_enabled                      = "true"
  fw_default_rule_logging_enabled = "true"
  fw_default_rule_action          = "accept"

  external_network {
    name = vcd_external_network.{{.NewExternalNetwork}}.name

    subnet {
      ip_address            = "192.168.30.51"
      gateway               = "192.168.30.49"
      netmask               = "255.255.255.240"
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
    name = data.vcd_external_network.ds-network.name

    subnet {
      # ip_address is skipped here on purpose to get dynamic IP
      use_for_default_route = false
      gateway               = data.vcd_external_network.ds-network.ip_scope[0].gateway
      netmask               = data.vcd_external_network.ds-network.ip_scope[0].netmask
    }
  }
}

data "vcd_edgegateway" "egw" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name       = vcd_edgegateway.egw.name
  depends_on = [vcd_edgegateway.egw]
}

# Use data source of existing external network to get needed gateway and netmask
# for subnet participation details
data "vcd_external_network" "ds-network" {
  name = "{{.ExternalNetwork}}"
}
`

const testAccEdgeGatewayNetworks2 = testAccEdgeGatewayComplexNetwork + `
resource "vcd_edgegateway" "egw" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name          = "simple-edge-with-complex-networks"
  configuration = "compact"

  external_network {
    name = vcd_external_network.{{.NewExternalNetwork}}.name
    subnet {
      gateway               = "192.168.30.49"
      netmask               = "255.255.255.240"
      use_for_default_route = true
    }
  }
}
`

const testAccEdgeGatewayParallel = testAccEdgeGatewayComplexNetwork + `
resource "vcd_edgegateway" "egw" {
  count = 2

  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name          = "parallel-${count.index}"
  configuration = "compact"

  external_network {
    name = vcd_external_network.{{.NewExternalNetwork}}.name
    subnet {
      gateway = "192.168.30.49"
      netmask = "255.255.255.240"
    }
  }
}
`
