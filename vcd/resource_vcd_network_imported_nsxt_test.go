// +build network nsxt ALL functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdNetworkRoutedV2Nsxt tests out NSX-T backed Org VDC networking capabilities
func TestAccVcdNetworkImportedV2Nsxt(t *testing.T) {
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	vcdClient := createTemporaryVCDConnection()
	if vcdClient.Client.APIVCDMaxVersionIs("< 34.0") {
		t.Skip(t.Name() + " requires at least API v34.0 (vCD 10.1.1+)")
	}
	if !vcdClient.Client.IsSysAdmin {
		t.Skip(t.Name() + " only System Administrator can create Imported networks")
	}

	skipNoNsxtConfiguration(t)

	// String map to fill the template
	var params = StringMap{
		"Org":               testConfig.VCD.Org,
		"NsxtVdc":           testConfig.Nsxt.Vdc,
		"EdgeGw":            testConfig.Nsxt.EdgeGateway,
		"NetworkName":       t.Name(),
		"NsxtImportSegment": testConfig.Nsxt.NsxtImportSegment,
		"Tags":              "network",
	}

	configText := templateFill(TestAccVcdNetworkImportedV2NsxtStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText)

	// params["FuncName"] = t.Name() + "-step2"
	// configText2 := templateFill(TestAccVcdNetworkRoutedV2NsxtStep2, params)
	// debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)
	//
	// params["FuncName"] = t.Name() + "-step3"
	// configText3 := templateFill(TestAccVcdNetworkRoutedV2NsxtStep3, params)
	// debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

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
					cachedId.cacheTestResourceFieldValue("vcd_network_imported.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_imported.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_imported.net1", "nsxt_logical_switch_id"),
					resource.TestCheckResourceAttr("vcd_network_imported.net1", "name", "nsxt-imported-test-initial"),
					resource.TestCheckResourceAttr("vcd_network_imported.net1", "description", "NSX-T imported network test OpenAPI"),
					resource.TestCheckResourceAttr("vcd_network_imported.net1", "gateway", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_network_imported.net1", "prefix_length", "24"),
					resource.TestCheckResourceAttr("vcd_network_imported.net1", "static_ip_pool.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_network_imported.net1", "static_ip_pool.*", map[string]string{
						"start_address": "1.1.1.10",
						"end_address":   "1.1.1.20",
					}),
				),
			},
			// resource.TestStep{ // step 2
			// 	Config: configText2,
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		cachedId.testCheckCachedResourceFieldValue("vcd_network_routed_v2.net1", "id"),
			// 		resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "id"),
			// 		resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "name", t.Name()),
			// 		resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "description", "Updated"),
			// 		resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "edge_gateway_id"),
			// 		resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "gateway", "1.1.1.1"),
			// 		resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "prefix_length", "24"),
			// 		resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "static_ip_pool.#", "3"),
			// 		resource.TestCheckTypeSetElemNestedAttrs("vcd_network_routed_v2.net1", "static_ip_pool.*", map[string]string{
			// 			"start_address": "1.1.1.10",
			// 			"end_address":   "1.1.1.20",
			// 		}),
			//
			// 		resource.TestCheckTypeSetElemNestedAttrs("vcd_network_routed_v2.net1", "static_ip_pool.*", map[string]string{
			// 			"start_address": "1.1.1.40",
			// 			"end_address":   "1.1.1.50",
			// 		}),
			//
			// 		resource.TestCheckTypeSetElemNestedAttrs("vcd_network_routed_v2.net1", "static_ip_pool.*", map[string]string{
			// 			"start_address": "1.1.1.60",
			// 			"end_address":   "1.1.1.70",
			// 		}),
			// 	),
			// },
			//
			// Check that import works
			resource.TestStep{ // step 3
				ResourceName:      "vcd_network_imported.net1",
				ImportState:       true,
				ImportStateVerify: true,
				// It is impossible to read 'nsxt_logical_switch_name' for already consumed NSX-T segment (API returns
				// only unused segments) therefore this field cannot be set during read operations.
				ImportStateVerifyIgnore: []string{"nsxt_logical_switch_name"},
				ImportStateIdFunc:       importStateIdOrgNsxtVdcObject(testConfig, "nsxt-imported-test-initial"),
			},
			//
			// resource.TestStep{ // step 4
			// 	Config: configText3,
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		cachedId.testCheckCachedResourceFieldValue("vcd_network_routed_v2.net1", "id"),
			// 		resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "id"),
			// 		resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "name", t.Name()),
			// 		resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "description", "Updated"),
			// 		resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "edge_gateway_id"),
			// 		resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "gateway", "1.1.1.1"),
			// 		resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "prefix_length", "24"),
			// 		resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "static_ip_pool.#", "0"),
			// 	),
			// },
		},
	})
}

const TestAccVcdNetworkImportedV2NsxtStep1 = `
resource "vcd_network_imported" "net1" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "nsxt-imported-test-initial"
  description = "NSX-T imported network test OpenAPI"

  nsxt_logical_switch_name = "{{.NsxtImportSegment}}"

  gateway = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
	start_address = "1.1.1.10"
    end_address = "1.1.1.20"
  }
}
`

//
// const TestAccVcdNetworkRoutedV2NsxtStep2 = `
// data "vcd_nsxt_edgegateway" "existing" {
//   org  = "{{.Org}}"
//   vdc  = "{{.NsxtVdc}}"
//   name = "{{.EdgeGw}}"
// }
//
// resource "vcd_network_routed_v2" "net1" {
//   org  = "{{.Org}}"
//   vdc  = "{{.NsxtVdc}}"
//   name = "{{.NetworkName}}"
//   description = "Updated"
//
//   edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
//
//   gateway = "1.1.1.1"
//   prefix_length = 24
//
//   static_ip_pool {
// 	start_address = "1.1.1.10"
//     end_address = "1.1.1.20"
//   }
//
//   static_ip_pool {
// 	start_address = "1.1.1.40"
//     end_address = "1.1.1.50"
//   }
//
//   static_ip_pool {
// 	start_address = "1.1.1.60"
//     end_address = "1.1.1.70"
//   }
// }
// `
//
// const TestAccVcdNetworkRoutedV2NsxtStep3 = `
// data "vcd_nsxt_edgegateway" "existing" {
//   org  = "{{.Org}}"
//   vdc  = "{{.NsxtVdc}}"
//   name = "{{.EdgeGw}}"
// }
//
// resource "vcd_network_routed_v2" "net1" {
//   org  = "{{.Org}}"
//   vdc  = "{{.NsxtVdc}}"
//   name = "{{.NetworkName}}"
//   description = "Updated"
//
//   edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
//
//   gateway = "1.1.1.1"
//   prefix_length = 24
// }
// `
