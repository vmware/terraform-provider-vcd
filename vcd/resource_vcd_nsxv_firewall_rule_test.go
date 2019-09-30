// +build gateway firewall ALL functional

package vcd

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdNsxvEdgeFirewall(t *testing.T) {

	// String map to fill the template
	var params = StringMap{
		"Org":              testConfig.VCD.Org,
		"Vdc":              testConfig.VCD.Vdc,
		"EdgeGateway":      testConfig.Networking.EdgeGateway,
		"ExternalIp":       testConfig.Networking.ExternalIp,
		"InternalIp":       testConfig.Networking.InternalIp,
		"NetworkName":      testConfig.Networking.ExternalNetwork,
		"RouteNetworkName": "TestAccVcdVAppVmNet",
		"Catalog":          testSuiteCatalogName,
		"CatalogItem":      testSuiteCatalogOVAItem,
		"VappName":         vappName2,
		"VmName":           vmName,
		"Tags":             "gateway firewall",
	}

	configText := templateFill(testAccVcdEdgeFirewallRule0, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 0: %s", configText)

	params["FuncName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccVcdEdgeFirewallRule1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdEdgeFirewallRule2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcdEdgeFirewallRule3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(testAccVcdEdgeFirewallRule4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	params["FuncName"] = t.Name() + "-step5"
	configText5 := templateFill(testAccVcdEdgeFirewallRule5, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s", configText5)

	// -step6 is "import"

	params["FuncName"] = t.Name() + "-step7"
	configText7 := templateFill(testAccVcdEdgeFirewallRule6, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 7: %s", configText7)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckVcdFirewallRuleDestroy("vcd_nsxv_firewall.rule6"),
		Steps: []resource.TestStep{
			resource.TestStep{ // Step 0 - configuration only with ip_addresses
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall.rule0", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0", "name", "test-rule"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0", "rule_tag", "30000"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0", "action", "deny"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0", "logging_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0", "source.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule0", "source.0.gateway_interfaces"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule0", "source.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule0", "source.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule0", "source.0.ipsets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule0", "source.0.security_groups"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0", "destination.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule0", "destination.0.gateway_interfaces"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule0", "destination.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule0", "destination.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule0", "destination.0.ipsets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule0", "destination.0.security_groups"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule0", "destination.0.ip_addresses"),
					// Test object counts
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0", "source.0.ip_addresses.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0", "destination.0.ip_addresses.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0", "service.#", "1"),
					// Test hash values. The hardcoded hash values ensures that hashing function is not altered
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0", "source.0.ip_addresses.2942403275", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0", "destination.0.ip_addresses.3932350214", "192.168.1.110"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0", "service.455563319.protocol", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0", "service.455563319.port", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0", "service.455563319.source_port", "any"),
					// Resource rule0-2
					resource.TestMatchResourceAttr("vcd_nsxv_firewall.rule0-2", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0-2", "name", "rule 123123"),
					resource.TestMatchResourceAttr("vcd_nsxv_firewall.rule0-2", "rule_tag", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall.rule0-2", "id", "vcd_nsxv_firewall.rule0-2", "rule_tag"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0-2", "action", "deny"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0-2", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0-2", "logging_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0-2", "source.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule0-2", "source.0.gateway_interfaces"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule0-2", "source.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule0-2", "source.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule0-2", "source.0.ipsets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule0-2", "source.0.security_groups"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0-2", "destination.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule0-2", "destination.0.gateway_interfaces"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule0-2", "destination.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule0-2", "destination.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule0-2", "destination.0.ipsets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule0-2", "destination.0.security_groups"),

					// Test object counts
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0-2", "source.0.ip_addresses.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0-2", "destination.0.ip_addresses.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0-2", "service.#", "1"),

					// Test hash values. The hardcoded hash values ensures that hashing function is not altered
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0-2", "source.0.ip_addresses.1569065534", "4.4.4.4"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0-2", "destination.0.ip_addresses.4225208097", "5.5.5.5"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0-2", "service.455563319.protocol", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0-2", "service.455563319.port", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule0-2", "service.455563319.source_port", "any"),

					// These two rules should go one after another because an explicit depends_on case is used
					// for "vcd_nsxv_firewall.rule0-2" and above_rule_id field is not used
					checkfirewallRuleOrder("vcd_nsxv_firewall.rule0", "vcd_nsxv_firewall.rule0-2"),

					// Check that data source has all the fields and their values the same as resource
					resourceFieldsEqual("vcd_nsxv_firewall.rule0", "data.vcd_nsxv_firewall.rule0", []string{"rule_id"}),
				),
			},
			resource.TestStep{ // Step 1 - configuration only with gateway_interfaces (internal, external)
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall.rule1", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule1", "name", "test-rule-1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule1", "action", "deny"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule1", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule1", "logging_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule1", "source.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule1", "source.0.ip_addresses"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule1", "source.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule1", "source.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule1", "source.0.ipsets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule1", "source.0.security_groups"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule1", "destination.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule1", "destination.0.ip_addresses"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule1", "destination.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule1", "destination.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule1", "destination.0.ipsets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule1", "destination.0.security_groups"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule1", "destination.0.ip_addresses"),
					// Test object counts
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule1", "source.0.gateway_interfaces.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule1", "destination.0.gateway_interfaces.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule1", "service.#", "1"),
					// Test hash values. The hardcoded hash values ensures that hashing function is not altered
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule1", "source.0.gateway_interfaces.4195066894", "internal"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule1", "destination.0.gateway_interfaces.2800447414", "external"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule1", "service.2361247303.protocol", "tcp"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule1", "service.2361247303.port", "443"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule1", "service.2361247303.source_port", "any"),
				),
			},
			resource.TestStep{ // Step 2 - configuration only with gateway_interfaces (lookup)
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall.rule2", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule2", "name", "test-rule-2"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule2", "action", "deny"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule2", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule2", "logging_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule2", "source.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule2", "source.0.ip_addresses"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule2", "source.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule2", "source.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule2", "source.0.ipsets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule2", "source.0.security_groups"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule2", "destination.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule2", "destination.0.ip_addresses"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule2", "destination.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule2", "destination.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule2", "destination.0.ipsets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule2", "destination.0.security_groups"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule2", "destination.0.ip_addresses"),
					// Test object counts
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule2", "source.0.gateway_interfaces.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule2", "destination.0.gateway_interfaces.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule2", "service.#", "1"),
					// Test hash values. The hardcoded hash values ensures that hashing function is not altered
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule2", "source.0.gateway_interfaces.2418442387", "vse"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule2", "destination.0.gateway_interfaces.1696078302",
						testConfig.Networking.ExternalNetwork),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule2", "service.1333861436.protocol", "tcp"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule2", "service.1333861436.port", "443-543"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule2", "service.1333861436.source_port", "2000-4000"),
				),
			},
			resource.TestStep{ // Step 3 - only org networks
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall.rule3", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule3", "name", ""),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule3", "action", "deny"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule3", "enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule3", "logging_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule3", "source.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule3", "source.0.ip_addresses"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule3", "source.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule3", "source.0.gateway_interfaces"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule3", "source.0.ipsets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule3", "source.0.security_groups"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule3", "destination.0.exclude", "true"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule3", "destination.0.ip_addresses"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule3", "destination.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule3", "destination.0.gateway_interfaces"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule3", "destination.0.ipsets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule3", "destination.0.security_groups"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule3", "destination.0.ip_addresses"),
					// Test object counts
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule3", "source.0.org_networks.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule3", "destination.0.org_networks.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule3", "service.#", "1"),
					// Test hash values. The hardcoded hash values ensures that hashing function is not altered
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule3", "source.0.org_networks.1013247540", "firewall-test-0"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule3", "destination.0.org_networks.629137269", "firewall-test-1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule3", "service.2361247303.protocol", "tcp"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule3", "service.2361247303.port", "443"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule3", "service.2361247303.source_port", "any"),
				),
			},
			resource.TestStep{ // Step 4
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall.rule4", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule4", "name", "test-rule-4"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule4", "action", "deny"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule4", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule4", "logging_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule4", "source.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule4", "source.0.ip_addresses"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule4", "source.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule4", "source.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule4", "source.0.ipsets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule4", "source.0.security_groups"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule4", "destination.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule4", "destination.0.ip_addresses"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule4", "destination.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule4", "destination.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule4", "destination.0.ipsets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule4", "destination.0.security_groups"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule4", "destination.0.ip_addresses"),
					// Test object counts
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule4", "source.0.gateway_interfaces.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule4", "destination.0.gateway_interfaces.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule4", "service.#", "5"),
					// Test hash values. The hardcoded hash values ensures that hashing function is not altered
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule4", "source.0.gateway_interfaces.4195066894", "internal"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule4", "destination.0.gateway_interfaces.2800447414", "external"),
					// Service 1
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule4", "service.2361247303.protocol", "tcp"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule4", "service.2361247303.port", "443"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule4", "service.2361247303.source_port", "any"),
					// Service 2
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule4", "service.2135266082.protocol", "tcp"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule4", "service.2135266082.port", "8443"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule4", "service.2135266082.source_port", "20000-40000"),
					// Service 3
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule4", "service.3674967142.protocol", "udp"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule4", "service.3674967142.port", "10000"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule4", "service.3674967142.source_port", "any"),
					// Service 4
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule4", "service.4080176191.protocol", "udp"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule4", "service.4080176191.port", "10000"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule4", "service.4080176191.source_port", "20000"),
					// Service 5
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule4", "service.1865210680.protocol", "icmp"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule4", "service.1865210680.port", ""),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule4", "service.1865210680.source_port", ""),

					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall.rule4", "id", "data.vcd_nsxv_firewall.rule4", "id"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall.rule4", "id", "data.vcd_nsxv_firewall.rule4", "rule_id"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall.rule4", "name", "data.vcd_nsxv_firewall.rule4", "name"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall.rule4", "rule_tag", "data.vcd_nsxv_firewall.rule4", "rule_tag"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall.rule4", "action", "data.vcd_nsxv_firewall.rule4", "action"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall.rule4", "enabled", "data.vcd_nsxv_firewall.rule4", "enabled"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall.rule4", "logging_enabled", "data.vcd_nsxv_firewall.rule4", "logging_enabled"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall.rule4", "source.0.exclude", "data.vcd_nsxv_firewall.rule4", "source.0.exclude"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall.rule4", "source.0.gateway_interfaces", "data.vcd_nsxv_firewall.rule4", "source.0.gateway_interfaces"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall.rule4", "source.0.virtual_machine_ids", "data.vcd_nsxv_firewall.rule4", "source.0.virtual_machine_ids"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall.rule4", "source.0.org_networks", "data.vcd_nsxv_firewall.rule4", "source.0.org_networks"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall.rule4", "source.0.ipsets", "data.vcd_nsxv_firewall.rule4", "source.0.ipsets"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall.rule4", "source.0.security_groups", "data.vcd_nsxv_firewall.rule4", "source.0.security_groups"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall.rule4", "destination.0.exclude", "data.vcd_nsxv_firewall.rule4", "destination.0.exclude"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall.rule4", "destination.0.gateway_interfaces", "data.vcd_nsxv_firewall.rule4", "destination.0.gateway_interfaces"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall.rule4", "destination.0.virtual_machine_ids", "data.vcd_nsxv_firewall.rule4", "destination.0.virtual_machine_ids"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall.rule4", "destination.0.org_networks", "data.vcd_nsxv_firewall.rule4", "destination.0.org_networks"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall.rule4", "destination.0.ipsets", "data.vcd_nsxv_firewall.rule4", "destination.0.ipsets"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall.rule4", "destination.0.security_groups", "data.vcd_nsxv_firewall.rule4", "destination.0.security_groups"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall.rule4", "destination.0.ip_addresses", "data.vcd_nsxv_firewall.rule4", "destination.0.ip_addresses"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall.rule4", "source.0.ip_addresses.#", "data.vcd_nsxv_firewall.rule4", "source.0.ip_addresses.#"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall.rule4", "destination.0.ip_addresses.#", "data.vcd_nsxv_firewall.rule4", "destination.0.ip_addresses.#"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall.rule4", "service.#", "data.vcd_nsxv_firewall.rule4", "service.#"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall.rule4", "service", "data.vcd_nsxv_firewall.rule4", "service"),

					// Check that data source has all the fields and their values the same as resource
					resourceFieldsEqual("vcd_nsxv_firewall.rule4", "data.vcd_nsxv_firewall.rule4", []string{"rule_id"}),
				),
			},
			resource.TestStep{ // Step 5 -
				Config: configText5,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall.rule5", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule5", "name", "test-rule-5"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule5", "action", "accept"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule5", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule5", "logging_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule5", "source.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule5", "source.0.ip_addresses"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule5", "source.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule5", "source.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule5", "source.0.ipsets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule5", "source.0.security_groups"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule5", "destination.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule5", "destination.0.ip_addresses"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule5", "destination.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule5", "destination.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule5", "destination.0.ipsets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule5", "destination.0.security_groups"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule5", "destination.0.ip_addresses"),
					// Test object counts
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule5", "source.0.gateway_interfaces.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule5", "destination.0.gateway_interfaces.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule5", "service.#", "2"),
					// Test hash values. The hardcoded hash values ensures that hashing function is not altered
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule5", "source.0.gateway_interfaces.4195066894", "internal"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule5", "destination.0.gateway_interfaces.2800447414", "external"),

					// Service 1
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule5", "service.3088950294.protocol", "tcp"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule5", "service.3088950294.port", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule5", "service.3088950294.source_port", "any"),
					// Service 2
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule5", "service.176422394.protocol", "udp"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule5", "service.176422394.port", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule5", "service.176422394.source_port", "any"),
				),
			},
			resource.TestStep{ // Step 6 - resource import
				ResourceName:      "vcd_nsxv_firewall.imported",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdByResourceName("vcd_nsxv_firewall.rule5"),
			},
			resource.TestStep{ // Step 7 - two rules - one above another
				Config: configText7,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall.rule6", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6", "name", "below-rule"),
					resource.TestMatchResourceAttr("vcd_nsxv_firewall.rule6", "rule_tag", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall.rule6", "id", "vcd_nsxv_firewall.rule6", "rule_tag"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6", "action", "accept"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6", "logging_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6", "source.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule6", "source.0.gateway_interfaces"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule6", "source.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule6", "source.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule6", "source.0.ipsets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule6", "source.0.security_groups"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6", "destination.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule6", "destination.0.gateway_interfaces"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule6", "destination.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule6", "destination.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule6", "destination.0.ipsets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule6", "destination.0.security_groups"),

					// Test object counts
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6", "source.0.ip_addresses.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6", "destination.0.ip_addresses.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6", "service.#", "1"),

					// Test hash values. The hardcoded hash values ensures that hashing function is not altered
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6", "source.0.ip_addresses.1914947629", "10.10.10.0/24"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6", "source.0.ip_addresses.2947879336", "11.10.10.0/24"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6", "destination.0.ip_addresses.239267318", "20.10.10.0/24"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6", "destination.0.ip_addresses.3553899635", "21.10.10.0/24"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6", "service.455563319.protocol", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6", "service.455563319.port", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6", "service.455563319.source_port", "any"),

					resource.TestMatchResourceAttr("vcd_nsxv_firewall.rule6-6", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6-6", "name", "above-rule"),
					resource.TestMatchResourceAttr("vcd_nsxv_firewall.rule6-6", "rule_tag", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall.rule6-6", "id", "vcd_nsxv_firewall.rule6-6", "rule_tag"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6-6", "action", "accept"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6-6", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6-6", "logging_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6-6", "source.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule6-6", "source.0.gateway_interfaces"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule6-6", "source.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule6-6", "source.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule6-6", "source.0.ipsets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule6-6", "source.0.security_groups"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6-6", "destination.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule6-6", "destination.0.gateway_interfaces"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule6-6", "destination.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule6-6", "destination.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule6-6", "destination.0.ipsets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall.rule6-6", "destination.0.security_groups"),

					// Test object counts
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6-6", "source.0.ip_addresses.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6-6", "destination.0.ip_addresses.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6-6", "service.#", "1"),

					// Test hash values. The hardcoded hash values ensures that hashing function is not altered
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6-6", "source.0.ip_addresses.2471300224", "30.10.10.0/24"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6-6", "source.0.ip_addresses.1323029765", "31.10.10.0/24"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6-6", "destination.0.ip_addresses.4135626304", "40.10.10.0/24"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6-6", "destination.0.ip_addresses.722894789", "41.10.10.0/24"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6-6", "service.455563319.protocol", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6-6", "service.455563319.port", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall.rule6-6", "service.455563319.source_port", "any"),

					// vcd_nsxv_firewall.rule6-6 should be above vcd_nsxv_firewall.rule6
					// although it has depends_on = ["vcd_nsxv_firewall.rule6"] which puts its
					// provisioning on later stage, but it uses the explicit positioning field
					// "above_rule_id =  vcd_nsxv_firewall.rule6.id"
					checkfirewallRuleOrder("vcd_nsxv_firewall.rule6-6", "vcd_nsxv_firewall.rule6"),
				),
			},
		},
	})
}

