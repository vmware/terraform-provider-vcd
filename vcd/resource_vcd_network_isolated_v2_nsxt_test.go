// +build network nsxt ALL functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdNetworkIsolatedV2Nsxt tests out NSX-T backed Org VDC networking capabilities
func TestAccVcdNetworkIsolatedV2Nsxt(t *testing.T) {
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	vcdClient := createTemporaryVCDConnection()
	if vcdClient.Client.APIVCDMaxVersionIs("< 34.0") {
		t.Skip(t.Name() + " requires at least API v34.0 (vCD 10.1.1+)")
	}
	skipNoNsxtConfiguration(t)

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"NsxtVdc":     testConfig.Nsxt.Vdc,
		"NetworkName": t.Name(),
		"Tags":        "network nsxt",
	}

	configText := templateFill(TestAccVcdNetworkIsolatedV2NsxtStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(TestAccVcdNetworkIsolatedV2NsxtStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(TestAccVcdNetworkIsolatedV2NsxtStep3, params)
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
					cachedId.cacheTestResourceFieldValue("vcd_network_isolated_v2.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_isolated_v2.net1", "id"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "name", "nsxt-isolated-test-initial"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "description", "NSX-T isolated network test"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "gateway", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "prefix_length", "24"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "static_ip_pool.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_network_isolated_v2.net1", "static_ip_pool.*", map[string]string{
						"start_address": "1.1.1.10",
						"end_address":   "1.1.1.20",
					}),
				),
			},
			resource.TestStep{ // step 2
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.cacheTestResourceFieldValue("vcd_network_isolated_v2.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_isolated_v2.net1", "id"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "description", "updated NSX-T isolated network test"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "gateway", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "prefix_length", "24"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "static_ip_pool.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_network_isolated_v2.net1", "static_ip_pool.*", map[string]string{
						"start_address": "1.1.1.10",
						"end_address":   "1.1.1.20",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_network_isolated_v2.net1", "static_ip_pool.*", map[string]string{
						"start_address": "1.1.1.30",
						"end_address":   "1.1.1.40",
					}),
				),
			},
			// Check that import works
			resource.TestStep{ // step 3
				ResourceName:      "vcd_network_isolated_v2.net1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcObject(testConfig, t.Name()),
			},
			resource.TestStep{ // step 4
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.cacheTestResourceFieldValue("vcd_network_isolated_v2.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_isolated_v2.net1", "id"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "description", ""),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "gateway", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "prefix_length", "24"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "static_ip_pool.#", "0"),
				),
			},
		},
	})
}

const TestAccVcdNetworkIsolatedV2NsxtStep1 = `
resource "vcd_network_isolated_v2" "net1" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  
  name        = "nsxt-isolated-test-initial"
  description = "NSX-T isolated network test"

  gateway = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
	start_address = "1.1.1.10"
	end_address = "1.1.1.20"
  }
}
`

const TestAccVcdNetworkIsolatedV2NsxtStep2 = `
resource "vcd_network_isolated_v2" "net1" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  
  name        = "{{.NetworkName}}"
  description = "updated NSX-T isolated network test"

  gateway       = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
	start_address = "1.1.1.10"
	end_address   = "1.1.1.20"
  }

  static_ip_pool {
	start_address = "1.1.1.30"
	end_address   = "1.1.1.40"
  }
}
`

const TestAccVcdNetworkIsolatedV2NsxtStep3 = `
resource "vcd_network_isolated_v2" "net1" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  
  name          = "{{.NetworkName}}"
  gateway       = "1.1.1.1"
  prefix_length = 24
}
`
