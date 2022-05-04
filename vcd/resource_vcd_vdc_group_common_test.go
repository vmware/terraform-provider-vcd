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

	params["FuncName"] = t.Name() + "step4"
	configText4 := templateFill(testAccVcdNsxVdcGroupCompleteMigrationStep4DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

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
	firewallId := testCachedFieldValue{}
	snatRuleId := testCachedFieldValue{}
	dnatRuleId := testCachedFieldValue{}
	ipSecVpnTunnelId := testCachedFieldValue{}

	parentVdcGroupName := t.Name()

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

					firewallId.cacheTestResourceFieldValue("vcd_nsxt_firewall.testing", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "vdc", fmt.Sprintf("%s-%s", t.Name(), "0")),

					snatRuleId.cacheTestResourceFieldValue("vcd_nsxt_nat_rule.snat", "id"),
					dnatRuleId.cacheTestResourceFieldValue("vcd_nsxt_nat_rule.snat", "id"),

					ipSecVpnTunnelId.cacheTestResourceFieldValue("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "id"),

					routedNetId.cacheTestResourceFieldValue("vcd_network_routed_v2.nsxt-backed", "id"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.nsxt-backed", "vdc", fmt.Sprintf("%s-%s", t.Name(), "0")),
					resource.TestCheckResourceAttrPair("vcd_network_routed_v2.nsxt-backed", "owner_id", "vcd_org_vdc.newVdc.0", "id"),

					isolatedNetId.cacheTestResourceFieldValue("vcd_network_isolated_v2.nsxt-backed", "id"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.nsxt-backed", "vdc", fmt.Sprintf("%s-%s", t.Name(), "0")),
					resource.TestCheckResourceAttrPair("vcd_network_isolated_v2.nsxt-backed", "owner_id", "vcd_org_vdc.newVdc.0", "id"),

					importedNetId.cacheTestResourceFieldValue("vcd_nsxt_network_imported.nsxt-backed", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.nsxt-backed", "vdc", fmt.Sprintf("%s-%s", t.Name(), "0")),
					resource.TestCheckResourceAttrPair("vcd_nsxt_network_imported.nsxt-backed", "owner_id", "vcd_org_vdc.newVdc.0", "id"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					edgeGatewayId.testCheckCachedResourceFieldValue("vcd_nsxt_edgegateway.nsxt-edge", "id"),
					resource.TestCheckResourceAttrPair("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", "vcd_vdc_group.test1", "id"),
					resource.TestCheckResourceAttrPair("vcd_nsxt_edgegateway.nsxt-edge", "vdc", "vcd_vdc_group.test1", "name"),

					firewallId.testCheckCachedResourceFieldValue("vcd_nsxt_firewall.testing", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "vdc", fmt.Sprintf("%s-%s", t.Name(), "0")),

					snatRuleId.testCheckCachedResourceFieldValue("vcd_nsxt_nat_rule.snat", "id"),
					dnatRuleId.testCheckCachedResourceFieldValue("vcd_nsxt_nat_rule.snat", "id"),

					ipSecVpnTunnelId.testCheckCachedResourceFieldValue("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "id"),

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

					firewallId.testCheckCachedResourceFieldValue("vcd_nsxt_firewall.testing", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "vdc", fmt.Sprintf("%s-%s", t.Name(), "0")),

					snatRuleId.testCheckCachedResourceFieldValue("vcd_nsxt_nat_rule.snat", "id"),
					dnatRuleId.testCheckCachedResourceFieldValue("vcd_nsxt_nat_rule.snat", "id"),

					ipSecVpnTunnelId.testCheckCachedResourceFieldValue("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "id"),

					routedNetId.testCheckCachedResourceFieldValue("vcd_network_routed_v2.nsxt-backed", "id"),
					resource.TestCheckResourceAttrPair("vcd_network_routed_v2.nsxt-backed", "owner_id", "vcd_vdc_group.test1", "id"),
					resource.TestCheckResourceAttrPair("vcd_network_routed_v2.nsxt-backed", "vdc", "vcd_vdc_group.test1", "name"),

					isolatedNetId.testCheckCachedResourceFieldValue("vcd_network_isolated_v2.nsxt-backed", "id"),
					resource.TestCheckResourceAttrPair("vcd_network_isolated_v2.nsxt-backed", "owner_id", "vcd_vdc_group.test1", "id"),
					resource.TestCheckResourceAttrPair("vcd_network_isolated_v2.nsxt-backed", "vdc", "vcd_vdc_group.test1", "name"),

					importedNetId.testCheckCachedResourceFieldValue("vcd_nsxt_network_imported.nsxt-backed", "id"),
					resource.TestCheckResourceAttrPair("vcd_nsxt_network_imported.nsxt-backed", "owner_id", "vcd_vdc_group.test1", "id"),
					resource.TestCheckResourceAttrPair("vcd_nsxt_network_imported.nsxt-backed", "vdc", "vcd_vdc_group.test1", "name"),
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("data.vcd_nsxt_firewall.testing", "vcd_nsxt_firewall.testing", nil),
					resourceFieldsEqual("data.vcd_nsxt_nat_rule.snat", "vcd_nsxt_nat_rule.snat", nil),
					resourceFieldsEqual("data.vcd_nsxt_nat_rule.no-snat", "vcd_nsxt_nat_rule.no-snat", nil),
					resourceFieldsEqual("data.vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "vcd_nsxt_ipsec_vpn_tunnel.tunnel1", nil),
				),
			},
			{
				ResourceName:            "vcd_nsxt_firewall.testing",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdOrgNsxtVdcGroupObject(testConfig, parentVdcGroupName, t.Name()),
				ImportStateVerifyIgnore: []string{"vdc"},
			},
			{
				ResourceName:            "vcd_nsxt_nat_rule.snat",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdNsxtEdgeGatewayObjectUsingVdcGroup(parentVdcGroupName, t.Name(), "SNAT rule"),
				ImportStateVerifyIgnore: []string{"vdc"},
			},
			{
				ResourceName:            "vcd_nsxt_nat_rule.no-snat",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdNsxtEdgeGatewayObjectUsingVdcGroup(parentVdcGroupName, t.Name(), "test-no-snat-rule"),
				ImportStateVerifyIgnore: []string{"vdc"},
			},
			{
				ResourceName:            "vcd_nsxt_ipsec_vpn_tunnel.tunnel1",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdNsxtEdgeGatewayObjectUsingVdcGroup(parentVdcGroupName, t.Name(), "First"),
				ImportStateVerifyIgnore: []string{"vdc"},
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

resource "vcd_nsxt_firewall" "testing" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.newVdc.0.name

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  rule {
    action      = "ALLOW"
    name        = "allow all IPv4 traffic"
    direction   = "IN_OUT"
    ip_protocol = "IPV4"
  }
}

resource "vcd_nsxt_nat_rule" "snat" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.newVdc.0.name

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  name        = "SNAT rule"
  rule_type   = "SNAT"
  description = "description"

  # Using primary_ip from edge gateway
  external_address         = tolist(vcd_nsxt_edgegateway.nsxt-edge.subnet)[0].primary_ip
  internal_address         = "11.11.11.0/24"
  snat_destination_address = "8.8.8.8"
  logging                  = true
}

resource "vcd_nsxt_nat_rule" "no-snat" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.newVdc.0.name

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  name        = "test-no-snat-rule"
  rule_type   = "NO_SNAT"
  description = "description"

  internal_address = "11.11.11.0/24"
}

resource "vcd_nsxt_ipsec_vpn_tunnel" "tunnel1" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.newVdc.0.name

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  name        = "First"
  description = "testing tunnel"

  pre_shared_key = "my-presharaed-key"
  # Primary IP address of Edge Gateway pulled from data source
  local_ip_address = tolist(vcd_nsxt_edgegateway.nsxt-edge.subnet)[0].primary_ip
  local_networks   = ["10.10.10.0/24", "30.30.30.0/28", "40.40.40.1/32"]
  # That is a fake remote IP address
  remote_ip_address = "1.2.3.4"
  remote_networks   = ["192.168.1.0/24", "192.168.10.0/24", "192.168.20.0/28"]
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

resource "vcd_nsxt_firewall" "testing" {
  org = "{{.Org}}"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  rule {
    action      = "ALLOW"
    name        = "allow all IPv4 traffic"
    direction   = "IN_OUT"
    ip_protocol = "IPV4"
  }
}

resource "vcd_nsxt_nat_rule" "snat" {
  org = "{{.Org}}"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  name        = "SNAT rule"
  rule_type   = "SNAT"
  description = "description"

  # Using primary_ip from edge gateway
  external_address         = tolist(vcd_nsxt_edgegateway.nsxt-edge.subnet)[0].primary_ip
  internal_address         = "11.11.11.0/24"
  snat_destination_address = "8.8.8.8"
  logging                  = true
}

resource "vcd_nsxt_nat_rule" "no-snat" {
  org = "{{.Org}}"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  name        = "test-no-snat-rule"
  rule_type   = "NO_SNAT"
  description = "description"

  internal_address = "11.11.11.0/24"
}

resource "vcd_nsxt_ipsec_vpn_tunnel" "tunnel1" {
  org = "{{.Org}}"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  name        = "First"
  description = "testing tunnel"

  pre_shared_key = "my-presharaed-key"
  # Primary IP address of Edge Gateway pulled from data source
  local_ip_address = tolist(vcd_nsxt_edgegateway.nsxt-edge.subnet)[0].primary_ip
  local_networks   = ["10.10.10.0/24", "30.30.30.0/28", "40.40.40.1/32"]
  # That is a fake remote IP address
  remote_ip_address = "1.2.3.4"
  remote_networks   = ["192.168.1.0/24", "192.168.10.0/24", "192.168.20.0/28"]
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

const testAccVcdNsxVdcGroupCompleteMigrationStep4DS = testAccVcdNsxVdcGroupCompleteMigrationStep3 + `
# skip-binary-test: Data Source test
data "vcd_nsxt_firewall" "testing" {
  org = "{{.Org}}"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id
}

data "vcd_nsxt_nat_rule" "snat" {
  org = "{{.Org}}"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id
  name            = "SNAT rule"
}

data "vcd_nsxt_nat_rule" "no-snat" {
  org = "{{.Org}}"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id
  name            = "test-no-snat-rule"
}

data "vcd_nsxt_ipsec_vpn_tunnel" "tunnel1" {
  org = "{{.Org}}"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id
  name            = "First"
}

`
