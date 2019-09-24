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
	name                  string
	gateway               string
	startStaticIpAddress1 string
	endStaticIpAddress1   string
	startStaticIpAddress2 string
	endStaticIpAddress2   string
	startDhcpIpAddress    string
	endDhcpIpAddress      string
	externalNetwork       string
	configText            string
	resourceName          string
}

const (
	isolatedStaticNetwork1 string = "TestAccVcdNetworkIsoStatic1"
	isolatedStaticNetwork2 string = "TestAccVcdNetworkIsoStatic2"
	isolatedDhcpNetwork    string = "TestAccVcdNetworkIsoDhcp"
	isolatedMixedNetwork1  string = "TestAccVcdNetworkIsoMixed1"
	isolatedMixedNetwork2  string = "TestAccVcdNetworkIsoMixed2"
	routedStaticNetwork1   string = "TestAccVcdNetworkRoutedStatic1"
	routedStaticNetwork2   string = "TestAccVcdNetworkRoutedStatic2"
	routedDhcpNetwork      string = "TestAccVcdNetworkRoutedDhcp"
	routedMixedNetwork     string = "TestAccVcdNetworkRoutedMixed"
	directNetwork          string = "TestAccVcdNetworkDirect"
)

func TestAccVcdNetworkIsolatedStatic1(t *testing.T) {
	var def = networkDef{
		name:                  isolatedStaticNetwork1,
		gateway:               "192.168.2.1",
		startStaticIpAddress1: "192.168.2.2",
		endStaticIpAddress1:   "192.168.2.50",
		configText:            testAccCheckVcdNetworkIsolatedStatic1,
		resourceName:          "vcd_network_isolated",
	}

	runTest(def, t)
}

func TestAccVcdNetworkIsolatedStatic2(t *testing.T) {
	var def = networkDef{
		name:                  isolatedStaticNetwork2,
		gateway:               "192.168.2.1",
		startStaticIpAddress1: "192.168.2.2",
		endStaticIpAddress1:   "192.168.2.50",
		startStaticIpAddress2: "192.168.2.52",
		endStaticIpAddress2:   "192.168.2.100",
		configText:            testAccCheckVcdNetworkIsolatedStatic2,
		resourceName:          "vcd_network_isolated",
	}

	runTest(def, t)
}

func TestAccVcdNetworkIsolatedDhcp(t *testing.T) {
	var def = networkDef{
		name:               isolatedDhcpNetwork,
		gateway:            "192.168.2.1",
		startDhcpIpAddress: "192.168.2.51",
		endDhcpIpAddress:   "192.168.2.100",
		configText:         testAccCheckVcdNetworkIsolatedDhcp,
		resourceName:       "vcd_network_isolated",
	}
	runTest(def, t)
}

func TestAccVcdNetworkIsolatedMixed1(t *testing.T) {
	var def = networkDef{
		name:                  isolatedMixedNetwork1,
		gateway:               "192.168.2.1",
		startStaticIpAddress1: "192.168.2.2",
		endStaticIpAddress1:   "192.168.2.50",
		startDhcpIpAddress:    "192.168.2.51",
		endDhcpIpAddress:      "192.168.2.100",
		configText:            testAccCheckVcdNetworkIsolatedMixed1,
		resourceName:          "vcd_network_isolated",
	}
	runTest(def, t)
}
func TestAccVcdNetworkIsolatedMixed2(t *testing.T) {
	var def = networkDef{
		name:                  isolatedMixedNetwork2,
		gateway:               "192.168.2.1",
		startStaticIpAddress1: "192.168.2.2",
		endStaticIpAddress1:   "192.168.2.50",
		startStaticIpAddress2: "192.168.2.52",
		endStaticIpAddress2:   "192.168.2.100",
		startDhcpIpAddress:    "192.168.2.151",
		endDhcpIpAddress:      "192.168.2.200",
		configText:            testAccCheckVcdNetworkIsolatedMixed2,
		resourceName:          "vcd_network_isolated",
	}
	runTest(def, t)
}

func TestAccVcdNetworkRoutedStatic1(t *testing.T) {
	var def = networkDef{
		name:                  routedStaticNetwork1,
		gateway:               "10.10.102.1",
		startStaticIpAddress1: "10.10.102.2",
		endStaticIpAddress1:   "10.10.102.50",
		configText:            testAccCheckVcdNetworkRoutedStatic1,
		resourceName:          "vcd_network_routed",
	}
	runTest(def, t)
}

