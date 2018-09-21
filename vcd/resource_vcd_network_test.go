package vcd

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/go-vcloud-director/govcd"
)

func TestAccVcdNetwork_Basic(t *testing.T) {
	var network govcd.OrgVDCNetwork
	generatedHrefRegexp := regexp.MustCompile("^https://")

	networkName := "TestAccVcdNetwork"
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"NetworkName": networkName,
	}

	configText := templateFill(testAccCheckVcdNetwork_basic, params)
	if os.Getenv("GOVCD_DEBUG") != "" {
		log.Printf("#[DEBUG] CONFIGURATION: %s", configText)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdNetworkDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdNetworkExists("vcd_network."+networkName, &network),
					testAccCheckVcdNetworkAttributes(networkName, &network),
					resource.TestCheckResourceAttr(
						"vcd_network."+networkName, "name", networkName),
					resource.TestCheckResourceAttr(
						"vcd_network."+networkName, "static_ip_pool.#", "1"),
					resource.TestCheckResourceAttr(
						"vcd_network."+networkName, "gateway", "10.10.102.1"),
					resource.TestMatchResourceAttr(
						"vcd_network."+networkName, "href", generatedHrefRegexp),
				),
			},
		},
	})
}

func testAccCheckVcdNetworkExists(n string, network *govcd.OrgVDCNetwork) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VAPP ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)
		org, err := govcd.GetOrgByName(conn.VCDClient, testConfig.VCD.Org)
		if err != nil || org == (govcd.Org{}) {
			return fmt.Errorf("Could not find test Org")
		}
		vdc, err := org.GetVdcByName(testConfig.VCD.Vdc)
		if err != nil || vdc == (govcd.Vdc{}) {
			return fmt.Errorf("Could not find test Vdc")
		}
		resp, err := vdc.FindVDCNetwork(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Network does not exist.")
		}

		*network = resp

		return nil
	}
}

func testAccCheckVcdNetworkDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_network" {
			continue
		}
		org, err := govcd.GetOrgByName(conn.VCDClient, testConfig.VCD.Org)
		if err != nil || org == (govcd.Org{}) {
			return fmt.Errorf("Could not find test Org")
		}
		vdc, err := org.GetVdcByName(testConfig.VCD.Vdc)
		if err != nil || vdc == (govcd.Vdc{}) {
			return fmt.Errorf("Could not find test Vdc")
		}
		_, err = vdc.FindVDCNetwork(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("Network still exists.")
		}

		return nil
	}

	return nil
}

func testAccCheckVcdNetworkAttributes(name string, network *govcd.OrgVDCNetwork) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if network.OrgVDCNetwork.Name != name {
			return fmt.Errorf("Bad name: %s", network.OrgVDCNetwork.Name)
		}

		return nil
	}
}

const testAccCheckVcdNetwork_basic = `
resource "vcd_network" "{{.NetworkName}}" {
	name = "{{.NetworkName}}"
	org = "{{.Org}}"
	vdc = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	gateway = "10.10.102.1"
	static_ip_pool {
		start_address = "10.10.102.2"
		end_address = "10.10.102.254"
	}
}
`
