//go:build gateway || nsxt || ALL || functional || vdcGroup

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdDistributedFirewallRule(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	// String map to fill the template
	var params = StringMap{
		"Org":                       testConfig.VCD.Org,
		"Name":                      t.Name(),
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"ProviderVdcStorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"Dfw":                       "true",
		"DefaultPolicy":             "true",
		"RemoveDefaultFirewallRule": "true", // will remove default firewall rule in VDC Group
		"TestName":                  t.Name(),

		"NsxtManager":     testConfig.Nsxt.Manager,
		"ExternalNetwork": testConfig.Nsxt.ExternalNetwork,

		"Tags": "vdcGroup gateway nsxt",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "-newVdc"
	configTextPre := templateFill(testAccVcdVdcGroupNew, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configTextPre)

	params["FuncName"] = t.Name() + "-newVdcGroupValidation"
	configTextPre2 := templateFill(dfwRuleStep11, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configTextPre2)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(dfwRuleStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(dfwRuleStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "-step4"
	configText4DS := templateFill(dfwRuleStep4DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4DS)

	params["FuncName"] = t.Name() + "-step5"
	configText5 := templateFill(dfwRuleStep5, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s", configText5)

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
				// This step validates that `remove_default_firewall_rule` in resource
				// `vcd_vdc_group` worked. The expectation is to get 0 rules in `vcd_nsxt_firewall`
				// data source as otherwise it would be one (default rule gets added when enabling
				// DFW)
				Config: configTextPre2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.vcd_nsxt_firewall.rule-count", "rule.#", "0"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_distributed_firewall_rule.r1", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_distributed_firewall_rule.r2", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_distributed_firewall_rule.r3", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_distributed_firewall_rule.r4", "id"),

					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r1", "name", "rule1"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r1", "action", "ALLOW"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r1", "description", "description"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r1", "source_ids.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r1", "destination_ids.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r1", "app_port_profile_ids.#", "3"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r1", "network_context_profile_ids.#", "0"),

					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r2", "name", "rule2"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r2", "action", "DROP"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r2", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r2", "ip_protocol", "IPV4"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r2", "source_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r2", "destination_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r2", "app_port_profile_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r2", "network_context_profile_ids.#", "0"),

					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r3", "name", "rule3"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r3", "action", "DROP"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r3", "enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r3", "logging", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r3", "direction", "IN_OUT"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r3", "source_ids.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r3", "destination_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r3", "app_port_profile_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r3", "network_context_profile_ids.#", "3"),

					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r4", "name", "rule4"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r4", "action", "ALLOW"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r4", "ip_protocol", "IPV6"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r4", "direction", "OUT"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r4", "source_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r4", "destination_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r4", "app_port_profile_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r4", "network_context_profile_ids.#", "0"),
				),
			},
			{
				// Using 'vcd_nsxt_distributed_firewall' data source to get all rules, verify their
				// count and their order to validate that standalone
				// 'vcd_nsxt_distributed_firewall_rule' resources and their field 'above_rule_id'
				// works
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.vcd_nsxt_distributed_firewall.all-rules", "rule.#", "4"),
					resource.TestCheckResourceAttrPair("data.vcd_nsxt_distributed_firewall.all-rules", "rule.0.name", "vcd_nsxt_distributed_firewall_rule.r3", "name"),
					resource.TestCheckResourceAttrPair("data.vcd_nsxt_distributed_firewall.all-rules", "rule.1.name", "vcd_nsxt_distributed_firewall_rule.r2", "name"),
					resource.TestCheckResourceAttrPair("data.vcd_nsxt_distributed_firewall.all-rules", "rule.2.name", "vcd_nsxt_distributed_firewall_rule.r1", "name"),
					resource.TestCheckResourceAttrPair("data.vcd_nsxt_distributed_firewall.all-rules", "rule.3.name", "vcd_nsxt_distributed_firewall_rule.r4", "name"),
				),
			},
			{
				ResourceName:      "vcd_nsxt_distributed_firewall_rule.r1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcGroupObject(t.Name(), "rule1"),
			},
			{
				Config: configText4DS,
				Check: resource.ComposeAggregateTestCheckFunc(
					// field count differs in resource and data source because resource has `above_rule_id` field
					resourceFieldsEqual("data.vcd_nsxt_distributed_firewall_rule.r1", "vcd_nsxt_distributed_firewall_rule.r1", []string{"%"}),
					resourceFieldsEqual("data.vcd_nsxt_distributed_firewall_rule.r2", "vcd_nsxt_distributed_firewall_rule.r2", []string{"%"}),
					resourceFieldsEqual("data.vcd_nsxt_distributed_firewall_rule.r3", "vcd_nsxt_distributed_firewall_rule.r3", []string{"%"}),
				),
			},
			{
				Config: configText5,
				Check: resource.ComposeAggregateTestCheckFunc(
					// field count differs in resource and data source because resource has `above_rule_id` field
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r1", "name", "rule1-updated"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r1", "action", "ALLOW"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r1", "description", "description-updated"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r1", "source_ids.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r1", "destination_ids.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r1", "app_port_profile_ids.#", "3"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r1", "network_context_profile_ids.#", "0"),

					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r2", "name", "rule2-updated"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r2", "action", "ALLOW"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r2", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r2", "ip_protocol", "IPV4"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r2", "source_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r2", "destination_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r2", "app_port_profile_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r2", "network_context_profile_ids.#", "0"),
				),
			},
		},
	})
	postTestChecks(t)
}

