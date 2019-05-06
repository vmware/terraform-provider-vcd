package vcd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

var TestAccVcdExternalNetwork = "TestAccVcdExternalNetworkBasic"
var externalNetwork govcd.ExternalNetwork

func TestAccVcdExternalNetworkBasic(t *testing.T) {

	var params = StringMap{
		"ExternalNetworkName": TestAccVcdExternalNetwork,
		"FenceMode":           "isolated",
		"Type":                testConfig.Networking.ExternalNetworkPortGroupType,
		"PortGroup":           testConfig.Networking.ExternalNetworkPortGroup,
		"Vcenter":             testConfig.Networking.Vcenter,
	}

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	/*	if !usingSysAdmin() {
		t.Skip("TestAccVcdExternalNetworkBasic requires system admin privileges")
		return
	}*/

	configText := templateFill(testAccCheckVcdExternalNetwork_basic, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckExternalNetworkDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdExternalNetworkExists("vcd_external_network."+TestAccVcdExternalNetwork, &externalNetwork),
					resource.TestCheckResourceAttr(
						"vcd_external_network."+TestAccVcdExternalNetwork, "name", TestAccVcdExternalNetwork),
					resource.TestCheckResourceAttr(
						"vcd_external_network."+TestAccVcdExternalNetwork, "description", "Test External Network"),
					//resource.TestCheckResourceAttr(
					//"vcd_external_network."+TestAccVcdExternalNetwork, "ip_scope[0].is_inherited", "false"),
					resource.TestCheckResourceAttr(
						"vcd_external_network."+TestAccVcdExternalNetwork, "fence_mode", "isolated"),
					resource.TestCheckResourceAttr(
						"vcd_external_network."+TestAccVcdExternalNetwork, "retain_net_info_across_deployments", "false"),
				),
			},
		},
	})
}

func testAccCheckVcdExternalNetworkExists(name string, vdc *govcd.ExternalNetwork) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no external network ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)
		externalNetwork, err := govcd.GetExternalNetwork(conn.VCDClient, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("vdc %s does not exist (%#v)", rs.Primary.ID, externalNetwork.ExternalNetwork)
		}

		return nil
	}
}

func testAccCheckExternalNetworkDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_external_network" && rs.Primary.Attributes["name"] != TestAccVcdExternalNetwork {
			continue
		}

		conn := testAccProvider.Meta().(*VCDClient)
		_, err := govcd.GetExternalNetwork(conn.VCDClient, rs.Primary.ID)
		if err != nil && !strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("external network %s still exists or other error: %#v", rs.Primary.ID, err)
		}
	}

	return nil
}

const testAccCheckVcdExternalNetwork_basic = `
resource "vcd_external_network" "{{.ExternalNetworkName}}" {
  name        = "{{.ExternalNetworkName}}"
  description = "Test External Network"

  vsphere_networks {
    vcenter         = "{{.Vcenter}}"
    vsphere_network = "{{.PortGroup}}"
    type            = "{{.Type}}"
  }

  ip_scope {
    is_inherited = "false"
    gateway      = "192.168.30.49"
    netmask      = "255.255.255.240"
    dns1         = "192.168.0.164"
    dns2         = "192.168.0.196"
    dns_suffix   = "company.biz"

    static_ip_pool {
      start_address = "192.168.30.51"
      end_address   = "192.168.30.62"
    }
  }

  fence_mode                         = "{{.FenceMode}}"
  retain_net_info_across_deployments = "false"
}
`
