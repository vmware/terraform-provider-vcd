// +build gateway nat ALL functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdEdgeSnat(t *testing.T) {
	preTestChecks(t)

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"ExternalIp":  testConfig.Networking.ExternalIp,
		"NetworkName": "my-vdc-int-net",
		"Tags":        "gateway nat",
	}

	configText := templateFill(testAccVcdEdgeSnatRule, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 0: %s", configText)

	params["FuncName"] = t.Name() + "-step1"
	configText2 := templateFill(testAccVcdEdgeSnatRuleUpdate, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckVcdNatRuleDestroy("vcd_nsxv_snat.test"),
		Steps: []resource.TestStep{
			resource.TestStep{ // Step 0 - minimal configuration and data source
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_snat.test", "id", regexp.MustCompile(`\d*`)),
					// When rule_tag is not specified - we expect it to be the same as ID
					resource.TestCheckResourceAttrPair("vcd_nsxv_snat.test", "rule_tag", "vcd_nsxv_snat.test", "id"),
					resource.TestCheckResourceAttr("vcd_nsxv_snat.test", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxv_snat.test", "original_address", "4.4.4.160"),
					resource.TestCheckResourceAttr("vcd_nsxv_snat.test", "translated_address", testConfig.Networking.ExternalIp),
					resource.TestCheckResourceAttr("vcd_nsxv_snat.test", "rule_type", "user"),
					resource.TestCheckResourceAttr("vcd_nsxv_snat.test", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_snat.test", "logging_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_snat.test", "network_name", "test-org-for-snat"),
					resource.TestCheckResourceAttr("vcd_nsxv_snat.test", "network_type", "org"),

					// Data source testing - it must expose all fields which resource has
					resource.TestCheckResourceAttrPair("vcd_nsxv_snat.test", "id", "data.vcd_nsxv_snat.data-test", "id"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_snat.test", "rule_type", "data.vcd_nsxv_snat.data-test", "rule_type"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_snat.test", "rule_tag", "data.vcd_nsxv_snat.data-test", "rule_tag"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_snat.test", "enabled", "data.vcd_nsxv_snat.data-test", "enabled"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_snat.test", "logging_enabled", "data.vcd_nsxv_snat.data-test", "logging_enabled"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_snat.test", "description", "data.vcd_nsxv_snat.data-test", "description"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_snat.test", "original_address", "data.vcd_nsxv_snat.data-test", "original_address"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_snat.test", "translated_address", "data.vcd_nsxv_snat.data-test", "translated_address"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_snat.test", "network_name", "data.vcd_nsxv_snat.data-test", "network_name"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_snat.test", "network_type", "data.vcd_nsxv_snat.data-test", "network_type"),
				),
			},
			resource.TestStep{ // Step 1 - update
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_snat.test", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxv_snat.test", "original_address", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_nsxv_snat.test", "translated_address", "4.4.4.170"),
					resource.TestCheckResourceAttr("vcd_nsxv_snat.test", "rule_type", "user"),
					resource.TestCheckResourceAttr("vcd_nsxv_snat.test", "enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_snat.test", "logging_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_snat.test", "description", "test suite snat rule"),
					resource.TestCheckResourceAttr("vcd_nsxv_snat.test", "network_name", "test-org-for-snat"),
					resource.TestCheckResourceAttr("vcd_nsxv_snat.test", "network_type", "org"),
				),
			},
			resource.TestStep{ // Step 2 - resource import
				ResourceName:      "vcd_nsxv_snat.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdByResourceName("vcd_nsxv_snat.test"),
			},
		},
	})
	postTestChecks(t)
}

const testAccNetForNat = `
resource "vcd_network_routed" "net" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"

  name         = "test-org-for-snat"
  gateway      = "4.4.4.1"

  static_ip_pool {
    start_address = "4.4.4.152"
    end_address   = "4.4.4.254"
  }
}
`

const testAccVcdEdgeSnatRule = testAccNetForNat + `
resource "vcd_nsxv_snat" "test" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"

  network_type = "org"
  network_name = vcd_network_routed.net.name

  original_address   = "4.4.4.160"
  translated_address = "{{.ExternalIp}}"
}

data "vcd_nsxv_snat" "data-test" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  rule_id      = vcd_nsxv_snat.test.id
}
`

const testAccVcdEdgeSnatRuleUpdate = testAccNetForNat + `
resource "vcd_nsxv_snat" "test" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"

  description = "test suite snat rule"

  enabled         = false
  logging_enabled = true

  network_type = "org"
  network_name = vcd_network_routed.net.name

  original_address   = "1.1.1.1"
  translated_address = "4.4.4.170"
}
`
