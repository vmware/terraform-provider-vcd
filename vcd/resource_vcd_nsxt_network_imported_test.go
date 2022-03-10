//go:build network || nsxt || ALL || functional
// +build network nsxt ALL functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNsxtNetworkImported(t *testing.T) {
	preTestChecks(t)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	vcdClient := createTemporaryVCDConnection(false)
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
		"Tags":              "network nsxt",
	}

	configText := templateFill(TestAccVcdNetworkImportedV2NsxtStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(TestAccVcdNetworkImportedV2NsxtStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	// Ensure the resource is never recreated - ID stays the same
	cachedId := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckOpenApiVcdNetworkDestroy(testConfig.Nsxt.Vdc, t.Name()),
		Steps: []resource.TestStep{
			{ // step 1
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.cacheTestResourceFieldValue("vcd_nsxt_network_imported.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_imported.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_imported.net1", "nsxt_logical_switch_id"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.net1", "name", "nsxt-imported-test-initial"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.net1", "description", "NSX-T imported network test OpenAPI"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.net1", "gateway", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.net1", "prefix_length", "24"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.net1", "static_ip_pool.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_imported.net1", "static_ip_pool.*", map[string]string{
						"start_address": "1.1.1.10",
						"end_address":   "1.1.1.20",
					}),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_imported.net1", "owner_id"),
				),
			},
			{ // step 2
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.cacheTestResourceFieldValue("vcd_nsxt_network_imported.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_imported.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_imported.net1", "nsxt_logical_switch_id"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.net1", "name", "updated-nsxt-imported-test-initial"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.net1", "description", "Updated NSX-T imported network test OpenAPI"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.net1", "gateway", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.net1", "prefix_length", "24"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.net1", "static_ip_pool.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_imported.net1", "static_ip_pool.*", map[string]string{
						"start_address": "1.1.1.10",
						"end_address":   "1.1.1.20",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_imported.net1", "static_ip_pool.*", map[string]string{
						"start_address": "1.1.1.40",
						"end_address":   "1.1.1.50",
					}),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_imported.net1", "owner_id"),
				),
			},
			// Check that import works
			{ // step 3
				ResourceName:      "vcd_nsxt_network_imported.net1",
				ImportState:       true,
				ImportStateVerify: true,
				// It is impossible to read 'nsxt_logical_switch_name' for already consumed NSX-T segment (API returns
				// only unused segments) therefore this field cannot be set during read operations.
				ImportStateVerifyIgnore: []string{"nsxt_logical_switch_name", "vdc"},
				ImportStateIdFunc:       importStateIdOrgNsxtVdcObject(testConfig, "updated-nsxt-imported-test-initial"),
			},
		},
	})
	postTestChecks(t)
}

const TestAccVcdNetworkImportedV2NsxtStep1 = `
resource "vcd_nsxt_network_imported" "net1" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "nsxt-imported-test-initial"
  description = "NSX-T imported network test OpenAPI"

  nsxt_logical_switch_name = "{{.NsxtImportSegment}}"

  gateway       = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
	start_address = "1.1.1.10"
	end_address   = "1.1.1.20"
  }
}
`

const TestAccVcdNetworkImportedV2NsxtStep2 = `
resource "vcd_nsxt_network_imported" "net1" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "updated-nsxt-imported-test-initial"
  description = "Updated NSX-T imported network test OpenAPI"

  nsxt_logical_switch_name = "{{.NsxtImportSegment}}"

  gateway       = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
	start_address = "1.1.1.10"
	end_address   = "1.1.1.20"
  }

  static_ip_pool {
	start_address = "1.1.1.40"
	end_address   = "1.1.1.50"
  }
}
`

