package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

var itemName string = "TestAccVcdFirewallRules_basic"

func TestAccVcdFirewallRules_basic(t *testing.T) {
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	if testConfig.Networking.ExternalIp == "" {
		t.Skip("Variable networking.externalIp must be set to run DNAT tests")
		return
	}
	var existingRules, fwRules govcd.EdgeGateway
	newConfig := createFirewallRulesConfigs(&existingRules)

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
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		edgeGateway, err := conn.GetEdgeGateway(testConfig.VCD.Org, testConfig.VCD.Vdc, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, err)
		}

		*gateway = edgeGateway

		return nil
	}
}

func testAccCheckVcdFirewallRulesAttributes(newRules, existingRules *govcd.EdgeGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if len(newRules.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.FirewallService.FirewallRule) != len(existingRules.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.FirewallService.FirewallRule)+1 {
			return fmt.Errorf("New firewall rule not added: %d != %d",
				len(newRules.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.FirewallService.FirewallRule),
				len(existingRules.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.FirewallService.FirewallRule)+1)
		}

		return nil
	}
}

func createFirewallRulesConfigs(existingRules *govcd.EdgeGateway) string {
	config := Config{
		User:            testConfig.Provider.User,
		Password:        testConfig.Provider.Password,
		SysOrg:          testConfig.Provider.SysOrg,
		Org:             testConfig.VCD.Org,
		Vdc:             testConfig.VCD.Vdc,
		Href:            testConfig.Provider.Url,
		InsecureFlag:    testConfig.Provider.AllowInsecure,
		MaxRetryTimeout: 140,
	}

	conn, err := config.Client()
	if err != nil {
		panic(err)
	}

	edgeGatewayName := testConfig.Networking.EdgeGateway
	if edgeGatewayName == "" {
		panic(fmt.Errorf("could not get an Edge Gateway. Variable networking.edgeGateway is not set"))
	}
	//edgeGateway, err := vdc.FindEdgeGateway(edgeGatewayName)
	edgeGateway, err := conn.GetEdgeGateway(testConfig.VCD.Org, testConfig.VCD.Vdc, edgeGatewayName)
	if err != nil {
		panic(err)
	}
	*existingRules = edgeGateway
	debugPrintf("[DEBUG] Edge gateway: %#v", edgeGateway)
	firewallRules := *edgeGateway.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.FirewallService
	var params = StringMap{
		"Org":           testConfig.VCD.Org,
		"Vdc":           testConfig.VCD.Vdc,
		"EdgeGateway":   edgeGatewayName,
		"DefaultAction": firewallRules.DefaultAction,
		"FuncName":      itemName,
	}
	configText := templateFill(testAccCheckVcdFirewallRules_add, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	return configText
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
