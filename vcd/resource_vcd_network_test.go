// +build network ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

type networkDef struct {
	name                  string
	description           string
	gateway               string
	startStaticIpAddress1 string
	endStaticIpAddress1   string
	startStaticIpAddress2 string
	endStaticIpAddress2   string
	startDhcpIpAddress    string
	endDhcpIpAddress      string
	maxLeaseTime          int
	defaultLeaseTime      int
	externalNetwork       string
	configText            string
	resourceName          string
	interfaceName         string
}

const (
	isolatedStaticNetwork1   string = "TestAccVcdNetworkIsoStatic1"
	isolatedStaticNetwork2   string = "TestAccVcdNetworkIsoStatic2"
	isolatedDhcpNetwork      string = "TestAccVcdNetworkIsoDhcp"
	isolatedMixedNetwork1    string = "TestAccVcdNetworkIsoMixed1"
	isolatedMixedNetwork2    string = "TestAccVcdNetworkIsoMixed2"
	routedStaticNetwork1     string = "TestAccVcdNetworkRoutedStatic1"
	routedStaticNetwork2     string = "TestAccVcdNetworkRoutedStatic2"
	routedDhcpNetwork        string = "TestAccVcdNetworkRoutedDhcp"
	routedMixedNetwork       string = "TestAccVcdNetworkRoutedMixed"
	routedStaticNetworkSub2  string = "TestAccVcdNetworkRoutedStaticSub2"
	routedStaticNetworkDist  string = "TestAccVcdNetworkRoutedStaticDist"
	routedStaticNetworkDist2 string = "TestAccVcdNetworkRoutedStaticDist2"
	routedDhcpNetworkSub     string = "TestAccVcdNetworkRoutedDhcpSub"
	routedMixedNetworkSub    string = "TestAccVcdNetworkRoutedMixedSub"
	directNetwork            string = "TestAccVcdNetworkDirect"
	groupStartLabel          string = "start_address"
	groupEndLabel            string = "end_address"
	groupDefaultLease        string = "default_lease_time"
	groupMaxLease            string = "max_lease_time"
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
	var updateDef = networkDef{
		name:                  isolatedStaticNetwork1 + "-update",
		gateway:               "192.168.2.1",
		startStaticIpAddress1: "192.168.2.5",
		endStaticIpAddress1:   "192.168.2.45",
		configText:            testAccCheckVcdNetworkIsolatedStatic1,
		resourceName:          "vcd_network_isolated",
	}

	runTest(def, updateDef, t)
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
	var updateDef = networkDef{
		name:                  isolatedStaticNetwork2 + "-update",
		gateway:               "192.168.2.1",
		startStaticIpAddress1: "192.168.2.5",
		endStaticIpAddress1:   "192.168.2.45",
		startStaticIpAddress2: "192.168.2.53",
		endStaticIpAddress2:   "192.168.2.99",
		configText:            testAccCheckVcdNetworkIsolatedStatic2,
		resourceName:          "vcd_network_isolated",
	}
	runTest(def, updateDef, t)
}

func TestAccVcdNetworkIsolatedDhcp(t *testing.T) {
	var def = networkDef{
		name:               isolatedDhcpNetwork,
		gateway:            "192.168.2.1",
		startDhcpIpAddress: "192.168.2.51",
		endDhcpIpAddress:   "192.168.2.100",
		defaultLeaseTime:   4000,
		maxLeaseTime:       86400,
		configText:         testAccCheckVcdNetworkIsolatedDhcp,
		resourceName:       "vcd_network_isolated",
	}
	var updateDef = networkDef{
		name:               isolatedDhcpNetwork + "-update",
		gateway:            "192.168.2.1",
		startDhcpIpAddress: "192.168.2.53",
		endDhcpIpAddress:   "192.168.2.99",
		defaultLeaseTime:   8000,
		maxLeaseTime:       604800,
		configText:         testAccCheckVcdNetworkIsolatedDhcp,
		resourceName:       "vcd_network_isolated",
	}
	runTest(def, updateDef, t)
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
	var updateDef = networkDef{
		name:                  isolatedMixedNetwork1 + "-update",
		gateway:               "192.168.2.1",
		startStaticIpAddress1: "192.168.2.5",
		endStaticIpAddress1:   "192.168.2.45",
		startDhcpIpAddress:    "192.168.2.53",
		endDhcpIpAddress:      "192.168.2.99",
		configText:            testAccCheckVcdNetworkIsolatedMixed1,
		resourceName:          "vcd_network_isolated",
	}

	runTest(def, updateDef, t)
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
	var updateDef = networkDef{
		name:                  isolatedMixedNetwork2 + "-update",
		gateway:               "192.168.2.1",
		startStaticIpAddress1: "192.168.2.5",
		endStaticIpAddress1:   "192.168.2.45",
		startStaticIpAddress2: "192.168.2.53",
		endStaticIpAddress2:   "192.168.2.99",
		startDhcpIpAddress:    "192.168.2.153",
		endDhcpIpAddress:      "192.168.2.198",
		configText:            testAccCheckVcdNetworkIsolatedMixed2,
		resourceName:          "vcd_network_isolated",
	}
	runTest(def, updateDef, t)
}

