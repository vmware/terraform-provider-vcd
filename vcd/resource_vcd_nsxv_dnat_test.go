// +build gateway nat ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdEdgeDnat(t *testing.T) {

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"ExternalIp":  testConfig.Networking.ExternalIp,
		"InternalIp":  testConfig.Networking.InternalIp,
		"NetworkName": testConfig.Networking.ExternalNetwork,
		"Tags":        "gateway nat",
	}

	configText := templateFill(testAccVcdEdgeDnatRule, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 0: %s", configText)

	params["FuncName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccVcdEdgeDnatRuleUpdate, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdEdgeDnatRuleUpdate2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcdEdgeDnatRuleUpdateOrg, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "-step5"
	configText5 := templateFill(testAccVcdEdgeDnatRuleIcmp, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s", configText5)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckVcdNatRuleDestroy("vcd_nsxv_dnat.test2"),
		Steps: []resource.TestStep{
			resource.TestStep{ // Step 0 - minimal configuration and data source
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_dnat.test", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttrPair("vcd_nsxv_dnat.test", "rule_tag", "vcd_nsxv_dnat.test", "id"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "network_type", "ext"),
					// resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "vnic", "0"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "protocol", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "original_port", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "translated_port", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "original_address", testConfig.Networking.ExternalIp),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "translated_address", testConfig.Networking.InternalIp),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "rule_type", "user"),

					// Data source testing - it must expose all fields which resource has
					resource.TestCheckResourceAttrPair("vcd_nsxv_dnat.test", "id", "data.vcd_nsxv_dnat.data-test", "id"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_dnat.test", "rule_type", "data.vcd_nsxv_dnat.data-test", "rule_type"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_dnat.test", "network_type", "data.vcd_nsxv_dnat.data-test", "network_type"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_dnat.test", "network_name", "data.vcd_nsxv_dnat.data-test", "network_name"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_dnat.test", "rule_tag", "data.vcd_nsxv_dnat.data-test", "rule_tag"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_dnat.test", "enabled", "data.vcd_nsxv_dnat.data-test", "enabled"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_dnat.test", "logging_enabled", "data.vcd_nsxv_dnat.data-test", "logging_enabled"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_dnat.test", "description", "data.vcd_nsxv_dnat.data-test", "description"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_dnat.test", "original_address", "data.vcd_nsxv_dnat.data-test", "original_address"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_dnat.test", "protocol", "data.vcd_nsxv_dnat.data-test", "protocol"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_dnat.test", "icmp_type", "data.vcd_nsxv_dnat.data-test", "icmp_type"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_dnat.test", "original_port", "data.vcd_nsxv_dnat.data-test", "original_port"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_dnat.test", "translated_address", "data.vcd_nsxv_dnat.data-test", "translated_address"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_dnat.test", "translated_port", "data.vcd_nsxv_dnat.data-test", "translated_port"),
				),
			},
			resource.TestStep{ // Step 1 - update
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_dnat.test", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "network_type", "ext"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "protocol", "tcp"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "original_port", "443"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "translated_port", "8443"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "original_address", testConfig.Networking.ExternalIp),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "translated_address", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "rule_type", "user"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "description", "sending quote \""),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "logging_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "enabled", "false"),
				),
			},

			resource.TestStep{ // Step 2 - update with majority defaulted fields
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_dnat.test", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "network_type", "ext"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "protocol", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "original_port", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "translated_port", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "original_address", testConfig.Networking.ExternalIp),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "translated_address", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "rule_type", "user"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "logging_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "enabled", "true"),
				),
			},
			resource.TestStep{ // Step 3 - switch nat rule to org network
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_dnat.test", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "network_type", "org"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "network_name", "test-org-for-dnat"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "protocol", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "original_port", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "translated_port", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "original_address", "10.10.0.180"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "translated_address", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "rule_type", "user"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "logging_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "enabled", "true"),
				),
			},
			resource.TestStep{ // Step 4 - resource import
				ResourceName:      "vcd_nsxv_dnat.imported",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdByResourceName("vcd_nsxv_dnat.test"),
			},
			resource.TestStep{ // Step 5 - Another resource with different settings
				Config: configText5,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_dnat.test2", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test2", "rule_tag", "70000"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test2", "protocol", "icmp"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test2", "icmp_type", "router-advertisement"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test2", "original_port", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test2", "translated_port", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test2", "original_address", testConfig.Networking.ExternalIp),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test2", "translated_address", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test2", "rule_type", "user"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test2", "logging_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test2", "enabled", "true"),
				),
			},
		},
	})
}