// resourceFieldsEqual checks if secondObject has all the fields and their values set as the
// firstObject except `[]excludeFields`. This is very useful to check if data sources have all
// the same values as resources
func resourceFieldsEqual(firstObject, secondObject string, excludeFields []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource1, ok := s.RootModule().Resources[firstObject]
		if !ok {
			return fmt.Errorf("unable to find %s", firstObject)
		}

		resource2, ok := s.RootModule().Resources[secondObject]
		if !ok {
			return fmt.Errorf("unable to find %s", secondObject)
		}

		for fieldName := range resource1.Primary.Attributes {
			// Do not validate the fields marked for exclusion
			if stringInSlice(fieldName, excludeFields) {
				continue
			}

			if os.Getenv(testVerbose) != "" {
				fmt.Printf("field %s %s (value %s) and %s (value %s))\n", fieldName, firstObject,
					resource1.Primary.Attributes[fieldName], secondObject, resource2.Primary.Attributes[fieldName])
			}
			if !reflect.DeepEqual(resource1.Primary.Attributes[fieldName], resource2.Primary.Attributes[fieldName]) {
				return fmt.Errorf("field %s differs in resources %s (value %s) and %s (value %s)",
					fieldName, firstObject, resource1.Primary.Attributes[fieldName], secondObject, resource2.Primary.Attributes[fieldName])
			}
		}
		return nil
	}
}

