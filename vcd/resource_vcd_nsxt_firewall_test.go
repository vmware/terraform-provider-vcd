// +build network nsxt ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdNsxtFirewall(t *testing.T) {
	preTestChecks(t)
	skipNoNsxtConfiguration(t)

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"NsxtVdc":     testConfig.Nsxt.Vdc,
		"EdgeGw":      testConfig.Nsxt.EdgeGateway,
		"NetworkName": t.Name(),
		"Tags":        "network nsxt",
	}

	configText1 := templateFill(testAccVcdNsxtFirewall, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdNsxtFirewallAndDs, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcdNsxtFirewall2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(testAccVcdNsxtFirewall2AndDs, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckNsxtFirewallRulesDestroy(testConfig.Nsxt.Vdc, testConfig.Nsxt.EdgeGateway),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_firewall.testing", "id", regexp.MustCompile(`^urn:vcloud:gateway:.*$`)),
					resource.TestCheckResourceAttrPair("vcd_nsxt_firewall.testing", "id", "vcd_nsxt_firewall.testing", "edge_gateway_id"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.name", "test_rule"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.direction", "IN"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.ip_protocol", "IPV4"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.logging", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.action", "ALLOW"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.enabled", "true"),
				),
			},

			resource.TestStep{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_firewall.testing", "id", regexp.MustCompile(`^urn:vcloud:gateway:.*$`)),
					// Validating that datasource populates the same fields as resource
					resourceFieldsEqual("vcd_nsxt_firewall.testing", "data.vcd_nsxt_firewall.testing", nil),
				),
			},
			resource.TestStep{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_firewall.testing", "id", regexp.MustCompile(`^urn:vcloud:gateway:.*$`)),
					resource.TestCheckResourceAttrPair("vcd_nsxt_firewall.testing", "id", "vcd_nsxt_firewall.testing", "edge_gateway_id"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.#", "3"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.name", "test_rule"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.direction", "IN"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.ip_protocol", "IPV4"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.logging", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.action", "ALLOW"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.sources.#", "3"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.destinations.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.applications.#", "0"),

					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.1.name", "test_rule-2"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.1.direction", "OUT"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.1.ip_protocol", "IPV6"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.1.logging", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.1.action", "DROP"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.1.enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.1.sources.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.1.destinations.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.1.applications.#", "0"),

					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.2.name", "test_rule-3"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.2.direction", "IN_OUT"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.2.ip_protocol", "IPV4_IPV6"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.2.logging", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.2.action", "ALLOW"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.2.enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.2.sources.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.2.destinations.#", "3"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.2.applications.#", "0"),
				),
			},
			resource.TestStep{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_firewall.testing", "id", regexp.MustCompile(`^urn:vcloud:gateway:.*$`)), // Firewall resource holds ID of Edge Gateway
					resource.TestCheckResourceAttrPair("vcd_nsxt_firewall.testing", "id", "vcd_nsxt_firewall.testing", "edge_gateway_id"),
					// Validating that datasource populates the same fields as resource
					resourceFieldsEqual("vcd_nsxt_firewall.testing", "data.vcd_nsxt_firewall.testing", nil),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_nsxt_firewall.testing",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcObject(testConfig, testConfig.Nsxt.EdgeGateway),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtFirewallPrereqs = `
data "vcd_nsxt_edgegateway" "testing" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name = "{{.EdgeGw}}"
}
`

const testAccVcdNsxtFirewallDS = `
# skip-binary-test: cannot define resource and datasource in the same file
data "vcd_nsxt_firewall" "testing" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id
}
`

const testAccVcdNsxtFirewall = testAccVcdNsxtFirewallPrereqs + `
resource "vcd_nsxt_firewall" "testing" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  rule {
    name        = "test_rule"
    direction   = "IN"
    ip_protocol = "IPV4"
  }
}
`

const testAccVcdNsxtFirewallAndDs = testAccVcdNsxtFirewall + testAccVcdNsxtFirewallDS

const testAccVcdNsxtFirewall2 = testAccVcdNsxtFirewallPrereqs + `
resource "vcd_nsxt_security_group" "group" {
  count = 3
  org   = "{{.Org}}"
  vdc   = "{{.NsxtVdc}}"

  # Referring to a datasource for testing NSX-T Edge Gateway

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  name        = "test security group - ${count.index}"
  description = "Members to be added later"
}


resource "vcd_nsxt_firewall" "testing" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  rule {
    name        = "test_rule"
    direction   = "IN"
    ip_protocol = "IPV4"
    sources     = vcd_nsxt_security_group.group.*.id
  }

  rule {
    name         = "test_rule-2"
    direction    = "OUT"
    ip_protocol  = "IPV6"
    destinations = [vcd_nsxt_security_group.group.2.id]
    action       = "DROP"
    logging      = true
  }

  rule {
    name         = "test_rule-3"
    direction    = "IN_OUT"
    ip_protocol  = "IPV4_IPV6"
    sources      = [vcd_nsxt_security_group.group.1.id]
    destinations = vcd_nsxt_security_group.group.*.id
  }
}
`

const testAccVcdNsxtFirewall2AndDs = testAccVcdNsxtFirewall2 + testAccVcdNsxtFirewallDS

func testAccCheckNsxtFirewallRulesDestroy(vdcName, edgeGatewayName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, vdcName)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, vdcName, testConfig.VCD.Org, err)
		}

		edge, err := vdc.GetNsxtEdgeGatewayByName(edgeGatewayName)
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, edgeGatewayName)
		}

		fwRules, err := edge.GetNsxtFirewall()
		if err != nil {
			return fmt.Errorf("error retrieving NSX-T Firewall Rules for Edge Gateway '%s': %s", edgeGatewayName, err)
		}

		if len(fwRules.NsxtFirewallRuleContainer.UserDefinedRules) > 0 {
			return fmt.Errorf("there are still some firewall rules remaining (%d)",
				len(fwRules.NsxtFirewallRuleContainer.UserDefinedRules))
		}

		return nil
	}
}