// TestAccVcdNetworkRoutedStatic1 tests a routed network with static IP pool
// and implicit internal interface
func TestAccVcdNetworkRoutedStatic1(t *testing.T) {
	var def = networkDef{
		name:                  routedStaticNetwork1,
		gateway:               "10.10.102.1",
		startStaticIpAddress1: "10.10.102.2",
		endStaticIpAddress1:   "10.10.102.50",
		configText:            testAccCheckVcdNetworkRoutedStatic1,
		resourceName:          "vcd_network_routed",
	}
	var updateDef = networkDef{
		name:                  routedStaticNetwork1 + "-update",
		gateway:               "10.10.102.1",
		startStaticIpAddress1: "10.10.102.5",
		endStaticIpAddress1:   "10.10.102.45",
		configText:            testAccCheckVcdNetworkRoutedStatic1,
		resourceName:          "vcd_network_routed",
	}
	runTest(def, updateDef, t)
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
		interfaceName:         "internal",
	}
	var updateDef = networkDef{
		name:                  routedStaticNetwork2 + "-update",
		gateway:               "10.10.102.1",
		startStaticIpAddress1: "10.10.102.5",
		endStaticIpAddress1:   "10.10.102.45",
		startStaticIpAddress2: "10.10.102.53",
		endStaticIpAddress2:   "10.10.102.99",
		configText:            testAccCheckVcdNetworkRoutedStatic2,
		resourceName:          "vcd_network_routed",
		interfaceName:         "internal",
	}
	runTest(def, updateDef, t)
}

func TestAccVcdNetworkRoutedStaticSub2(t *testing.T) {
	var def = networkDef{
		name:                  routedStaticNetworkSub2,
		gateway:               "10.10.102.1",
		startStaticIpAddress1: "10.10.102.2",
		endStaticIpAddress1:   "10.10.102.50",
		startStaticIpAddress2: "10.10.102.52",
		endStaticIpAddress2:   "10.10.102.100",
		configText:            testAccCheckVcdNetworkRoutedStatic2,
		resourceName:          "vcd_network_routed",
		interfaceName:         "subinterface",
	}
	var updateDef = networkDef{
		name:                  routedStaticNetworkSub2 + "-update",
		gateway:               "10.10.102.1",
		startStaticIpAddress1: "10.10.102.5",
		endStaticIpAddress1:   "10.10.102.45",
		startStaticIpAddress2: "10.10.102.53",
		endStaticIpAddress2:   "10.10.102.99",
		configText:            testAccCheckVcdNetworkRoutedStatic2,
		resourceName:          "vcd_network_routed",
		interfaceName:         "subinterface",
	}
	runTest(def, updateDef, t)
}

func TestAccVcdNetworkRoutedStaticDist(t *testing.T) {
	if !testDistributedNetworks {
		t.Skip("Distributed test skipped: not enabled")
	}
	var def = networkDef{
		name:                  routedStaticNetworkDist,
		gateway:               "10.10.103.1",
		startStaticIpAddress1: "10.10.103.2",
		endStaticIpAddress1:   "10.10.103.50",
		startStaticIpAddress2: "10.10.103.52",
		endStaticIpAddress2:   "10.10.103.100",
		configText:            testAccCheckVcdNetworkRoutedStatic2,
		resourceName:          "vcd_network_routed",
		interfaceName:         "distributed",
	}
	var updateDef = networkDef{
		name:                  routedStaticNetworkDist + "-update",
		gateway:               "10.10.103.1",
		startStaticIpAddress1: "10.10.103.5",
		endStaticIpAddress1:   "10.10.103.45",
		startStaticIpAddress2: "10.10.103.53",
		endStaticIpAddress2:   "10.10.103.99",
		configText:            testAccCheckVcdNetworkRoutedStatic2,
		resourceName:          "vcd_network_routed",
		interfaceName:         "distributed",
	}
	runTest(def, updateDef, t)
}

func TestAccVcdNetworkRoutedStaticDist2(t *testing.T) {
	if !testDistributedNetworks {
		t.Skip("Distributed test skipped: not enabled")
	}
	var def = networkDef{
		name:                  routedStaticNetworkDist2,
		gateway:               "10.10.102.1",
		startStaticIpAddress1: "10.10.102.2",
		endStaticIpAddress1:   "10.10.102.50",
		startStaticIpAddress2: "10.10.102.52",
		endStaticIpAddress2:   "10.10.102.100",
		configText:            testAccCheckVcdNetworkRoutedStatic2,
		resourceName:          "vcd_network_routed",
		interfaceName:         "distributed",
	}
	var updateDef = networkDef{
		name:                  routedStaticNetworkDist2 + "update",
		gateway:               "10.10.102.1",
		startStaticIpAddress1: "10.10.102.5",
		endStaticIpAddress1:   "10.10.102.45",
		startStaticIpAddress2: "10.10.102.53",
		endStaticIpAddress2:   "10.10.102.99",
		configText:            testAccCheckVcdNetworkRoutedStatic2,
		resourceName:          "vcd_network_routed",
		interfaceName:         "distributed",
	}
	runTest(def, updateDef, t)
}

