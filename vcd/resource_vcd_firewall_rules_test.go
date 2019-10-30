// +build functional gateway ALL

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

var itemName string = "TestAccVcdFirewallRules_basic"

func TestAccVcdFirewallRules_basic(t *testing.T) {
	if testConfig.Networking.ExternalIp == "" {
		t.Skip("Variable networking.externalIp must be set to run DNAT tests")
		return
	}
	var existingRules, fwRules govcd.EdgeGateway
	newConfig := createFirewallRulesConfigs(&existingRules)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: newConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdFirewallRulesExists("vcd_firewall_rules."+itemName, &fwRules),
					testAccCheckVcdFirewallRulesAttributes(&fwRules, &existingRules),
				),
			},
		},
	})

}

func testAccCheckVcdFirewallRulesExists(n string, gateway *govcd.EdgeGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no record ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		edgeGateway, err := conn.GetEdgeGateway(testConfig.VCD.Org, testConfig.VCD.Vdc, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, err)
		}

		*gateway = *edgeGateway

		return nil
	}
}

func testAccCheckVcdFirewallRulesAttributes(newRules, existingRules *govcd.EdgeGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if len(newRules.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.FirewallService.FirewallRule) != len(existingRules.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.FirewallService.FirewallRule)+1 {
			return fmt.Errorf("new firewall rule not added: %d != %d",
				len(newRules.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.FirewallService.FirewallRule),
				len(existingRules.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.FirewallService.FirewallRule)+1)
		}

		return nil
	}
}

func createFirewallRulesConfigs(existingRules *govcd.EdgeGateway) string {
	defaultAction := "drop"
	var edgeGateway *govcd.EdgeGateway
	var err error
	edgeGatewayName := testConfig.Networking.EdgeGateway
	if !vcdShortTest {
		conn := createTemporaryVCDConnection()

		if edgeGatewayName == "" {
			panic(fmt.Errorf("could not get an Edge Gateway. Variable networking.edgeGateway is not set"))
		}
		edgeGateway, err = conn.GetEdgeGateway(testConfig.VCD.Org, testConfig.VCD.Vdc, edgeGatewayName)
		if err != nil {
			panic(err)
		}
		*existingRules = *edgeGateway
		debugPrintf("[DEBUG] Edge gateway: %#v", edgeGateway)
		firewallRules := edgeGateway.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.FirewallService
		defaultAction = firewallRules.DefaultAction
	}
	var params = StringMap{
		"Org":           testConfig.VCD.Org,
		"Vdc":           testConfig.VCD.Vdc,
		"EdgeGateway":   edgeGatewayName,
		"DefaultAction": defaultAction,
		"FuncName":      itemName,
		"Tags":          "gateway",
	}
	configText := templateFill(testAccCheckVcdFirewallRules_add, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	return configText
}

func init() {
	testingTags["gateway"] = "resource_firewall_rules_test.go"
}

const testAccCheckVcdFirewallRules_add = `
resource "vcd_firewall_rules" "{{.FuncName}}" {
  org            = "{{.Org}}"
  vdc            = "{{.Vdc}}"
  edge_gateway   = "{{.EdgeGateway}}"
  default_action = "{{.DefaultAction}}"

  rule {
    description      = "Test rule"
    policy           = "allow"
    protocol         = "any"
    destination_port = "any"
    destination_ip   = "any"
    source_port      = "any"
    source_ip        = "any"
  }
}
`