func TestAccVcdNetworkRoutedStatic2(t *testing.T) {
	var def = networkDef{
		name:                  routedStaticNetwork2,
		gateway:               "10.10.102.1",
		startStaticIpAddress1: "10.10.102.2",
		endStaticIpAddress1:   "10.10.102.50",
		startStaticIpAddress2: "10.10.102.52",
		endStaticIpAddress2:   "10.10.102.100",
		configText:            testAccCheckVcdNetworkRoutedStatic2,
		resourceName:          "vcd_network_routed",
	}
	runTest(def, t)
}

func TestAccVcdNetworkRoutedDhcp(t *testing.T) {
	var def = networkDef{
		name:               routedDhcpNetwork,
		gateway:            "10.10.102.1",
		startDhcpIpAddress: "10.10.102.51",
		endDhcpIpAddress:   "10.10.102.100",
		configText:         testAccCheckVcdNetworkRoutedDhcp,
		resourceName:       "vcd_network_routed",
	}
	runTest(def, t)
}

func TestAccVcdNetworkRoutedMixed(t *testing.T) {

	var def = networkDef{
		name:                  routedMixedNetwork,
		gateway:               "10.10.102.1",
		startStaticIpAddress1: "10.10.102.2",
		endStaticIpAddress1:   "10.10.102.50",
		startDhcpIpAddress:    "10.10.102.51",
		endDhcpIpAddress:      "10.10.102.100",
		configText:            testAccCheckVcdNetworkRoutedMixed,
		resourceName:          "vcd_network_routed",
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
		"Org":                   testConfig.VCD.Org,
		"Vdc":                   testConfig.VCD.Vdc,
		"EdgeGateway":           testConfig.Networking.EdgeGateway,
		"NetworkName":           networkName,
		"Gateway":               def.gateway,
		"StartStaticIpAddress1": def.startStaticIpAddress1,
		"EndStaticIpAddress1":   def.endStaticIpAddress1,
		"StartStaticIpAddress2": def.startStaticIpAddress2,
		"EndStaticIpAddress2":   def.endStaticIpAddress2,
		"StartDhcpIpAddress":    def.startDhcpIpAddress,
		"EndDhcpIpAddress":      def.endDhcpIpAddress,
		"ExternalNetwork":       def.externalNetwork,
		"FuncName":              networkName,
		"Tags":                  "network",
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
	case isolatedMixedNetwork2:
		steps = []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdNetworkExists(networkName, &network),
					testAccCheckVcdNetworkAttributes(networkName, &network),
					resource.TestCheckResourceAttr(
						def.resourceName+"."+networkName, "name", networkName),
					resource.TestCheckOutput("start_dhcp_address", def.startDhcpIpAddress),
					resource.TestCheckOutput("end_dhcp_address", def.endDhcpIpAddress),
					resource.TestCheckOutput("start_static_address1", def.startStaticIpAddress1),
					resource.TestCheckOutput("end_static_address1", def.endStaticIpAddress1),
					resource.TestCheckOutput("start_static_address2", def.startStaticIpAddress2),
					resource.TestCheckOutput("end_static_address2", def.endStaticIpAddress2),
					resource.TestCheckResourceAttr(
						def.resourceName+"."+networkName, "gateway", def.gateway),
					resource.TestMatchResourceAttr(
						def.resourceName+"."+networkName, "href", generatedHrefRegexp),
				),
			},
		}
	case isolatedStaticNetwork2, routedStaticNetwork2:
		steps = []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdNetworkExists(networkName, &network),
					testAccCheckVcdNetworkAttributes(networkName, &network),
					resource.TestCheckResourceAttr(
						def.resourceName+"."+networkName, "name", networkName),
					resource.TestCheckOutput("start_static_address1", def.startStaticIpAddress1),
					resource.TestCheckOutput("end_static_address1", def.endStaticIpAddress1),
					resource.TestCheckOutput("start_static_address2", def.startStaticIpAddress2),
					resource.TestCheckOutput("end_static_address2", def.endStaticIpAddress2),
					resource.TestCheckResourceAttr(
						def.resourceName+"."+networkName, "gateway", def.gateway),
					resource.TestMatchResourceAttr(
						def.resourceName+"."+networkName, "href", generatedHrefRegexp),
				),
			},
		}
	case routedStaticNetwork1, isolatedStaticNetwork1:
		steps = []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdNetworkExists(networkName, &network),
					testAccCheckVcdNetworkAttributes(networkName, &network),
					resource.TestCheckResourceAttr(
						def.resourceName+"."+networkName, "name", networkName),
					resource.TestCheckOutput("start_static_address1", def.startStaticIpAddress1),
					resource.TestCheckOutput("end_static_address1", def.endStaticIpAddress1),
					resource.TestCheckResourceAttr(
						def.resourceName+"."+networkName, "gateway", def.gateway),
					resource.TestMatchResourceAttr(
						def.resourceName+"."+networkName, "href", generatedHrefRegexp),
				),
			},
		}
	case routedMixedNetwork, isolatedMixedNetwork1:
		steps = []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdNetworkExists(networkName, &network),
					testAccCheckVcdNetworkAttributes(networkName, &network),
					resource.TestCheckResourceAttr(
						def.resourceName+"."+networkName, "name", networkName),
					resource.TestCheckOutput("start_static_address1", def.startStaticIpAddress1),
					resource.TestCheckOutput("end_static_address1", def.endStaticIpAddress1),
					resource.TestCheckOutput("start_dhcp_address", def.startDhcpIpAddress),
					resource.TestCheckOutput("end_dhcp_address", def.endDhcpIpAddress),
					resource.TestCheckResourceAttr(
						def.resourceName+"."+networkName, "gateway", def.gateway),
					resource.TestMatchResourceAttr(
						def.resourceName+"."+networkName, "href", generatedHrefRegexp),
				),
			},
		}
	case isolatedDhcpNetwork, routedDhcpNetwork:
		steps = []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdNetworkExists(networkName, &network),
					testAccCheckVcdNetworkAttributes(networkName, &network),
					resource.TestCheckResourceAttr(
						def.resourceName+"."+networkName, "name", networkName),
					resource.TestCheckOutput("start_dhcp_address", def.startDhcpIpAddress),
					resource.TestCheckOutput("end_dhcp_address", def.endDhcpIpAddress),
					resource.TestCheckResourceAttr(
						def.resourceName+"."+networkName, "gateway", def.gateway),
					resource.TestMatchResourceAttr(
						def.resourceName+"."+networkName, "href", generatedHrefRegexp),
				),
			},
		}
	default:
		// Let's make sure we are handling all tests
		fmt.Printf("*** Unhandled test %s\n", def.name)
		t.Fail()
		return
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

