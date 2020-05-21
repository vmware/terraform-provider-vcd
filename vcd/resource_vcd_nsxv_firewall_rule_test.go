// +build gateway firewall ALL functional

package vcd

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdNsxvEdgeFirewallRule(t *testing.T) {
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

	// -step6 is "import" by real ID
	// -step7 is "import" by UI ID

	params["FuncName"] = t.Name() + "-step8"
	configText8 := templateFill(testAccVcdEdgeFirewallRule8, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 8: %s", configText8)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckVcdFirewallRuleDestroy("vcd_nsxv_firewall_rule.rule6"),
		Steps: []resource.TestStep{
			resource.TestStep{ // Step 0 - configuration only with ip_addresses
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.rule0", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0", "name", "test-rule"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0", "rule_tag", "30000"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0", "action", "deny"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0", "logging_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0", "source.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule0", "source.0.gateway_interfaces"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule0", "source.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule0", "source.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule0", "source.0.ip_sets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule0", "source.0.security_groups"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0", "destination.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule0", "destination.0.gateway_interfaces"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule0", "destination.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule0", "destination.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule0", "destination.0.ip_sets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule0", "destination.0.security_groups"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule0", "destination.0.ip_addresses"),
					// Test object counts
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0", "source.0.ip_addresses.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0", "destination.0.ip_addresses.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0", "service.#", "1"),
					// Test hash values. The hardcoded hash values ensures that hashing function is not altered
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0", "source.0.ip_addresses.2942403275", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0", "destination.0.ip_addresses.3932350214", "192.168.1.110"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0", "service.455563319.protocol", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0", "service.455563319.port", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0", "service.455563319.source_port", "any"),
					// Resource rule0-2
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.rule0-2", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0-2", "name", "rule 123123"),
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.rule0-2", "rule_tag", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall_rule.rule0-2", "id", "vcd_nsxv_firewall_rule.rule0-2", "rule_tag"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0-2", "action", "deny"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0-2", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0-2", "logging_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0-2", "source.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule0-2", "source.0.gateway_interfaces"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule0-2", "source.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule0-2", "source.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule0-2", "source.0.ip_sets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule0-2", "source.0.security_groups"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0-2", "destination.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule0-2", "destination.0.gateway_interfaces"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule0-2", "destination.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule0-2", "destination.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule0-2", "destination.0.ip_sets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule0-2", "destination.0.security_groups"),

					// Test object counts
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0-2", "source.0.ip_addresses.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0-2", "destination.0.ip_addresses.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0-2", "service.#", "1"),

					// Test hash values. The hardcoded hash values ensures that hashing function is not altered
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0-2", "source.0.ip_addresses.1569065534", "4.4.4.4"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0-2", "destination.0.ip_addresses.4225208097", "5.5.5.5"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0-2", "service.455563319.protocol", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0-2", "service.455563319.port", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule0-2", "service.455563319.source_port", "any"),

					// These two rules should go one after another because an explicit depends_on case is used
					// for "vcd_nsxv_firewall_rule.rule0-2" and above_rule_id field is not used
					checkfirewallRuleOrder("vcd_nsxv_firewall_rule.rule0", "vcd_nsxv_firewall_rule.rule0-2"),

					// Check that data source has all the fields and their values the same as resource
					resourceFieldsEqual("vcd_nsxv_firewall_rule.rule0", "data.vcd_nsxv_firewall_rule.rule0", []string{"rule_id"}),
				),
			},
			resource.TestStep{ // Step 1 - configuration only with gateway_interfaces (internal, external)
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.rule1", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule1", "name", "test-rule-1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule1", "action", "deny"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule1", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule1", "logging_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule1", "source.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule1", "source.0.ip_addresses"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule1", "source.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule1", "source.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule1", "source.0.ip_sets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule1", "source.0.security_groups"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule1", "destination.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule1", "destination.0.ip_addresses"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule1", "destination.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule1", "destination.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule1", "destination.0.ip_sets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule1", "destination.0.security_groups"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule1", "destination.0.ip_addresses"),
					// Test object counts
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule1", "source.0.gateway_interfaces.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule1", "destination.0.gateway_interfaces.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule1", "service.#", "1"),
					// Test hash values. The hardcoded hash values ensures that hashing function is not altered
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule1", "source.0.gateway_interfaces.4195066894", "internal"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule1", "destination.0.gateway_interfaces.2800447414", "external"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule1", "service.2361247303.protocol", "tcp"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule1", "service.2361247303.port", "443"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule1", "service.2361247303.source_port", "any"),
				),
			},
			resource.TestStep{ // Step 2 - configuration only with gateway_interfaces (lookup)
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.rule2", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule2", "name", "test-rule-2"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule2", "action", "deny"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule2", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule2", "logging_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule2", "source.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule2", "source.0.ip_addresses"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule2", "source.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule2", "source.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule2", "source.0.ip_sets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule2", "source.0.security_groups"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule2", "destination.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule2", "destination.0.ip_addresses"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule2", "destination.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule2", "destination.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule2", "destination.0.ip_sets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule2", "destination.0.security_groups"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule2", "destination.0.ip_addresses"),
					// Test object counts
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule2", "source.0.gateway_interfaces.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule2", "destination.0.gateway_interfaces.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule2", "service.#", "1"),
					// Test hash values. The hardcoded hash values ensures that hashing function is not altered
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule2", "source.0.gateway_interfaces.2418442387", "vse"),
					resource.TestCheckOutput("destination_gateway_interface", testConfig.Networking.ExternalNetwork),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule2", "service.1333861436.protocol", "TCP"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule2", "service.1333861436.port", "443-543"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule2", "service.1333861436.source_port", "2000-4000"),
				),
			},
			resource.TestStep{ // Step 3 - only org networks
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.rule3", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule3", "name", ""),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule3", "action", "deny"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule3", "enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule3", "logging_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule3", "source.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule3", "source.0.ip_addresses"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule3", "source.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule3", "source.0.gateway_interfaces"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule3", "source.0.ip_sets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule3", "source.0.security_groups"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule3", "destination.0.exclude", "true"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule3", "destination.0.ip_addresses"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule3", "destination.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule3", "destination.0.gateway_interfaces"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule3", "destination.0.ip_sets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule3", "destination.0.security_groups"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule3", "destination.0.ip_addresses"),
					// Test object counts
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule3", "source.0.org_networks.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule3", "destination.0.org_networks.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule3", "service.#", "1"),
					// Test hash values. The hardcoded hash values ensures that hashing function is not altered
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule3", "source.0.org_networks.1013247540", "firewall-test-0"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule3", "destination.0.org_networks.629137269", "firewall-test-1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule3", "service.2361247303.protocol", "tcp"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule3", "service.2361247303.port", "443"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule3", "service.2361247303.source_port", "any"),
				),
			},
			resource.TestStep{ // Step 4
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.rule4", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule4", "name", "test-rule-4"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule4", "action", "deny"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule4", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule4", "logging_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule4", "source.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule4", "source.0.ip_addresses"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule4", "source.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule4", "source.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule4", "source.0.ip_sets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule4", "source.0.security_groups"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule4", "destination.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule4", "destination.0.ip_addresses"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule4", "destination.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule4", "destination.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule4", "destination.0.ip_sets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule4", "destination.0.security_groups"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule4", "destination.0.ip_addresses"),
					// Test object counts
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule4", "source.0.gateway_interfaces.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule4", "destination.0.gateway_interfaces.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule4", "service.#", "5"),
					// Test hash values. The hardcoded hash values ensures that hashing function is not altered
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule4", "source.0.gateway_interfaces.4195066894", "internal"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule4", "destination.0.gateway_interfaces.2800447414", "external"),
					// Service 1
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule4", "service.2361247303.protocol", "tcp"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule4", "service.2361247303.port", "443"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule4", "service.2361247303.source_port", "any"),
					// Service 2
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule4", "service.2135266082.protocol", "tcp"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule4", "service.2135266082.port", "8443"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule4", "service.2135266082.source_port", "20000-40000"),
					// Service 3
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule4", "service.3674967142.protocol", "UDP"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule4", "service.3674967142.port", "10000"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule4", "service.3674967142.source_port", "any"),
					// Service 4
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule4", "service.4080176191.protocol", "udp"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule4", "service.4080176191.port", "10000"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule4", "service.4080176191.source_port", "20000"),
					// Service 5
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule4", "service.1865210680.protocol", "ICMP"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule4", "service.1865210680.port", ""),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule4", "service.1865210680.source_port", ""),

					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall_rule.rule4", "id", "data.vcd_nsxv_firewall_rule.rule4", "id"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall_rule.rule4", "id", "data.vcd_nsxv_firewall_rule.rule4", "rule_id"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall_rule.rule4", "name", "data.vcd_nsxv_firewall_rule.rule4", "name"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall_rule.rule4", "rule_tag", "data.vcd_nsxv_firewall_rule.rule4", "rule_tag"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall_rule.rule4", "action", "data.vcd_nsxv_firewall_rule.rule4", "action"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall_rule.rule4", "enabled", "data.vcd_nsxv_firewall_rule.rule4", "enabled"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall_rule.rule4", "logging_enabled", "data.vcd_nsxv_firewall_rule.rule4", "logging_enabled"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall_rule.rule4", "source.0.exclude", "data.vcd_nsxv_firewall_rule.rule4", "source.0.exclude"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall_rule.rule4", "source.0.gateway_interfaces", "data.vcd_nsxv_firewall_rule.rule4", "source.0.gateway_interfaces"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall_rule.rule4", "source.0.virtual_machine_ids", "data.vcd_nsxv_firewall_rule.rule4", "source.0.virtual_machine_ids"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall_rule.rule4", "source.0.org_networks", "data.vcd_nsxv_firewall_rule.rule4", "source.0.org_networks"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall_rule.rule4", "source.0.ip_sets", "data.vcd_nsxv_firewall_rule.rule4", "source.0.ip_sets"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall_rule.rule4", "source.0.security_groups", "data.vcd_nsxv_firewall_rule.rule4", "source.0.security_groups"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall_rule.rule4", "destination.0.exclude", "data.vcd_nsxv_firewall_rule.rule4", "destination.0.exclude"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall_rule.rule4", "destination.0.gateway_interfaces", "data.vcd_nsxv_firewall_rule.rule4", "destination.0.gateway_interfaces"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall_rule.rule4", "destination.0.virtual_machine_ids", "data.vcd_nsxv_firewall_rule.rule4", "destination.0.virtual_machine_ids"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall_rule.rule4", "destination.0.org_networks", "data.vcd_nsxv_firewall_rule.rule4", "destination.0.org_networks"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall_rule.rule4", "destination.0.ip_sets", "data.vcd_nsxv_firewall_rule.rule4", "destination.0.ip_sets"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall_rule.rule4", "destination.0.security_groups", "data.vcd_nsxv_firewall_rule.rule4", "destination.0.security_groups"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall_rule.rule4", "destination.0.ip_addresses", "data.vcd_nsxv_firewall_rule.rule4", "destination.0.ip_addresses"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall_rule.rule4", "source.0.ip_addresses.#", "data.vcd_nsxv_firewall_rule.rule4", "source.0.ip_addresses.#"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall_rule.rule4", "destination.0.ip_addresses.#", "data.vcd_nsxv_firewall_rule.rule4", "destination.0.ip_addresses.#"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall_rule.rule4", "service.#", "data.vcd_nsxv_firewall_rule.rule4", "service.#"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall_rule.rule4", "service", "data.vcd_nsxv_firewall_rule.rule4", "service"),

					// Check that data source has all the fields and their values the same as resource
					resourceFieldsEqual("vcd_nsxv_firewall_rule.rule4", "data.vcd_nsxv_firewall_rule.rule4", []string{"rule_id"}),
				),
			},
			resource.TestStep{ // Step 5 -
				Config: configText5,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.rule5", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule5", "name", "test-rule-5"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule5", "action", "accept"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule5", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule5", "logging_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule5", "source.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule5", "source.0.ip_addresses"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule5", "source.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule5", "source.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule5", "source.0.ip_sets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule5", "source.0.security_groups"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule5", "destination.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule5", "destination.0.ip_addresses"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule5", "destination.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule5", "destination.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule5", "destination.0.ip_sets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule5", "destination.0.security_groups"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule5", "destination.0.ip_addresses"),
					// Test object counts
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule5", "source.0.gateway_interfaces.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule5", "destination.0.gateway_interfaces.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule5", "service.#", "2"),
					// Test hash values. The hardcoded hash values ensures that hashing function is not altered
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule5", "source.0.gateway_interfaces.4195066894", "internal"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule5", "destination.0.gateway_interfaces.2800447414", "external"),

					// Service 1
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule5", "service.3088950294.protocol", "tcp"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule5", "service.3088950294.port", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule5", "service.3088950294.source_port", "any"),
					// Service 2
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule5", "service.176422394.protocol", "udp"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule5", "service.176422394.port", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule5", "service.176422394.source_port", "any"),
				),
			},
			resource.TestStep{ // Step 6 - resource import by real ID
				ResourceName:      "vcd_nsxv_firewall_rule.imported",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdByResourceName("vcd_nsxv_firewall_rule.rule5"),
			},
			resource.TestStep{ // Step 7 - resource import by UI Number
				ResourceName:      "vcd_nsxv_firewall_rule.imported",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateFirewallUiNumberByResourceName("vcd_nsxv_firewall_rule.rule5"),
			},
			resource.TestStep{ // Step 8 - two rules - one above another
				Config: configText8,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.rule6", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6", "name", "below-rule"),
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.rule6", "rule_tag", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall_rule.rule6", "id", "vcd_nsxv_firewall_rule.rule6", "rule_tag"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6", "action", "accept"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6", "logging_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6", "source.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule6", "source.0.gateway_interfaces"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule6", "source.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule6", "source.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule6", "source.0.ip_sets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule6", "source.0.security_groups"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6", "destination.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule6", "destination.0.gateway_interfaces"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule6", "destination.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule6", "destination.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule6", "destination.0.ip_sets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule6", "destination.0.security_groups"),

					// Test object counts
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6", "source.0.ip_addresses.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6", "destination.0.ip_addresses.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6", "service.#", "1"),

					// Test hash values. The hardcoded hash values ensures that hashing function is not altered
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6", "source.0.ip_addresses.1914947629", "10.10.10.0/24"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6", "source.0.ip_addresses.2947879336", "11.10.10.0/24"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6", "destination.0.ip_addresses.239267318", "20.10.10.0/24"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6", "destination.0.ip_addresses.3553899635", "21.10.10.0/24"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6", "service.455563319.protocol", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6", "service.455563319.port", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6", "service.455563319.source_port", "any"),

					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "name", "above-rule"),
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "rule_tag", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttrPair("vcd_nsxv_firewall_rule.rule6-6", "id", "vcd_nsxv_firewall_rule.rule6-6", "rule_tag"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "action", "accept"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "logging_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "source.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "source.0.gateway_interfaces"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "source.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "source.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "source.0.ip_sets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "source.0.security_groups"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "destination.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "destination.0.gateway_interfaces"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "destination.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "destination.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "destination.0.ip_sets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "destination.0.security_groups"),

					// Test object counts
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "source.0.ip_addresses.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "destination.0.ip_addresses.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "service.#", "1"),

					// Test hash values. The hardcoded hash values ensures that hashing function is not altered
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "source.0.ip_addresses.2471300224", "30.10.10.0/24"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "source.0.ip_addresses.1323029765", "31.10.10.0/24"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "destination.0.ip_addresses.4135626304", "40.10.10.0/24"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "destination.0.ip_addresses.722894789", "41.10.10.0/24"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "service.455563319.protocol", "ANY"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "service.455563319.port", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "service.455563319.source_port", "any"),

					// vcd_nsxv_firewall_rule.rule6-6 should be above vcd_nsxv_firewall_rule.rule6
					// although it has depends_on = ["vcd_nsxv_firewall_rule.rule6"] which puts its
					// provisioning on later stage, but it uses the explicit positioning field
					// "above_rule_id =  vcd_nsxv_firewall_rule.rule6.id"
					checkfirewallRuleOrder("vcd_nsxv_firewall_rule.rule6-6", "vcd_nsxv_firewall_rule.rule6"),
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

			if vcdTestVerbose {
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

// importStateFirewallUiNumberByResourceName constructs an import path (ID in Terraform import terms) in the format of:
// organization.vdc.edge-gateway-name.ui-no:X
// It uses terraform.State to find existing object's UI ID by 'resource.resource-name'
func importStateFirewallUiNumberByResourceName(resource string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return "", fmt.Errorf("not found resource: %s", resource)
		}

		if rs.Primary.ID == "" {
			return "", fmt.Errorf("no ID is set for %s resource", resource)
		}

		// Find UI Number by having only real ID in the system
		conn := testAccProvider.Meta().(*VCDClient)
		edgeGateway, err := conn.GetEdgeGateway(testConfig.VCD.Org, testConfig.VCD.Vdc, testConfig.Networking.EdgeGateway)
		if err != nil {
			return "", fmt.Errorf(errorUnableToFindEdgeGateway, err)
		}

		allRules, err := edgeGateway.GetAllNsxvFirewallRules()
		if err != nil {
			return "", fmt.Errorf("could not get all firewall rules: %s", err)
		}

		// firewallRuleIndex is used to store firewall rule number in UI
		var firewallRuleIndex int
		var found bool
		for ruleIndex, rule := range allRules {
			// if the rule with reald ID is found
			if rule.ID == rs.Primary.ID {
				firewallRuleIndex = ruleIndex + 1
				found = true
				break
			}
		}

		if !found {
			return "", fmt.Errorf("could not find firewall rule by ID %s", rs.Primary.ID)
		}

		importId := fmt.Sprintf("%s.%s.%s.ui-no.%d", testConfig.VCD.Org, testConfig.VCD.Vdc, testConfig.Networking.EdgeGateway, firewallRuleIndex)
		if testConfig.VCD.Org == "" || testConfig.VCD.Vdc == "" || testConfig.Networking.EdgeGateway == "" {
			return "", fmt.Errorf("missing information to generate import path: %s", importId)
		}

		return importId, nil
	}
}

