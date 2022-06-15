//go:build network || nsxt || ALL || functional
// +build network nsxt ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdNetworkRoutedV2Nsxt tests out NSX-T backed Org VDC networking capabilities
func TestAccVcdNetworkRoutedV2Nsxt(t *testing.T) {
	preTestChecks(t)

	// String map to fill the template
	var params = StringMap{
		"Org":                  testConfig.VCD.Org,
		"NsxtVdc":              testConfig.Nsxt.Vdc,
		"EdgeGw":               testConfig.Nsxt.EdgeGateway,
		"NetworkName":          t.Name(),
		"Tags":                 "network",
		"MetadataKey":          "key1",
		"MetadataValue":        "value1",
		"MetadataKeyUpdated":   "key2",
		"MetadataValueUpdated": "value2",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(TestAccVcdNetworkRoutedV2NsxtStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(TestAccVcdNetworkRoutedV2NsxtStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(TestAccVcdNetworkRoutedV2NsxtStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	// Ensure the resource is never recreated - ID stays the same
	cachedId := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckOpenApiVcdNetworkDestroy(testConfig.Nsxt.Vdc, t.Name()),
		Steps: []resource.TestStep{
			{ // step 1
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
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "metadata."+params["MetadataKey"].(string), params["MetadataValue"].(string)),
					resource.TestCheckResourceAttrPair("data.vcd_nsxt_edgegateway.existing", "owner_id", "vcd_network_routed_v2.net1", "owner_id"),
				),
			},
			{ // step 2
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
					resource.TestCheckNoResourceAttr("vcd_network_routed_v2.net1", "metadata."+params["MetadataKey"].(string)),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "metadata."+params["MetadataKeyUpdated"].(string), params["MetadataValueUpdated"].(string)),

					resource.TestCheckResourceAttrPair("data.vcd_nsxt_edgegateway.existing", "owner_id", "vcd_network_routed_v2.net1", "owner_id"),
				),
			},

			// Check that import works
			{ // step 3
				ResourceName:      "vcd_network_routed_v2.net1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcObject(testConfig, t.Name()),
			},

			{ // step 4
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
					resource.TestCheckResourceAttrPair("data.vcd_nsxt_edgegateway.existing", "owner_id", "vcd_network_routed_v2.net1", "owner_id"),
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

  metadata = {
    {{.MetadataKey}} = "{{.MetadataValue}}"
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

  metadata = {
    {{.MetadataKeyUpdated}} = "{{.MetadataValueUpdated}}"
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

// TestAccVcdNetworkRoutedV2NsxtOwnerVdc checks that a routed network can be created without specifying
// `vdc` field
func TestAccVcdNetworkRoutedV2NsxtOwnerVdc(t *testing.T) {
	preTestChecks(t)

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"NsxtVdc":     testConfig.Nsxt.Vdc,
		"EdgeGw":      testConfig.Nsxt.EdgeGateway,
		"NetworkName": t.Name(),
		"Tags":        "network",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdNetworkRoutedV2NsxtOwnerVdcStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcdNetworkRoutedV2NsxtOwnerVdcStep1DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	// Ensure the resource is never recreated - ID stays the same
	cachedId := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckOpenApiVcdNetworkDestroy(testConfig.Nsxt.Vdc, t.Name()),
		Steps: []resource.TestStep{
			{ // step 1
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.cacheTestResourceFieldValue("vcd_network_routed_v2.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "id"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "name", t.Name()),
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "edge_gateway_id"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "gateway", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "prefix_length", "24"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "static_ip_pool.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_network_routed_v2.net1", "static_ip_pool.*", map[string]string{
						"start_address": "1.1.1.10",
						"end_address":   "1.1.1.20",
					}),
					resource.TestCheckResourceAttrPair("data.vcd_nsxt_edgegateway.existing", "owner_id", "vcd_network_routed_v2.net1", "owner_id"),
				),
			},

			// Check that import works
			{ // step 2
				ResourceName:      "vcd_network_routed_v2.net1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcObject(testConfig, t.Name()),
			},
			{ // step 1
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.cacheTestResourceFieldValue("vcd_network_routed_v2.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "id"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "name", t.Name()),
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "edge_gateway_id"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "gateway", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "prefix_length", "24"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "static_ip_pool.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_network_routed_v2.net1", "static_ip_pool.*", map[string]string{
						"start_address": "1.1.1.10",
						"end_address":   "1.1.1.20",
					}),
					resource.TestCheckResourceAttrPair("data.vcd_nsxt_edgegateway.existing", "owner_id", "vcd_network_routed_v2.net1", "owner_id"),
					// data source has `filter` field therefore total field number '%s' is ignored
					resourceFieldsEqual("data.vcd_network_routed_v2.net1", "vcd_network_routed_v2.net1", []string{"%"}),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNetworkRoutedV2NsxtOwnerVdcStep1 = `
data "vcd_nsxt_edgegateway" "existing" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "{{.EdgeGw}}"
}

resource "vcd_network_routed_v2" "net1" {
  org  = "{{.Org}}"
  name = "{{.NetworkName}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  gateway = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
	start_address = "1.1.1.10"
    end_address = "1.1.1.20"
  }
}
`

const testAccVcdNetworkRoutedV2NsxtOwnerVdcStep1DS = testAccVcdNetworkRoutedV2NsxtOwnerVdcStep1 + `
data "vcd_network_routed_v2" "net1" {
  name            = vcd_network_routed_v2.net1.name
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
}
`

// TestAccVcdNetworkRoutedV2NsxtMigration attempts to check migration path from legacy VDC
// configuration to new configuration which makes the NSX-T Edge Gateway follow membership of parent
// NSX-T Edge Gateway
// * Step 1 - creates prerequisites - VDC Group and 2 VDCs
// * Step 2 - creates an Edge Gateway and a routed network attached to it
// * Step 3 - leaves the Edge Gateway as it is, but removed `vdc` field
// * Step 4 - migrates the Edge Gateway to VDC Group and observes that routed networks moves
// together and reflects it
func TestAccVcdNetworkRoutedV2NsxtMigration(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges to create VDCs")
		return
	}

	// String map to fill the template
	var params = StringMap{
		"Org":                       testConfig.VCD.Org,
		"NsxtVdc":                   testConfig.Nsxt.Vdc,
		"EdgeGw":                    testConfig.Nsxt.EdgeGateway,
		"NetworkName":               t.Name(),
		"Name":                      t.Name(),
		"Dfw":                       "false",
		"DefaultPolicy":             "false",
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"ProviderVdcStorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"ExternalNetwork":           testConfig.Nsxt.ExternalNetwork,
		"TestName":                  t.Name(),
		"NsxtEdgeGatewayVcd":        t.Name() + "-edge",
		"MetadataKey":               "key1",
		"MetadataValue":             "value1",
		"Tags":                      "network",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "-newVdc"
	configTextPre := templateFill(testAccVcdVdcGroupNew, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configTextPre)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdNetworkRoutedV2NsxtMigrationStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcdNetworkRoutedV2NsxtMigrationStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(testAccVcdNetworkRoutedV2NsxtMigrationStep4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	cachedId := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckOpenApiVcdNetworkDestroy(testConfig.Nsxt.Vdc, t.Name()),
		Steps: []resource.TestStep{
			{ // step 1 - setup prerequisites
				Config: configTextPre,
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.cacheTestResourceFieldValue("vcd_network_routed_v2.net1", "id"),
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
					resource.TestCheckResourceAttrPair("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", "vcd_network_routed_v2.net1", "owner_id"),
					resource.TestMatchResourceAttr("vcd_network_routed_v2.net1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "metadata."+params["MetadataKey"].(string), params["MetadataValue"].(string)),
				),
			},
			{
				Config: configText3,
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
					resource.TestCheckResourceAttrPair("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", "vcd_network_routed_v2.net1", "owner_id"),
					resource.TestMatchResourceAttr("vcd_network_routed_v2.net1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "metadata."+params["MetadataKey"].(string), params["MetadataValue"].(string)),
				),
			},
			{
				Config: configText4,
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
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "metadata."+params["MetadataKey"].(string), params["MetadataValue"].(string)),
				),
			},
			{ // Applying the same step once more to be sure that vcd_network_routed_v2 has refreshed its fields after edge gateway was moved
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", "vcd_network_routed_v2.net1", "owner_id"),
					resource.TestMatchResourceAttr("vcd_network_routed_v2.net1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdcGroup:`)),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "metadata."+params["MetadataKey"].(string), params["MetadataValue"].(string)),
				),
			},

			// Check that import works
			{ // step 3
				ResourceName:            "vcd_network_routed_v2.net1",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata"}, // Network is in a VDC Group as the Edge Gateway moved, so it can't import metadata
				ImportStateId:           fmt.Sprintf("%s.%s.%s", testConfig.VCD.Org, params["Name"].(string), params["Name"].(string)),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNetworkRoutedV2NsxtMigrationStep2 = testAccVcdVdcGroupNew + `
data "vcd_external_network_v2" "existing-extnet" {
	name = "{{.ExternalNetwork}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org         = "{{.Org}}"
  owner_id    = vcd_org_vdc.newVdc.0.id
  name        = "{{.NsxtEdgeGatewayVcd}}"
  description = "Description"

  external_network_id = data.vcd_external_network_v2.existing-extnet.id

  subnet {
     gateway       = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].gateway
     prefix_length = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].prefix_length

     primary_ip = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }
}

resource "vcd_network_routed_v2" "net1" {
  org  = "{{.Org}}"
  vdc  = vcd_org_vdc.newVdc.0.name
  name = "{{.NetworkName}}"
  description = "Updated"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

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

  metadata = {
   {{.MetadataKey}}  = "{{.MetadataValue}}"
  }
}
`

const testAccVcdNetworkRoutedV2NsxtMigrationStep3 = testAccVcdVdcGroupNew + `
data "vcd_external_network_v2" "existing-extnet" {
	name = "{{.ExternalNetwork}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org         = "{{.Org}}"
  owner_id    = vcd_org_vdc.newVdc.0.id
  name        = "{{.NsxtEdgeGatewayVcd}}"
  description = "Description"

  external_network_id = data.vcd_external_network_v2.existing-extnet.id

  subnet {
     gateway       = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].gateway
     prefix_length = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].prefix_length

     primary_ip = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }
}

resource "vcd_network_routed_v2" "net1" {
  org  = "{{.Org}}"

  name = "{{.NetworkName}}"
  description = "Updated"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

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

  metadata = {
   {{.MetadataKey}}  = "{{.MetadataValue}}"
  }
}
`

const testAccVcdNetworkRoutedV2NsxtMigrationStep4 = testAccVcdVdcGroupNew + `
data "vcd_external_network_v2" "existing-extnet" {
	name = "{{.ExternalNetwork}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org         = "{{.Org}}"
  owner_id    = vcd_vdc_group.test1.id
  name        = "{{.NsxtEdgeGatewayVcd}}"
  description = "Description"

  external_network_id = data.vcd_external_network_v2.existing-extnet.id

  subnet {
     gateway       = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].gateway
     prefix_length = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].prefix_length

     primary_ip = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }
}

resource "vcd_network_routed_v2" "net1" {
  org  = "{{.Org}}"

  name = "{{.NetworkName}}"
  description = "Updated"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

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

  metadata = {
   {{.MetadataKey}}  = "{{.MetadataValue}}"
  }
}
`

// TestAccVcdNetworkRoutedV2InheritedVdc tests that NSX-T Edge Gateway network can be created by
// using `vdc` field inherited from provider in NSX-T VDC
// * Step 1 - Rely on configuration comming from `provider` configuration for `vdc` value
// * Step 2 - Test that import works correctly
// * Step 3 - Test that data source works correctly
// * Step 4 - Start using `vdc` fields in resource and make sure it is not recreated
// * Step 5 - Test that import works correctly
// * Step 6 - Test data source
// Note. It does not test `org` field inheritance because our import sets it by default.
func TestAccVcdNetworkRoutedV2InheritedVdc(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	// String map to fill the template
	var params = StringMap{
		"Org":                testConfig.VCD.Org,
		"NsxtVdc":            testConfig.Nsxt.Vdc,
		"NetworkName":        t.Name(),
		"NsxtEdgeGatewayVcd": "nsxt-edge-test",
		"ExternalNetwork":    testConfig.Nsxt.ExternalNetwork,

		// This particular field is consumed by `templateFill` to generate binary tests with correct
		// default VDC (NSX-T)
		"PrVdc": testConfig.Nsxt.Vdc,

		"Tags": "network",
	}
	testParamsNotEmpty(t, params)

	// This test explicitly tests that `vdc` field inherited from provider works correctly therefore
	// it must override default `vdc` field value at provider level to be NSX-T VDC and restore it
	// after this test.
	restoreDefaultVdcFunc := overrideDefaultVdcForTest(testConfig.Nsxt.Vdc)
	defer restoreDefaultVdcFunc()

	params["FuncName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccVcdNetworkRoutedV2InheritedVdcStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcdNetworkRoutedV2InheritedVdcStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText1)

	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(testAccVcdNetworkRoutedV2InheritedVdcStep4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	params["FuncName"] = t.Name() + "-step6"
	configText6 := templateFill(testAccVcdNetworkRoutedV2InheritedVdcStep6, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 6: %s", configText6)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	cacheEdgeGatewaydId := &testCachedFieldValue{}
	cacheRoutedNetId := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckOpenApiVcdNetworkDestroy(testConfig.Nsxt.Vdc, t.Name()),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					cacheEdgeGatewaydId.cacheTestResourceFieldValue("vcd_nsxt_edgegateway.nsxt-edge", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.nsxt-edge", "id"),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "vdc", testConfig.Nsxt.Vdc),

					cacheRoutedNetId.cacheTestResourceFieldValue("vcd_network_routed_v2.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "id"),
					resource.TestMatchResourceAttr("vcd_network_routed_v2.net1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "vdc", testConfig.Nsxt.Vdc),
				),
			},
			{
				ResourceName:      "vcd_network_routed_v2.net1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcObject(testConfig, params["NetworkName"].(string)),
				// field nsxt_logical_switch_name cannot be read during import because VCD does not
				// provider API for reading it after being consumed
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					cacheEdgeGatewaydId.testCheckCachedResourceFieldValue("vcd_nsxt_edgegateway.nsxt-edge", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.nsxt-edge", "id"),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "vdc", testConfig.Nsxt.Vdc),
					resourceFieldsEqual("data.vcd_nsxt_edgegateway.nsxt-edge", "vcd_nsxt_edgegateway.nsxt-edge", []string{"%"}),

					cacheRoutedNetId.testCheckCachedResourceFieldValue("vcd_network_routed_v2.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "id"),
					resource.TestMatchResourceAttr("vcd_network_routed_v2.net1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "vdc", testConfig.Nsxt.Vdc),
					resourceFieldsEqual("data.vcd_network_routed_v2.net1", "vcd_network_routed_v2.net1", []string{"%"}),
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					cacheEdgeGatewaydId.testCheckCachedResourceFieldValue("vcd_nsxt_edgegateway.nsxt-edge", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.nsxt-edge", "id"),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "vdc", testConfig.Nsxt.Vdc),

					cacheRoutedNetId.testCheckCachedResourceFieldValue("vcd_network_routed_v2.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "id"),
					resource.TestMatchResourceAttr("vcd_network_routed_v2.net1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "vdc", testConfig.Nsxt.Vdc),
				),
			},
			{
				ResourceName:      "vcd_network_routed_v2.net1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcObject(testConfig, params["NetworkName"].(string)),
				// field nsxt_logical_switch_name cannot be read during import because VCD does not
				// provide API for reading it after being consumed
				ImportStateVerifyIgnore: []string{"nsxt_logical_switch_name"},
			},
			{
				Config: configText6,
				Check: resource.ComposeAggregateTestCheckFunc(
					cacheEdgeGatewaydId.testCheckCachedResourceFieldValue("vcd_nsxt_edgegateway.nsxt-edge", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.nsxt-edge", "id"),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "vdc", testConfig.Nsxt.Vdc),
					resourceFieldsEqual("data.vcd_nsxt_edgegateway.nsxt-edge", "vcd_nsxt_edgegateway.nsxt-edge", []string{"%"}),

					cacheRoutedNetId.testCheckCachedResourceFieldValue("vcd_network_routed_v2.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "id"),
					resource.TestMatchResourceAttr("vcd_network_routed_v2.net1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.net1", "vdc", testConfig.Nsxt.Vdc),
					resourceFieldsEqual("data.vcd_network_routed_v2.net1", "vcd_network_routed_v2.net1", []string{"%"}),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNetworkRoutedV2InheritedVdcStep1 = `
data "vcd_external_network_v2" "existing-extnet" {
	name = "{{.ExternalNetwork}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org  = "{{.Org}}"
  name = "{{.NsxtEdgeGatewayVcd}}"

  external_network_id = data.vcd_external_network_v2.existing-extnet.id

  subnet {
     gateway               = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].gateway
     prefix_length         = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].prefix_length

     primary_ip            = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }
}

resource "vcd_network_routed_v2" "net1" {
  org  = "{{.Org}}"
  name = "{{.NetworkName}}"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  gateway = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
	start_address = "1.1.1.10"
    end_address   = "1.1.1.20"
  }
}
`

const testAccVcdNetworkRoutedV2InheritedVdcStep3 = testAccVcdNetworkRoutedV2InheritedVdcStep1 + `
# skip-binary-test: Data Source test
data "vcd_network_routed_v2" "net1" {
  org  = "{{.Org}}"
  name = "{{.NetworkName}}"
}

data "vcd_nsxt_edgegateway" "nsxt-edge" {
  org  = "{{.Org}}"
  name = "{{.NsxtEdgeGatewayVcd}}"
}
`

const testAccVcdNetworkRoutedV2InheritedVdcStep4 = `
data "vcd_external_network_v2" "existing-extnet" {
	name = "{{.ExternalNetwork}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "{{.NsxtEdgeGatewayVcd}}"

  external_network_id = data.vcd_external_network_v2.existing-extnet.id

  subnet {
     gateway               = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].gateway
     prefix_length         = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].prefix_length

     primary_ip            = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }
}

resource "vcd_network_routed_v2" "net1" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "{{.NetworkName}}"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  gateway = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
	start_address = "1.1.1.10"
    end_address   = "1.1.1.20"
  }
}
`

const testAccVcdNetworkRoutedV2InheritedVdcStep6 = testAccVcdNetworkRoutedV2InheritedVdcStep4 + `
# skip-binary-test: Data Source test
data "vcd_network_routed_v2" "net1" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "{{.NetworkName}}"
}

data "vcd_nsxt_edgegateway" "nsxt-edge" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "{{.NsxtEdgeGatewayVcd}}"
}
`
