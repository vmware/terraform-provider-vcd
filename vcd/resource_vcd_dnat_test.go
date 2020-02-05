// +build functional gateway ALL

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

var baseDnatName string = "TestAccVcdDNAT"
var orgVdcNetworkName = "TestAccVcdDNAT_BasicNetwork"

func TestAccVcdDNAT_WithOrgNetw(t *testing.T) {
	if testConfig.Networking.ExternalIp == "" {
		t.Skip("Variable networking.externalIp must be set to run DNAT tests")
		return
	}
	var e govcd.EdgeGateway

	var dnatName = baseDnatName + "_WithOrgNetw"
	var params = StringMap{
		"Org":               testConfig.VCD.Org,
		"Vdc":               testConfig.VCD.Vdc,
		"EdgeGateway":       testConfig.Networking.EdgeGateway,
		"ExternalIp":        testConfig.Networking.ExternalIp,
		"DnatName":          dnatName,
		"OrgVdcNetworkName": orgVdcNetworkName,
		"Gateway":           "10.10.102.1",
		"StartIpAddress":    "10.10.102.51",
		"EndIpAddress":      "10.10.102.100",
		"Tags":              "gateway",
		"Description":       "test run1",
	}

	configText := templateFill(testAccCheckVcdDnatWithOrgNetw, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdDNATDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdDNATExists("vcd_dnat."+dnatName, &e),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "network_name", orgVdcNetworkName),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "network_type", "org"),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "external_ip", testConfig.Networking.ExternalIp),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "port", "7777"),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "protocol", "tcp"),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "internal_ip", "10.10.102.60"),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "description", "test run1"),
				),
			},
		},
	})
}

func TestAccVcdDNAT_WithExtNetw(t *testing.T) {
	if !usingSysAdmin() {
		t.Skip("TestAccVcdDNAT_WithExtNetw requires system admin privileges")
	}
	if testConfig.Networking.ExternalIp == "" {
		t.Skip("Variable networking.externalIp must be set to run DNAT tests")
		return
	}
	var e govcd.EdgeGateway

	var dnatName = baseDnatName + "_WithExtNetw"
	var params = StringMap{
		"Org":                 testConfig.VCD.Org,
		"Vdc":                 testConfig.VCD.Vdc,
		"EdgeGateway":         testConfig.Networking.EdgeGateway,
		"ExternalIp":          testConfig.Networking.ExternalIp,
		"DnatName":            dnatName,
		"OrgVdcNetworkName":   "TestAccVcdDNAT_BasicNetwork",
		"ExternalNetworkName": testConfig.Networking.ExternalNetwork,
		"Gateway":             "10.10.102.1",
		"StartIpAddress":      "10.10.102.51",
		"EndIpAddress":        "10.10.102.100",
		"Tags":                "gateway",
		"Description":         "test run2",
	}

	configText := templateFill(testAccCheckVcdDnatWithExtNetw, params)
	params["FuncName"] = t.Name() + "-Update"
	updateText := templateFill(testAccCheckVcdDnatWithExtNetwUpdate, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdDNATDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdDNATExists("vcd_dnat."+dnatName, &e),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "network_name", testConfig.Networking.ExternalNetwork),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "network_type", "ext"),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "external_ip", testConfig.Networking.ExternalIp),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "port", "7777"),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "protocol", "tcp"),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "internal_ip", "10.10.102.60"),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "translated_port", "77"),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "description", "test run2"),
				),
			},
			resource.TestStep{
				Config: updateText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdDNATExists("vcd_dnat."+dnatName, &e),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "network_name", testConfig.Networking.ExternalNetwork),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "network_type", "ext"),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "external_ip", testConfig.Networking.ExternalIp),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "port", "8888"),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "protocol", "udp"),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "internal_ip", "10.10.102.80"),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "translated_port", "88"),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "description", "test run2"),
				),
			},
		},
	})
}

func testAccCheckVcdDNATExists(n string, gateway *govcd.EdgeGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("DNAT ID is not set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		gatewayName := rs.Primary.Attributes["edge_gateway"]

		edgeGateway, err := conn.GetEdgeGateway(testConfig.VCD.Org, testConfig.VCD.Vdc, gatewayName)

		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, err)
		}

		natRule, err := edgeGateway.GetNatRule(rs.Primary.ID)
		if err != nil {
			return err
		}

		if nil == natRule {
			return fmt.Errorf("rule isn't found")
		}

		gateway = edgeGateway

		return nil
	}
}

func testAccCheckVcdDNATDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_dnat" {
			continue
		}

		gatewayName := rs.Primary.Attributes["edge_gateway"]
		edgeGateway, err := conn.GetEdgeGateway(testConfig.VCD.Org, testConfig.VCD.Vdc, gatewayName)

		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, err)
		}

		_, err = edgeGateway.GetNatRule(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("DNAT rule still exists")
		}
	}

	return nil
}

func TestAccVcdDNAT_ForBackCompability(t *testing.T) {
	if testConfig.Networking.ExternalIp == "" {
		t.Skip("Variable networking.externalIp must be set to run DNAT tests")
		return
	}
	var e govcd.EdgeGateway

	var dnatName string = baseDnatName + "_ForBackCompabilit"
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"ExternalIp":  testConfig.Networking.ExternalIp,
		"DnatName":    dnatName,
		"Tags":        "gateway",
	}

	configText := templateFill(testAccCheckVcdDnat_ForBackCompability, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdDNATDestroyForBackCompability,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdDNATExistsForBackCompability("vcd_dnat."+dnatName, &e),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "external_ip", testConfig.Networking.ExternalIp),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "port", "7777"),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "internal_ip", "10.10.102.60"),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "translated_port", "77"),
				),
			},
		},
	})
}

