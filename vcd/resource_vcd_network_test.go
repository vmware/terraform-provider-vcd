// +build network ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

type networkDef struct {
	name            string
	gateway         string
	startIpAddress  string
	endIpAddress    string
	startIpAddress2 string
	endIpAddress2   string
	externalNetwork string
	configText      string
	resourceName    string
}

const (
	isolatedStaticNetwork string = "TestAccVcdNetworkIsoStatic"
	isolatedDhcpNetwork   string = "TestAccVcdNetworkIsoDhcp"
	isolatedMixedNetwork  string = "TestAccVcdNetworkIsoMixed"
	routedNetworkStatic   string = "TestAccVcdNetworkRoutedStatic"
	routedNetworkDhcp     string = "TestAccVcdNetworkRoutedDhcp"
	routedNetworkMixed    string = "TestAccVcdNetworkRoutedMixed"
	directNetwork         string = "TestAccVcdNetworkDirect"
)

func TestAccVcdNetworkIsolatedStatic(t *testing.T) {
	var def = networkDef{
		name:           isolatedStaticNetwork,
		gateway:        "192.168.2.1",
		startIpAddress: "192.168.2.2",
		endIpAddress:   "192.168.2.50",
		configText:     testAccCheckVcdNetworkIsolatedStatic,
		resourceName:   "vcd_network_isolated",
	}

	runTest(def, t)
}

func TestAccVcdNetworkIsolatedDhcp(t *testing.T) {
	var def = networkDef{
		name:           isolatedDhcpNetwork,
		gateway:        "192.168.2.1",
		startIpAddress: "192.168.2.51",
		endIpAddress:   "192.168.2.100",
		configText:     testAccCheckVcdNetworkIsolatedDhcp,
		resourceName:   "vcd_network_isolated",
	}
	runTest(def, t)
}

func TestAccVcdNetworkIsolatedMixed(t *testing.T) {
	var def = networkDef{
		name:            isolatedMixedNetwork,
		gateway:         "192.168.2.1",
		startIpAddress:  "192.168.2.2",
		endIpAddress:    "192.168.2.50",
		startIpAddress2: "192.168.2.51",
		endIpAddress2:   "192.168.2.100",
		configText:      testAccCheckVcdNetworkIsolatedMixed,
		resourceName:    "vcd_network_isolated",
	}
	runTest(def, t)
}

func TestAccVcdNetworkRoutedStatic(t *testing.T) {
	var def = networkDef{
		name:           routedNetworkStatic,
		gateway:        "10.10.102.1",
		startIpAddress: "10.10.102.2",
		endIpAddress:   "10.10.102.50",
		configText:     testAccCheckVcdNetworkRoutedStatic,
		resourceName:   "vcd_network_routed",
	}
	runTest(def, t)
}

func TestAccVcdNetworkRoutedDhcp(t *testing.T) {
	var def = networkDef{
		name:           routedNetworkDhcp,
		gateway:        "10.10.102.1",
		startIpAddress: "10.10.102.51",
		endIpAddress:   "10.10.102.100",
		configText:     testAccCheckVcdNetworkRoutedDhcp,
		resourceName:   "vcd_network_routed",
	}
	runTest(def, t)
}

func TestAccVcdNetworkRoutedMixed(t *testing.T) {

	var def = networkDef{
		name:            routedNetworkMixed,
		gateway:         "10.10.102.1",
		startIpAddress:  "10.10.102.2",
		endIpAddress:    "10.10.102.50",
		startIpAddress2: "10.10.102.51",
		endIpAddress2:   "10.10.102.100",
		configText:      testAccCheckVcdNetworkRoutedMixed,
		resourceName:    "vcd_network_routed",
	}
	runTest(def, t)
}

func TestAccVcdNetworkDirect(t *testing.T) {
	if !usingSysAdmin() {
		t.Skip("TestAccVcdNetworkDirect requires system admin privileges")
		return
	}

	var def = networkDef{
		name:            directNetwork,
		externalNetwork: testConfig.Networking.ExternalNetwork,
		configText:      testAccCheckVcdNetworkDirect,
		resourceName:    "vcd_network_direct",
	}
	runTest(def, t)
}

