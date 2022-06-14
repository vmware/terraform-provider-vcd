//go:build gateway || nsxt || ALL || functional || vdcGroup
// +build gateway nsxt ALL functional vdcGroup

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdDistributedFirewall goal is to try out diverse configuration of rules. This test should
// support running on all supported versions of VCD. There are a few additional tests which
// explicitly check new features introduced in newer VCD versions having versions in their names.
func TestAccVcdDistributedFirewall(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	// String map to fill the template
	var params = StringMap{
		"Org":                       testConfig.VCD.Org,
		"Name":                      t.Name(),
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"Dfw":                       "true",
		"DefaultPolicy":             "true",
		"TestName":                  t.Name(),

		"NsxtManager":     testConfig.Nsxt.Manager,
		"ExternalNetwork": testConfig.Nsxt.ExternalNetwork,

		"Tags": "vdcGroup gateway nsxt",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "-newVdc"
	configTextPre := templateFill(testAccVcdVdcGroupNew, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configTextPre)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(dfwStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(dfwStep3DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText4)

	params["FuncName"] = t.Name() + "-step5"
	configText5 := templateFill(dfwStep4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText5)

	params["FuncName"] = t.Name() + "-step6"
	configText6 := templateFill(dfwStep6DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 6: %s", configText6)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				// Setup prerequisites
				Config: configTextPre,
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_distributed_firewall.t1", "id"),

					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.#", "5"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.name", "rule1"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.direction", "IN_OUT"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.action", "ALLOW"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.ip_protocol", "IPV4_IPV6"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.logging", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.source_ids.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.destination_ids.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.app_port_profile_ids.#", "3"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.network_context_profile_ids.#", "0"),

					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.name", "rule2"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.direction", "IN_OUT"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.ip_protocol", "IPV4_IPV6"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.action", "DROP"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.source_ids.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.destination_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.app_port_profile_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.network_context_profile_ids.#", "3"),

					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.2.name", "rule3"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.2.direction", "IN_OUT"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.2.ip_protocol", "IPV4"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.2.logging", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.2.action", "DROP"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.2.enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.2.source_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.2.destination_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.2.app_port_profile_ids.#", "0"),

					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.3.name", "rule4"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.3.direction", "OUT"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.3.ip_protocol", "IPV6"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.3.logging", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.3.action", "ALLOW"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.3.enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.3.source_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.3.destination_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.3.app_port_profile_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.3.network_context_profile_ids.#", "0"),

					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.4.name", "rule5"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.4.direction", "IN"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.4.ip_protocol", "IPV6"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.4.logging", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.4.action", "ALLOW"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.4.enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.4.source_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.4.destination_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.4.app_port_profile_ids.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.4.network_context_profile_ids.#", "0"),
				),
			},
			{
				ResourceName:      "vcd_nsxt_distributed_firewall.t1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgObject(testConfig, t.Name()),
			},
			{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("vcd_nsxt_distributed_firewall.t1", "data.vcd_nsxt_distributed_firewall.t1", nil),
				),
			},
			{
				Config: configText5,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_distributed_firewall.t1", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.name", "rule1"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.action", "DROP"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.direction", "IN_OUT"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.ip_protocol", "IPV4_IPV6"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.logging", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.source_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.destination_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.app_port_profile_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.network_context_profile_ids.#", "3"),
					resource.TestMatchResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.network_context_profile_ids.0", regexp.MustCompile(`^urn:vcloud:networkContextProfile:`)),
					resource.TestMatchResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.network_context_profile_ids.1", regexp.MustCompile(`^urn:vcloud:networkContextProfile:`)),
					resource.TestMatchResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.network_context_profile_ids.2", regexp.MustCompile(`^urn:vcloud:networkContextProfile:`)),
				),
			},
			{
				Config: configText6,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("vcd_nsxt_distributed_firewall.t1", "data.vcd_nsxt_distributed_firewall.t1", nil),
				),
			},
		},
	})
	postTestChecks(t)
}

