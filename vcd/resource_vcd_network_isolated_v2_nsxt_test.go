//go:build network || nsxt || ALL || functional
// +build network nsxt ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdNetworkIsolatedV2Nsxt tests out NSX-T backed Org VDC networking capabilities
func TestAccVcdNetworkIsolatedV2Nsxt(t *testing.T) {
	preTestChecks(t)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	skipNoNsxtConfiguration(t)

	// String map to fill the template
	var params = StringMap{
		"Org":                  testConfig.VCD.Org,
		"NsxtVdc":              testConfig.Nsxt.Vdc,
		"NetworkName":          t.Name(),
		"Tags":                 "network nsxt",
		"MetadataKey":          "key1",
		"MetadataValue":        "value1",
		"MetadataKeyUpdated":   "key2",
		"MetadataValueUpdated": "value2",
	}

	configText := templateFill(testAccVcdNetworkIsolatedV2NsxtStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdNetworkIsolatedV2NsxtStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcdNetworkIsolatedV2NsxtStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	// Ensure the resource is never recreated - ID stays the same
	cachedId := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testParamsNotEmpty(t, params) },
		CheckDestroy:      testAccCheckOpenApiVcdNetworkDestroy(testConfig.Nsxt.Vdc, t.Name()),
		Steps: []resource.TestStep{
			{ // step 1
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
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "metadata."+params["MetadataKey"].(string), params["MetadataValue"].(string)),
				),
			},
			{ // step 2
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.testCheckCachedResourceFieldValue("vcd_network_isolated_v2.net1", "id"),
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
					resource.TestCheckNoResourceAttr("vcd_network_isolated_v2.net1", "metadata."+params["MetadataKey"].(string)),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "metadata."+params["MetadataKeyUpdated"].(string), params["MetadataValueUpdated"].(string)),
				),
			},
			// Check that import works
			{ // step 3
				ResourceName:      "vcd_network_isolated_v2.net1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcObject(testConfig, t.Name()),
			},
			{ // step 4
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.testCheckCachedResourceFieldValue("vcd_network_isolated_v2.net1", "id"),
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
	postTestChecks(t)
}

const testAccVcdNetworkIsolatedV2NsxtStep1 = `
resource "vcd_network_isolated_v2" "net1" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  
  name        = "nsxt-isolated-test-initial"
  description = "NSX-T isolated network test"

  gateway       = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
	start_address = "1.1.1.10"
	end_address   = "1.1.1.20"
  }

  metadata = {
    {{.MetadataKey}} = "{{.MetadataValue}}"
  }
}
`

const testAccVcdNetworkIsolatedV2NsxtStep2 = `
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

  metadata = {
    {{.MetadataKeyUpdated}} = "{{.MetadataValueUpdated}}"
  }
}
`

const testAccVcdNetworkIsolatedV2NsxtStep3 = `
resource "vcd_network_isolated_v2" "net1" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  
  name          = "{{.NetworkName}}"
  gateway       = "1.1.1.1"
  prefix_length = 24
}
`