const testAccCheckVcdNetworkIsolatedStatic1 = `
resource "vcd_network_isolated" "{{.NetworkName}}" {
  name       = "{{.NetworkName}}"
  org        = "{{.Org}}"
  vdc        = "{{.Vdc}}"
  gateway    = "{{.Gateway}}"
  dns1       = "192.168.2.1"
  static_ip_pool {
    start_address = "{{.StartStaticIpAddress1}}"
    end_address   = "{{.EndStaticIpAddress1}}"
  }
}

output "start_static_address1" {
  value = tolist(vcd_network_isolated.{{.NetworkName}}.static_ip_pool)[0].start_address
  depends_on = [vcd_network_isolated.{{.NetworkName}}]
}
output "end_static_address1" {
  value = tolist(vcd_network_isolated.{{.NetworkName}}.static_ip_pool)[0].end_address
  depends_on = [vcd_network_isolated.{{.NetworkName}}]
}
`

const testAccCheckVcdNetworkIsolatedStatic2 = `
resource "vcd_network_isolated" "{{.NetworkName}}" {
  name       = "{{.NetworkName}}"
  org        = "{{.Org}}"
  vdc        = "{{.Vdc}}"
  gateway    = "{{.Gateway}}"
  dns1       = "192.168.2.1"
  static_ip_pool {
    start_address = "{{.StartStaticIpAddress1}}"
    end_address   = "{{.EndStaticIpAddress1}}"
  }
  static_ip_pool {
    start_address = "{{.StartStaticIpAddress2}}"
    end_address   = "{{.EndStaticIpAddress2}}"
  }
}

output "start_static_address2" {
  value = tolist(vcd_network_isolated.{{.NetworkName}}.static_ip_pool)[0].start_address
  depends_on = [vcd_network_isolated.{{.NetworkName}}]
}
output "end_static_address2" {
  value = tolist(vcd_network_isolated.{{.NetworkName}}.static_ip_pool)[0].end_address
  depends_on = [vcd_network_isolated.{{.NetworkName}}]
}
output "start_static_address1" {
  value = tolist(vcd_network_isolated.{{.NetworkName}}.static_ip_pool)[1].start_address
  depends_on = [vcd_network_isolated.{{.NetworkName}}]
}
output "end_static_address1" {
  value = tolist(vcd_network_isolated.{{.NetworkName}}.static_ip_pool)[1].end_address
  depends_on = [vcd_network_isolated.{{.NetworkName}}]
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
    start_address = "{{.StartDhcpIpAddress}}"
    end_address   = "{{.EndDhcpIpAddress}}"
  }
}
output "start_dhcp_address" {
  value = tolist(vcd_network_isolated.{{.NetworkName}}.dhcp_pool)[0].start_address
  depends_on = [vcd_network_isolated.{{.NetworkName}}]
}
output "end_dhcp_address" {
  value = tolist(vcd_network_isolated.{{.NetworkName}}.dhcp_pool)[0].end_address
  depends_on = [vcd_network_isolated.{{.NetworkName}}]
}
`

