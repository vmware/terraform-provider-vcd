// +build gateway nat ALL functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccVcdEdgeSnat(t *testing.T) {

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"ExternalIp":  testConfig.Networking.ExternalIp,
		"InternalIp":  testConfig.Networking.InternalIp,
		"NetworkName": "my-vdc-int-net",
		"Tags":        "egatewaydge nat",
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
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckVcdNatRuleDestroy("vcd_nsxv_snat.test"),
		Steps: []resource.TestStep{
			resource.TestStep{ // Step 0 - minimal configuration and data source
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_snat.test", "id", regexp.MustCompile(`\d*`)),
					// When rule_tag is not specified - we expect it to be the same as ID
					resource.TestCheckResourceAttrPair("vcd_nsxv_snat.test", "rule_tag", "vcd_nsxv_snat.test", "id"),
					resource.TestCheckResourceAttr("vcd_nsxv_snat.test", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxv_snat.test", "original_address", testConfig.Networking.InternalIp),
					resource.TestCheckResourceAttr("vcd_nsxv_snat.test", "translated_address", testConfig.Networking.ExternalIp),
					resource.TestCheckResourceAttr("vcd_nsxv_snat.test", "rule_type", "user"),
					resource.TestCheckResourceAttr("vcd_nsxv_snat.test", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_snat.test", "logging_enabled", "false"),

					// Data source testing - it must expose all fields which resource has
					resource.TestCheckResourceAttrPair("vcd_nsxv_snat.test", "id", "data.vcd_nsxv_snat.data-test", "id"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_snat.test", "rule_type", "data.vcd_nsxv_snat.data-test", "rule_type"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_snat.test", "rule_tag", "data.vcd_nsxv_snat.data-test", "rule_tag"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_snat.test", "enabled", "data.vcd_nsxv_snat.data-test", "enabled"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_snat.test", "logging_enabled", "data.vcd_nsxv_snat.data-test", "logging_enabled"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_snat.test", "description", "data.vcd_nsxv_snat.data-test", "description"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_snat.test", "original_address", "data.vcd_nsxv_snat.data-test", "original_address"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_snat.test", "translated_address", "data.vcd_nsxv_snat.data-test", "translated_address"),
				),
			},
			resource.TestStep{ // Step 1 - update
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_snat.test", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxv_snat.test", "original_address", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_nsxv_snat.test", "translated_address", testConfig.Networking.InternalIp),
					resource.TestCheckResourceAttr("vcd_nsxv_snat.test", "rule_type", "user"),
					resource.TestCheckResourceAttr("vcd_nsxv_snat.test", "enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_snat.test", "logging_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_snat.test", "description", "test suite snat rule"),
				),
			},
			resource.TestStep{ // Step 2 - resource import
				ResourceName:      "vcd_nsxv_snat.imported",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdByOrgVdcEdgeUnknownId("vcd_nsxv_snat.test"),
			},
		},
	})
}

const testAccVcdEdgeSnatRule = `
resource "vcd_nsxv_snat" "test" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"

  network_type = "org"
  network_name = "{{.NetworkName}}"

  original_address   = "{{.InternalIp}}"
  translated_address = "{{.ExternalIp}}"
}

data "vcd_nsxv_snat" "data-test" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  rule_id      = "${vcd_nsxv_snat.test.id}"
}
`

const testAccVcdEdgeSnatRuleUpdate = `
resource "vcd_nsxv_snat" "test" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"

  description = "test suite snat rule"

  enabled         = false
  logging_enabled = true

  network_type = "org"
  network_name = "{{.NetworkName}}"

  original_address   = "1.1.1.1"
  translated_address = "{{.InternalIp}}"
}
`
