//go:build network || nsxt || ALL || functional
// +build network nsxt ALL functional

package vcd

import (
	"fmt"
	"regexp"
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

func TestAccVcdNsxtAlbVdcGroupIntegration(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	skipNoNsxtAlbConfiguration(t)

	if testConfig.Certificates.Certificate1Path == "" || testConfig.Certificates.Certificate2Path == "" ||
		testConfig.Certificates.Certificate1PrivateKeyPath == "" || testConfig.Certificates.Certificate1Pass == "" {
		t.Skip("Variables Certificates.Certificate1Path, Certificates.Certificate2Path, " +
			"Certificates.Certificate1PrivateKeyPath, Certificates.Certificate1Pass must be set")
	}

	// String map to fill the template
	var params = StringMap{
		"VirtualServiceName":        t.Name(),
		"ControllerName":            t.Name(),
		"ControllerUrl":             testConfig.Nsxt.NsxtAlbControllerUrl,
		"ControllerUsername":        testConfig.Nsxt.NsxtAlbControllerUser,
		"ControllerPassword":        testConfig.Nsxt.NsxtAlbControllerPassword,
		"ImportableCloud":           testConfig.Nsxt.NsxtAlbImportableCloud,
		"ReservationModel":          "DEDICATED",
		"Org":                       testConfig.VCD.Org,
		"NsxtVdc":                   testConfig.Nsxt.Vdc,
		"EdgeGw":                    testConfig.Nsxt.EdgeGateway,
		"IsActive":                  "true",
		"AliasPrivate":              t.Name() + "-cert",
		"Certificate1Path":          testConfig.Certificates.Certificate1Path,
		"CertPrivateKey1":           testConfig.Certificates.Certificate1PrivateKeyPath,
		"CertPassPhrase1":           testConfig.Certificates.Certificate1Pass,
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

		"Tags": "nsxt alb vdcGroup",
	}

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccVcdNsxtAlbVdcGroupIntegration1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdNsxtAlbVdcGroupIntegration2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	// params["FuncName"] = t.Name() + "step3"
	// configText3 := templateFill(testAccVcdNsxtAlbVdcGroupIntegration3, params)
	// debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVcdAlbControllerDestroy("vcd_nsxt_alb_controller.first"),
			testAccCheckVcdAlbServiceEngineGroupDestroy("vcd_nsxt_alb_cloud.first"),
			testAccCheckVcdAlbCloudDestroy("vcd_nsxt_alb_cloud.first"),
			testAccCheckVcdNsxtEdgeGatewayAlbSettingsDestroy(params["EdgeGw"].(string)),
			testAccCheckVcdAlbVirtualServiceDestroy("vcd_nsxt_alb_virtual_service.test"),
		),

		Steps: []resource.TestStep{
			{
				Config: configText1, // Setup prerequisites - configure NSX-T ALB in Provider
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "id", regexp.MustCompile(`^urn:vcloud:loadBalancerVirtualService:`)),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "id", regexp.MustCompile(`^urn:vcloud:loadBalancerVirtualService:`)),
				),
			},
		},
	})
	postTestChecks(t)
}

