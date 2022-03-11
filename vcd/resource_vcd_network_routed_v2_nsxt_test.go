//go:build network || nsxt || ALL || functional
// +build network nsxt ALL functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdNetworkRoutedV2Nsxt tests out NSX-T backed Org VDC networking capabilities
func TestAccVcdNetworkRoutedV2Nsxt(t *testing.T) {
	preTestChecks(t)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	skipNoNsxtConfiguration(t)

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"NsxtVdc":     testConfig.Nsxt.Vdc,
		"EdgeGw":      testConfig.Nsxt.EdgeGateway,
		"NetworkName": t.Name(),
		"Tags":        "network",
	}

	configText := templateFill(TestAccVcdNetworkRoutedV2NsxtStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(TestAccVcdNetworkRoutedV2NsxtStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(TestAccVcdNetworkRoutedV2NsxtStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	// Ensure the resource is never recreated - ID stays the same
	cachedId := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckOpenApiVcdNetworkDestroy(testConfig.Nsxt.Vdc, t.Name()),
		Steps: []resource.TestStep{
			resource.TestStep{ // step 1
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.cacheTestResourceFieldValue("vcd_network_routed_v2.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "id"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "name", "nsxt-routed-test-initial"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "description", "NSX-T routed network test OpenAPI"),
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "edge_gateway_id"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "gateway", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "prefix_length", "24"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "static_ip_pool.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_network_routed_v2.net1", "static_ip_pool.*", map[string]string{
						"start_address": "1.1.1.10",
						"end_address":   "1.1.1.20",
					}),
				),
			},
			resource.TestStep{ // step 2
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.testCheckCachedResourceFieldValue("vcd_network_routed_v2.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "id"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "description", "Updated"),
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "edge_gateway_id"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "gateway", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "prefix_length", "24"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "static_ip_pool.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_network_routed_v2.net1", "static_ip_pool.*", map[string]string{
						"start_address": "1.1.1.10",
						"end_address":   "1.1.1.20",
					}),

					resource.TestCheckTypeSetElemNestedAttrs("vcd_network_routed_v2.net1", "static_ip_pool.*", map[string]string{
						"start_address": "1.1.1.40",
						"end_address":   "1.1.1.50",
					}),

					resource.TestCheckTypeSetElemNestedAttrs("vcd_network_routed_v2.net1", "static_ip_pool.*", map[string]string{
						"start_address": "1.1.1.60",
						"end_address":   "1.1.1.70",
					}),
				),
			},

			// Check that import works
			resource.TestStep{ // step 3
				ResourceName:      "vcd_network_routed_v2.net1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcObject(testConfig, t.Name()),
			},

			resource.TestStep{ // step 4
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.testCheckCachedResourceFieldValue("vcd_network_routed_v2.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "id"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "description", "Updated"),
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "edge_gateway_id"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "gateway", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "prefix_length", "24"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "static_ip_pool.#", "0"),
				),
			},
		},
	})
	postTestChecks(t)
}

const TestAccVcdNetworkRoutedV2NsxtStep1 = `
data "vcd_nsxt_edgegateway" "existing" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "{{.EdgeGw}}"
}

resource "vcd_network_routed_v2" "net1" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "nsxt-routed-test-initial"
  description = "NSX-T routed network test OpenAPI"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  gateway = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
	start_address = "1.1.1.10"
    end_address = "1.1.1.20"
  }
}
`

const TestAccVcdNetworkRoutedV2NsxtStep2 = `
data "vcd_nsxt_edgegateway" "existing" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "{{.EdgeGw}}"
}

resource "vcd_network_routed_v2" "net1" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "{{.NetworkName}}"
  description = "Updated"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  gateway = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
	start_address = "1.1.1.10"
    end_address = "1.1.1.20"
  }

  static_ip_pool {
	start_address = "1.1.1.40"
    end_address = "1.1.1.50"
  }

  static_ip_pool {
	start_address = "1.1.1.60"
    end_address = "1.1.1.70"
  }
}
`

const TestAccVcdNetworkRoutedV2NsxtStep3 = `
data "vcd_nsxt_edgegateway" "existing" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "{{.EdgeGw}}"
}

resource "vcd_network_routed_v2" "net1" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "{{.NetworkName}}"
  description = "Updated"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  gateway = "1.1.1.1"
  prefix_length = 24
}
`