const constDfwPrereqs = testAccVcdVdcGroupNew + `
data "vcd_nsxt_network_context_profile" "cp1" {
  context_id = vcd_vdc_group.test1.id
  name       = "360ANTIV"
}

data "vcd_nsxt_network_context_profile" "cp2" {
  context_id = vcd_vdc_group.test1.id
  name         = "AMQP"
  scope        = "SYSTEM"
}

data "vcd_nsxt_network_context_profile" "cp3" {
  context_id = vcd_vdc_group.test1.id
  name         = "AVAST"
}

data "vcd_external_network_v2" "existing-extnet" {
  name = "{{.ExternalNetwork}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org                 = "{{.Org}}"
  owner_id            = vcd_vdc_group.test1.id
  name                = "{{.Name}}-edge"
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

resource "vcd_nsxt_ip_set" "set1" {
  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  name        = "{{.Name}}-ipset1"
  description = "IP Set containing IPv4 and IPv6 ranges"

  ip_addresses = [
    "12.12.12.1",
    "10.10.10.0/24",
    "11.11.11.1-11.11.11.2",
    "2001:db8::/48",
    "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff",
  ]
}

resource "vcd_nsxt_ip_set" "set2" {
  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  name        = "{{.Name}}-ipset2"
  description = "Empty IP Set"
}

resource "vcd_nsxt_security_group" "g1-empty" {
  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  name = "{{.Name}} empty group"
}

resource "vcd_nsxt_security_group" "g2" {
  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  name = "{{.Name}} group with members"
  member_org_network_ids = [vcd_network_routed_v2.nsxt-backed.id]
}

resource "vcd_network_routed_v2" "nsxt-backed" {
  name            = "{{.Name}}-routed-net"
  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  gateway       = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "1.1.1.10"
    end_address   = "1.1.1.20"
  }
}

resource "vcd_nsxt_app_port_profile" "p1" {
  org = "{{.Org}}"
  context_id = vcd_vdc_group.test1.id
  name = "{{.Name}}-app-profile"

  scope = "TENANT"

  app_port {
    protocol = "ICMPv6"
  }

  app_port {
    protocol = "TCP"
    port     = ["2000", "2010-2020"]
  }

  app_port {
    protocol = "UDP"
    port     = ["40000-60000"]
  }
}

data "vcd_nsxt_manager" "main" {
  name = "{{.NsxtManager}}"
}

data "vcd_nsxt_app_port_profile" "WINS" {
  context_id = data.vcd_nsxt_manager.main.id
  name       = "WINS"
  scope      = "SYSTEM"
}

data "vcd_nsxt_app_port_profile" "FTP" {
  context_id = data.vcd_nsxt_manager.main.id
  name       = "FTP"
  scope      = "SYSTEM"
}
`

const dfwStep2 = constDfwPrereqs + `
resource "vcd_nsxt_distributed_firewall" "t1" {
  org          = "{{.Org}}"
  vdc_group_id = vcd_vdc_group.test1.id

  rule {
	name        = "rule1"
	action      = "ALLOW"
	description = "description"

	source_ids           = [vcd_nsxt_ip_set.set1.id, vcd_nsxt_ip_set.set2.id]
	destination_ids      = [vcd_nsxt_security_group.g1-empty.id, vcd_nsxt_security_group.g2.id]
	app_port_profile_ids = [vcd_nsxt_app_port_profile.p1.id, data.vcd_nsxt_app_port_profile.WINS.id, data.vcd_nsxt_app_port_profile.FTP.id]
  }

  rule {
	name      = "rule2"
	action    = "DROP"
	enabled   = false
	logging   = true
	direction = "IN_OUT"

	source_ids                  = [vcd_nsxt_ip_set.set1.id, vcd_nsxt_ip_set.set2.id]
	network_context_profile_ids = [
		data.vcd_nsxt_network_context_profile.cp1.id,
		data.vcd_nsxt_network_context_profile.cp2.id,
		data.vcd_nsxt_network_context_profile.cp3.id
	]
  }

  rule {
	name        = "rule3"
	action      = "DROP"
	ip_protocol = "IPV4"
  }

  rule {
	name        = "rule4"
	action      = "ALLOW"
	ip_protocol = "IPV6"
	direction   = "OUT"
  }

  rule {
	name        = "rule5"
	action      = "ALLOW"
	ip_protocol = "IPV6"
	direction   = "IN"

	app_port_profile_ids = [vcd_nsxt_app_port_profile.p1.id]
  }
}
`