// Config merges required prerequisites for ALB and VDC Group creation
const testAccVcdNsxtAlbVdcGroupIntegration1 = testAccVcdVdcGroupNew + `
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

resource "vcd_nsxt_alb_settings" "test" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.newVdc.0.name

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id
  is_active       = true

  # This dependency is required to make sure that provider part of operations is done
  depends_on = [vcd_nsxt_alb_service_engine_group.first]
}


locals {
  controller_id = vcd_nsxt_alb_controller.first.id
}

data "vcd_nsxt_alb_importable_cloud" "cld" {
  name          = "{{.ImportableCloud}}"
  controller_id = local.controller_id
}

resource "vcd_nsxt_alb_controller" "first" {
  name         = "{{.ControllerName}}"
  url          = "{{.ControllerUrl}}"
  username     = "{{.ControllerUsername}}"
  password     = "{{.ControllerPassword}}"
  license_type = "ENTERPRISE"
}

resource "vcd_nsxt_alb_cloud" "first" {
  name        = "nsxt-cloud"
  description = "first alb cloud"

  controller_id       = vcd_nsxt_alb_controller.first.id
  importable_cloud_id = data.vcd_nsxt_alb_importable_cloud.cld.id
  network_pool_id     = data.vcd_nsxt_alb_importable_cloud.cld.network_pool_id
}

resource "vcd_nsxt_alb_service_engine_group" "first" {
  name                                 = "{{.Name}}-service-engine-group"
  alb_cloud_id                         = vcd_nsxt_alb_cloud.first.id
  importable_service_engine_group_name = "Default-Group"
  reservation_model                    = "DEDICATED"
}

resource "vcd_nsxt_alb_edgegateway_service_engine_group" "assignment" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.newVdc.0.name

  edge_gateway_id         = vcd_nsxt_alb_settings.test.edge_gateway_id
  service_engine_group_id = vcd_nsxt_alb_service_engine_group.first.id
}

resource "vcd_nsxt_alb_pool" "test" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.newVdc.0.name

  name            = "{{.Name}}-pool"
  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id
}

resource "vcd_nsxt_alb_virtual_service" "test" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.newVdc.0.name

  name            = "{{.Name}}-vs"
  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id

  pool_id                  = vcd_nsxt_alb_pool.test.id
  service_engine_group_id  = vcd_nsxt_alb_edgegateway_service_engine_group.assignment.service_engine_group_id
  virtual_ip_address       = tolist(vcd_nsxt_edgegateway.nsxt-edge.subnet)[0].primary_ip
  application_profile_type = "HTTP"
  service_port {
    start_port = 80
    end_port   = 81
    type       = "TCP_PROXY"
  }
}
`

const testAccVcdNsxtAlbVdcGroupIntegration2 = testAccVcdVdcGroupNew + `
data "vcd_external_network_v2" "nsxt-ext-net" {
  name = "{{.NsxtExternalNetworkName}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org      = "{{.Org}}"
  owner_id = vcd_vdc_group.test1.id

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

resource "vcd_nsxt_alb_settings" "test" {
  org = "{{.Org}}"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id
  is_active       = true

  # This dependency is required to make sure that provider part of operations is done
  depends_on = [vcd_nsxt_alb_service_engine_group.first]
}


locals {
  controller_id = vcd_nsxt_alb_controller.first.id
}

data "vcd_nsxt_alb_importable_cloud" "cld" {
  name          = "{{.ImportableCloud}}"
  controller_id = local.controller_id
}

resource "vcd_nsxt_alb_controller" "first" {
  name         = "{{.ControllerName}}"
  url          = "{{.ControllerUrl}}"
  username     = "{{.ControllerUsername}}"
  password     = "{{.ControllerPassword}}"
  license_type = "ENTERPRISE"
}

resource "vcd_nsxt_alb_cloud" "first" {
  name        = "nsxt-cloud"
  description = "first alb cloud"

  controller_id       = vcd_nsxt_alb_controller.first.id
  importable_cloud_id = data.vcd_nsxt_alb_importable_cloud.cld.id
  network_pool_id     = data.vcd_nsxt_alb_importable_cloud.cld.network_pool_id
}

resource "vcd_nsxt_alb_service_engine_group" "first" {
  name                                 = "{{.Name}}-service-engine-group"
  alb_cloud_id                         = vcd_nsxt_alb_cloud.first.id
  importable_service_engine_group_name = "Default-Group"
  reservation_model                    = "DEDICATED"
}

resource "vcd_nsxt_alb_edgegateway_service_engine_group" "assignment" {
  org = "{{.Org}}"

  edge_gateway_id         = vcd_nsxt_alb_settings.test.edge_gateway_id
  service_engine_group_id = vcd_nsxt_alb_service_engine_group.first.id
}

resource "vcd_nsxt_alb_pool" "test" {
  org = "{{.Org}}"

  name            = "{{.Name}}-pool"
  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id
}

resource "vcd_nsxt_alb_virtual_service" "test" {
  org = "{{.Org}}"

  name            = "{{.Name}}-vs"
  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id

  pool_id                  = vcd_nsxt_alb_pool.test.id
  service_engine_group_id  = vcd_nsxt_alb_edgegateway_service_engine_group.assignment.service_engine_group_id
  virtual_ip_address       = tolist(vcd_nsxt_edgegateway.nsxt-edge.subnet)[0].primary_ip
  application_profile_type = "HTTP"
  service_port {
    start_port = 80
    end_port   = 81
    type       = "TCP_PROXY"
  }
}
`

const testAccVcdNsxtAlbVdcGroupIntegration3 = testAccVcdNsxtAlbVirtualServiceStep1 + testAccVcdVdcGroupNew
