//go:build network || nsxt || ALL || functional
// +build network nsxt ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdNsxVdcGroupCompleteMigration test aims to check integration of resource migration from
// old configuration to new configured.
// * Step 1 - creates prerequisites - two VDCs and a VDC Group
// * Step 2 - creates
// All the checks carried out in steps are related to vdc/owner related fields
// TODO Remove this test when 4.0 is released
func TestAccVcdNsxVdcGroupCompleteMigration(t *testing.T) {
	preTestChecks(t)

	// This test requires access to the vCD before filling templates
	// Thus it won't run in the short test
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	vcdClient := createTemporaryVCDConnection(false)

	if !vcdClient.Client.IsSysAdmin {
		t.Skip(t.Name() + " only System Administrator can run test of VDC Group")
	}

	if testConfig.Nsxt.Vdc == "" || testConfig.VCD.NsxtProviderVdc.Name == "" ||
		testConfig.VCD.NsxtProviderVdc.NetworkPool == "" || testConfig.VCD.ProviderVdc.StorageProfile == "" {
		t.Skip("Variables Nsxt.Vdc, VCD.NsxtProviderVdc.NetworkPool, VCD.NsxtProviderVdc.Name," +
			" VCD.ProviderVdc.StorageProfile  must be set")
	}

	// String map to fill the template
	var params = StringMap{
		"Org":                       testConfig.VCD.Org,
		"VDC":                       testConfig.Nsxt.Vdc,
		"NameUpdated":               "TestAccVcdVdcGroupResourceUpdated",
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"Allocated":                 "1024",
		"Limit":                     "1024",
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"Dfw":                       "false",
		"DefaultPolicy":             "false",
		"NsxtImportSegment":         testConfig.Nsxt.NsxtImportSegment,
		"Name":                      t.Name(),
		"TestName":                  t.Name(),
		"NsxtExternalNetworkName":   testConfig.Nsxt.ExternalNetwork,

		"Tags": "vdc nsxt vdcGroup",
	}

	params["FuncName"] = t.Name() + "-newVdc"
	configTextPre := templateFill(testAccVcdVdcGroupNew, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configTextPre)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdNsxVdcGroupCompleteMigrationStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "step3"
	configText3 := templateFill(testAccVcdNsxVdcGroupCompleteMigrationStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "step5"
	configText5 := templateFill(testAccVcdNsxVdcGroupCompleteMigrationStep5DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s", configText5)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// use cache fields to check that IDs remain the same accross multiple steps (this proved that
	// resources were not recreated)
	edgeGatewayId := testCachedFieldValue{}
	routedNetId := testCachedFieldValue{}
	isolatedNetId := testCachedFieldValue{}
	importedNetId := testCachedFieldValue{}
	networkDhcpPools := testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: configTextPre,
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					edgeGatewayId.cacheTestResourceFieldValue("vcd_nsxt_edgegateway.nsxt-edge", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "vdc", fmt.Sprintf("%s-%s", t.Name(), "0")),
					resource.TestCheckResourceAttrPair("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", "vcd_org_vdc.newVdc.0", "id"),

					routedNetId.cacheTestResourceFieldValue("vcd_network_routed_v2.nsxt-backed", "id"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.nsxt-backed", "vdc", fmt.Sprintf("%s-%s", t.Name(), "0")),
					resource.TestCheckResourceAttrPair("vcd_network_routed_v2.nsxt-backed", "owner_id", "vcd_org_vdc.newVdc.0", "id"),

					isolatedNetId.cacheTestResourceFieldValue("vcd_network_isolated_v2.nsxt-backed", "id"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.nsxt-backed", "vdc", fmt.Sprintf("%s-%s", t.Name(), "0")),
					resource.TestCheckResourceAttrPair("vcd_network_isolated_v2.nsxt-backed", "owner_id", "vcd_org_vdc.newVdc.0", "id"),

					importedNetId.cacheTestResourceFieldValue("vcd_nsxt_network_imported.nsxt-backed", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.nsxt-backed", "vdc", fmt.Sprintf("%s-%s", t.Name(), "0")),
					resource.TestCheckResourceAttrPair("vcd_nsxt_network_imported.nsxt-backed", "owner_id", "vcd_org_vdc.newVdc.0", "id"),

					networkDhcpPools.cacheTestResourceFieldValue("vcd_nsxt_network_dhcp.pools", "id"),
					resource.TestCheckResourceAttrPair("vcd_nsxt_network_dhcp.pools", "id", "vcd_network_routed_v2.nsxt-backed", "id"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					edgeGatewayId.testCheckCachedResourceFieldValue("vcd_nsxt_edgegateway.nsxt-edge", "id"),
					resource.TestCheckResourceAttrPair("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", "vcd_vdc_group.test1", "id"),
					resource.TestCheckResourceAttrPair("vcd_nsxt_edgegateway.nsxt-edge", "vdc", "vcd_vdc_group.test1", "name"),

					// This test is explicitly skipped during apply because routed network migrates
					// with Edge Gateway and first read might happen earlier than parent Edge
					// Gateway is moved. These fields will stay undocumented, but are useful for
					// testing.
					//
					// resource.TestCheckResourceAttrPair("vcd_network_routed_v2.nsxt-backed", "owner_id", "vcd_vdc_group.test1", "id"),
					// resource.TestCheckResourceAttrPair("vcd_network_routed_v2.nsxt-backed", "vdc", "vcd_vdc_group.test1", "name"),
					routedNetId.testCheckCachedResourceFieldValue("vcd_network_routed_v2.nsxt-backed", "id"),

					isolatedNetId.testCheckCachedResourceFieldValue("vcd_network_isolated_v2.nsxt-backed", "id"),
					resource.TestCheckResourceAttrPair("vcd_network_isolated_v2.nsxt-backed", "owner_id", "vcd_vdc_group.test1", "id"),
					resource.TestCheckResourceAttrPair("vcd_network_isolated_v2.nsxt-backed", "vdc", "vcd_vdc_group.test1", "name"),

					importedNetId.testCheckCachedResourceFieldValue("vcd_nsxt_network_imported.nsxt-backed", "id"),
					resource.TestCheckResourceAttrPair("vcd_nsxt_network_imported.nsxt-backed", "owner_id", "vcd_vdc_group.test1", "id"),
					resource.TestCheckResourceAttrPair("vcd_nsxt_network_imported.nsxt-backed", "vdc", "vcd_vdc_group.test1", "name"),

					networkDhcpPools.testCheckCachedResourceFieldValue("vcd_nsxt_network_dhcp.pools", "id"),
					resource.TestCheckResourceAttrPair("vcd_nsxt_network_dhcp.pools", "id", "vcd_network_routed_v2.nsxt-backed", "id"),
				),
			},
			{
				// The same config is applied once more to verify that routed network finally
				// reports 'owner_id' and 'vdc' fields to correct name.
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					edgeGatewayId.testCheckCachedResourceFieldValue("vcd_nsxt_edgegateway.nsxt-edge", "id"),
					resource.TestCheckResourceAttrPair("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", "vcd_vdc_group.test1", "id"),
					resource.TestCheckResourceAttrPair("vcd_nsxt_edgegateway.nsxt-edge", "vdc", "vcd_vdc_group.test1", "name"),

					routedNetId.testCheckCachedResourceFieldValue("vcd_network_routed_v2.nsxt-backed", "id"),
					resource.TestCheckResourceAttrPair("vcd_network_routed_v2.nsxt-backed", "owner_id", "vcd_vdc_group.test1", "id"),
					resource.TestCheckResourceAttrPair("vcd_network_routed_v2.nsxt-backed", "vdc", "vcd_vdc_group.test1", "name"),

					isolatedNetId.testCheckCachedResourceFieldValue("vcd_network_isolated_v2.nsxt-backed", "id"),
					resource.TestCheckResourceAttrPair("vcd_network_isolated_v2.nsxt-backed", "owner_id", "vcd_vdc_group.test1", "id"),
					resource.TestCheckResourceAttrPair("vcd_network_isolated_v2.nsxt-backed", "vdc", "vcd_vdc_group.test1", "name"),

					importedNetId.testCheckCachedResourceFieldValue("vcd_nsxt_network_imported.nsxt-backed", "id"),
					resource.TestCheckResourceAttrPair("vcd_nsxt_network_imported.nsxt-backed", "owner_id", "vcd_vdc_group.test1", "id"),
					resource.TestCheckResourceAttrPair("vcd_nsxt_network_imported.nsxt-backed", "vdc", "vcd_vdc_group.test1", "name"),

					networkDhcpPools.testCheckCachedResourceFieldValue("vcd_nsxt_network_dhcp.pools", "id"),
					resource.TestCheckResourceAttrPair("vcd_nsxt_network_dhcp.pools", "id", "vcd_network_routed_v2.nsxt-backed", "id"),
				),
			},
			{
				// Data source testing
				Config: configText5,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcd_nsxt_network_dhcp.pools", "data.vcd_nsxt_network_dhcp.pools", []string{"vdc"}),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxVdcGroupCompleteMigrationStep2 = testAccVcdVdcGroupNew + `
data "vcd_external_network_v2" "nsxt-ext-net" {
  name = "{{.NsxtExternalNetworkName}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.newVdc.0.name

  name = "{{.Name}}"

  external_network_id = data.vcd_external_network_v2.nsxt-ext-net.id

  subnet {
     gateway       = tolist(data.vcd_external_network_v2.nsxt-ext-net.ip_scope)[0].gateway
     prefix_length = tolist(data.vcd_external_network_v2.nsxt-ext-net.ip_scope)[0].prefix_length

     primary_ip = tolist(tolist(data.vcd_external_network_v2.nsxt-ext-net.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.nsxt-ext-net.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.nsxt-ext-net.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }
}

resource "vcd_network_routed_v2" "nsxt-backed" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.newVdc.0.name

  name = "{{.Name}}-routed"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  gateway       = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "1.1.1.10"
    end_address   = "1.1.1.20"
  }
}

resource "vcd_nsxt_network_dhcp" "pools" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.newVdc.0.name

  org_network_id = vcd_network_routed_v2.nsxt-backed.id

  pool {
    start_address = "1.1.1.100"
    end_address   = "1.1.1.110"
  }

  pool {
    start_address = "1.1.1.111"
    end_address   = "1.1.1.112"
  }
}

resource "vcd_network_isolated_v2" "nsxt-backed" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.newVdc.0.name

  name = "{{.Name}}-isolated"

  gateway       = "2.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "2.1.1.10"
    end_address   = "2.1.1.20"
  }
}

resource "vcd_nsxt_network_imported" "nsxt-backed" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.newVdc.0.name

  name = "{{.Name}}-imported"

  nsxt_logical_switch_name = "{{.NsxtImportSegment}}"

  gateway       = "4.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "4.1.1.10"
    end_address   = "4.1.1.20"
  }
}
`

const testAccVcdNsxVdcGroupCompleteMigrationStep3 = testAccVcdVdcGroupNew + `
data "vcd_external_network_v2" "nsxt-ext-net" {
  name = "{{.NsxtExternalNetworkName}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org      = "{{.Org}}"
  owner_id = vcd_vdc_group.test1.id
  name     = "{{.Name}}"

  external_network_id = data.vcd_external_network_v2.nsxt-ext-net.id

  subnet {
     gateway       = tolist(data.vcd_external_network_v2.nsxt-ext-net.ip_scope)[0].gateway
     prefix_length = tolist(data.vcd_external_network_v2.nsxt-ext-net.ip_scope)[0].prefix_length

     primary_ip = tolist(tolist(data.vcd_external_network_v2.nsxt-ext-net.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.nsxt-ext-net.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.nsxt-ext-net.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }
}

resource "vcd_network_routed_v2" "nsxt-backed" {
  org             = "{{.Org}}"
  name            = "{{.Name}}-routed"
  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  gateway       = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "1.1.1.10"
    end_address   = "1.1.1.20"
  }
}

resource "vcd_nsxt_network_dhcp" "pools" {
  org = "{{.Org}}"

  org_network_id = vcd_network_routed_v2.nsxt-backed.id

  pool {
    start_address = "1.1.1.100"
    end_address   = "1.1.1.110"
  }

  pool {
    start_address = "1.1.1.111"
    end_address   = "1.1.1.112"
  }
}

resource "vcd_network_isolated_v2" "nsxt-backed" {
  org      = "{{.Org}}"
  owner_id = vcd_vdc_group.test1.id

  name = "{{.Name}}-isolated"

  gateway       = "2.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "2.1.1.10"
    end_address   = "2.1.1.20"
  }
}

resource "vcd_nsxt_network_imported" "nsxt-backed" {
  org      = "{{.Org}}"
  owner_id = vcd_vdc_group.test1.id

  name = "{{.Name}}-imported"

  nsxt_logical_switch_name = "{{.NsxtImportSegment}}"

  gateway       = "4.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "4.1.1.10"
    end_address   = "4.1.1.20"
  }
}
`

const testAccVcdNsxVdcGroupCompleteMigrationStep5DS = testAccVcdNsxVdcGroupCompleteMigrationStep3 + `
# skip-binary-test: Data Source test

data "vcd_nsxt_network_dhcp" "pools" {
  org = "{{.Org}}"

  org_network_id = vcd_network_routed_v2.nsxt-backed.id
}
`