func TestAccVcdNetworkRoutedDhcp(t *testing.T) {
	var def = networkDef{
		name:               routedDhcpNetwork,
		gateway:            "10.10.102.1",
		startDhcpIpAddress: "10.10.102.51",
		endDhcpIpAddress:   "10.10.102.100",
		maxLeaseTime:       86400,
		configText:         testAccCheckVcdNetworkRoutedDhcp,
		resourceName:       "vcd_network_routed",
		interfaceName:      "internal",
	}
	var updateDef = networkDef{
		name:               routedDhcpNetwork + "-update",
		startDhcpIpAddress: "10.10.102.52",
		endDhcpIpAddress:   "10.10.102.99",
		maxLeaseTime:       604800,
		configText:         testAccCheckVcdNetworkRoutedDhcp,
		resourceName:       "vcd_network_routed",
		interfaceName:      "internal",
	}
	runTest(def, updateDef, t)
}

func TestAccVcdNetworkRoutedDhcpSub(t *testing.T) {
	var def = networkDef{
		name:               routedDhcpNetworkSub,
		gateway:            "10.10.102.1",
		startDhcpIpAddress: "10.10.102.51",
		endDhcpIpAddress:   "10.10.102.100",
		configText:         testAccCheckVcdNetworkRoutedDhcp,
		resourceName:       "vcd_network_routed",
		interfaceName:      "subinterface",
	}
	var updateDef = networkDef{
		name:               routedDhcpNetworkSub + "-update",
		gateway:            "10.10.102.1",
		startDhcpIpAddress: "10.10.102.52",
		endDhcpIpAddress:   "10.10.102.99",
		configText:         testAccCheckVcdNetworkRoutedDhcp,
		resourceName:       "vcd_network_routed",
		interfaceName:      "subinterface",
	}
	runTest(def, updateDef, t)
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
		interfaceName:         "internal",
	}
	var updateDef = networkDef{
		name:                  routedMixedNetwork + "-update",
		gateway:               "10.10.102.1",
		startStaticIpAddress1: "10.10.102.5",
		endStaticIpAddress1:   "10.10.102.45",
		startDhcpIpAddress:    "10.10.102.52",
		endDhcpIpAddress:      "10.10.102.99",
		configText:            testAccCheckVcdNetworkRoutedMixed,
		resourceName:          "vcd_network_routed",
		interfaceName:         "internal",
	}
	runTest(def, updateDef, t)
}

func TestAccVcdNetworkRoutedMixedSub(t *testing.T) {
	var def = networkDef{
		name:                  routedMixedNetworkSub,
		gateway:               "10.10.102.1",
		startStaticIpAddress1: "10.10.102.2",
		endStaticIpAddress1:   "10.10.102.50",
		startDhcpIpAddress:    "10.10.102.51",
		endDhcpIpAddress:      "10.10.102.100",
		configText:            testAccCheckVcdNetworkRoutedMixed,
		resourceName:          "vcd_network_routed",
		interfaceName:         "subinterface",
	}
	var updateDef = networkDef{
		name:                  routedMixedNetworkSub + "-update",
		gateway:               "10.10.102.1",
		startStaticIpAddress1: "10.10.102.5",
		endStaticIpAddress1:   "10.10.102.45",
		startDhcpIpAddress:    "10.10.102.52",
		endDhcpIpAddress:      "10.10.102.99",
		configText:            testAccCheckVcdNetworkRoutedMixed,
		resourceName:          "vcd_network_routed",
		interfaceName:         "subinterface",
	}
	runTest(def, updateDef, t)
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
	var updateDef = networkDef{
		name:            directNetwork + "-update",
		externalNetwork: testConfig.Networking.ExternalNetwork,
		configText:      testAccCheckVcdNetworkDirect,
		resourceName:    "vcd_network_direct",
	}
	runTest(def, updateDef, t)
}