// importStateIdByResourceName constructs an import path (ID in Terraform import terms) in the format of:
// organization.vdc.edge-gateway-name.import-object-id (i.e. my-org.my-vdc.my-edge-gw.objectId)
// It uses terraform.State to find existing object's ID by 'resource.resource-name'
func importStateIdByResourceName(resource string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return "", fmt.Errorf("not found resource: %s", resource)
		}

		if rs.Primary.ID == "" {
			return "", fmt.Errorf("no ID is set for %s resource", resource)
		}

		importId := testConfig.VCD.Org +
			ImportSeparator + testConfig.VCD.Vdc +
			ImportSeparator + testConfig.Networking.EdgeGateway +
			ImportSeparator + rs.Primary.ID
		if testConfig.VCD.Org == "" || testConfig.VCD.Vdc == "" || testConfig.Networking.EdgeGateway == "" || rs.Primary.ID == "" {
			return "", fmt.Errorf("missing information to generate import path: %s", importId)
		}

		return importId, nil
	}
}

func testAccCheckVcdNatRuleDestroy(resource string) resource.TestCheckFunc {
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

		rule, err := edgeGateway.GetNsxvNatRuleById(rs.Primary.ID)

		if !govcd.IsNotFound(err) || rule != nil {
			return fmt.Errorf("NAT rule (ID: %s) was not deleted: %s", rs.Primary.ID, err)
		}
		return nil
	}
}

const testAccVcdEdgeDnatRule = `
resource "vcd_nsxv_dnat" "test" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  
  network_type = "ext"
  network_name = "{{.NetworkName}}"

  original_address   = "{{.ExternalIp}}"
  translated_address = "{{.InternalIp}}"
}

data "vcd_nsxv_dnat" "data-test" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  rule_id      = vcd_nsxv_dnat.test.id
}
`

const testAccVcdEdgeDnatRuleUpdate = `
resource "vcd_nsxv_dnat" "test" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"

  network_type = "ext"
  network_name = "{{.NetworkName}}"

  enabled         = false
  logging_enabled = true

  original_address   = "{{.ExternalIp}}"
  translated_address = "1.1.1.1"

  protocol        = "tcp"
  original_port   = 443
  translated_port = 8443

  description = "sending quote \""
}
`

const testAccVcdEdgeDnatRuleUpdate2 = `
resource "vcd_nsxv_dnat" "test" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"

  network_type = "ext"
  network_name = "{{.NetworkName}}"

  original_address   = "{{.ExternalIp}}"
  translated_address = "1.1.1.1"
}
`

const testAccVcdEdgeDnatRuleUpdateOrg = `
resource "vcd_network_routed" "net" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"

  name         = "test-org-for-dnat"
  gateway      = "10.10.0.1"

  static_ip_pool {
    start_address = "10.10.0.152"
    end_address   = "10.10.0.254"
  }
}

resource "vcd_nsxv_dnat" "test" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"

  network_type = "org"
  network_name = vcd_network_routed.net.name

  original_address   = "10.10.0.180"
  translated_address = "1.1.1.1"
}
`

const testAccVcdEdgeDnatRuleIcmp = `
resource "vcd_nsxv_dnat" "test2" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"

  network_type = "ext"
  network_name = "{{.NetworkName}}"

  rule_tag = "70000"

  logging_enabled = true

  protocol  = "icmp"
  icmp_type = "router-advertisement"

  original_address   = "{{.ExternalIp}}"
  translated_address = "1.1.1.1"
}
`
