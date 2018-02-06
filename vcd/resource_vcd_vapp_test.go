package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func testAccCheckVcdVAppDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_vm" {
			continue
		}

		_, err := conn.OrgVdc.GetVAppByHREF(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("VPCs still exist")
		}

		return nil
	}

	return nil
}

func TestAccVcdVApp_Basic(t *testing.T) {
	// var vapp govcd.VApp
	// var vm govcd.VM

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdVAppDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckVcdVApp_basic),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "name", "kdalby-dev-vapp-test-only-vapp-basic"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.0", "FCI-IRT_ISN6_ORG-SRV"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.1", "FCI-IRT_ISN6_ORG-MGT"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.0.name", "test"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.1.name", "test2"),
				),
			},
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckVcdVApp_basic2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "name", "kdalby-dev-vapp-test-only-vapp-basic"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.0", "FCI-IRT_ISN6_ORG-SRV"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.1", "FCI-IRT_ISN6_ORG-MGT"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.0.name", "test"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.1.name", "test2"),
				),
			},
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckVcdVApp_basic3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "name", "kdalby-dev-vapp-test-only-vapp-basic"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.0", "FCI-IRT_ISN6_ORG-SRV"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.1", "FCI-IRT_ISN6_ORG-MGT"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.0.name", "test"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.1.name", "test2"),
				),
			},
		},
	})

}

const testAccCheckVcdVApp_basic = `
resource "vcd_vapp" "test-vapp" {
  name     = "kdalby-dev-vapp-test-only-vapp-basic"

  organization_network = [
    "FCI-IRT_ISN6_ORG-SRV",
    "FCI-IRT_ISN6_ORG-MGT",
  ]

  vapp_network {
     name = "test"
     description = ""
     gateway = "192.168.2.1"
     netmask = "255.255.255.0"
     dns1 = "8.8.8.8"
     dns2 = "8.8.4.4"
     start = "192.168.2.100"
     end = "192.168.2.199"
     nat = false
     parent = "FCI-IRT_ISN6_ORG-SRV"
     dhcp = false
  }
  vapp_network {
     name = "test2"
     description = ""
     gateway = "192.168.3.1"
     netmask = "255.255.255.0"
     dns1 = "8.8.8.8"
     dns2 = "8.8.4.4"
     start = "192.168.3.100"
     end = "192.168.3.199"
     nat = true
     parent = "FCI-IRT_ISN6_ORG-SRV"
     dhcp = true
     dhcp_start = "192.168.3.200"
     dhcp_end = "192.168.3.249"
  }

}
`

const testAccCheckVcdVApp_basic2 = `
resource "vcd_vapp" "test-vapp" {
  name     = "kdalby-dev-vapp-test-only-vapp-basic"

  organization_network = [
    "FCI-IRT_ISN6_ORG-SRV",
    "FCI-IRT_ISN6_ORG-MGT",
  ]

  vapp_network {
     name = "test"
     description = ""
     gateway = "192.168.2.1"
     netmask = "255.255.255.0"
     dns1 = "8.8.8.8"
     dns2 = "8.8.4.4"
     start = "192.168.2.100"
     end = "192.168.2.199"
     nat = false
     parent = "FCI-IRT_ISN6_ORG-SRV"
     dhcp = false
  }
  vapp_network {
     name = "test2"
     description = ""
     gateway = "192.168.3.1"
     netmask = "255.255.255.0"
     dns1 = "8.8.8.8"
     dns2 = "8.8.4.4"
     start = "192.168.3.100"
     end = "192.168.3.199"
     nat = true
     parent = "FCI-IRT_ISN6_ORG-SRV"
     dhcp = true
     dhcp_start = "192.168.3.200"
     dhcp_end = "192.168.3.249"
  }

}
`

const testAccCheckVcdVApp_basic3 = `
resource "vcd_vapp" "test-vapp" {
  name     = "kdalby-dev-vapp-test-only-vapp-basic"

  organization_network = [
    "FCI-IRT_ISN6_ORG-SRV",
    "FCI-IRT_ISN6_ORG-MGT",
  ]

  vapp_network {
     name = "test"
     description = ""
     gateway = "192.168.2.1"
     netmask = "255.255.255.0"
     dns1 = "8.8.8.8"
     dns2 = "8.8.4.4"
     start = "192.168.2.100"
     end = "192.168.2.199"
     nat = false
     parent = "FCI-IRT_ISN6_ORG-SRV"
     dhcp = false
  }
  vapp_network {
     name = "test2"
     description = ""
     gateway = "192.168.3.1"
     netmask = "255.255.255.0"
     dns1 = "8.8.8.8"
     dns2 = "8.8.4.4"
     start = "192.168.3.100"
     end = "192.168.3.199"
     nat = true
     parent = "FCI-IRT_ISN6_ORG-SRV"
     dhcp = true
     dhcp_start = "192.168.3.200"
     dhcp_end = "192.168.3.249"
  }

}
`

