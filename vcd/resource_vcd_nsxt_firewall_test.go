//go:build network || nsxt || ALL || functional
// +build network nsxt ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdNsxtFirewall(t *testing.T) {
	preTestChecks(t)
	if noTestCredentials() {
		t.Skip("Skipping test run as no credentials are provided and this test needs to lookup VCD version")
		return
	}

	client := createTemporaryVCDConnection(false)

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"NsxtVdc":     testConfig.Nsxt.Vdc,
		"EdgeGw":      testConfig.Nsxt.EdgeGateway,
		"NetworkName": t.Name(),
		"Enabled":     "true",
		// Logging can not be enabled with Org Admin by default therefore this value depends on tests being run as
		// sysadmin (set to 'true') or an org admin (set to 'false')
		"Logging": strconv.FormatBool(client.Client.IsSysAdmin),
		"Tags":    "network nsxt",
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

	params["FuncName"] = t.Name() + "-step5"
	params["Enabled"] = "false"
	configText5 := templateFill(testAccVcdNsxtFirewall2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s", configText5)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testParamsNotEmpty(t, params) },
		CheckDestroy:      testAccCheckNsxtFirewallRulesDestroy(testConfig.Nsxt.Vdc, testConfig.Nsxt.EdgeGateway),
		Steps: []resource.TestStep{
			{
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
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.source_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.destination_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.app_port_profile_ids.#", "0"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_firewall.testing", "id", regexp.MustCompile(`^urn:vcloud:gateway:.*$`)),
					// Validating that datasource populates the same fields as resource
					resourceFieldsEqual("vcd_nsxt_firewall.testing", "data.vcd_nsxt_firewall.testing", nil),
				),
			},
			{
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
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.source_ids.#", "3"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.destination_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.app_port_profile_ids.#", "0"),

					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.1.name", "test_rule-2"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.1.direction", "OUT"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.1.ip_protocol", "IPV6"),
					// Logging can only be true with Sysadmin and this test conditionally sets it to true for Sysadmin tests
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.1.logging", strconv.FormatBool(client.Client.IsSysAdmin)),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.1.action", "DROP"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.1.enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.1.source_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.1.destination_ids.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.1.app_port_profile_ids.#", "1"),

					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.2.name", "test_rule-3"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.2.direction", "IN_OUT"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.2.ip_protocol", "IPV4_IPV6"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.2.logging", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.2.action", "ALLOW"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.2.enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.2.source_ids.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.2.destination_ids.#", "3"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.2.app_port_profile_ids.#", "2"),
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_firewall.testing", "id", regexp.MustCompile(`^urn:vcloud:gateway:.*$`)), // Firewall resource holds ID of Edge Gateway
					resource.TestCheckResourceAttrPair("vcd_nsxt_firewall.testing", "id", "vcd_nsxt_firewall.testing", "edge_gateway_id"),
					// Validating that datasource populates the same fields as resource
					resourceFieldsEqual("vcd_nsxt_firewall.testing", "data.vcd_nsxt_firewall.testing", nil),
				),
			},
			{
				Config: configText5,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_firewall.testing", "id", regexp.MustCompile(`^urn:vcloud:gateway:.*$`)),
					resource.TestCheckResourceAttrPair("vcd_nsxt_firewall.testing", "id", "vcd_nsxt_firewall.testing", "edge_gateway_id"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.#", "3"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.name", "test_rule"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.direction", "IN"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.ip_protocol", "IPV4"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.logging", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.action", "ALLOW"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.source_ids.#", "3"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.destination_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.0.app_port_profile_ids.#", "0"),

					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.1.name", "test_rule-2"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.1.direction", "OUT"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.1.ip_protocol", "IPV6"),
					// Logging can only be true with Sysadmin and this test conditionally sets it to true for Sysadmin tests
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.1.logging", strconv.FormatBool(client.Client.IsSysAdmin)),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.1.action", "DROP"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.1.enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.1.source_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.1.destination_ids.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.1.app_port_profile_ids.#", "1"),

					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.2.name", "test_rule-3"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.2.direction", "IN_OUT"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.2.ip_protocol", "IPV4_IPV6"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.2.logging", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.2.action", "ALLOW"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.2.enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.2.source_ids.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.2.destination_ids.#", "3"),
					resource.TestCheckResourceAttr("vcd_nsxt_firewall.testing", "rule.2.app_port_profile_ids.#", "2"),
				),
			},
			{
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
    action      = "ALLOW"
    name        = "test_rule"
    direction   = "IN"
    ip_protocol = "IPV4"
  }
}
`

const testAccVcdNsxtFirewallAndDs = testAccVcdNsxtFirewall + testAccVcdNsxtFirewallDS

const testAccVcdNsxtFirewall2 = testAccVcdNsxtFirewallPrereqs + `

data "vcd_nsxt_app_port_profile" "ssh" {
  scope = "SYSTEM"
  name  = "SSH"
}

resource "vcd_nsxt_app_port_profile" "custom-app" {
  org   = "{{.Org}}"
  vdc   = "{{.NsxtVdc}}"

  name        = "custom app profile"
  description = "Application port profile for custom application"

  scope       = "TENANT"

  app_port {
    protocol = "ICMPv6"
  }

  app_port {
    protocol = "TCP"
    port     = ["2000", "2010-2020", "12345", "65000"]
  }

  app_port {
    protocol = "UDP"
    port     = ["40000-60000"]
  }
}


resource "vcd_nsxt_security_group" "group" {
  count = 3
  org   = "{{.Org}}"
  vdc   = "{{.NsxtVdc}}"

  # Referring to a data source for testing NSX-T Edge Gateway
  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  name        = "test security group - ${count.index}"
  description = "Members to be added later"
}


resource "vcd_nsxt_firewall" "testing" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  rule {
    action      = "ALLOW"
    name        = "test_rule"
    direction   = "IN"
    ip_protocol = "IPV4"
    source_ids  = vcd_nsxt_security_group.group.*.id
    enabled     = {{.Enabled}}
  }

  rule {
    action               = "DROP"
    name                 = "test_rule-2"
    direction            = "OUT"
    ip_protocol          = "IPV6"
    destination_ids      = [vcd_nsxt_security_group.group.2.id]
    app_port_profile_ids = [data.vcd_nsxt_app_port_profile.ssh.id]
    logging              = {{.Logging}}
    enabled              = {{.Enabled}}
  }

  rule {
    action               = "ALLOW"
    name                 = "test_rule-3"
    direction            = "IN_OUT"
    ip_protocol          = "IPV4_IPV6"
    source_ids           = [vcd_nsxt_security_group.group.1.id]
    destination_ids      = vcd_nsxt_security_group.group.*.id
    app_port_profile_ids = [data.vcd_nsxt_app_port_profile.ssh.id, vcd_nsxt_app_port_profile.custom-app.id]
    enabled              = {{.Enabled}}
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
