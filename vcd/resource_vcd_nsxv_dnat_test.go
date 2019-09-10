// +build gateway nat ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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
		"Tags":        "egatewaydge nat",
	}

	configText := templateFill(testAccVcdEdgeNatRule, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 0: %s", configText)

	params["FuncName"] = t.Name() + "-step1"
	configText2 := templateFill(testAccVcdEdgeNatRule2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckVcdNatRuleDestroy("vcd_nsxv_dnat.test"),
		Steps: []resource.TestStep{
			resource.TestStep{ // Step 0 - minimal configuration and data source
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_dnat.test", "id", regexp.MustCompile(`\d*`)),
					// When rule_tag is not specified - we expect it to be the same as ID
					resource.TestCheckResourceAttrPair("vcd_nsxv_dnat.test", "rule_tag", "vcd_nsxv_dnat.test", "id"),
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
					resource.TestCheckResourceAttrPair("vcd_nsxv_dnat.test", "rule_tag", "data.vcd_nsxv_dnat.data-test", "rule_tag"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_dnat.test", "enabled", "data.vcd_nsxv_dnat.data-test", "enabled"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_dnat.test", "logging_enabled", "data.vcd_nsxv_dnat.data-test", "logging_enabled"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_dnat.test", "description", "data.vcd_nsxv_dnat.data-test", "description"),
					// resource.TestCheckResourceAttrPair("vcd_nsxv_dnat.test", "vnic", "data.vcd_nsxv_dnat.data-test", "vnic"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_dnat.test", "original_address", "data.vcd_nsxv_dnat.data-test", "original_address"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_dnat.test", "protocol", "data.vcd_nsxv_dnat.data-test", "protocol"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_dnat.test", "icmp_type", "data.vcd_nsxv_dnat.data-test", "icmp_type"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_dnat.test", "original_port", "data.vcd_nsxv_dnat.data-test", "original_port"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_dnat.test", "translated_address", "data.vcd_nsxv_dnat.data-test", "translated_address"),
					resource.TestCheckResourceAttrPair("vcd_nsxv_dnat.test", "translated_port", "data.vcd_nsxv_dnat.data-test", "translated_port"),
				),
			},
			resource.TestStep{ // Step 1 - update
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_dnat.test", "id", regexp.MustCompile(`\d*`)),
					// resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "vnic", "0"),p
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "protocol", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "original_port", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "translated_port", "any"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "original_address", testConfig.Networking.ExternalIp),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "translated_address", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_nsxv_dnat.test", "rule_type", "user"),
				),
			},
			resource.TestStep{ // Step 2 - resource import
				ResourceName:      "vcd_nsxv_dnat.imported",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdByOrgVdcEdgeUnknownId("vcd_nsxv_dnat.test"),
			},
		},
	})
}

// importStateIdByOrgVdcEdgeUnknownId constructs an import path (ID in Terraform import terms) in the format of:
// organization.vdc.edge-gateway-nane.import-object-id (i.e. my-org.my-vdc.my-edge-gw.objectId)
// It uses terraform.State to find existing object's ID inside state by
func importStateIdByOrgVdcEdgeUnknownId(resource string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return "", fmt.Errorf("not found resource: %s", resource)
		}

		if rs.Primary.ID == "" {
			return "", fmt.Errorf("no ID is set for %s resource", resource)
		}

		importId := testConfig.VCD.Org + "." + testConfig.VCD.Vdc + "." + testConfig.Networking.EdgeGateway + "." + rs.Primary.ID
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

const testAccVcdEdgeNatRule = `
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
  rule_id      = "${vcd_nsxv_dnat.test.id}"
}
`

const testAccVcdEdgeNatRule2 = `
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