const dfwRuleStep11 = constDfwPrereqs + `
# skip-binary-test: Only for rule count validation in Acceptance test
data "vcd_nsxt_firewall" "rule-count" {
  org = "{{.Org}}"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id
}
`

const dfwRuleStep2 = constDfwPrereqs + `
resource "vcd_nsxt_distributed_firewall_rule" "r1" {
  org          = "{{.Org}}"
  vdc_group_id = vcd_vdc_group.test1.id

  name        = "rule1"
  action      = "ALLOW"
  description = "description"

  source_ids           = [vcd_nsxt_ip_set.set1.id, vcd_nsxt_ip_set.set2.id]
  destination_ids      = [vcd_nsxt_security_group.g1-empty.id, vcd_nsxt_security_group.g2.id]
  app_port_profile_ids = [vcd_nsxt_app_port_profile.p1.id, data.vcd_nsxt_app_port_profile.WINS.id, data.vcd_nsxt_app_port_profile.FTP.id]
}

resource "vcd_nsxt_distributed_firewall_rule" "r2" {
  org          = "{{.Org}}"
  vdc_group_id = vcd_vdc_group.test1.id

  above_rule_id = vcd_nsxt_distributed_firewall_rule.r1.id # Order management element
  name        = "rule2"
  action      = "DROP"
  ip_protocol = "IPV4"
}

resource "vcd_nsxt_distributed_firewall_rule" "r3" {
  org          = "{{.Org}}"
  vdc_group_id = vcd_vdc_group.test1.id

  above_rule_id = vcd_nsxt_distributed_firewall_rule.r2.id # Order management element

  name      = "rule3"
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

resource "vcd_nsxt_distributed_firewall_rule" "r4" {
  org          = "{{.Org}}"
  vdc_group_id = vcd_vdc_group.test1.id

  name        = "rule4"
  action      = "ALLOW"
  ip_protocol = "IPV6"
  direction   = "OUT"

  # Simulate a Firewall rule addition when some other ordered rules already exist.
  # This rule should be added to the bottom of rule list as it has no specific 'above_rule_id'
  depends_on = [vcd_nsxt_distributed_firewall_rule.r1, vcd_nsxt_distributed_firewall_rule.r2, vcd_nsxt_distributed_firewall_rule.r3]
}
`

const dfwRuleStep3 = dfwRuleStep2 + `
# skip-binary-test: Data Source test
data "vcd_nsxt_distributed_firewall" "all-rules" {
  org          = "{{.Org}}"
  vdc_group_id = vcd_vdc_group.test1.id
}
`

