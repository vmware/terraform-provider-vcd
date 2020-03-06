// +build network vm ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// Private type used to collect networks available for data source tests
type networkRec struct {
	network *types.OrgVDCNetwork
	parent  string
}

// Structure used to collect networks
var availableNetworks = make(map[string]networkRec)

// Marker for collected networks
var networksSearched = false

// Templates for network data source tests
var networkTemplates = map[string]string{
	"vcd_network_routed":   datasourceTestNetworkRouted,
	"vcd_network_isolated": datasourceTestNetworkIsolated,
	"vcd_network_direct":   datasourceTestNetworkDirect,
}

// getAvailableNetworks collects available networks to use in data source tests
// It stores its results in availableNetworks.
// If called more than once, skips the search. The caller will then use
// the previous results.
func getAvailableNetworks() error {
	if networksSearched {
		return nil
	}
	// Get the data from configuration file. This client is still inactive at this point
	vcdClient, err := getTestVCDFromJson(testConfig)
	if err != nil {
		return fmt.Errorf("error getting client configuration: %s", err)
	}
	err = ProviderAuthenticate(vcdClient, testConfig.Provider.User, testConfig.Provider.Password, testConfig.Provider.Token, testConfig.Provider.SysOrg)
	if err != nil {
		return fmt.Errorf("authentication error: %s", err)
	}
	org, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
	if err != nil {
		return fmt.Errorf("org not found : %s", err)
	}
	vdc, err := org.GetVDCByName(testConfig.VCD.Vdc, false)
	if err != nil {
		return fmt.Errorf("vdc not found : %s", err)
	}
	networkList, err := vdc.GetNetworkList()
	if err != nil {
		return fmt.Errorf("error getting network list for VDC %s: %s", vdc.Vdc.Name, err)
	}

	for _, net := range networkList {

		network, err := vdc.GetOrgVdcNetworkByName(net.Name, false)
		if err != nil {
			return fmt.Errorf("error getting network %s: %s", net.Name, err)
		}
		networkType := ""
		parent := ""
		switch net.LinkType {
		case 0:
			networkType = "vcd_network_direct"
			parent = net.ConnectedTo
		case 1:
			networkType = "vcd_network_routed"
			parent = net.ConnectedTo
		case 2:
			networkType = "vcd_network_isolated"
		}

		_, ok := availableNetworks[networkType]
		if !ok {
			if networkType == "vcd_network_isolated" {
				// Make sure the IPScope structure is reachable for isolated networks
				if network.OrgVDCNetwork.Configuration != nil &&
					network.OrgVDCNetwork.Configuration.IPScopes != nil &&
					len(network.OrgVDCNetwork.Configuration.IPScopes.IPScope) > 0 {
					availableNetworks[networkType] = networkRec{network.OrgVDCNetwork, parent}
				}
			} else {
				availableNetworks[networkType] = networkRec{network.OrgVDCNetwork, parent}
			}
		}
	}

	networksSearched = true
	return nil
}

func TestAccVcdNetworkDirectDS(t *testing.T) {
	// This test requires access to the vCD before filling templates
	// Thus it won't run in the short test
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	err := getAvailableNetworks()
	if err != nil {
		t.Skip("error getting available networks")
		return
	}
	if len(availableNetworks) == 0 {
		t.Skip("No networks found - data source test skipped")
		return
	}

	networkType := "vcd_network_direct"
	data, ok := availableNetworks[networkType]
	if !ok {
		t.Skip("no direct network found ")
		return
	}

	template := networkTemplates[networkType]
	var params = StringMap{
		"Org":             testConfig.VCD.Org,
		"VDC":             testConfig.VCD.Vdc,
		"ExternalNetwork": data.parent,
		"NetworkName":     data.network.Name,
		"FuncName":        "TestNetworkDirectDS",
		"Tags":            "network",
	}
	configText := templateFill(template, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(

					resource.TestCheckOutput("network_org", testConfig.VCD.Org),
					resource.TestCheckOutput("network_vdc", testConfig.VCD.Vdc),
					resource.TestCheckOutput("network_name", data.network.Name),
					resource.TestCheckOutput("network_description", data.network.Description),
					resource.TestCheckOutput("external_network", data.parent),
				),
			},
		},
	})
}

func TestAccVcdNetworkRoutedDS(t *testing.T) {
	// This test requires access to the vCD before filling templates
	// Thus it won't run in the short test
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	err := getAvailableNetworks()

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

	template := networkTemplates[networkType]
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"VDC":         testConfig.VCD.Vdc,
		"EdgeGateway": data.parent,
		"NetworkName": data.network.Name,
		"FuncName":    "TestNetworkRoutedDS",
		"Tags":        "network",
	}
	configText := templateFill(template, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(

					resource.TestCheckOutput("network_org", testConfig.VCD.Org),
					resource.TestCheckOutput("network_vdc", testConfig.VCD.Vdc),
					resource.TestCheckOutput("network_name", data.network.Name),
					resource.TestCheckOutput("network_description", data.network.Description),
					resource.TestCheckOutput("network_edge", data.parent),
					resource.TestCheckOutput("default_gateway", testConfig.Networking.ExternalNetwork),
				),
			},
		},
	})
}