// checkfirewallRuleOrder function accepts firewall rule HCL address (in format 'vcd_nsxv_firewall.rule-name')
// and checks that its order is as specified (firstRule goes above secondRule)
func checkfirewallRuleOrder(firstRule, secondRule string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)
		edgeGateway, err := conn.GetEdgeGateway(testConfig.VCD.Org, testConfig.VCD.Vdc, testConfig.Networking.EdgeGateway)
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, err)
		}

		rule1, ok := s.RootModule().Resources[firstRule]
		if !ok {
			return fmt.Errorf("not found resource: %s", firstRule)
		}

		rule2, ok := s.RootModule().Resources[secondRule]
		if !ok {
			return fmt.Errorf("not found resource: %s", secondRule)
		}

		rule1Id := rule1.Primary.ID
		rule2Id := rule2.Primary.ID

		allFirewallRules, err := edgeGateway.GetAllNsxvFirewallRules()
		if err != nil {
			return fmt.Errorf("could not get all firewall rules: %s", err)
		}

		// This loop checks that at first rule1Id is found and then rule2Id is found in the ordered
		// list of firewall rules
		var foundSecondRule, foundFirstRule bool
		for _, rule := range allFirewallRules {
			if rule.ID == rule1Id {
				foundFirstRule = true
			}
			if foundFirstRule && rule.ID == rule2Id {
				foundSecondRule = true
			}
		}

		if !foundSecondRule || !foundFirstRule {
			return fmt.Errorf("incorrect rule order. %s is above %s. Should be reverse.", firstRule, secondRule)
		}

		return nil
	}
}