const dfwStep3DS = dfwStep2 + `
# skip-binary-test: Data Source test
data "vcd_nsxt_distributed_firewall" "t1" {
  org          = "{{.Org}}"
  vdc_group_id = vcd_vdc_group.test1.id
}
`

const dfwStep4 = constDfwPrereqs + `
resource "vcd_nsxt_distributed_firewall" "t1" {
  org          = "{{.Org}}"
  vdc_group_id = vcd_vdc_group.test1.id

  rule {
	name   = "rule1"
	action = "DROP"
	network_context_profile_ids = [
		data.vcd_nsxt_network_context_profile.cp1.id,
		data.vcd_nsxt_network_context_profile.cp2.id,
		data.vcd_nsxt_network_context_profile.cp3.id
		]
  }
}
`

const dfwStep6DS = dfwStep4 + `
# skip-binary-test: Data Source test
data "vcd_nsxt_distributed_firewall" "t1" {
  org          = "{{.Org}}"
  vdc_group_id = vcd_vdc_group.test1.id
}
`

// TestAccVcdDistributedFirewallVCD10_2_2 complements TestAccVcdDistributedFirewall and tests out
// new 10.2.2+ feature:
// * Firewall rule 'action' REJECT
func TestAccVcdDistributedFirewallVCD10_2_2(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	vcdClient := createTemporaryVCDConnection(false)
	if vcdClient.Client.APIVCDMaxVersionIs("< 35.2") {
		t.Skipf("This test tests VCD 10.2.2+ (API V35.2+) features. Skipping.")
	}

	// String map to fill the template
	var params = StringMap{
		"Org":                       testConfig.VCD.Org,
		"Name":                      t.Name(),
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"Dfw":                       "true",
		"DefaultPolicy":             "true",
		"TestName":                  t.Name(),

		"NsxtEdgeGatewayVcd": t.Name() + "-edge",
		"ExternalNetwork":    testConfig.Nsxt.ExternalNetwork,

		"Tags": "vdcGroup gateway nsxt",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "-newVdc"
	configTextPre := templateFill(testAccVcdVdcGroupNew, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configTextPre)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(dfwStep2VCD10_2_2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				// Setup prerequisites
				Config: configTextPre,
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_distributed_firewall.t1", "id"),

					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.name", "rule1"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.action", "REJECT"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.description", "description"),
				),
			},
		},
	})
	postTestChecks(t)
}

const dfwStep2VCD10_2_2 = testAccVcdVdcGroupNew + `
resource "vcd_nsxt_distributed_firewall" "t1" {
  org          = "{{.Org}}"
  vdc_group_id = vcd_vdc_group.test1.id

  rule {
	name        = "rule1"
	action      = "REJECT"
	description = "description"
  }
}
`

