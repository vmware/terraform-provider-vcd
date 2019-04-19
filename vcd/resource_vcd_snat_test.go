package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

var orgVdcNetworkNameForSnat = "TestAccVcdDNAT_BasicNetworkForSnat"
var startIpAddress = "10.10.102.51"

func TestAccVcdSNAT_Basic(t *testing.T) {
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	if testConfig.Networking.ExternalIp == "" {
		t.Skip("Variable networking.extarnalIp must be set to run SNAT tests")
		return
	}

	var e govcd.EdgeGateway

	snatName := "TestAccVcdSNAT"
	var params = StringMap{
		"Org":               testConfig.VCD.Org,
		"Vdc":               testConfig.VCD.Vdc,
		"EdgeGateway":       testConfig.Networking.EdgeGateway,
		"ExternalIp":        testConfig.Networking.ExternalIp,
		"ExternalNetwork":   testConfig.Networking.ExternalNetwork,
		"SnatName":          snatName,
		"OrgVdcNetworkName": orgVdcNetworkNameForSnat,
		"Gateway":           "10.10.102.1",
		"StartIpAddress":    startIpAddress,
		"EndIpAddress":      "10.10.102.100",
	}

	configText := templateFill(testAccCheckVcdSnat_basic, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdSNATDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config:      configText,
				ExpectError: regexp.MustCompile(`After applying this step and refreshing, the plan was not empty:`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdSNATExists("vcd_snat."+snatName, &e),
					resource.TestCheckResourceAttr(
						"vcd_snat."+snatName, "network_name", orgVdcNetworkNameForSnat),
					resource.TestCheckResourceAttr(
						"vcd_snat."+snatName, "external_ip", testConfig.Networking.ExternalIp),
					resource.TestCheckResourceAttr(
						"vcd_snat."+snatName, "internal_ip", startIpAddress),
				),
			},
		},
	})
}

func testAccCheckVcdSNATExists(n string, gateway *govcd.EdgeGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no SNAT ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		gatewayName := rs.Primary.Attributes["edge_gateway"]

		edgeGateway, err := conn.GetEdgeGateway(testConfig.VCD.Org, testConfig.VCD.Vdc, gatewayName)

		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.VCD.Vdc, testConfig.VCD.Org, err)
		}

		var found bool
		for _, v := range edgeGateway.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule {
			if v.RuleType == "SNAT" && v.GatewayNatRule.Interface.Name == orgVdcNetworkNameForSnat &&
				v.GatewayNatRule.OriginalIP == testConfig.Networking.ExternalIp &&
				v.GatewayNatRule.TranslatedIP == startIpAddress {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("SNAT rule was not found")
		}

		*gateway = edgeGateway

		return nil
	}
}

func testAccCheckVcdSNATDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_snat" {
			continue
		}

		gatewayName := rs.Primary.Attributes["edge_gateway"]

		edgeGateway, err := conn.GetEdgeGateway(testConfig.VCD.Org, testConfig.VCD.Vdc, gatewayName)
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, err)
		}

		var found bool
		for _, v := range edgeGateway.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule {
			if v.RuleType == "SNAT" && v.GatewayNatRule.Interface.Name == orgVdcNetworkNameForSnat &&
				v.GatewayNatRule.OriginalIP == startIpAddress &&
				v.GatewayNatRule.TranslatedIP == testConfig.Networking.ExternalIp {
				found = true
			}
		}

		if found {
			return fmt.Errorf("SNAT rule still exists.")
		}
	}

	return nil
}

const testAccCheckVcdSnat_basic = `
resource "vcd_network_routed" "{{.OrgVdcNetworkName}}" {
  name         = "{{.OrgVdcNetworkName}}"
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "{{.Gateway}}"

  static_ip_pool {
    start_address = "{{.StartIpAddress}}"
    end_address   = "{{.EndIpAddress}}"
  }
}


resource "vcd_snat" "{{.SnatName}}" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  network_name = "{{.OrgVdcNetworkName}}"
  external_ip  = "{{.ExternalIp}}"
  internal_ip  = "{{.StartIpAddress}}"
  depends_on      = ["vcd_network_routed.{{.OrgVdcNetworkName}}"]
}
`

func TestAccVcdSNAT_BackCompability(t *testing.T) {
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	if testConfig.Networking.ExternalIp == "" {
		t.Skip("Variable networking.extarnalIp must be set to run SNAT tests")
		return
	}

	var e govcd.EdgeGateway

	snatName := "TestAccVcdSNAT"
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"ExternalIp":  testConfig.Networking.ExternalIp,
		"SnatName":    snatName,
		"LocalIp":     testConfig.Networking.InternalIp,
	}

	configText := templateFill(testAccCheckVcdSnat_forBackCompability, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdSNATDestroyForBackCompability,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config:      configText,
				ExpectError: regexp.MustCompile(`After applying this step and refreshing, the plan was not empty:`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdSNATExistsForBackCompability("vcd_snat."+snatName, &e),
					resource.TestCheckResourceAttr(
						"vcd_snat."+snatName, "external_ip", testConfig.Networking.ExternalIp),
					resource.TestCheckResourceAttr(
						"vcd_snat."+snatName, "internal_ip", "10.10.102.0/24"),
				),
			},
		},
	})
}

func testAccCheckVcdSNATExistsForBackCompability(n string, gateway *govcd.EdgeGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no SNAT ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		gatewayName := rs.Primary.Attributes["edge_gateway"]

		edgeGateway, err := conn.GetEdgeGateway(testConfig.VCD.Org, testConfig.VCD.Vdc, gatewayName)

		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.VCD.Vdc, testConfig.VCD.Org, err)
		}

		var found bool
		for _, v := range edgeGateway.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule {
			if v.RuleType == "SNAT" &&
				v.GatewayNatRule.OriginalIP == "10.10.102.0/24" &&
				v.GatewayNatRule.OriginalPort == "" &&
				v.GatewayNatRule.TranslatedIP == testConfig.Networking.ExternalIp {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("SNAT rule was not found")
		}

		*gateway = edgeGateway

		return nil
	}
}

func testAccCheckVcdSNATDestroyForBackCompability(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_snat" {
			continue
		}

		gatewayName := rs.Primary.Attributes["edge_gateway"]

		edgeGateway, err := conn.GetEdgeGateway(testConfig.VCD.Org, testConfig.VCD.Vdc, gatewayName)
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, err)
		}

		var found bool
		for _, v := range edgeGateway.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule {
			if v.RuleType == "SNAT" &&
				v.GatewayNatRule.OriginalIP == "10.10.102.0/24" &&
				v.GatewayNatRule.OriginalPort == "" &&
				v.GatewayNatRule.TranslatedIP == testConfig.Networking.ExternalIp {
				found = true
			}
		}

		if found {
			return fmt.Errorf("SNAT rule still exists.")
		}
	}

	return nil
}

const testAccCheckVcdSnat_forBackCompability = `
resource "vcd_snat" "{{.SnatName}}" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  external_ip  = "{{.ExternalIp}}"
  internal_ip  = "10.10.102.0/24"
}
`