// checkfirewallRuleOrder function accepts firewall rule HCL address (in format 'vcd_nsxv_firewall_rule.rule-name')
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

		rule, err := edgeGateway.GetNsxvFirewallRuleById(rs.Primary.ID)

		if !govcd.IsNotFound(err) || rule != nil {
			return fmt.Errorf("firewall rule (ID: %s) was not deleted: %s", rs.Primary.ID, err)
		}
		return nil
	}
}

const testAccVcdEdgeFirewallRule0 = `
resource "vcd_nsxv_firewall_rule" "rule0" {
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

resource "vcd_nsxv_firewall_rule" "rule0-2" {
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
	# Dependency helps to ensure provisioning order (which becomes rule processing order)
	depends_on = ["vcd_nsxv_firewall_rule.rule0"]
}

data "vcd_nsxv_firewall_rule" "rule0" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"

	rule_id      = vcd_nsxv_firewall_rule.rule0.id
}
`

const testAccVcdEdgeFirewallRule1 = `
resource "vcd_nsxv_firewall_rule" "rule1" {
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
resource "vcd_nsxv_firewall_rule" "rule2" {
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
		protocol    = "TCP"
		port        = "443-543"
		source_port = "2000-4000"
	}
}

output "destination_gateway_interface" {
	value = tolist(vcd_nsxv_firewall_rule.rule2.destination[0].gateway_interfaces)[0]
}
`