func testAccCheckVcdDNATExistsForBackCompability(n string, gateway *govcd.EdgeGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no DNAT ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		gatewayName := rs.Primary.Attributes["edge_gateway"]

		edgeGateway, err := conn.GetEdgeGateway(testConfig.VCD.Org, testConfig.VCD.Vdc, gatewayName)
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, err)
		}

		var found bool
		for _, v := range edgeGateway.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule {
			if v.RuleType == "DNAT" &&
				v.GatewayNatRule.OriginalIP == testConfig.Networking.ExternalIp &&
				v.GatewayNatRule.OriginalPort == "7777" &&
				v.GatewayNatRule.TranslatedIP == "10.10.102.60" &&
				v.GatewayNatRule.TranslatedPort == "77" {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("DNAT rule was not found")
		}

		gateway = edgeGateway

		return nil
	}
}

func testAccCheckVcdDNATDestroyForBackCompability(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_dnat" {
			continue
		}

		gatewayName := rs.Primary.Attributes["edge_gateway"]
		edgeGateway, err := conn.GetEdgeGateway(testConfig.VCD.Org, testConfig.VCD.Vdc, gatewayName)

		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, err)
		}

		var found bool
		for _, v := range edgeGateway.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule {
			if v.RuleType == "DNAT" &&
				v.GatewayNatRule.OriginalIP == testConfig.Networking.ExternalIp &&
				v.GatewayNatRule.OriginalPort == "7777" &&
				v.GatewayNatRule.TranslatedIP == "10.10.102.60" &&
				v.GatewayNatRule.TranslatedPort == "77" {
				found = true
			}
		}

		if found {
			return fmt.Errorf("DNAT rule still exists.")
		}
	}

	return nil
}

const testAccCheckVcdDnatWithOrgNetw = `
resource "vcd_network_routed" "{{.OrgVdcNetworkName}}" {
  name         = "{{.OrgVdcNetworkName}}"
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "{{.Gateway}}"

  dhcp_pool {
    start_address = "{{.StartIpAddress}}"
    end_address   = "{{.EndIpAddress}}"
  }
}
resource "vcd_dnat" "{{.DnatName}}" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  network_name = "{{.OrgVdcNetworkName}}"
  network_type = "org"
  edge_gateway = "{{.EdgeGateway}}"
  external_ip  = "{{.ExternalIp}}"
  port         = 7777
  internal_ip  = "10.10.102.60"
  description  = "{{.Description}}"
  depends_on   = ["vcd_network_routed.{{.OrgVdcNetworkName}}"]
}
`

const testAccCheckVcdDnat_ForBackCompability = `
resource "vcd_dnat" "{{.DnatName}}" {
  org             = "{{.Org}}"
  vdc             = "{{.Vdc}}"
  edge_gateway    = "{{.EdgeGateway}}"
  external_ip     = "{{.ExternalIp}}"
  port            = 7777
  internal_ip     = "10.10.102.60"
  translated_port = 77
}
`
const testAccCheckVcdDnatWithExtNetw = `
data "vcd_external_network" "{{.ExternalNetworkName}}" {
  name = "{{.ExternalNetworkName}}"
}

resource "vcd_dnat" "{{.DnatName}}" {
  org             = "{{.Org}}"
  vdc             = "{{.Vdc}}"
  network_name    = data.vcd_external_network.{{.ExternalNetworkName}}.name
  network_type    = "ext"
  edge_gateway    = "{{.EdgeGateway}}"
  external_ip     = data.vcd_external_network.{{.ExternalNetworkName}}.ip_scope[0].static_ip_pool[0].start_address
  port            = 7777
  protocol        = "tcp"
  internal_ip     = "10.10.102.60"
  translated_port = 77
  description     = "{{.Description}}"
}
`
const testAccCheckVcdDnatWithExtNetwUpdate = `
# skip-binary-test: only for updates
resource "vcd_dnat" "{{.DnatName}}" {
  org             = "{{.Org}}"
  vdc             = "{{.Vdc}}"
  network_name    = "{{.ExternalNetworkName}}"
  network_type    = "ext"
  edge_gateway    = "{{.EdgeGateway}}"
  external_ip     = "{{.ExternalIp}}"
  protocol        = "udp"
  port            = 8888
  internal_ip     = "10.10.102.80"
  translated_port = 88
  description     = "{{.Description}}"
}
`

// testDeleteExistingCatalogMedia deletes catalog with name from test or returns a failure
func testDeleteExistingDnatRule(t *testing.T, dnatRuleId string) func() {
	return func() {
		vcdClient := createTemporaryVCDConnection()
		_, vdc, err := vcdClient.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
		if err != nil {
			t.Errorf(err.Error())
		}

		egw, err := vdc.GetEdgeGatewayByNameOrId(testConfig.Networking.EdgeGateway, false)
		if err != nil {
			t.Errorf(err.Error())
		}

		natRule, err := egw.GetNatRule(dnatRuleId)
		if err != nil || natRule == nil {
			t.Error(err.Error())
		}

		err = egw.RemoveNATRule(natRule.ID)
		if err != nil {
			t.Error(err.Error())
		}
	}
}