// TestAccVcdNetworkIsolatedV2NsxtMigration aims to test backwards compatibility of `vdc` field (in
// resource and inherited from provider configuration) and the possibility to migrate configuration
// from `vdc` field to `owner_id` without recreating resource.
// * Step 1 - creates prerequisites - 2 VDCs and a VDC Group
// * Step 2 - creates an Isolated network using legacy (pre 3.6.0) configuration by using a VDC field
// * Step 3 - replaces field `vdc` with `owner_id` using ID for the same VDC which was used to create in Step 2
// * Step 4 - migrates Isolated network to a VDC Group by using VDC Group ID for `owner_id`
// * Step 5 - verifies that `terraform import` works when an imported network is a member of VDC
// Group
// * Step 6 - migrates Isolated network from VDC Group back to VDC (using the same configuration as in Step 3)
// * Step 7 - checks out that import of network being in different VDC still works
func TestAccVcdNetworkIsolatedV2NsxtMigration(t *testing.T) {
	preTestChecks(t)
	skipNoNsxtConfiguration(t)
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
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"ExternalNetwork":           testConfig.Nsxt.ExternalNetwork,
		"TestName":                  t.Name(),
		"NsxtEdgeGatewayVcd":        t.Name() + "-edge",
		"MetadataKey":               "key1",
		"MetadataValue":             "value1",
		"Tags":                      "network",
	}

	params["FuncName"] = t.Name() + "-newVdc"
	configTextPre := templateFill(testAccVcdVdcGroupNew, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configTextPre)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdNetworkIsolatedV2NsxtMigrationStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcdNetworkIsolatedV2NsxtMigrationStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(testAccVcdNetworkIsolatedV2NsxtMigrationStep4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	params["FuncName"] = t.Name() + "-step6"
	configText6 := templateFill(testAccVcdNetworkIsolatedV2NsxtMigrationStep6, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s", configText6)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	cachedId := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testParamsNotEmpty(t, params) },
		CheckDestroy:      testAccCheckOpenApiVcdNetworkDestroy(testConfig.Nsxt.Vdc, t.Name()),
		Steps: []resource.TestStep{
			{ // step 1 - setup prerequisites
				Config: configTextPre,
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.cacheTestResourceFieldValue("vcd_network_isolated_v2.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_isolated_v2.net1", "id"),
					resource.TestMatchResourceAttr("vcd_network_isolated_v2.net1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "is_shared", "false"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "metadata."+params["MetadataKey"].(string), params["MetadataValue"].(string)),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.testCheckCachedResourceFieldValue("vcd_network_isolated_v2.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_isolated_v2.net1", "id"),
					resource.TestMatchResourceAttr("vcd_network_isolated_v2.net1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "is_shared", "false"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "metadata."+params["MetadataKey"].(string), params["MetadataValue"].(string)),
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.testCheckCachedResourceFieldValue("vcd_network_isolated_v2.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_isolated_v2.net1", "id"),
					resource.TestMatchResourceAttr("vcd_network_isolated_v2.net1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdcGroup:`)),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "is_shared", "true"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "metadata."+params["MetadataKey"].(string), params["MetadataValue"].(string)),
				),
			},
			{
				ResourceName:            "vcd_network_isolated_v2.net1",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata"}, // Network is in a VDC Group, so it can't import metadata
				ImportStateIdFunc:       importStateIdOrgNsxtVdcGroupObject(testConfig, t.Name(), t.Name()),
			},
			{
				Config: configText6,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.testCheckCachedResourceFieldValue("vcd_network_isolated_v2.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_isolated_v2.net1", "id"),
					resource.TestMatchResourceAttr("vcd_network_isolated_v2.net1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "is_shared", "false"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "metadata."+params["MetadataKey"].(string), params["MetadataValue"].(string)),
				),
			},
			{
				ResourceName:      "vcd_network_isolated_v2.net1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcGroupObject(testConfig, t.Name()+"-1", t.Name()),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNetworkIsolatedV2NsxtMigrationStep2 = testAccVcdVdcGroupNew + `
resource "vcd_network_isolated_v2" "net1" {
  org  = "{{.Org}}"
  vdc  = vcd_org_vdc.newVdc.0.name
  name = "{{.NetworkName}}"

  gateway = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
	start_address = "1.1.1.10"
    end_address = "1.1.1.20"
  }

  metadata = {
   {{.MetadataKey}}  = "{{.MetadataValue}}"
  }
}
`

const testAccVcdNetworkIsolatedV2NsxtMigrationStep3 = testAccVcdVdcGroupNew + `
resource "vcd_network_isolated_v2" "net1" {
  org      = "{{.Org}}"
  owner_id = vcd_org_vdc.newVdc.0.id

  name = "{{.NetworkName}}"
  
  gateway = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
	start_address = "1.1.1.10"
	end_address = "1.1.1.20"
  }

  metadata = {
   {{.MetadataKey}}  = "{{.MetadataValue}}"
  }
}
`

const testAccVcdNetworkIsolatedV2NsxtMigrationStep4 = testAccVcdVdcGroupNew + `
resource "vcd_network_isolated_v2" "net1" {
	org      = "{{.Org}}"
	owner_id = vcd_vdc_group.test1.id
  
	name        = "{{.NetworkName}}"
	description = "step4"
	
	gateway = "1.1.1.1"
	prefix_length = 24
  
	static_ip_pool {
	  start_address = "1.1.1.10"
	  end_address = "1.1.1.20"
	}

  metadata = {
   {{.MetadataKey}}  = "{{.MetadataValue}}"
  }
}
`

const testAccVcdNetworkIsolatedV2NsxtMigrationStep6 = testAccVcdVdcGroupNew + `
resource "vcd_network_isolated_v2" "net1" {
  org      = "{{.Org}}"
  owner_id = vcd_org_vdc.newVdc.1.id

  name = "{{.NetworkName}}"
  
  gateway = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
	start_address = "1.1.1.10"
	end_address = "1.1.1.20"
  }

  metadata = {
   {{.MetadataKey}}  = "{{.MetadataValue}}"
  }
}
`

func TestAccVcdNetworkIsolatedV2NsxtOwnerVdc(t *testing.T) {
	preTestChecks(t)
	skipNoNsxtConfiguration(t)

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
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"ExternalNetwork":           testConfig.Nsxt.ExternalNetwork,
		"TestName":                  t.Name(),

		"Tags": "network",
	}

	params["FuncName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccVcdNetworkIsolatedV2NsxtInVdc, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcdNetworkIsolatedV2NsxtInVdcDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	cachedId := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testParamsNotEmpty(t, params) },
		CheckDestroy:      testAccCheckOpenApiVcdNetworkDestroy(testConfig.Nsxt.Vdc, t.Name()),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.cacheTestResourceFieldValue("vcd_network_isolated_v2.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_isolated_v2.net1", "id"),
					resource.TestMatchResourceAttr("vcd_network_isolated_v2.net1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "is_shared", "false"),
				),
			},
			// Check that import works
			{ // step 2
				ResourceName:      "vcd_network_isolated_v2.net1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     fmt.Sprintf("%s.%s.%s", testConfig.VCD.Org, params["NsxtVdc"].(string), params["Name"].(string)),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.testCheckCachedResourceFieldValue("vcd_network_isolated_v2.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_isolated_v2.net1", "id"),
					resource.TestMatchResourceAttr("vcd_network_isolated_v2.net1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "is_shared", "false"),
					// Data source contains "filter" field
					resourceFieldsEqual("data.vcd_network_isolated_v2.net1", "vcd_network_isolated_v2.net1", []string{"%"}),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNetworkIsolatedV2NsxtInVdc = `
data "vcd_org_vdc" "existing" {
  org  = "{{.Org}}"
  name = "{{.NsxtVdc}}"
}

resource "vcd_network_isolated_v2" "net1" {
  org  = "{{.Org}}"
  
  owner_id = data.vcd_org_vdc.existing.id
  name     = "{{.NetworkName}}"

  gateway = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
	start_address = "1.1.1.10"
    end_address = "1.1.1.20"
  }
}
`

const testAccVcdNetworkIsolatedV2NsxtInVdcDS = testAccVcdNetworkIsolatedV2NsxtInVdc + `
data "vcd_network_isolated_v2" "net1" {
  org  = "{{.Org}}"
  
  owner_id = vcd_network_isolated_v2.net1.owner_id
  name     = vcd_network_isolated_v2.net1.name
}
`

// TestAccVcdNetworkIsolatedV2InheritedVdc tests that Isolated network can be created by using `vdc`
// field inherited from provider in NSX-T VDC
// * Step 1 - Rely on configuration comming from `provider` configuration for `vdc` value
// * Step 2 - Test that import works correctly
// * Step 3 - Test that data source works correctly
// * Step 4 - Start using `vdc` fields in resource and make sure it is not recreated
// * Step 5 - Test that import works correctly
// * Step 6 - Test data source
// Note. It does not test `org` field inheritance because our import sets it by default.
func TestAccVcdNetworkIsolatedV2InheritedVdc(t *testing.T) {
	preTestChecks(t)
	skipNoNsxtConfiguration(t)

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"NsxtVdc":     testConfig.Nsxt.Vdc,
		"NetworkName": t.Name(),

		// This particular field is consumed by `templateFill` to generate binary tests with correct
		// default VDC (NSX-T)
		"PrVdc": testConfig.Nsxt.Vdc,

		"Tags": "network",
	}

	// This test explicitly tests that `vdc` field inherited from provider works correctly therefore
	// it must override default `vdc` field value at provider level to be NSX-T VDC and restore it
	// after this test.
	restoreDefaultVdcFunc := overrideDefaultVdcForTest(testConfig.Nsxt.Vdc)
	defer restoreDefaultVdcFunc()

	params["FuncName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccVcdNetworkIsolatedV2InheritedVdcStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcdNetworkIsolatedV2InheritedVdcStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(testAccVcdNetworkIsolatedV2InheritedVdcStep4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	params["FuncName"] = t.Name() + "-step6"
	configText6 := templateFill(testAccVcdNetworkIsolatedV2InheritedVdcStep6, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 6: %s", configText6)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	cachedId := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,

		PreCheck:     func() { testParamsNotEmpty(t, params) },
		CheckDestroy: testAccCheckOpenApiVcdNetworkDestroy(testConfig.Nsxt.Vdc, t.Name()),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.cacheTestResourceFieldValue("vcd_network_isolated_v2.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_isolated_v2.net1", "id"),
					resource.TestMatchResourceAttr("vcd_network_isolated_v2.net1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "vdc", testConfig.Nsxt.Vdc),
				),
			},
			{
				ResourceName:      "vcd_network_isolated_v2.net1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcObject(testConfig, params["NetworkName"].(string)),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.testCheckCachedResourceFieldValue("vcd_network_isolated_v2.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_isolated_v2.net1", "id"),
					resource.TestMatchResourceAttr("vcd_network_isolated_v2.net1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "vdc", testConfig.Nsxt.Vdc),
					// Total field count ('%') differs because data source has additional field 'filter'
					resourceFieldsEqual("data.vcd_network_isolated_v2.net1", "vcd_network_isolated_v2.net1", []string{"%"}),
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.testCheckCachedResourceFieldValue("vcd_network_isolated_v2.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_isolated_v2.net1", "id"),
					resource.TestMatchResourceAttr("vcd_network_isolated_v2.net1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "vdc", testConfig.Nsxt.Vdc),
				),
			},
			{
				ResourceName:      "vcd_network_isolated_v2.net1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcObject(testConfig, params["NetworkName"].(string)),
			},
			{
				Config: configText6,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.testCheckCachedResourceFieldValue("vcd_network_isolated_v2.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_isolated_v2.net1", "id"),
					resource.TestMatchResourceAttr("vcd_network_isolated_v2.net1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "vdc", testConfig.Nsxt.Vdc),
					resourceFieldsEqual("data.vcd_network_isolated_v2.net1", "vcd_network_isolated_v2.net1", []string{"%"}),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNetworkIsolatedV2InheritedVdcStep1 = `
resource "vcd_network_isolated_v2" "net1" {
  org  = "{{.Org}}"
  name = "{{.NetworkName}}"

  gateway = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "1.1.1.10"
    end_address = "1.1.1.20"
  }
}
`

const testAccVcdNetworkIsolatedV2InheritedVdcStep3 = testAccVcdNetworkIsolatedV2InheritedVdcStep1 + `
# skip-binary-test: Data Source test
data "vcd_network_isolated_v2" "net1" {
  org  = "{{.Org}}"
  name = "{{.NetworkName}}"
}
`

const testAccVcdNetworkIsolatedV2InheritedVdcStep4 = `
resource "vcd_network_isolated_v2" "net1" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "{{.NetworkName}}"

  gateway = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "1.1.1.10"
    end_address = "1.1.1.20"
  }
}
`

const testAccVcdNetworkIsolatedV2InheritedVdcStep6 = testAccVcdNetworkIsolatedV2InheritedVdcStep4 + `
# skip-binary-test: Data Source test
data "vcd_network_isolated_v2" "net1" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "{{.NetworkName}}"
}
`