// TestAccVcdNsxtNetworkImportedOwnerIsVdc tests a case where VDC ID is specified as `owner_id`
// on the first run
func TestAccVcdNsxtNetworkImportedOwnerIsVdc(t *testing.T) {
	preTestChecks(t)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	skipNoNsxtConfiguration(t)

	// String map to fill the template
	var params = StringMap{
		"Name":              t.Name(),
		"Org":               testConfig.VCD.Org,
		"NsxtVdc":           testConfig.Nsxt.Vdc,
		"EdgeGw":            testConfig.Nsxt.EdgeGateway,
		"NetworkName":       t.Name(),
		"NsxtImportSegment": testConfig.Nsxt.NsxtImportSegment,

		"Tags": "network nsxt",
	}

	configText := templateFill(testAccVcdNetworkImportedV2NsxtOwnerIsVdcStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdNetworkImportedV2NsxtOwnerIsVdcStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	// Ensure the resource is never recreated - ID stays the same
	cachedId := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckOpenApiVcdNetworkDestroy(testConfig.Nsxt.Vdc, t.Name()),
		Steps: []resource.TestStep{
			{ // step 1
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.cacheTestResourceFieldValue("vcd_nsxt_network_imported.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_imported.net1", "id"),
					resource.TestCheckResourceAttrPair("vcd_nsxt_network_imported.net1", "owner_id", "data.vcd_org_vdc.nsxtvdc", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_imported.net1", "nsxt_logical_switch_id"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.net1", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.net1", "description", "NSX-T imported network test OpenAPI"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.net1", "gateway", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.net1", "prefix_length", "24"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.net1", "static_ip_pool.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_imported.net1", "static_ip_pool.*", map[string]string{
						"start_address": "1.1.1.10",
						"end_address":   "1.1.1.20",
					}),
				),
			},
			{ // step 2
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.cacheTestResourceFieldValue("vcd_nsxt_network_imported.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_imported.net1", "id"),
					resource.TestCheckResourceAttrPair("vcd_nsxt_network_imported.net1", "owner_id", "data.vcd_org_vdc.nsxtvdc", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_imported.net1", "nsxt_logical_switch_id"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.net1", "name", t.Name()+"-updated"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.net1", "description", "Updated NSX-T imported network test OpenAPI"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.net1", "gateway", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.net1", "prefix_length", "24"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.net1", "static_ip_pool.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_imported.net1", "static_ip_pool.*", map[string]string{
						"start_address": "1.1.1.10",
						"end_address":   "1.1.1.20",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_imported.net1", "static_ip_pool.*", map[string]string{
						"start_address": "1.1.1.40",
						"end_address":   "1.1.1.50",
					}),
				),
			},
			// Check that import works
			{ // step 3
				ResourceName:      "vcd_nsxt_network_imported.net1",
				ImportState:       true,
				ImportStateVerify: true,
				// It is impossible to read 'nsxt_logical_switch_name' for already consumed NSX-T segment (API returns
				// only unused segments) therefore this field cannot be set during read operations.
				ImportStateVerifyIgnore: []string{"nsxt_logical_switch_name", "vdc"},
				ImportStateIdFunc:       importStateIdOrgNsxtVdcGroupObject(testConfig, testConfig.Nsxt.Vdc, t.Name()+"-updated"),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNetworkImportedV2NsxtOwnerIsVdcStep1 = `
data "vcd_org_vdc" "nsxtvdc" {
  org  = "{{.Org}}"
  name = "{{.NsxtVdc}}"
}

resource "vcd_nsxt_network_imported" "net1" {
  org         = "{{.Org}}"
  owner_id    = data.vcd_org_vdc.nsxtvdc.id
  name        = "{{.Name}}"
  description = "NSX-T imported network test OpenAPI"

  nsxt_logical_switch_name = "{{.NsxtImportSegment}}"

  gateway       = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
	start_address = "1.1.1.10"
	end_address   = "1.1.1.20"
  }
}
`

const testAccVcdNetworkImportedV2NsxtOwnerIsVdcStep2 = `
data "vcd_org_vdc" "nsxtvdc" {
  org  = "{{.Org}}"
  name = "{{.NsxtVdc}}"
}

resource "vcd_nsxt_network_imported" "net1" {
  org         = "{{.Org}}"
  owner_id    = data.vcd_org_vdc.nsxtvdc.id
  name        = "{{.Name}}-updated"
  description = "Updated NSX-T imported network test OpenAPI"

  nsxt_logical_switch_name = "{{.NsxtImportSegment}}"

  gateway       = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
	start_address = "1.1.1.10"
	end_address   = "1.1.1.20"
  }

  static_ip_pool {
	start_address = "1.1.1.40"
	end_address   = "1.1.1.50"
  }
}
`

// TestAccVcdNsxtNetworkImportedInVdcGroup tests a case where network is created directly in VDC
// Group using owner_id reference.
func TestAccVcdNsxtNetworkImportedInVdcGroup(t *testing.T) {
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
		"NsxtImportSegment":         testConfig.Nsxt.NsxtImportSegment,

		"Tags": "network",
	}

	params["FuncName"] = t.Name() + "-newVdc"
	configTextPre := templateFill(testAccVcdVdcGroupNew, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configTextPre)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdNsxtNetworkImportedInVdcGroup, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	// Ensure the resource is never recreated - ID stays the same
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
					cachedId.cacheTestResourceFieldValue("vcd_nsxt_network_imported.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_imported.net1", "id"),
					resource.TestCheckResourceAttrPair("vcd_nsxt_network_imported.net1", "owner_id", "vcd_vdc_group.test1", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_imported.net1", "nsxt_logical_switch_id"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.net1", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.net1", "description", "NSX-T imported network test OpenAPI"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.net1", "gateway", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.net1", "prefix_length", "24"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.net1", "static_ip_pool.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_imported.net1", "static_ip_pool.*", map[string]string{
						"start_address": "1.1.1.10",
						"end_address":   "1.1.1.20",
					}),
				),
			},

			{ // step 3
				ResourceName:      "vcd_nsxt_network_imported.net1",
				ImportState:       true,
				ImportStateVerify: true,
				// It is impossible to read 'nsxt_logical_switch_name' for already consumed NSX-T segment (API returns
				// only unused segments) therefore this field cannot be set during read operations.
				ImportStateVerifyIgnore: []string{"nsxt_logical_switch_name", "vdc"},
				ImportStateIdFunc:       importStateIdOrgNsxtVdcGroupObject(testConfig, t.Name(), t.Name()),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtNetworkImportedInVdcGroup = testAccVcdVdcGroupNew + `
resource "vcd_nsxt_network_imported" "net1" {
  org         = "{{.Org}}"
  owner_id    = vcd_vdc_group.test1.id
  name        = "{{.Name}}"
  description = "NSX-T imported network test OpenAPI"

  nsxt_logical_switch_name = "{{.NsxtImportSegment}}"

  gateway       = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
	start_address = "1.1.1.10"
	end_address   = "1.1.1.20"
  }
}
`

// TestAccVcdNetworkImportedNsxtMigration aims to test migration scenario from pre 3.6.0 configuration
// to new one using `owner_id` and VDC group support
// * Step 1 - creates prerequisite VDCs and VDC Group
// * Step 2 - creates an imported network using `vdc` field
// * Step 3 - removes `vdc` field from configuration and uses `owner_id` pointing to the same VDC
// * Step 4 - changes `owner_id` field value from a VDC to a VDC Group to migrate the network
// * Step 5 - verifies that `terraform import` works when an imported network is a member of VDC
// Group
// * Step 6 - migrates the network to different VDC in VDC Group
// * Step 7 - checks out that import of network being in different VDC still works
func TestAccVcdNetworkImportedNsxtMigration(t *testing.T) {
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
		"NsxtImportSegment":         testConfig.Nsxt.NsxtImportSegment,

		"Tags": "network",
	}

	params["FuncName"] = t.Name() + "-newVdc"
	configTextPre := templateFill(testAccVcdVdcGroupNew, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configTextPre)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdNetworkImportedNsxtMigrationStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcdNetworkImportedNsxtMigrationStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(testAccVcdNetworkImportedNsxtMigrationStep4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	params["FuncName"] = t.Name() + "-step5"
	configText6 := templateFill(testAccVcdNetworkImportedNsxtMigrationStep6, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s", configText6)

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
					cachedId.cacheTestResourceFieldValue("vcd_nsxt_network_imported.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_imported.net1", "id"),
					resource.TestMatchResourceAttr("vcd_nsxt_network_imported.net1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.testCheckCachedResourceFieldValue("vcd_nsxt_network_imported.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_imported.net1", "id"),
					resource.TestMatchResourceAttr("vcd_nsxt_network_imported.net1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttrPair("vcd_nsxt_network_imported.net1", "owner_id", "vcd_org_vdc.newVdc.0", "id"),
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.testCheckCachedResourceFieldValue("vcd_nsxt_network_imported.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_imported.net1", "id"),
					resource.TestMatchResourceAttr("vcd_nsxt_network_imported.net1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdcGroup:`)),
					resource.TestCheckResourceAttrPair("vcd_nsxt_network_imported.net1", "owner_id", "vcd_vdc_group.test1", "id"),
				),
			},
			{
				ResourceName:      "vcd_nsxt_network_imported.net1",
				ImportState:       true,
				ImportStateVerify: true,
				// It is impossible to read 'nsxt_logical_switch_name' for already consumed NSX-T segment (API returns
				// only unused segments) therefore this field cannot be set during read operations.
				ImportStateVerifyIgnore: []string{"nsxt_logical_switch_name", "vdc"},
				ImportStateIdFunc:       importStateIdOrgNsxtVdcGroupObject(testConfig, t.Name(), t.Name()),
			},
			{
				Config: configText6,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.testCheckCachedResourceFieldValue("vcd_nsxt_network_imported.net1", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_imported.net1", "id"),
					resource.TestMatchResourceAttr("vcd_nsxt_network_imported.net1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttrPair("vcd_nsxt_network_imported.net1", "owner_id", "vcd_org_vdc.newVdc.1", "id"),
				),
			},
			{
				ResourceName:      "vcd_nsxt_network_imported.net1",
				ImportState:       true,
				ImportStateVerify: true,
				// It is impossible to read 'nsxt_logical_switch_name' for already consumed NSX-T
				// segment (API returns only unused segments) therefore this field cannot be set
				// during read operations.
				ImportStateVerifyIgnore: []string{"nsxt_logical_switch_name", "vdc"},
				ImportStateIdFunc:       importStateIdOrgNsxtVdcGroupObject(testConfig, t.Name()+"-1", t.Name()),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNetworkImportedNsxtMigrationStep2 = testAccVcdVdcGroupNew + `
resource "vcd_nsxt_network_imported" "net1" {
  org         = "{{.Org}}"
  vdc         = vcd_org_vdc.newVdc.0.name
  name        = "{{.Name}}"
  description = "NSX-T imported network test OpenAPI"

  nsxt_logical_switch_name = "{{.NsxtImportSegment}}"

  gateway       = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
	start_address = "1.1.1.10"
	end_address   = "1.1.1.20"
  }
}
`

const testAccVcdNetworkImportedNsxtMigrationStep3 = testAccVcdVdcGroupNew + `
resource "vcd_nsxt_network_imported" "net1" {
  org       = "{{.Org}}"
  owner_id  = vcd_org_vdc.newVdc.0.id  
  name        = "{{.Name}}"
  description = "NSX-T imported network test OpenAPI"
  
  nsxt_logical_switch_name = "{{.NsxtImportSegment}}"
  
  gateway       = "1.1.1.1"
  prefix_length = 24
  
  static_ip_pool {
    start_address = "1.1.1.10"
    end_address   = "1.1.1.20"
  }
}
`

const testAccVcdNetworkImportedNsxtMigrationStep4 = testAccVcdVdcGroupNew + `
resource "vcd_nsxt_network_imported" "net1" {
  org       = "{{.Org}}"
  owner_id = vcd_vdc_group.test1.id
  name        = "{{.Name}}"
  description = "NSX-T imported network test OpenAPI"
  
  nsxt_logical_switch_name = "{{.NsxtImportSegment}}"
  
  gateway       = "1.1.1.1"
  prefix_length = 24
  
  static_ip_pool {
    start_address = "1.1.1.10"
    end_address   = "1.1.1.20"
  }
}
`

const testAccVcdNetworkImportedNsxtMigrationStep6 = testAccVcdVdcGroupNew + `
resource "vcd_nsxt_network_imported" "net1" {
  org       = "{{.Org}}"
  owner_id = vcd_org_vdc.newVdc.1.id  
  name        = "{{.Name}}"
  description = "NSX-T imported network test OpenAPI"
  
  nsxt_logical_switch_name = "{{.NsxtImportSegment}}"
  
  gateway       = "1.1.1.1"
  prefix_length = 24
  
  static_ip_pool {
    start_address = "1.1.1.10"
    end_address   = "1.1.1.20"
  }
}
`
