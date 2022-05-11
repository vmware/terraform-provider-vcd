//go:build network || ALL || functional
// +build network ALL functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdNetworkRoutedV2NsxvInterfaceTypes attempts to test all supported interface types (except distributed) for
// NSX-V Org VDC routed network
func TestAccVcdNetworkRoutedV2NsxvInterfaceTypes(t *testing.T) {
	preTestChecks(t)
	var params = StringMap{
		"Org":           testConfig.VCD.Org,
		"Vdc":           testConfig.VCD.Vdc,
		"EdgeGw":        testConfig.Networking.EdgeGateway,
		"InterfaceType": "internal",
		"NetworkName":   t.Name(),
		"Tags":          "network",
	}

	configText := templateFill(testAccVcdNetworkRoutedV2Nsxv, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 0: %s", configText)

	params["FuncName"] = t.Name() + "-step1"
	params["InterfaceType"] = "subinterface"
	configText1 := templateFill(testAccVcdNetworkRoutedV2Nsxv, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testParamsNotEmpty(t, params) },
		CheckDestroy:      testAccCheckOpenApiVcdNetworkDestroy(testConfig.VCD.Vdc, t.Name()),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "id"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "interface_type", "INTERNAL"),
				),
			},
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "id"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "interface_type", "SUBINTERFACE"),
				),
			},
			// Check that import works
			{
				ResourceName:      "vcd_network_routed_v2.net1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgVdcObject(testConfig, t.Name()),
			},
		},
	})
	postTestChecks(t)
}

func TestAccVcdNetworkRoutedV2NsxvDistributedInterface(t *testing.T) {
	preTestChecks(t)
	if !testDistributedNetworksEnabled() {
		t.Skip("Distributed test skipped: not enabled")
	}

	var params = StringMap{
		"Org":           testConfig.VCD.Org,
		"Vdc":           testConfig.VCD.Vdc,
		"EdgeGw":        testConfig.Networking.EdgeGateway,
		"InterfaceType": "distributed",
		"NetworkName":   t.Name(),
		"Tags":          "network",
	}

	params["FuncName"] = t.Name()
	configText := templateFill(testAccVcdNetworkRoutedV2Nsxv, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 0: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testParamsNotEmpty(t, params) },
		CheckDestroy:      testAccCheckOpenApiVcdNetworkDestroy(testConfig.VCD.Vdc, t.Name()),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "id"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "name", "TestAccVcdNetworkRoutedV2NsxvDistributedInterface"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "description", "NSX-V routed network test OpenAPI"),
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "edge_gateway_id"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "interface_type", "DISTRIBUTED"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "gateway", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "prefix_length", "24"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "static_ip_pool.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_network_routed_v2.net1", "static_ip_pool.*", map[string]string{
						"start_address": "1.1.1.10",
						"end_address":   "1.1.1.20",
					}),
				),
			},
			// Check that import works
			{
				ResourceName:      "vcd_network_routed_v2.net1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgVdcObject(testConfig, t.Name()),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNetworkRoutedV2Nsxv = `
data "vcd_edgegateway" "existing" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  name = "{{.EdgeGw}}"
}

resource "vcd_network_routed_v2" "net1" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  name = "{{.NetworkName}}"
  description = "NSX-V routed network test OpenAPI"

  interface_type = "{{.InterfaceType}}"

  edge_gateway_id = data.vcd_edgegateway.existing.id
  
  gateway = "1.1.1.1"
  prefix_length = 24


  static_ip_pool {
	start_address = "1.1.1.10"
    end_address = "1.1.1.20"
  }
}
`
