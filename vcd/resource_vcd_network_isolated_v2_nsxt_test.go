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
	vcdClient := createTemporaryVCDConnection(false)
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

func TestAccVcdNetworkIsolatedV2NsxtMigration(t *testing.T) {
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
		"NsxtEdgeGatewayVcd":        t.Name() + "-edge",

		"Tags": "network",
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

	params["FuncName"] = t.Name() + "-step5"
	configText5 := templateFill(testAccVcdNetworkIsolatedV2NsxtMigrationStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s", configText5)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	cachedId := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
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
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					// stateDumper(),
					cachedId.testCheckCachedResourceFieldValue("vcd_network_isolated_v2.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_isolated_v2.net1", "id"),
					resource.TestMatchResourceAttr("vcd_network_isolated_v2.net1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "is_shared", "false"),
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					// stateDumper(),
					cachedId.testCheckCachedResourceFieldValue("vcd_network_isolated_v2.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_isolated_v2.net1", "id"),
					resource.TestMatchResourceAttr("vcd_network_isolated_v2.net1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdcGroup:`)),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "is_shared", "true"),
				),
			},
			{
				Config: configText5,
				Check: resource.ComposeAggregateTestCheckFunc(
					// stateDumper(),
					cachedId.testCheckCachedResourceFieldValue("vcd_network_isolated_v2.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_network_isolated_v2.net1", "id"),
					resource.TestMatchResourceAttr("vcd_network_isolated_v2.net1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "is_shared", "false"),
				),
			},

			// // Check that import works
			// { // step 3
			// 	ResourceName:      "vcd_network_isolated_v2.net1",
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// 	// ImportStateIdFunc: importStateIdOrgNsxtVdcObject(testConfig, t.Name()),
			// 	ImportStateId: fmt.Sprintf("%s.%s.%s", testConfig.VCD.Org, params["Name"].(string), params["Name"].(string)),
			// },

			// { // step 4
			// 	Config: configText3,
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		cachedId.testCheckCachedResourceFieldValue("vcd_network_isolated_v2.net1", "id"),
			// 		resource.TestCheckResourceAttrSet("vcd_network_isolated_v2.net1", "id"),
			// 		resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "name", t.Name()),
			// 		resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "description", "Updated"),
			// 		resource.TestCheckResourceAttrSet("vcd_network_isolated_v2.net1", "edge_gateway_id"),
			// 		resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "gateway", "1.1.1.1"),
			// 		resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "prefix_length", "24"),
			// 		resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "static_ip_pool.#", "0"),
			// 		resource.TestCheckResourceAttrPair("data.vcd_nsxt_edgegateway.existing", "owner_id", "vcd_network_isolated_v2.net1", "owner_id"),
			// 	),
			// },
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
}
`

const testAccVcdNetworkIsolatedV2NsxtMigrationStep3 = testAccVcdVdcGroupNew + `
resource "vcd_network_isolated_v2" "net1" {
  org       = "{{.Org}}"
  owner_id  = vcd_org_vdc.newVdc.0.id

  name = "{{.NetworkName}}"
  
  gateway = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
	start_address = "1.1.1.10"
	end_address = "1.1.1.20"
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
		PreCheck:          func() { testAccPreCheck(t) },
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
				// ImportStateIdFunc: importStateIdOrgNsxtVdcObject(testConfig, t.Name()),
				ImportStateId: fmt.Sprintf("%s.%s.%s", testConfig.VCD.Org, params["Name"].(string), params["Name"].(string)),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.cacheTestResourceFieldValue("vcd_network_isolated_v2.net1", "id"),
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
resource "vcd_org_vdc" "newVdc" {

  name = "{{.TestName}}"
  org  = "{{.Org}}"

  allocation_model  = "Flex"
  network_pool_name = "{{.NetworkPool}}"
  provider_vdc_name = "{{.ProviderVdc}}"

  compute_capacity {
    cpu {
      allocated = "1024"
      limit     = "1024"
    }

    memory {
      allocated = "1024"
      limit     = "1024"
    }
  }

  storage_profile {
    name    = "{{.ProviderVdcStorageProfile}}"
    enabled = true
    limit   = 10240
    default = true
  }

  network_quota = 100

  enabled                    = true
  enable_thin_provisioning   = true
  enable_fast_provisioning   = true
  delete_force               = true
  delete_recursive           = true
  elasticity      			     = true
  include_vm_memory_overhead = true
}

resource "vcd_network_isolated_v2" "net1" {
  org  = "{{.Org}}"
  
  owner_id = vcd_org_vdc.newVdc.id
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