func testAccCheckVcdFirewallRuleDestroy(resource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found resource: %s", resource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set for %s resource", resource)
		}

		conn := testAccProvider.Meta().(*VCDClient)

		edgeGateway, err := conn.GetEdgeGateway(testConfig.VCD.Org, testConfig.VCD.Vdc, testConfig.Networking.EdgeGateway)
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, err)
		}

		rule, err := edgeGateway.GetNsxvFirewallById(rs.Primary.ID)

		if !govcd.IsNotFound(err) || rule != nil {
			return fmt.Errorf("firewall rule (ID: %s) was not deleted: %s", rs.Primary.ID, err)
		}
		return nil
	}
}

const testAccVcdEdgeFirewallRule0 = `
resource "vcd_nsxv_firewall" "rule0" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name = "test-rule"
	rule_tag = "30000"
	action = "deny"

	source {
		ip_addresses = ["any"]
	}
  
	destination {
		ip_addresses = ["192.168.1.110"]
	}
  
	service {
		protocol = "any"
	}
}

resource "vcd_nsxv_firewall" "rule0-2" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name = "rule 123123"
	action = "deny"

	source {
		ip_addresses = ["4.4.4.4"]
	}
  
	destination {
		ip_addresses = ["5.5.5.5"]
	}
  
	service {
		protocol = "any"
	}
	depends_on = ["vcd_nsxv_firewall.rule0"]
}

data "vcd_nsxv_firewall" "rule0" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"

	rule_id      = "${vcd_nsxv_firewall.rule0.id}"
}
`

