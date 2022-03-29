//go:build gateway || nsxt || ALL || functional || vdcGroup
// +build gateway nsxt ALL functional vdcGroup

package vcd

import (
	"regexp"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdDistributedFirewall(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	skipNoNsxtConfiguration(t)

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
		PreCheck:          func() { testAccPreCheck(t) },

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
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.source_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.destination_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.app_port_profile_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.network_context_profile_ids.#", "0"),

					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.name", "rule2"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.direction", "IN_OUT"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.ip_protocol", "IPV4_IPV6"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.action", "DROP"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.source_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.destination_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.app_port_profile_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.network_context_profile_ids.#", "0"),

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
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.4.app_port_profile_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.4.network_context_profile_ids.#", "0"),
					// stateDumper(),
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
func stateDumper() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		spew.Dump(s)
		return nil
	}
}

const dfwStep2 = testAccVcdVdcGroupNew + `
resource "vcd_nsxt_distributed_firewall" "t1" {
  org          = "{{.Org}}"
  vdc_group_id = vcd_vdc_group.test1.id

  rule {
	name        = "rule1"
	action      = "ALLOW"
	description = "description"
  }

  rule {
	name      = "rule2"
	action    = "DROP"
	enabled   = false
	logging   = true
	direction = "IN_OUT"
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
  }
}
`

const dfwStep3DS = dfwStep2 + `
data "vcd_nsxt_distributed_firewall" "t1" {
  org          = "{{.Org}}"
  vdc_group_id = vcd_vdc_group.test1.id
}
`

const dfwStep4 = testAccVcdVdcGroupNew + `
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
data "vcd_nsxt_distributed_firewall" "t1" {
  org          = "{{.Org}}"
  vdc_group_id = vcd_vdc_group.test1.id
}
`

// TestAccVcdDistributedFirewallVCD10_2_2 tests out new 10.2.2+ feature:
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

	skipNoNsxtConfiguration(t)

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
		PreCheck:          func() { testAccPreCheck(t) },

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

	skipNoNsxtConfiguration(t)

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
		PreCheck:          func() { testAccPreCheck(t) },

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
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.source_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.destination_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.app_port_profile_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.0.network_context_profile_ids.#", "0"),
					// resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "source_groups_excluded", "true"),
					// resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "destination_groups_excluded", "true"),

					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.name", "rule2"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.direction", "IN_OUT"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.ip_protocol", "IPV4_IPV6"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.action", "DROP"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.source_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.destination_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.app_port_profile_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.network_context_profile_ids.#", "1"),
					resource.TestMatchResourceAttr("vcd_nsxt_distributed_firewall.t1", "rule.1.network_context_profile_ids.0", regexp.MustCompile(`^urn:vcloud:networkContextProfile:`)),
					// resource.TestCheckResourceAttr("vcd_nsxt_distributed_firewall.t1", "source_groups_excluded", "true"),

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

resource "vcd_nsxt_distributed_firewall" "t1" {
  org          = "{{.Org}}"
  vdc_group_id = vcd_vdc_group.test1.id

  rule {
	name        = "rule1"
	action      = "REJECT"
	description = "description"
	comment     = "longer text comment field filled"

	// source_groups_excluded      = true
	// destination_groups_excluded = true
  }

  rule {
	name      = "rule2"
	action    = "DROP"
	enabled   = false
	logging   = true
	direction = "IN_OUT"
	network_context_profile_ids = [data.vcd_nsxt_network_context_profile.cp1.id]

	// source_groups_excluded = true
  }

  rule {
	name        = "rule3"
	action      = "DROP"
	ip_protocol = "IPV4"

	// destination_groups_excluded = true
  }
}
`