func runTest(def networkDef, t *testing.T) {

	generatedHrefRegexp := regexp.MustCompile("^https://")

	networkName := def.name
	var params = StringMap{
		"Org":             testConfig.VCD.Org,
		"Vdc":             testConfig.VCD.Vdc,
		"EdgeGateway":     testConfig.Networking.EdgeGateway,
		"NetworkName":     networkName,
		"Gateway":         def.gateway,
		"StartIpAddress":  def.startIpAddress,
		"EndIpAddress":    def.endIpAddress,
		"StartIpAddress2": def.startIpAddress2,
		"EndIpAddress2":   def.endIpAddress2,
		"ExternalNetwork": def.externalNetwork,
		"FuncName":        networkName,
		"Tags":            "network",
	}
	var network govcd.OrgVDCNetwork
	configText := templateFill(def.configText, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	// steps for external network
	var steps []resource.TestStep

	switch def.name {
	case directNetwork:
		steps = []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdNetworkExists(networkName, &network),
					testAccCheckVcdNetworkAttributes(networkName, &network),
					resource.TestCheckResourceAttr(
						def.resourceName+"."+networkName, "name", networkName),
					resource.TestMatchResourceAttr(
						def.resourceName+"."+networkName, "href", generatedHrefRegexp),
				),
			},
		}
	case routedNetworkStatic, routedNetworkMixed, isolatedStaticNetwork, isolatedMixedNetwork:
		steps = []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdNetworkExists(networkName, &network),
					testAccCheckVcdNetworkAttributes(networkName, &network),
					resource.TestCheckResourceAttr(
						def.resourceName+"."+networkName, "name", networkName),
					resource.TestCheckResourceAttr(
						def.resourceName+"."+networkName, "static_ip_pool.#", "1"),
					resource.TestCheckResourceAttr(
						def.resourceName+"."+networkName, "static_ip_pool.0.start_address", def.startIpAddress),
					resource.TestCheckResourceAttr(
						def.resourceName+"."+networkName, "static_ip_pool.0.end_address", def.endIpAddress),
					resource.TestCheckResourceAttr(
						def.resourceName+"."+networkName, "gateway", def.gateway),
					resource.TestMatchResourceAttr(
						def.resourceName+"."+networkName, "href", generatedHrefRegexp),
				),
			},
		}
	case isolatedDhcpNetwork, routedNetworkDhcp:
		steps = []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdNetworkExists(networkName, &network),
					testAccCheckVcdNetworkAttributes(networkName, &network),
					resource.TestCheckResourceAttr(
						def.resourceName+"."+networkName, "name", networkName),
					resource.TestCheckResourceAttr(
						def.resourceName+"."+networkName, "dhcp_pool.#", "1"),
					resource.TestCheckResourceAttr(
						def.resourceName+"."+networkName, "dhcp_pool.0.start_address", def.startIpAddress),
					resource.TestCheckResourceAttr(
						def.resourceName+"."+networkName, "dhcp_pool.0.end_address", def.endIpAddress),
					resource.TestCheckResourceAttr(
						def.resourceName+"."+networkName, "gateway", def.gateway),
					resource.TestMatchResourceAttr(
						def.resourceName+"."+networkName, "href", generatedHrefRegexp),
				),
			},
		}

	}

	steps = append(steps, resource.TestStep{
		ResourceName:      def.resourceName + "." + networkName + "-import",
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateIdFunc: importStateIdByNetwork(testConfig, networkName),
	})

	// Don't convert this test to parallel, as it will cause IP ranges conflicts
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: func(s *terraform.State) error { return testAccCheckVcdNetworkDestroy(s, def.resourceName, networkName) },
		Steps:        steps,
	})
}

func importStateIdByNetwork(vcd TestConfig, objectName string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		if testConfig.VCD.Org == "" || testConfig.VCD.Vdc == "" || objectName == "" {
			return "", fmt.Errorf("missing information to generate import path")
		}
		return testConfig.VCD.Org + "." + testConfig.VCD.Vdc + "." + objectName, nil
	}
}