// TestAccVcdDistributedFirewallVCD10_3_2 complements TestAccVcdDistributedFirewall and aims to test
// our 3 new fields of VCD 10.3.2+ in distributed firewall:
// * comment (this one is shown in UI, not like `description`)
// * source_groups_excluded (negates the values specified in source_ids)
// * destination_groups_excluded (negates the values specified in destinations_ids)
func TestAccVcdDistributedFirewallVCD10_3_2(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	vcdClient := createTemporaryVCDConnection(false)
	if vcdClient.Client.APIVCDMaxVersionIs("< 36.2") {
		t.Skipf("This test tests VCD 10.3.2+ (API V36.2+) features. Skipping.")
	}

	// String map to fill the template
	var params = StringMap{
		"Org":                       testConfig.VCD.Org,
		"Name":                      t.Name(),
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"Dfw":                       "true",
		"DefaultPolicy":             "true",
		"TestName":                  t.Name(),
		"NsxtManager":               testConfig.Nsxt.Manager,

		"NsxtEdgeGatewayVcd": t.Name() + "-edge",
		"ExternalNetwork":    testConfig.Nsxt.ExternalNetwork,

		"Tags": "vdcGroup gateway nsxt",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "-newVdc"
	configTextPre := templateFill(testAccVcdVdcGroupNew, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configTextPre)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(dfwStep2VCD10_3_2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				// Setup prerequisites
				Config: configTextPre,
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_distributed_firewall.t1", "id"),

					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.#", "3"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.name", "rule1"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.comment", "longer text comment field filled"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.direction", "IN_OUT"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.action", "REJECT"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.ip_protocol", "IPV4_IPV6"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.logging", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.source_ids.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.destination_ids.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.app_port_profile_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.network_context_profile_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.source_groups_excluded", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.destination_groups_excluded", "true"),

					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.name", "rule2"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.direction", "IN_OUT"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.ip_protocol", "IPV4_IPV6"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.action", "DROP"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.source_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.destination_ids.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.app_port_profile_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.network_context_profile_ids.#", "1"),
					resource.TestMatchResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.network_context_profile_ids.0", regexp.MustCompile(`^urn:vcloud:networkContextProfile:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.source_groups_excluded", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.destination_groups_excluded", "true"),

					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.2.name", "rule3"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.2.direction", "IN_OUT"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.2.ip_protocol", "IPV4"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.2.logging", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.2.action", "DROP"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.2.enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.2.source_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.2.destination_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.2.app_port_profile_ids.#", "0"),
				),
			},
		},
	})
	postTestChecks(t)
}

const dfwStep2VCD10_3_2 = testAccVcdVdcGroupNew + `
data "vcd_nsxt_network_context_profile" "cp1" {
  context_id = vcd_vdc_group.test1.id
  name       = "360ANTIV"
}

data "vcd_external_network_v2" "existing-extnet" {
  name = "{{.ExternalNetwork}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org                 = "{{.Org}}"
  owner_id            = vcd_vdc_group.test1.id
  name                = "{{.Name}}-edge"
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

resource "vcd_nsxt_security_group" "g2" {
  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  name = "{{.Name}} group with members"
  member_org_network_ids = [vcd_network_routed_v2.nsxt-backed.id]
}

resource "vcd_network_routed_v2" "nsxt-backed" {
  name            = "{{.Name}}-routed-net"
  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  gateway       = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "1.1.1.10"
    end_address   = "1.1.1.20"
  }
}

resource "vcd_nsxt_ip_set" "set1" {
  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  name        = "{{.Name}}-ipset1"
  description = "IP Set containing IPv4 and IPv6 ranges"

  ip_addresses = [
    "12.12.12.1",
    "10.10.10.0/24",
    "11.11.11.1-11.11.11.2",
    "2001:db8::/48",
    "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff",
  ]
}

resource "vcd_nsxt_distributed_firewall" "t1" {
  org          = "{{.Org}}"
  vdc_group_id = vcd_vdc_group.test1.id

  rule {
	name        = "rule1"
	action      = "REJECT"
	description = "description"
	comment     = "longer text comment field filled"

	source_ids             = [vcd_nsxt_security_group.g2.id]
	source_groups_excluded = true

	destination_ids             = [vcd_nsxt_ip_set.set1.id]
	destination_groups_excluded = true
  }

  rule {
	name      = "rule2"
	action    = "DROP"
	enabled   = false
	logging   = true
	destination_ids             = [vcd_nsxt_security_group.g2.id,vcd_nsxt_ip_set.set1.id]
    destination_groups_excluded = true

	network_context_profile_ids = [data.vcd_nsxt_network_context_profile.cp1.id]
  }

  rule {
	name        = "rule3"
	action      = "DROP"
	ip_protocol = "IPV4"
  }
}
`
