package vcd

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/go-vcloud-director/govcd"
)

func TestAccVcdSNAT_Basic(t *testing.T) {
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
	}

	configText := templateFill(testAccCheckVcdSnat_basic, params)
	if os.Getenv("GOVCD_DEBUG") != "" {
		log.Printf("#[DEBUG] CONFIGURATION: %s", configText)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdSNATDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdSNATExists("vcd_snat."+snatName, &e),
					resource.TestCheckResourceAttr(
						"vcd_snat."+snatName, "external_ip", testConfig.Networking.ExternalIp),
					resource.TestCheckResourceAttr(
						"vcd_snat."+snatName, "internal_ip", "10.10.102.0/24"),
				),
			},
		},
	})
}

func testAccCheckVcdSNATExists(n string, gateway *govcd.EdgeGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SNAT ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		gatewayName := rs.Primary.Attributes["edge_gateway"]
		org, err := govcd.GetOrgByName(conn.VCDClient, testConfig.VCD.Org)
		if err != nil || org == (govcd.Org{}) {
			return fmt.Errorf("Could not find test Org")
		}
		vdc, err := org.GetVdcByName(testConfig.VCD.Vdc)
		if err != nil || vdc == (govcd.Vdc{}) {
			return fmt.Errorf("Could not find test Vdc")
		}
		edgeGateway, err := vdc.FindEdgeGateway(gatewayName)

		if err != nil {
			return fmt.Errorf("Could not find edge gateway")
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

func testAccCheckVcdSNATDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_snat" {
			continue
		}

		gatewayName := rs.Primary.Attributes["edge_gateway"]
		org, err := govcd.GetOrgByName(conn.VCDClient, testConfig.VCD.Org)
		if err != nil || org == (govcd.Org{}) {
			return fmt.Errorf("Could not find test Org")
		}
		vdc, err := org.GetVdcByName(testConfig.VCD.Vdc)
		if err != nil || vdc == (govcd.Vdc{}) {
			return fmt.Errorf("Could not find test Vdc")
		}
		edgeGateway, err := vdc.FindEdgeGateway(gatewayName)

		if err != nil {
			return fmt.Errorf("Could not find edge gateway")
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

const testAccCheckVcdSnat_basic = `
resource "vcd_snat" "{{.SnatName}}" {
	org = "{{.Org}}"
	vdc = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	external_ip = "{{.ExternalIp}}"
	internal_ip = "10.10.102.0/24"
}
`
