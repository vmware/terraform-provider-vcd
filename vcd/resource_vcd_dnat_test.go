package vcd

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	govcd "github.com/vmware/go-vcloud-director/govcd"
)

var baseDnatName string = "TestAccVcdDNAT"

func TestAccVcdDNAT_Basic(t *testing.T) {
	if testConfig.Networking.ExternalIp == "" {
		t.Skip("Variable networking.externalIp must be set to run DNAT tests")
		return
	}
	var e govcd.EdgeGateway

	var dnatName string = baseDnatName + "_Basic"
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"ExternalIp":  testConfig.Networking.ExternalIp,
		"DnatName":    dnatName,
	}
	configText := templateFill(testAccCheckVcdDnat_basic, params)
	if os.Getenv("GOVCD_DEBUG") != "" {
		log.Printf("#[DEBUG] CONFIGURATION: %s", configText)
	}
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
						"vcd_dnat."+dnatName, "external_ip", testConfig.Networking.ExternalIp),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "port", "7777"),
					resource.TestCheckResourceAttr(
						"vcd_dnat."+dnatName, "internal_ip", "10.10.102.60"),
				),
			},
		},
	})
}

func TestAccVcdDNAT_tlate(t *testing.T) {
	if testConfig.Networking.ExternalIp == "" {
		t.Skip("Variable networking.externalIp must be set to run DNAT tests")
		return
	}
	var e govcd.EdgeGateway

	var dnatName string = baseDnatName + "_tlate"
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"ExternalIp":  testConfig.Networking.ExternalIp,
		"DnatName":    dnatName,
	}

	configText := templateFill(testAccCheckVcdDnat_tlate, params)
	if os.Getenv("GOVCD_DEBUG") != "" {
		log.Printf("#[DEBUG] CONFIGURATION: %s", configText)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdDNATDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdDNATtlateExists("vcd_dnat."+dnatName, &e),
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

func testAccCheckVcdDNATExists(n string, gateway *govcd.EdgeGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DNAT ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		gatewayName := rs.Primary.Attributes["edge_gateway"]
		org, err := govcd.GetOrgByName(conn.VCDClient, testConfig.VCD.Org)
		if err != nil && org == (govcd.Org{}) {
			return fmt.Errorf("Could not find test Org")
		}
		vdc, err := org.GetVdcByName(testConfig.VCD.Vdc)
		if err != nil || vdc == (govcd.Vdc{}) {
			return fmt.Errorf("Could not find test Vdc %s", testConfig.VCD.Vdc)
		}
		edgeGateway, err := vdc.FindEdgeGateway(gatewayName)

		if err != nil {
			return fmt.Errorf("Could not find edge gateway")
		}

		var found bool
		for _, v := range edgeGateway.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule {
			if v.RuleType == "DNAT" &&
				v.GatewayNatRule.OriginalIP == testConfig.Networking.ExternalIp &&
				v.GatewayNatRule.OriginalPort == "7777" &&
				v.GatewayNatRule.TranslatedIP == "10.10.102.60" {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("DNAT rule was not found")
		}

		*gateway = edgeGateway

		return nil
	}
}

func testAccCheckVcdDNATtlateExists(n string, gateway *govcd.EdgeGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DNAT ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		gatewayName := rs.Primary.Attributes["edge_gateway"]
		// fmt.Printf(testConfig.VCD.Org)
		org, err := govcd.GetOrgByName(conn.VCDClient, testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf("Could not find test Org")
		}
		vdc, err := org.GetVdcByName(testConfig.VCD.Vdc)
		if err != nil {
			return fmt.Errorf("Could not find test Vdc")
		}
		edgeGateway, err := vdc.FindEdgeGateway(gatewayName)

		if err != nil {
			return fmt.Errorf("Could not find edge gateway")
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

		*gateway = edgeGateway

		return nil
	}
}

func testAccCheckVcdDNATDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_dnat" {
			continue
		}
		org, err := govcd.GetOrgByName(conn.VCDClient, testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf("Could not find test Org")
		}
		vdc, err := org.GetVdcByName(testConfig.VCD.Vdc)
		if err != nil {
			return fmt.Errorf("Could not find test Vdc")
		}
		gatewayName := rs.Primary.Attributes["edge_gateway"]
		edgeGateway, err := vdc.FindEdgeGateway(gatewayName)

		if err != nil {
			return fmt.Errorf("Could not find edge gateway")
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

const testAccCheckVcdDnat_basic = `
resource "vcd_dnat" "{{.DnatName}}" {
	org = "{{.Org}}"
	vdc = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	external_ip = "{{.ExternalIp}}"
	port = 7777
	internal_ip = "10.10.102.60"
}
`
const testAccCheckVcdDnat_tlate = `
resource "vcd_dnat" "{{.DnatName}}" {
	org = "{{.Org}}"
	vdc = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	external_ip = "{{.ExternalIp}}"
	port = 7777
	internal_ip = "10.10.102.60"
	translated_port = 77
}
`