func TestAccVcdVApp_Complex(t *testing.T) {
	// var vapp govcd.VApp
	// var vm govcd.VM

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdVAppDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckVcdVApp_complex),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "name", "kdalby-dev-vapp-test-only-vapp-complex"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.0", "FCI-IRT_ISN6_ORG-SRV"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.1", "FCI-IRT_ISN6_ORG-MGT"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.0.name", "test"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.1.name", "test2"),
				),
			},
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckVcdVApp_complex2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "name", "kdalby-dev-vapp-test-only-vapp-complex"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.0", "FCI-IRT_ISN6_ORG-MGT"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.1", "FCI-IRT_ISN6_ORG-SRV"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.0.name", "test2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.1.name", "test"),
				),
			},
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckVcdVApp_complex3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "name", "kdalby-dev-vapp-test-only-vapp-complex"),
					resource.TestCheckNoResourceAttr(
						"vcd_vapp.test-vapp", "organization_network"),
					resource.TestCheckNoResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network"),
				),
			},
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckVcdVApp_complex4),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "name", "kdalby-dev-vapp-test-only-vapp-complex"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.0", "FCI-IRT_ISN6_ORG-SRV"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.1", "FCI-IRT_ISN6_ORG-MGT"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.0.name", "test"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.1.name", "test2"),
				),
			},
		},
	})

}

const testAccCheckVcdVApp_complex = `
resource "vcd_vapp" "test-vapp" {
  name     = "kdalby-dev-vapp-test-only-vapp-complex"

  organization_network = [
    "FCI-IRT_ISN6_ORG-SRV",
    "FCI-IRT_ISN6_ORG-MGT",
  ]

  vapp_network {
     name = "test"
     description = ""
     gateway = "192.168.2.1"
     netmask = "255.255.255.0"
     dns1 = "8.8.8.8"
     dns2 = "8.8.4.4"
     start = "192.168.2.100"
     end = "192.168.2.199"
     nat = false
     parent = "FCI-IRT_ISN6_ORG-SRV"
     dhcp = false
  }
  vapp_network {
     name = "test2"
     description = ""
     gateway = "192.168.3.1"
     netmask = "255.255.255.0"
     dns1 = "8.8.8.8"
     dns2 = "8.8.4.4"
     start = "192.168.3.100"
     end = "192.168.3.199"
     nat = true
     parent = "FCI-IRT_ISN6_ORG-SRV"
     dhcp = true
     dhcp_start = "192.168.3.200"
     dhcp_end = "192.168.3.249"
  }

}
`

const testAccCheckVcdVApp_complex2 = `
resource "vcd_vapp" "test-vapp" {
  name     = "kdalby-dev-vapp-test-only-vapp-complex"

  organization_network = [
    "FCI-IRT_ISN6_ORG-MGT",
    "FCI-IRT_ISN6_ORG-SRV",
  ]

  vapp_network {
     name = "test2"
     description = ""
     gateway = "192.168.3.1"
     netmask = "255.255.255.0"
     dns1 = "8.8.8.8"
     dns2 = "8.8.4.4"
     start = "192.168.3.100"
     end = "192.168.3.199"
     nat = true
     parent = "FCI-IRT_ISN6_ORG-SRV"
     dhcp = true
     dhcp_start = "192.168.3.200"
     dhcp_end = "192.168.3.249"
  }
  vapp_network {
     name = "test"
     description = ""
     gateway = "192.168.2.1"
     netmask = "255.255.255.0"
     dns1 = "8.8.8.8"
     dns2 = "8.8.4.4"
     start = "192.168.2.100"
     end = "192.168.2.199"
     nat = false
     parent = "FCI-IRT_ISN6_ORG-SRV"
     dhcp = false
  }

}
`

const testAccCheckVcdVApp_complex3 = `
resource "vcd_vapp" "test-vapp" {
  name     = "kdalby-dev-vapp-test-only-vapp-complex"


}
`

const testAccCheckVcdVApp_complex4 = `
resource "vcd_vapp" "test-vapp" {
  name     = "kdalby-dev-vapp-test-only-vapp-complex"

  organization_network = [
    "FCI-IRT_ISN6_ORG-SRV",
    "FCI-IRT_ISN6_ORG-MGT",
  ]

  vapp_network {
     name = "test"
     description = ""
     gateway = "192.168.2.1"
     netmask = "255.255.255.0"
     dns1 = "8.8.8.8"
     dns2 = "8.8.4.4"
     start = "192.168.2.100"
     end = "192.168.2.199"
     nat = false
     parent = "FCI-IRT_ISN6_ORG-SRV"
     dhcp = false
  }
  vapp_network {
     name = "test2"
     description = ""
     gateway = "192.168.3.1"
     netmask = "255.255.255.0"
     dns1 = "8.8.8.8"
     dns2 = "8.8.4.4"
     start = "192.168.3.100"
     end = "192.168.3.199"
     nat = true
     parent = "FCI-IRT_ISN6_ORG-SRV"
     dhcp = true
     dhcp_start = "192.168.3.200"
     dhcp_end = "192.168.3.249"
  }

}
`