func runTest(def, updateDef networkDef, t *testing.T) {

	generatedHrefRegexp := regexp.MustCompile("^https://")

	networkName := def.name
	if def.description == "" {
		def.description = fmt.Sprintf("%s description", networkName)
	}
	if updateDef.description == "" {
		updateDef.description = fmt.Sprintf("%s updated description", networkName)
	}
	if def.maxLeaseTime == 0 {
		def.maxLeaseTime = 7200
	}
	if updateDef.maxLeaseTime == 0 {
		updateDef.maxLeaseTime = 7200
	}
	if def.defaultLeaseTime == 0 {
		def.defaultLeaseTime = 3600
	}
	if updateDef.defaultLeaseTime == 0 {
		updateDef.defaultLeaseTime = 3600
	}
	var params = StringMap{
		"Org":                   testConfig.VCD.Org,
		"Vdc":                   testConfig.VCD.Vdc,
		"Description":           def.description,
		"EdgeGateway":           testConfig.Networking.EdgeGateway,
		"ResourceName":          networkName,
		"NetworkName":           networkName,
		"Gateway":               def.gateway,
		"StartStaticIpAddress1": def.startStaticIpAddress1,
		"EndStaticIpAddress1":   def.endStaticIpAddress1,
		"StartStaticIpAddress2": def.startStaticIpAddress2,
		"EndStaticIpAddress2":   def.endStaticIpAddress2,
		"StartDhcpIpAddress":    def.startDhcpIpAddress,
		"EndDhcpIpAddress":      def.endDhcpIpAddress,
		"DefaultLeaseTime":      def.defaultLeaseTime,
		"MaxLeaseTime":          def.maxLeaseTime,
		"ExternalNetwork":       def.externalNetwork,
		"FuncName":              networkName,
		"InterfaceType":         def.interfaceName,
		"Tags":                  "network",
	}
	var network govcd.OrgVDCNetwork
	configText := templateFill(def.configText, params)

	updateDef.description = firstNonEmpty(updateDef.description, def.description)
	updateDef.name = firstNonEmpty(updateDef.name, def.name)
	updateDef.startStaticIpAddress1 = firstNonEmpty(updateDef.startStaticIpAddress1, def.startStaticIpAddress1)
	updateDef.startStaticIpAddress2 = firstNonEmpty(updateDef.startStaticIpAddress2, def.startStaticIpAddress2)
	updateDef.endStaticIpAddress1 = firstNonEmpty(updateDef.endStaticIpAddress1, def.endStaticIpAddress1)
	updateDef.endStaticIpAddress2 = firstNonEmpty(updateDef.endStaticIpAddress2, def.endStaticIpAddress2)
	updateDef.startDhcpIpAddress = firstNonEmpty(updateDef.startDhcpIpAddress, def.startDhcpIpAddress)
	updateDef.endDhcpIpAddress = firstNonEmpty(updateDef.endDhcpIpAddress, def.endDhcpIpAddress)

	params["Description"] = updateDef.description
	params["NetworkName"] = updateDef.name
	params["StartStaticIpAddress1"] = updateDef.startStaticIpAddress1
	params["StartStaticIpAddress2"] = updateDef.startStaticIpAddress2
	params["EndStaticIpAddress1"] = updateDef.endStaticIpAddress1
	params["EndStaticIpAddress2"] = updateDef.endStaticIpAddress2
	params["StartDhcpIpAddress"] = updateDef.startDhcpIpAddress
	params["EndDhcpIpAddress"] = updateDef.endDhcpIpAddress
	params["FuncName"] = updateDef.name
	params["MaxLeaseTime"] = updateDef.maxLeaseTime
	params["DefaultLeaseTime"] = updateDef.defaultLeaseTime

	updateConfigText := templateFill(fmt.Sprintf("\n# skip-binary-test only for updates\n%s", def.configText), params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	debugPrintf("#[DEBUG] UPDATE CONFIGURATION: %s", updateConfigText)
	// steps for external network
	var steps []resource.TestStep

	resourceDef := def.resourceName + "." + networkName
	switch def.name {
	case directNetwork:
		steps = []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdNetworkExists(networkName, &network),
					resource.TestCheckResourceAttr(
						resourceDef, "name", networkName),
					resource.TestCheckResourceAttr(
						resourceDef, "description", def.description),
					resource.TestMatchResourceAttr(
						resourceDef, "href", generatedHrefRegexp),
				),
			},
			resource.TestStep{
				Config: updateConfigText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdNetworkExists(updateDef.name, &network),
					resource.TestCheckResourceAttr(
						resourceDef, "name", updateDef.name),
					resource.TestCheckResourceAttr(
						resourceDef, "description", updateDef.description),
					resource.TestMatchResourceAttr(
						resourceDef, "href", generatedHrefRegexp),
				),
			},
		}
	case isolatedMixedNetwork2:
		steps = []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdNetworkExists(networkName, &network),
					checkNetWorkIpGroups(resourceDef, def, resourceVcdNetworkStaticIpPoolHash, resourceVcdNetworkIsolatedDhcpPoolHash),
					resource.TestCheckResourceAttr(
						resourceDef, "name", networkName),
					resource.TestCheckResourceAttr(
						resourceDef, "description", def.description),
					resource.TestCheckResourceAttr(
						resourceDef, "gateway", def.gateway),
					resource.TestMatchResourceAttr(
						resourceDef, "href", generatedHrefRegexp),
				),
			},
			resource.TestStep{
				Config: updateConfigText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdNetworkExists(updateDef.name, &network),
					checkIpGroup(resourceDef,
						"static_ip_pool",
						map[string]interface{}{
							"start_address": updateDef.startStaticIpAddress1,
							"end_address":   updateDef.endStaticIpAddress1,
						},
						resourceVcdNetworkStaticIpPoolHash,
					),
					checkNetWorkIpGroups(resourceDef, updateDef, resourceVcdNetworkStaticIpPoolHash, resourceVcdNetworkIsolatedDhcpPoolHash),
					resource.TestCheckResourceAttr(
						resourceDef, "name", updateDef.name),
					resource.TestCheckResourceAttr(
						resourceDef, "description", updateDef.description),
					resource.TestCheckResourceAttr(
						resourceDef, "gateway", def.gateway),
					resource.TestMatchResourceAttr(
						resourceDef, "href", generatedHrefRegexp),
				),
			},
		}
	case isolatedStaticNetwork2, routedStaticNetwork2, routedStaticNetworkSub2, routedStaticNetworkDist, routedStaticNetworkDist2:
		hashFunc := resourceVcdNetworkRoutedDhcpPoolHash
		if def.name == isolatedStaticNetwork2 {
			hashFunc = resourceVcdNetworkIsolatedDhcpPoolHash
		}
		steps = []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdNetworkExists(networkName, &network),
					resource.TestCheckResourceAttr(
						resourceDef, "name", networkName),
					resource.TestCheckResourceAttr(
						resourceDef, "description", def.description),
					checkNetWorkIpGroups(resourceDef, def, resourceVcdNetworkStaticIpPoolHash, hashFunc),
					resource.TestCheckResourceAttr(
						resourceDef, "gateway", def.gateway),
					resource.TestMatchResourceAttr(
						resourceDef, "href", generatedHrefRegexp),
				),
			},
			resource.TestStep{
				Config: updateConfigText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdNetworkExists(updateDef.name, &network),
					resource.TestCheckResourceAttr(
						resourceDef, "name", updateDef.name),
					resource.TestCheckResourceAttr(
						resourceDef, "description", updateDef.description),
					checkNetWorkIpGroups(resourceDef, updateDef, resourceVcdNetworkStaticIpPoolHash, hashFunc),
					resource.TestCheckResourceAttr(
						resourceDef, "gateway", def.gateway),
					resource.TestMatchResourceAttr(
						resourceDef, "href", generatedHrefRegexp),
				),
			},
		}
	case routedStaticNetwork1, isolatedStaticNetwork1:
		hashFunc := resourceVcdNetworkRoutedDhcpPoolHash
		if def.name == isolatedStaticNetwork1 {
			hashFunc = resourceVcdNetworkIsolatedDhcpPoolHash
		}
		steps = []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdNetworkExists(networkName, &network),
					resource.TestCheckResourceAttr(
						resourceDef, "name", networkName),
					resource.TestCheckResourceAttr(
						resourceDef, "description", def.description),
					checkNetWorkIpGroups(resourceDef, def, resourceVcdNetworkStaticIpPoolHash, hashFunc),
					resource.TestCheckResourceAttr(
						resourceDef, "gateway", def.gateway),
					resource.TestMatchResourceAttr(
						resourceDef, "href", generatedHrefRegexp),
				),
			},
			resource.TestStep{
				Config: updateConfigText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdNetworkExists(updateDef.name, &network),
					resource.TestCheckResourceAttr(
						resourceDef, "name", updateDef.name),
					resource.TestCheckResourceAttr(
						resourceDef, "description", updateDef.description),
					checkNetWorkIpGroups(resourceDef, updateDef, resourceVcdNetworkStaticIpPoolHash, hashFunc),
					resource.TestCheckResourceAttr(
						resourceDef, "gateway", def.gateway),
					resource.TestMatchResourceAttr(
						resourceDef, "href", generatedHrefRegexp),
				),
			},
		}
	case routedMixedNetwork, isolatedMixedNetwork1, routedMixedNetworkSub:
		hashFunc := resourceVcdNetworkRoutedDhcpPoolHash
		if def.name == isolatedMixedNetwork1 {
			hashFunc = resourceVcdNetworkIsolatedDhcpPoolHash
		}
		steps = []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdNetworkExists(networkName, &network),
					resource.TestCheckResourceAttr(
						resourceDef, "name", networkName),
					resource.TestCheckResourceAttr(
						resourceDef, "description", def.description),
					checkNetWorkIpGroups(resourceDef, def, resourceVcdNetworkStaticIpPoolHash, hashFunc),
					resource.TestCheckResourceAttr(
						resourceDef, "gateway", def.gateway),
					resource.TestMatchResourceAttr(
						resourceDef, "href", generatedHrefRegexp),
				),
			},
			resource.TestStep{
				Config: updateConfigText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdNetworkExists(updateDef.name, &network),
					resource.TestCheckResourceAttr(
						resourceDef, "name", updateDef.name),
					resource.TestCheckResourceAttr(
						resourceDef, "description", updateDef.description),
					checkNetWorkIpGroups(resourceDef, updateDef, resourceVcdNetworkStaticIpPoolHash, hashFunc),
					resource.TestCheckResourceAttr(
						resourceDef, "gateway", def.gateway),
					resource.TestMatchResourceAttr(
						resourceDef, "href", generatedHrefRegexp),
				),
			},
		}
	case isolatedDhcpNetwork, routedDhcpNetwork, routedDhcpNetworkSub:
		hashFunc := resourceVcdNetworkRoutedDhcpPoolHash
		if def.name == isolatedDhcpNetwork {
			hashFunc = resourceVcdNetworkIsolatedDhcpPoolHash
		}
		steps = []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdNetworkExists(networkName, &network),
					resource.TestCheckResourceAttr(
						resourceDef, "name", networkName),
					resource.TestCheckResourceAttr(
						resourceDef, "description", def.description),
					checkNetWorkIpGroups(resourceDef, def, resourceVcdNetworkStaticIpPoolHash, hashFunc),
					resource.TestCheckResourceAttr(
						resourceDef, "gateway", def.gateway),
					resource.TestMatchResourceAttr(
						resourceDef, "href", generatedHrefRegexp),
				),
			},
			resource.TestStep{
				Config: updateConfigText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdNetworkExists(networkName, &network),
					resource.TestCheckResourceAttr(
						resourceDef, "name", updateDef.name),
					resource.TestCheckResourceAttr(
						resourceDef, "description", updateDef.description),
					checkNetWorkIpGroups(resourceDef, updateDef, resourceVcdNetworkStaticIpPoolHash, hashFunc),
					resource.TestCheckResourceAttr(
						resourceDef, "gateway", def.gateway),
					resource.TestMatchResourceAttr(
						resourceDef, "href", generatedHrefRegexp),
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
		ImportStateIdFunc: importStateIdOrgVdcObject(testConfig, updateDef.name),
	})

	// Don't convert this test to parallel, as it will cause IP ranges conflicts
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: func(s *terraform.State) error { return testAccCheckVcdNetworkDestroy(s, def.resourceName, networkName) },
		Steps:        steps,
	})
}