const testAccVcdEdgeFirewallRule3 = `
resource "vcd_nsxv_firewall_rule" "rule3" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	action          = "deny"
	enabled         = "false"
	logging_enabled = "true"

	source {
		org_networks = [vcd_network_routed.test-routed[0].name]
	}
  
	destination {
		exclude      = "true"
		org_networks = [vcd_network_routed.test-routed[1].name]
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
resource "vcd_nsxv_firewall_rule" "rule4" {
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
		protocol = "UDP"
		port     = "10000"
	}

	service {
		protocol    = "udp"
		port        = "10000"
		source_port = "20000"
	}

	service {
		protocol = "ICMP"
	}
  }

data "vcd_nsxv_firewall_rule" "rule4" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"

	rule_id      = vcd_nsxv_firewall_rule.rule4.id
}
`

const testAccVcdEdgeFirewallRule5 = `
resource "vcd_nsxv_firewall_rule" "rule5" {
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

const testAccVcdEdgeFirewallRule8 = `
resource "vcd_nsxv_firewall_rule" "rule6" {
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

resource "vcd_nsxv_firewall_rule" "rule6-6" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name = "above-rule"
	action = "accept"
	above_rule_id = vcd_nsxv_firewall_rule.rule6.id


	source {
		ip_addresses = ["30.10.10.0/24", "31.10.10.0/24"]
	}
  
	destination {
		ip_addresses = ["40.10.10.0/24", "41.10.10.0/24"]
	}

	service {
		protocol = "ANY"
	}

	depends_on = ["vcd_nsxv_firewall_rule.rule6"]
}
`

func TestAccVcdNsxvEdgeFirewallRuleIpSets(t *testing.T) {
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
		"Tags":             "gateway firewall",
	}

	configText := templateFill(testAccVcdEdgeFirewallRuleIpSets, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 0: %s", configText)

	params["FuncName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccVcdEdgeFirewallRuleIpSetsUpdate, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckVcdFirewallRuleDestroy("vcd_nsxv_firewall_rule.ip_sets"),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "name", "rule-with-ip_sets"),
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "rule_tag", regexp.MustCompile(`\d*`)),

					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "action", "accept"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "logging_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "source.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "source.0.gateway_interfaces"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "source.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "source.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "source.0.ip_addresses"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "source.0.security_groups"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "destination.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "destination.0.gateway_interfaces"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "destination.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "destination.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "destination.0.ip_addresses"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "destination.0.security_groups"),

					// Test object counts
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "source.0.ip_sets.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "source.0.ip_sets.1692753458", "acceptance test IPset 1"),

					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "destination.0.ip_sets.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "destination.0.ip_sets.1338510833", "acceptance test IPset 2"),

					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "service.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "service.455563319.port", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "service.455563319.protocol", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "service.455563319.source_port", "any"),
				),
			},
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "name", "updated-rule-with-ip_sets"),
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "rule_tag", regexp.MustCompile(`\d*`)),

					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "action", "accept"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "logging_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "source.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "source.0.gateway_interfaces"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "source.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "source.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "source.0.ip_addresses"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "source.0.security_groups"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "destination.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "destination.0.gateway_interfaces"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "destination.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "destination.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "destination.0.ip_addresses"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "destination.0.security_groups"),

					// Test object counts
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "destination.0.ip_sets.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "destination.0.ip_sets.1692753458", "acceptance test IPset 1"),

					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "source.0.ip_sets.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "source.0.ip_sets.1338510833", "acceptance test IPset 2"),

					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "service.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.ip_sets", "service.1865210680.protocol", "icmp"),
				),
			},
		},
	})
}

const testAccVcdEdgeFirewallRuleIpSets = `
resource "vcd_nsxv_firewall_rule" "ip_sets" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name = "rule-with-ip_sets"
	action = "accept"

	source {
		ip_sets = [vcd_nsxv_ip_set.aceeptance-ipset-1.name]
	}
  
	destination {
		ip_sets = [vcd_nsxv_ip_set.aceeptance-ipset-2.name]
	}

	service {
		protocol = "any"
	}
}