const testAccCheckVcdNetworkIsolatedMixed1 = `
resource "vcd_network_isolated" "{{.NetworkName}}" {
  name       = "{{.NetworkName}}"
  org        = "{{.Org}}"
  vdc        = "{{.Vdc}}"
  gateway    = "{{.Gateway}}"
  dns1       = "192.168.2.1"
  static_ip_pool {
    start_address = "{{.StartStaticIpAddress1}}"
    end_address   = "{{.EndStaticIpAddress1}}"
  }
  dhcp_pool {
    start_address = "{{.StartDhcpIpAddress}}"
    end_address   = "{{.EndDhcpIpAddress}}"
  }
}

output "start_dhcp_address" {
  value = tolist(vcd_network_isolated.{{.NetworkName}}.dhcp_pool)[0].start_address
  depends_on = [vcd_network_isolated.{{.NetworkName}}]
}
output "end_dhcp_address" {
  value = tolist(vcd_network_isolated.{{.NetworkName}}.dhcp_pool)[0].end_address
  depends_on = [vcd_network_isolated.{{.NetworkName}}]
}
output "start_static_address1" {
  value = tolist(vcd_network_isolated.{{.NetworkName}}.static_ip_pool)[0].start_address
  depends_on = [vcd_network_isolated.{{.NetworkName}}]
}
output "end_static_address1" {
  value = tolist(vcd_network_isolated.{{.NetworkName}}.static_ip_pool)[0].end_address
  depends_on = [vcd_network_isolated.{{.NetworkName}}]
}
`