func testAccCheckVcdNetworkExists(name string, network *govcd.OrgVDCNetwork) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*VCDClient)

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.VCD.Vdc, testConfig.VCD.Org, err)
		}

		orgVDCNetwork, err := vdc.GetOrgVdcNetworkByName(name, false)
		if err != nil {
			return fmt.Errorf("network %s does not exist ", name)
		}

		*network = *orgVDCNetwork

		return nil
	}
}

func testAccCheckVcdNetworkDestroy(s *terraform.State, networkType string, networkName string) error {
	conn := testAccProvider.Meta().(*VCDClient)

	_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
	if err != nil {
		return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.VCD.Vdc, testConfig.VCD.Org, err)
	}

	_, err = vdc.GetOrgVdcNetworkByNameOrId(networkName, false)
	if err == nil {
		return fmt.Errorf("network %s still exists", networkName)
	}

	return nil
}

func testAccCheckVcdNetworkAttributes(name string, network *govcd.OrgVDCNetwork) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if network.OrgVDCNetwork.Name != name {
			return fmt.Errorf("bad name: %s", network.OrgVDCNetwork.Name)
		}

		return nil
	}
}

func init() {
	testingTags["network"] = "resource_vcd_network_test.go"
}

const testAccCheckVcdNetworkIsolatedStatic = `
resource "vcd_network_isolated" "{{.NetworkName}}" {
  name       = "{{.NetworkName}}"
  org        = "{{.Org}}"
  vdc        = "{{.Vdc}}"
  gateway    = "{{.Gateway}}"
  dns1       = "192.168.2.1"
  static_ip_pool {
    start_address = "{{.StartIpAddress}}"
    end_address   = "{{.EndIpAddress}}"
  }
}
`

const testAccCheckVcdNetworkIsolatedDhcp = `
resource "vcd_network_isolated" "{{.NetworkName}}" {
  name       = "{{.NetworkName}}"
  org        = "{{.Org}}"
  vdc        = "{{.Vdc}}"
  gateway    = "{{.Gateway}}"
  dns1       = "192.168.2.1"
  dhcp_pool {
    start_address = "{{.StartIpAddress}}"
    end_address   = "{{.EndIpAddress}}"
  }
}
`

const testAccCheckVcdNetworkIsolatedMixed = `
resource "vcd_network_isolated" "{{.NetworkName}}" {
  name       = "{{.NetworkName}}"
  org        = "{{.Org}}"
  vdc        = "{{.Vdc}}"
  gateway    = "{{.Gateway}}"
  dns1       = "192.168.2.1"
  static_ip_pool {
    start_address = "{{.StartIpAddress}}"
    end_address   = "{{.EndIpAddress}}"
  }
  dhcp_pool {
    start_address = "{{.StartIpAddress2}}"
    end_address   = "{{.EndIpAddress2}}"
  }
}
`

const testAccCheckVcdNetworkDirect = `
resource "vcd_network_direct" "{{.NetworkName}}" {
  name             = "{{.NetworkName}}"
  org              = "{{.Org}}"
  vdc              = "{{.Vdc}}"
  external_network = "{{.ExternalNetwork}}"
}
`

const testAccCheckVcdNetworkRoutedStatic = `
resource "vcd_network_routed" "{{.NetworkName}}" {
  name         = "{{.NetworkName}}"
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "{{.Gateway}}"

  static_ip_pool {
    start_address = "{{.StartIpAddress}}"
    end_address   = "{{.EndIpAddress}}"
  }
}
`

const testAccCheckVcdNetworkRoutedDhcp = `
resource "vcd_network_routed" "{{.NetworkName}}" {
  name         = "{{.NetworkName}}"
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "{{.Gateway}}"

  dhcp_pool {
    start_address = "{{.StartIpAddress}}"
    end_address   = "{{.EndIpAddress}}"
  }
}
`

const testAccCheckVcdNetworkRoutedMixed = `
resource "vcd_network_routed" "{{.NetworkName}}" {
  name         = "{{.NetworkName}}"
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "{{.Gateway}}"

  static_ip_pool {
    start_address = "{{.StartIpAddress}}"
    end_address   = "{{.EndIpAddress}}"
  }

  dhcp_pool {
    start_address = "{{.StartIpAddress2}}"
    end_address   = "{{.EndIpAddress2}}"
  }
}
`