// TestHashFunc makes sure that the hash used to compute the network static IP pool and DHCP pool
// doesn't change.
// It tests some IP pairs with the hard coded hash result as of version 2.8.0
// If this test fails, we may have introduced a breaking change that causes a plan update.
func TestHashFunc(t *testing.T) {
	var testsDhcp = []struct {
		startIp      string
		endIp        string
		maxLease     int
		defaultLease int
		wantIso      int
		wantRouted   int
	}{
		// Hard coded values obtained on 2020-03-18 (version 2.8.0)
		// Do not change
		{"10.10.10.2", "10.10.10.100", 8000, 3600, 1030199880, 1056625442},
		{"192.168.1.2", "192.168.1.20", 9000, 4000, 631317345, 159171876},
		{"10.10.1.2", "10.10.1.100", 9000, 4000, 756273417, 3024715874},
		{"10.10.1.2", "10.10.1.100", 86400, 3600, 3223781707, 13051714},
		{"10.10.0.101", "10.10.0.200", 9500, 5000, 1068413432, 77228404},
		{"10.10.0.1", "10.10.0.50", 3600, 7200, 3836851868, 4072727102},
		{"10.10.0.1", "10.10.0.50", 7200, 7200, 2964735978, 2283983785},
	}
	var testsStatic = []struct {
		startIp string
		endIp   string
		want    int
	}{
		// Hard coded values obtained on 2020-03-18 (version 2.8.0)
		// Do not change
		{"10.10.10.2", "10.10.10.100", 3116097209},
		{"192.168.1.2", "192.168.1.20", 3331633131},
		{"10.10.1.2", "10.10.1.100", 2850949493},
		{"10.10.0.101", "10.10.0.200", 4005846706},
	}
	t.Run("static", func(t *testing.T) {
		for _, tc := range testsStatic {
			value := map[string]interface{}{
				groupStartLabel: tc.startIp,
				groupEndLabel:   tc.endIp,
			}
			got := resourceVcdNetworkStaticIpPoolHash(value)
			if got != tc.want {
				t.Logf("startIp: %s, endIp: %s - want %d, got %d", tc.startIp, tc.endIp, tc.want, got)
				t.Fail()
			}
		}
	})
	// DHCP in isolated network
	t.Run("dhcp-isolated", func(t *testing.T) {
		for _, tc := range testsDhcp {
			value := map[string]interface{}{
				groupStartLabel:   tc.startIp,
				groupEndLabel:     tc.endIp,
				groupMaxLease:     tc.maxLease,
				groupDefaultLease: tc.defaultLease,
			}
			got := resourceVcdNetworkIsolatedDhcpPoolHash(value)
			if got != tc.wantIso {
				t.Logf("startIp: %s, endIp: %s, maxLease: %d, defaultLease: %d - want %d, got %d",
					tc.startIp, tc.endIp, tc.maxLease, tc.defaultLease, tc.wantIso, got)
				t.Fail()
			}
		}
	})
	// DHCP in routed network
	t.Run("dhcp-routed", func(t *testing.T) {
		for _, tc := range testsDhcp {
			value := map[string]interface{}{
				groupStartLabel: tc.startIp,
				groupEndLabel:   tc.endIp,
				groupMaxLease:   tc.maxLease,
			}
			got := resourceVcdNetworkRoutedDhcpPoolHash(value)
			if got != tc.wantRouted {
				t.Logf("startIp: %s, endIp: %s, maxLease: %d - want %d, got %d",
					tc.startIp, tc.endIp, tc.maxLease, tc.wantRouted, got)
				t.Fail()
			}
		}
	})
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
			// If the function was called after an update that changed the name, we need to
			// search the network by ID
			if network != nil {
				orgVDCNetwork, err = vdc.GetOrgVdcNetworkById(network.OrgVDCNetwork.ID, false)
				if err != nil {
					return fmt.Errorf("[test network exists] error retrieving network %s (id: %s) ", name, network.OrgVDCNetwork.ID)
				}
				*network = *orgVDCNetwork
				return nil
			}
			return fmt.Errorf("network %s does not exist ", name)
		}

		// Save the network for future use
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