const testAccVcdEdgeFirewallRule1 = `
resource "vcd_nsxv_firewall" "rule1" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name = "test-rule-1"
	action = "deny"

	source {
		gateway_interfaces = ["internal"]
	}
  
	destination {
		gateway_interfaces = ["external"]
	}
	service {
		protocol = "tcp"
		port     = "443"
	}
  }
`

const testAccVcdEdgeFirewallRule2 = `
resource "vcd_nsxv_firewall" "rule2" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name = "test-rule-2"
	action = "deny"

	source {
		gateway_interfaces = ["vse"]
	}
  
	destination {
		gateway_interfaces = ["{{.NetworkName}}"]
	}
	service {
		protocol    = "tcp"
		port        = "443-543"
		source_port = "2000-4000"
	}
  }
`

const testAccVcdEdgeFirewallRule3 = `
resource "vcd_nsxv_firewall" "rule3" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	action          = "deny"
	enabled         = "false"
	logging_enabled = "true"

	source {
		org_networks = ["${vcd_network_routed.test-routed[0].name}"]
	}
  
	destination {
		exclude      = "true"
		org_networks = ["${vcd_network_routed.test-routed[1].name}"]
	}
	service {
		protocol = "tcp"
		port     = "443"
	}
}
resource "vcd_network_routed" "test-routed" {
  count        = 2
  name         = "firewall-test-${count.index}"
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "10.201.${count.index}.1"
  netmask      = "255.255.255.0"

  static_ip_pool {
    start_address = "10.201.${count.index}.10"
    end_address   = "10.201.${count.index}.20"
  }
}
`