const testAccCheckVcdNetworkIsolatedMixed2 = `
resource "vcd_network_isolated" "{{.NetworkName}}" {
  name       = "{{.NetworkName}}"
  org        = "{{.Org}}"
  vdc        = "{{.Vdc}}"
  gateway    = "{{.Gateway}}"
  dns1       = "192.168.2.1"
  static_ip_pool {
    start_address = "{{.StartStaticIpAddress1}}"
    end_address   = "{{.EndStaticIpAddress1}}"
  }
  static_ip_pool {
    start_address = "{{.StartStaticIpAddress2}}"
    end_address   = "{{.EndStaticIpAddress2}}"
  }
  dhcp_pool {
    start_address = "{{.StartDhcpIpAddress}}"
    end_address   = "{{.EndDhcpIpAddress}}"
  }
}

output "start_dhcp_address" {
  value = tolist(vcd_network_isolated.{{.NetworkName}}.dhcp_pool)[0].start_address
  depends_on = [vcd_network_isolated.{{.NetworkName}}]
}
output "end_dhcp_address" {
  value = tolist(vcd_network_isolated.{{.NetworkName}}.dhcp_pool)[0].end_address
  depends_on = [vcd_network_isolated.{{.NetworkName}}]
}
output "start_static_address2" {
  value = tolist(vcd_network_isolated.{{.NetworkName}}.static_ip_pool)[0].start_address
  depends_on = [vcd_network_isolated.{{.NetworkName}}]
}
output "end_static_address2" {
  value = tolist(vcd_network_isolated.{{.NetworkName}}.static_ip_pool)[0].end_address
  depends_on = [vcd_network_isolated.{{.NetworkName}}]
}
output "start_static_address1" {
  value = tolist(vcd_network_isolated.{{.NetworkName}}.static_ip_pool)[1].start_address
  depends_on = [vcd_network_isolated.{{.NetworkName}}]
}
output "end_static_address1" {
  value = tolist(vcd_network_isolated.{{.NetworkName}}.static_ip_pool)[1].end_address
  depends_on = [vcd_network_isolated.{{.NetworkName}}]
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

const testAccCheckVcdNetworkRoutedStatic1 = `
resource "vcd_network_routed" "{{.NetworkName}}" {
  name         = "{{.NetworkName}}"
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "{{.Gateway}}"

  static_ip_pool {
    start_address = "{{.StartStaticIpAddress1}}"
    end_address   = "{{.EndStaticIpAddress1}}"
  }
}
output "end_static_address1" {
  value = tolist(vcd_network_routed.{{.NetworkName}}.static_ip_pool)[0].end_address
  depends_on = [vcd_network_routed.{{.NetworkName}}]
}
output "start_static_address1" {
  value = tolist(vcd_network_routed.{{.NetworkName}}.static_ip_pool)[0].start_address
  depends_on = [vcd_network_routed.{{.NetworkName}}]
}`

const testAccCheckVcdNetworkRoutedStatic2 = `
resource "vcd_network_routed" "{{.NetworkName}}" {
  name         = "{{.NetworkName}}"
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "{{.Gateway}}"

  static_ip_pool {
    start_address = "{{.StartStaticIpAddress1}}"
    end_address   = "{{.EndStaticIpAddress1}}"
  }
  static_ip_pool {
    start_address = "{{.StartStaticIpAddress2}}"
    end_address   = "{{.EndStaticIpAddress2}}"
  }
}
output "start_static_address2" {
  value = tolist(vcd_network_routed.{{.NetworkName}}.static_ip_pool)[0].start_address
  depends_on = [vcd_network_routed.{{.NetworkName}}]
}
output "end_static_address2" {
  value = tolist(vcd_network_routed.{{.NetworkName}}.static_ip_pool)[0].end_address
  depends_on = [vcd_network_routed.{{.NetworkName}}]
}
output "start_static_address1" {
  value = tolist(vcd_network_routed.{{.NetworkName}}.static_ip_pool)[1].start_address
  depends_on = [vcd_network_routed.{{.NetworkName}}]
}
output "end_static_address1" {
  value = tolist(vcd_network_routed.{{.NetworkName}}.static_ip_pool)[1].end_address
  depends_on = [vcd_network_routed.{{.NetworkName}}]
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
    start_address = "{{.StartDhcpIpAddress}}"
    end_address   = "{{.EndDhcpIpAddress}}"
  }
}
output "start_dhcp_address" {
  value = tolist(vcd_network_routed.{{.NetworkName}}.dhcp_pool)[0].start_address
  depends_on = [vcd_network_routed.{{.NetworkName}}]
}
output "end_dhcp_address" {
  value = tolist(vcd_network_routed.{{.NetworkName}}.dhcp_pool)[0].end_address
  depends_on = [vcd_network_routed.{{.NetworkName}}]
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
    start_address = "{{.StartStaticIpAddress1}}"
    end_address   = "{{.EndStaticIpAddress1}}"
  }

  dhcp_pool {
    start_address = "{{.StartDhcpIpAddress}}"
    end_address   = "{{.EndDhcpIpAddress}}"
  }
}
output "start_dhcp_address" {
  value = tolist(vcd_network_routed.{{.NetworkName}}.dhcp_pool)[0].start_address
  depends_on = [vcd_network_routed.{{.NetworkName}}]
}
output "end_dhcp_address" {
  value = tolist(vcd_network_routed.{{.NetworkName}}.dhcp_pool)[0].end_address
  depends_on = [vcd_network_routed.{{.NetworkName}}]
}
output "end_static_address1" {
  value = tolist(vcd_network_routed.{{.NetworkName}}.static_ip_pool)[0].end_address
  depends_on = [vcd_network_routed.{{.NetworkName}}]
}
output "start_static_address1" {
  value = tolist(vcd_network_routed.{{.NetworkName}}.static_ip_pool)[0].start_address
  depends_on = [vcd_network_routed.{{.NetworkName}}]
}`