// checkNetWorkIpGroups is a wrapper around checkIpGroup that generates
// a test for every pair of IPs in the network definition structure
func checkNetWorkIpGroups(resourceDef string, def networkDef, staticIpHashFunc, dhcpHashFunc schema.SchemaSetFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		var checks []resource.TestCheckFunc

		if def.startStaticIpAddress1 != "" {
			f := checkIpGroup(resourceDef, "static_ip_pool",
				map[string]interface{}{
					groupStartLabel: def.startStaticIpAddress1,
					groupEndLabel:   def.endStaticIpAddress1,
				}, staticIpHashFunc)

			checks = append(checks, f)
		}
		if def.startStaticIpAddress2 != "" {
			f := checkIpGroup(resourceDef, "static_ip_pool",
				map[string]interface{}{
					groupStartLabel: def.startStaticIpAddress2,
					groupEndLabel:   def.endStaticIpAddress2,
				}, staticIpHashFunc)

			checks = append(checks, f)
		}
		if def.startDhcpIpAddress != "" {
			f := checkIpGroup(resourceDef, "dhcp_pool",
				map[string]interface{}{
					groupStartLabel:   def.startDhcpIpAddress,
					groupEndLabel:     def.endDhcpIpAddress,
					groupDefaultLease: def.defaultLeaseTime,
					groupMaxLease:     def.maxLeaseTime,
				}, dhcpHashFunc)
			checks = append(checks, f)
		}

		for _, f := range checks {
			err := f(s)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// checkIpGroup will check the contents of a group of Ips in a TypeSet structure.
// This function computes the hash for the set and then calls TestCheckResourceAttr
// for each key in the map.
// It can be used in a test wherever a TestCheckFunc is allowed
// Sample call:
//  checkIpGroup("vcd_network_isolated.MyNetworkName",
//		"static_ip_pool",
//		map[string]interface{}{
//		    "start_address": "192.168.2.2",
//		    "end_address":   "192.168.2.50",
//		},
//      resourceVcdNetworkIPAddressHash,
//  ),
//
func checkIpGroup(resourceDef, groupName string, values map[string]interface{}, hashFunc schema.SchemaSetFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		var (
			hash       = hashFunc(values)
			ok         bool
			startValue interface{}
			endValue   interface{}
		)
		startValue, ok = values[groupStartLabel]
		if !ok {
			return fmt.Errorf("key '%s' not found in map", groupStartLabel)
		}
		endValue, ok = values[groupEndLabel]
		if !ok {
			return fmt.Errorf("key '%s' not found in map", groupEndLabel)
		}

		startKey := fmt.Sprintf("%s.%d.%s", groupName, hash, groupStartLabel)
		endKey := fmt.Sprintf("%s.%d.%s", groupName, hash, groupEndLabel)
		fStart := resource.TestCheckResourceAttr(resourceDef, startKey, startValue.(string))
		fEnd := resource.TestCheckResourceAttr(resourceDef, endKey, endValue.(string))

		err := fStart(s)
		if err != nil {
			return err
		}
		return fEnd(s)
	}
}

func init() {
	testingTags["network"] = "resource_vcd_network_test.go"
}

const testAccCheckVcdNetworkIsolatedStatic1 = `
resource "vcd_network_isolated" "{{.ResourceName}}" {
  name        = "{{.NetworkName}}"
  description = "{{.Description}}"
  org         = "{{.Org}}"
  vdc         = "{{.Vdc}}"
  gateway     = "{{.Gateway}}"
  dns1        = "192.168.2.1"
  static_ip_pool {
    start_address = "{{.StartStaticIpAddress1}}"
    end_address   = "{{.EndStaticIpAddress1}}"
  }
}
`

const testAccCheckVcdNetworkIsolatedStatic2 = `
resource "vcd_network_isolated" "{{.ResourceName}}" {
  name        = "{{.NetworkName}}"
  description = "{{.Description}}"
  org         = "{{.Org}}"
  vdc         = "{{.Vdc}}"
  gateway     = "{{.Gateway}}"
  dns1        = "192.168.2.1"
  static_ip_pool {
    start_address = "{{.StartStaticIpAddress1}}"
    end_address   = "{{.EndStaticIpAddress1}}"
  }
  static_ip_pool {
    start_address = "{{.StartStaticIpAddress2}}"
    end_address   = "{{.EndStaticIpAddress2}}"
  }
}
`

const testAccCheckVcdNetworkIsolatedDhcp = `
resource "vcd_network_isolated" "{{.ResourceName}}" {
  name        = "{{.NetworkName}}"
  description = "{{.Description}}"
  org         = "{{.Org}}"
  vdc         = "{{.Vdc}}"
  gateway     = "{{.Gateway}}"
  dns1        = "192.168.2.1"
  dhcp_pool {
    start_address      = "{{.StartDhcpIpAddress}}"
    end_address        = "{{.EndDhcpIpAddress}}"
    default_lease_time = "{{.DefaultLeaseTime}}"
    max_lease_time     = "{{.MaxLeaseTime}}"
  }
}
`

const testAccCheckVcdNetworkIsolatedMixed1 = `
resource "vcd_network_isolated" "{{.ResourceName}}" {
  name        = "{{.NetworkName}}"
  description = "{{.Description}}"
  org         = "{{.Org}}"
  vdc         = "{{.Vdc}}"
  gateway     = "{{.Gateway}}"
  dns1        = "192.168.2.1"
  static_ip_pool {
    start_address = "{{.StartStaticIpAddress1}}"
    end_address   = "{{.EndStaticIpAddress1}}"
  }
  dhcp_pool {
    start_address = "{{.StartDhcpIpAddress}}"
    end_address   = "{{.EndDhcpIpAddress}}"
  }
}
`

const testAccCheckVcdNetworkIsolatedMixed2 = `
resource "vcd_network_isolated" "{{.ResourceName}}" {
  name        = "{{.NetworkName}}"
  description = "{{.Description}}"
  org         = "{{.Org}}"
  vdc         = "{{.Vdc}}"
  gateway     = "{{.Gateway}}"
  dns1        = "192.168.2.1"
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
`

const testAccCheckVcdNetworkDirect = `
resource "vcd_network_direct" "{{.ResourceName}}" {
  name             = "{{.NetworkName}}"
  description      = "{{.Description}}"
  org              = "{{.Org}}"
  vdc              = "{{.Vdc}}"
  external_network = "{{.ExternalNetwork}}"
}
`

const testAccCheckVcdNetworkRoutedStatic1 = `
resource "vcd_network_routed" "{{.ResourceName}}" {
  name         = "{{.NetworkName}}"
  description  = "{{.Description}}"
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "{{.Gateway}}"

  static_ip_pool {
    start_address = "{{.StartStaticIpAddress1}}"
    end_address   = "{{.EndStaticIpAddress1}}"
  }
}
`

const testAccCheckVcdNetworkRoutedStatic2 = `
resource "vcd_network_routed" "{{.ResourceName}}" {
  name           = "{{.NetworkName}}"
  description    = "{{.Description}}"
  org            = "{{.Org}}"
  vdc            = "{{.Vdc}}"
  edge_gateway   = "{{.EdgeGateway}}"
  gateway        = "{{.Gateway}}"
  interface_type = "{{.InterfaceType}}"

  static_ip_pool {
    start_address = "{{.StartStaticIpAddress1}}"
    end_address   = "{{.EndStaticIpAddress1}}"
  }
  static_ip_pool {
    start_address = "{{.StartStaticIpAddress2}}"
    end_address   = "{{.EndStaticIpAddress2}}"
  }
}
`

const testAccCheckVcdNetworkRoutedDhcp = `
resource "vcd_network_routed" "{{.ResourceName}}" {
  name           = "{{.NetworkName}}"
  description    = "{{.Description}}"
  org            = "{{.Org}}"
  vdc            = "{{.Vdc}}"
  edge_gateway   = "{{.EdgeGateway}}"
  gateway        = "{{.Gateway}}"
  interface_type = "{{.InterfaceType}}"

  dhcp_pool {
    start_address      = "{{.StartDhcpIpAddress}}"
    end_address        = "{{.EndDhcpIpAddress}}"
    max_lease_time     = "{{.MaxLeaseTime}}"
  }
}
`

const testAccCheckVcdNetworkRoutedMixed = `
resource "vcd_network_routed" "{{.ResourceName}}" {
  name           = "{{.NetworkName}}"
  description    = "{{.Description}}"
  org            = "{{.Org}}"
  vdc            = "{{.Vdc}}"
  edge_gateway   = "{{.EdgeGateway}}"
  gateway        = "{{.Gateway}}"
  interface_type = "{{.InterfaceType}}"

  static_ip_pool {
    start_address = "{{.StartStaticIpAddress1}}"
    end_address   = "{{.EndStaticIpAddress1}}"
  }

  dhcp_pool {
    start_address = "{{.StartDhcpIpAddress}}"
    end_address   = "{{.EndDhcpIpAddress}}"
  }
}
`