const testAccVcdEdgeFirewallRule4 = `
resource "vcd_nsxv_firewall" "rule4" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name = "test-rule-4"
	action = "deny"

	source {
		gateway_interfaces = ["internal"]
	}
  
	destination {
		gateway_interfaces = ["external"]
	}

	service {
		protocol = "tcp"
		port     = "443"
	}
	
	service {
		protocol = "tcp"
		port     = "8443"
		source_port = "20000-40000"
	}

	service {
		protocol = "udp"
		port     = "10000"
	}

	service {
		protocol    = "udp"
		port        = "10000"
		source_port = "20000"
	}

	service {
		protocol = "icmp"
	}
  }

data "vcd_nsxv_firewall" "rule4" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"

	rule_id      = "${vcd_nsxv_firewall.rule4.id}"
}
`

const testAccVcdEdgeFirewallRule5 = `
resource "vcd_nsxv_firewall" "rule5" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name = "test-rule-5"

	source {
		gateway_interfaces = ["internal"]
	}
  
	destination {
		gateway_interfaces = ["external"]
	}

	service {
		protocol = "tcp"
	}

	service {
		protocol = "udp"
	}

  }
`

const testAccVcdEdgeFirewallRule6 = `
resource "vcd_nsxv_firewall" "rule6" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name = "below-rule"
	action = "accept"

	source {
		ip_addresses = ["10.10.10.0/24", "11.10.10.0/24"]
	}
  
	destination {
		ip_addresses = ["20.10.10.0/24", "21.10.10.0/24"]
	}

	service {
		protocol = "any"
	}
}

resource "vcd_nsxv_firewall" "rule6-6" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name = "above-rule"
	action = "accept"
	above_rule_id = "${vcd_nsxv_firewall.rule6.id}"


	source {
		ip_addresses = ["30.10.10.0/24", "31.10.10.0/24"]
	}
  
	destination {
		ip_addresses = ["40.10.10.0/24", "41.10.10.0/24"]
	}

	service {
		protocol = "any"
	}

	depends_on = ["vcd_nsxv_firewall.rule6"]
}
`