resource "vcd_nsxv_ip_set" "aceeptance-ipset-1" {
	name = "acceptance test IPset 1"
	ip_addresses = ["222.222.222.1/24"]
}

resource "vcd_nsxv_ip_set" "aceeptance-ipset-2" {
	name = "acceptance test IPset 2"
	ip_addresses = ["11.11.11.1-11.11.11.100", "12.12.12.1"]
}
`

const testAccVcdEdgeFirewallRuleIpSetsUpdate = `
resource "vcd_nsxv_firewall_rule" "ip_sets" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name = "updated-rule-with-ip_sets"
	action = "accept"

	source {
		ip_sets = [vcd_nsxv_ip_set.aceeptance-ipset-2.name]
	}
  
	destination {
		ip_sets = [vcd_nsxv_ip_set.aceeptance-ipset-1.name]
	}

	service {
		protocol = "icmp"
	}
}

resource "vcd_nsxv_ip_set" "aceeptance-ipset-1" {
	name = "acceptance test IPset 1"
	ip_addresses = ["222.222.222.1/24"]
}

resource "vcd_nsxv_ip_set" "aceeptance-ipset-2" {
	name = "acceptance test IPset 2"
	ip_addresses = ["11.11.11.1-11.11.11.100", "12.12.12.1"]
}
`

func TestAccVcdNsxvEdgeFirewallRuleVms(t *testing.T) {
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
		"Tags":             "gateway firewall",
	}

	configText := templateFill(testAccVcdEdgeFirewallRuleVms, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 0: %s", configText)

	params["FuncName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccVcdEdgeFirewallRuleVmsUpdate, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckVcdFirewallRuleDestroy("vcd_nsxv_firewall_rule.vms"),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.vms", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.vms", "name", "rule-with-ip_sets"),
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.vms", "rule_tag", regexp.MustCompile(`\d*`)),

					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.vms", "action", "accept"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.vms", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.vms", "logging_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.vms", "source.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.vms", "source.0.gateway_interfaces"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.vms", "source.0.ip_sets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.vms", "source.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.vms", "source.0.ip_addresses"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.vms", "source.0.security_groups"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.vms", "destination.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.vms", "destination.0.gateway_interfaces"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.vms", "destination.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.vms", "destination.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.vms", "destination.0.ip_sets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.vms", "destination.0.security_groups"),

					// Test object counts
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.vms", "source.0.virtual_machine_ids.#", "1"),

					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.vms", "destination.0.ip_addresses.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.vms", "destination.0.ip_addresses.2942403275", "any"),

					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.vms", "service.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.vms", "service.455563319.port", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.vms", "service.455563319.protocol", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.vms", "service.455563319.source_port", "any"),
				),
			},

			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.vms", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.vms", "name", "rule-with-ip_sets"),
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.vms", "rule_tag", regexp.MustCompile(`\d*`)),

					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.vms", "action", "accept"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.vms", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.vms", "logging_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.vms", "source.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.vms", "source.0.gateway_interfaces"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.vms", "source.0.ip_sets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.vms", "source.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.vms", "source.0.virtual_machine_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.vms", "source.0.security_groups"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.vms", "destination.0.exclude", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.vms", "destination.0.gateway_interfaces"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.vms", "destination.0.ip_addresses"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.vms", "destination.0.org_networks"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.vms", "destination.0.ip_sets"),
					resource.TestCheckNoResourceAttr("vcd_nsxv_firewall_rule.vms", "destination.0.security_groups"),

					// Test object counts
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.vms", "destination.0.virtual_machine_ids.#", "1"),

					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.vms", "source.0.ip_addresses.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.vms", "source.0.ip_addresses.2942403275", "any"),

					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.vms", "service.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.vms", "service.455563319.port", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.vms", "service.455563319.protocol", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_firewall_rule.vms", "service.455563319.source_port", "any"),
				),
			},
		},
	})
}

const testAccVcdEdgeFirewallRuleVmsPrereqs = `