func TestAccVcdNetworkIsolatedDS(t *testing.T) {
	// This test requires access to the vCD before filling templates
	// Thus it won't run in the short test
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	err := getAvailableNetworks()

	if err != nil {
		fmt.Printf("%s\n", err)
		t.Skip("error getting available networks")
		return
	}
	if len(availableNetworks) == 0 {
		t.Skip("No networks found - data source test skipped")
		return
	}

	networkType := "vcd_network_isolated"
	existingNetwork, ok := availableNetworks[networkType]
	if !ok {
		t.Skip("no isolated network found ")
		return
	}

	template := networkTemplates[networkType]
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"VDC":         testConfig.VCD.Vdc,
		"NetworkName": existingNetwork.network.Name,
		"FuncName":    "TestNetworkIsolatedDS",
		"Tags":        "network",
	}
	configText := templateFill(template, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(

					resource.TestCheckOutput("network_org", testConfig.VCD.Org),
					resource.TestCheckOutput("network_vdc", testConfig.VCD.Vdc),
					resource.TestCheckOutput("network_name", existingNetwork.network.Name),
					resource.TestCheckOutput("network_description", existingNetwork.network.Description),
					resource.TestCheckOutput("network_gateway", existingNetwork.network.Configuration.IPScopes.IPScope[0].Gateway),
					resource.TestCheckOutput("network_netmask", existingNetwork.network.Configuration.IPScopes.IPScope[0].Netmask),
					resource.TestCheckOutput("network_start_address", existingNetwork.network.Configuration.IPScopes.IPScope[0].IPRanges.IPRange[0].StartAddress),
				),
			},
		},
	})
}

const datasourceTestNetworkDirect = `
data "vcd_network_direct" "{{.NetworkName}}" {
  name             = "{{.NetworkName}}"
  org              = "{{.Org}}"
  vdc              = "{{.VDC}}"
}

output "network_name" {
  value = data.vcd_network_direct.{{.NetworkName}}.name
}

output "network_description" {
  value = data.vcd_network_direct.{{.NetworkName}}.description
}

output "network_org" {
  value = data.vcd_network_direct.{{.NetworkName}}.org
}

output "network_vdc" {
  value = data.vcd_network_direct.{{.NetworkName}}.vdc
}

output "external_network" {
  value = data.vcd_network_direct.{{.NetworkName}}.external_network
}
`

const datasourceTestNetworkRouted = `
data "vcd_edgegateway" "{{.EdgeGateway}}" {
  name = "{{.EdgeGateway}}"
  org  = "{{.Org}}"
  vdc  = "{{.VDC}}"
}

data "vcd_network_routed" "{{.NetworkName}}" {
  name         = "{{.NetworkName}}"
  org          = "{{.Org}}"
  vdc          = "{{.VDC}}"
}

output "default_gateway" {
  value = data.vcd_edgegateway.{{.EdgeGateway}}.default_gateway_network
}

output "network_name" {
  value = data.vcd_network_routed.{{.NetworkName}}.name
}

output "network_description" {
  value = data.vcd_network_routed.{{.NetworkName}}.description
}

output "network_org" {
  value = data.vcd_network_routed.{{.NetworkName}}.org
}

output "network_vdc" {
  value = data.vcd_network_routed.{{.NetworkName}}.vdc
}

output "network_edge" {
  value = data.vcd_network_routed.{{.NetworkName}}.edge_gateway
}
`

const datasourceTestNetworkIsolated = `
data "vcd_network_isolated" "{{.NetworkName}}" {
  name         = "{{.NetworkName}}"
  org          = "{{.Org}}"
  vdc          = "{{.VDC}}"
}

output "network_name" {
  value = data.vcd_network_isolated.{{.NetworkName}}.name
}

output "network_description" {
  value = data.vcd_network_isolated.{{.NetworkName}}.description
}

output "network_org" {
  value = data.vcd_network_isolated.{{.NetworkName}}.org
}

output "network_vdc" {
  value = data.vcd_network_isolated.{{.NetworkName}}.vdc
}

output "network_gateway" {
  value = data.vcd_network_isolated.{{.NetworkName}}.gateway
}

output "network_netmask" {
  value = data.vcd_network_isolated.{{.NetworkName}}.netmask
}

output "network_start_address" {
  value  = tolist(data.vcd_network_isolated.{{.NetworkName}}.static_ip_pool)[0].start_address
}
`
