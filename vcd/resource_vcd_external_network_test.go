// +build functional network extnetwork ALL

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

var TestAccVcdExternalNetwork = "TestAccVcdExternalNetworkBasic"
var externalNetwork govcd.ExternalNetwork

func TestAccVcdExternalNetworkBasic(t *testing.T) {

	if !usingSysAdmin() {
		t.Skip("TestAccVcdExternalNetworkBasic requires system admin privileges")
		return
	}

	var params = StringMap{
		"ExternalNetworkName": TestAccVcdExternalNetwork,
		"Type":                testConfig.Networking.ExternalNetworkPortGroupType,
		"PortGroup":           testConfig.Networking.ExternalNetworkPortGroup,
		"Vcenter":             testConfig.Networking.Vcenter,
		"Tags":                "network extnetwork",
	}

	configText := templateFill(testAccCheckVcdExternalNetwork_basic, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
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
					resource.TestCheckResourceAttr(
						"vcd_external_network."+TestAccVcdExternalNetwork, "retain_net_info_across_deployments", "false"),
				),
			},
		},
	})
}

func testAccCheckVcdExternalNetworkExists(name string, externalNetwork *govcd.ExternalNetwork) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no external network Id is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)
		newExternalNetwork, err := govcd.GetExternalNetwork(conn.VCDClient, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("external network %s does not exist (%#v)", rs.Primary.ID, err)
		}

		externalNetwork = newExternalNetwork
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
		if err == nil {
			return fmt.Errorf("external network %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func init() {
	testingTags["extnetwork"] = "resource_vcd_external_network_test.go"
}

const testAccCheckVcdExternalNetwork_basic = `
resource "vcd_external_network" "{{.ExternalNetworkName}}" {
  name        = "{{.ExternalNetworkName}}"
  description = "Test External Network"

  vsphere_network {
    vcenter = "{{.Vcenter}}"
    name    = "{{.PortGroup}}"
    type    = "{{.Type}}"
  }

  ip_scope {
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

  retain_net_info_across_deployments = "false"
}
`