resource "vcd_network_routed" "net" {
	org = "{{.Org}}"
	vdc = "{{.Vdc}}"
  
	name         = "fw-routed-net"
	edge_gateway = "{{.EdgeGateway}}"
	gateway      = "47.10.0.1"

	static_ip_pool {
	  start_address = "47.10.0.152"
	  end_address   = "47.10.0.254"
	}
}
resource "vcd_vapp" "fw-test" {
  name = "fw-test"
}

resource "vcd_vapp_org_network" "vappNetwork1" {
  org                = "{{.Org}}"
  vdc                = "{{.Vdc}}"
  vapp_name          = vcd_vapp.fw-test.name
  org_network_name   = vcd_network_routed.net.name 
}

resource "vcd_vapp_vm" "fw-vm" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  vapp_name     = vcd_vapp.fw-test.name
  name          = "fw-test"
  computer_name = "fw-test"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 1024
  cpus          = 2
  cpu_cores     = 1

  network {
    name               = vcd_vapp_org_network.vappNetwork1.org_network_name
    type               = "org"
    ip_allocation_mode = "POOL"
  }
}
`

const testAccVcdEdgeFirewallRuleVms = testAccVcdEdgeFirewallRuleVmsPrereqs + `
resource "vcd_nsxv_firewall_rule" "vms" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name = "rule-with-ip_sets"
	action = "accept"

	source {
		virtual_machine_ids = [vcd_vapp_vm.fw-vm.id]
	}
  
	destination {
		ip_addresses = ["any"]
	}

	service {
		protocol = "any"
	}
}
`

const testAccVcdEdgeFirewallRuleVmsUpdate = testAccVcdEdgeFirewallRuleVmsPrereqs + `
resource "vcd_nsxv_firewall_rule" "vms" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name = "rule-with-ip_sets"
	action = "accept"

	source {
		ip_addresses = ["any"]
	}
  
	destination {
		virtual_machine_ids = [vcd_vapp_vm.fw-vm.id]
	}

	service {
		protocol = "any"
	}
}
`