const dfwRuleStep4DS = dfwRuleStep2 + `
# skip-binary-test: Data Source test
data "vcd_nsxt_distributed_firewall_rule" "r1" {
  org          = "{{.Org}}"
  vdc_group_id = vcd_vdc_group.test1.id

  name = "rule1"
}

data "vcd_nsxt_distributed_firewall_rule" "r2" {
  org          = "{{.Org}}"
  vdc_group_id = vcd_vdc_group.test1.id

  name = "rule2"
}

data "vcd_nsxt_distributed_firewall_rule" "r3" {
  org          = "{{.Org}}"
  vdc_group_id = vcd_vdc_group.test1.id

  name = "rule3"
}
`
const dfwRuleStep5 = constDfwPrereqs + `
# skip-binary-test: Testing update capability
resource "vcd_nsxt_distributed_firewall_rule" "r1" {
  org          = "{{.Org}}"
  vdc_group_id = vcd_vdc_group.test1.id

  name        = "rule1-updated"
  action      = "ALLOW"
  description = "description-updated"

  source_ids           = [vcd_nsxt_ip_set.set1.id, vcd_nsxt_ip_set.set2.id]
  destination_ids      = [vcd_nsxt_security_group.g1-empty.id, vcd_nsxt_security_group.g2.id]
  app_port_profile_ids = [vcd_nsxt_app_port_profile.p1.id, data.vcd_nsxt_app_port_profile.WINS.id, data.vcd_nsxt_app_port_profile.FTP.id]
}

resource "vcd_nsxt_distributed_firewall_rule" "r2" {
  org          = "{{.Org}}"
  vdc_group_id = vcd_vdc_group.test1.id

  above_rule_id = vcd_nsxt_distributed_firewall_rule.r1.id # Order management element
  name        = "rule2-updated"
  action      = "ALLOW"
  ip_protocol = "IPV4"
}
`

// TestAccVcdDistributedFirewallRuleVCD10_3_2 complements TestAccVcdDistributedFirewallRule and aims
// to test our 3 new fields of VCD 10.3.2+ in distributed firewall:
// * comment (this one is shown in UI, not like `description`)
// * source_groups_excluded (negates the values specified in source_ids)
// * destination_groups_excluded (negates the values specified in destinations_ids)
func TestAccVcdDistributedFirewallRuleVCD10_3_2(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

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
		"ProviderVdcStorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"Dfw":                       "true",
		"DefaultPolicy":             "true",
		"RemoveDefaultFirewallRule": "true", // will remove default firewall rule in VDC Group
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
	configText2 := templateFill(dfwStep2RuleVCD10_3_2, params)
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
					resource.TestCheckResourceAttrSet("vcd_nsxt_distributed_firewall_rule.r1", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_distributed_firewall_rule.r2", "id"),

					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r1", "name", "rule2"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r1", "action", "DROP"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r1", "enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r1", "logging", "true"),

					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r1", "source_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r1", "destination_ids.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r1", "destination_groups_excluded", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r1", "app_port_profile_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r1", "network_context_profile_ids.#", "1"),

					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r2", "name", "rule1"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r2", "action", "REJECT"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r2", "description", "description"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r2", "comment", "longer text comment field filled"),

					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r2", "source_ids.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r2", "source_groups_excluded", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r2", "destination_ids.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r2", "destination_groups_excluded", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r2", "app_port_profile_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall_rule.r2", "network_context_profile_ids.#", "0"),
				),
			},
		},
	})
	postTestChecks(t)
}

const dfwStep2RuleVCD10_3_2 = testAccVcdVdcGroupNew + `
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

resource "vcd_nsxt_distributed_firewall_rule" "r1" {
  org          = "{{.Org}}"
  vdc_group_id = vcd_vdc_group.test1.id

  name      = "rule2"
  action    = "DROP"
  enabled   = false
  logging   = true
  destination_ids             = [vcd_nsxt_security_group.g2.id,vcd_nsxt_ip_set.set1.id]
  destination_groups_excluded = true

  network_context_profile_ids = [data.vcd_nsxt_network_context_profile.cp1.id]
}

resource "vcd_nsxt_distributed_firewall_rule" "r2" {
  org          = "{{.Org}}"
  vdc_group_id = vcd_vdc_group.test1.id

  name        = "rule1"
  action      = "REJECT"
  description = "description"
  comment     = "longer text comment field filled"

  source_ids             = [vcd_nsxt_security_group.g2.id]
  source_groups_excluded = true

  destination_ids             = [vcd_nsxt_ip_set.set1.id]
  destination_groups_excluded = true
}
`
