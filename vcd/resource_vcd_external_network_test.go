// +build functional network extnetwork ALL

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

var TestAccVcdExternalNetwork = "TestAccVcdExternalNetworkBasic"
var externalNetwork govcd.ExternalNetwork

func TestAccVcdExternalNetworkBasic(t *testing.T) {

	if !usingSysAdmin() {
		t.Skip("TestAccVcdExternalNetworkBasic requires system admin privileges")
		return
	}

	startAddress := "192.168.30.51"
	endAddress := "192.168.30.62"
	description := "Test External Network"
	gateway := "192.168.30.49"
	netmask := "255.255.255.240"
	dns1 := "192.168.0.164"
	dns2 := "192.168.0.196"
	var params = StringMap{
		"ExternalNetworkName": TestAccVcdExternalNetwork,
		"Type":                testConfig.Networking.ExternalNetworkPortGroupType,
		"PortGroup":           testConfig.Networking.ExternalNetworkPortGroup,
		"Vcenter":             testConfig.Networking.Vcenter,
		"StartAddress":        startAddress,
		"EndAddress":          endAddress,
		"Description":         description,
		"Gateway":             gateway,
		"Netmask":             netmask,
		"Dns1":                dns1,
		"Dns2":                dns2,
		"Tags":                "network extnetwork",
	}

	configText := templateFill(testAccCheckVcdExternalNetwork_basic, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resourceName := "vcd_external_network." + TestAccVcdExternalNetwork
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
						resourceName, "name", TestAccVcdExternalNetwork),
					resource.TestCheckResourceAttr(
						resourceName, "vsphere_network.0.vcenter", testConfig.Networking.Vcenter),
					resource.TestCheckResourceAttr(
						resourceName, "vsphere_network.0.name", testConfig.Networking.ExternalNetworkPortGroup),
					resource.TestCheckResourceAttr(
						resourceName, "vsphere_network.0.type", testConfig.Networking.ExternalNetworkPortGroupType),
					resource.TestCheckResourceAttr(
						resourceName, "ip_scope.0.gateway", gateway),
					resource.TestCheckResourceAttr(
						resourceName, "ip_scope.0.netmask", netmask),
					resource.TestCheckResourceAttr(
						resourceName, "ip_scope.0.dns1", dns1),
					resource.TestCheckResourceAttr(
						resourceName, "ip_scope.0.dns2", dns2),
					resource.TestCheckResourceAttr(
						resourceName, "ip_scope.0.static_ip_pool.0.start_address", startAddress),
					resource.TestCheckResourceAttr(
						resourceName, "ip_scope.0.static_ip_pool.0.end_address", endAddress),
					resource.TestCheckResourceAttr(
						resourceName, "description", description),
					resource.TestCheckResourceAttr(
						resourceName, "retain_net_info_across_deployments", "false"),
				),
			},
			resource.TestStep{
				ResourceName:      resourceName + "-import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdTopHierarchy(TestAccVcdExternalNetwork),
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
			return fmt.Errorf("no external network ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)
		newExternalNetwork, err := conn.GetExternalNetworkByNameOrId(rs.Primary.ID)
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
		_, err := conn.GetExternalNetworkByNameOrId(rs.Primary.ID)
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
  description = "{{.Description}}"

  vsphere_network {
    vcenter = "{{.Vcenter}}"
    name    = "{{.PortGroup}}"
    type    = "{{.Type}}"
  }

  ip_scope {
    gateway      = "{{.Gateway}}"
    netmask      = "{{.Netmask}}"
    dns1         = "{{.Dns1}}"
    dns2         = "{{.Dns2}}"
    dns_suffix   = "company.biz"

    static_ip_pool {
      start_address = "{{.StartAddress}}"
      end_address   = "{{.EndAddress}}"
    }
  }

  retain_net_info_across_deployments = "false"
}
`
