package vcd

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	govcd "github.com/vmware/go-vcloud-director/govcd"
)

func TestAccVcdVApp_PowerOff(t *testing.T) {
	var vapp govcd.VApp

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdVAppDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckVcdVApp_basic, testOrg, testVDC, os.Getenv("VCD_EDGE_GATEWAY"), testOrg, testVDC, os.Getenv("VCD_EDGE_GATEWAY"), testOrg, testVDC, testOrg, testVDC),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppExists("vcd_vapp.foobar", &vapp),
					testAccCheckVcdVAppAttributes(&vapp),
					resource.TestCheckResourceAttr(
						"vcd_vapp.foobar", "name", "foobar"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.foobar", "ip", "10.10.102.160"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.foobar", "power_on", "true"),
				),
			},

			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckVcdVApp_basic, testOrg, testVDC, os.Getenv("VCD_EDGE_GATEWAY"), testOrg, testVDC, os.Getenv("VCD_EDGE_GATEWAY"), testOrg, testVDC, testOrg, testVDC),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vcd_vapp.foobar_allocated", "name", "foobar-allocated"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.foobar_allocated", "ip", "allocated"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.foobar_allocated", "power_on", "true"),
				),
			},

			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckVcdVApp_powerOff, testOrg, testVDC, os.Getenv("VCD_EDGE_GATEWAY"), testOrg, testVDC),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppExists("vcd_vapp.foobar", &vapp),
					testAccCheckVcdVAppAttributes_off(&vapp),
					resource.TestCheckResourceAttr(
						"vcd_vapp.foobar", "name", "foobar"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.foobar", "ip", "10.10.103.160"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.foobar", "power_on", "false"),
				),
			},
		},
	})
}

func testAccCheckVcdVAppExists(n string, vapp *govcd.VApp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VAPP ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)
		org, err := govcd.GetOrgByName(conn.VCDClient, testOrg)
		if err != nil {
			return fmt.Errorf("Could not find test Org")
		}
		vdc, err := org.GetVdcByName(testVDC)
		if err != nil {
			return fmt.Errorf("Could not find test Vdc")
		}
		resp, err := vdc.FindVAppByName(rs.Primary.ID)
		if err != nil {
			return err
		}

		*vapp = resp

		return nil
	}
}

func testAccCheckVcdVAppDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_vapp" {
			continue
		}
		org, err := govcd.GetOrgByName(conn.VCDClient, testOrg)
		if err != nil {
			return fmt.Errorf("Could not find test Org")
		}
		vdc, err := org.GetVdcByName(testVDC)
		if err != nil {
			return fmt.Errorf("Could not find test Vdc")
		}
		_, err = vdc.FindVAppByName(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("VPCs still exist")
		}

		return nil
	}

	return nil
}

func testAccCheckVcdVAppAttributes(vapp *govcd.VApp) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if vapp.VApp.Name != "foobar" {
			return fmt.Errorf("Bad name: %s", vapp.VApp.Name)
		}

		if vapp.VApp.Name != vapp.VApp.Children.VM[0].Name {
			return fmt.Errorf("VApp and VM names do not match. %s != %s",
				vapp.VApp.Name, vapp.VApp.Children.VM[0].Name)
		}

		status, _ := vapp.GetStatus()
		if status != "POWERED_ON" {
			return fmt.Errorf("VApp is not powered on")
		}

		return nil
	}
}

func testAccCheckVcdVAppAttributes_off(vapp *govcd.VApp) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if vapp.VApp.Name != "foobar" {
			return fmt.Errorf("Bad name: %s", vapp.VApp.Name)
		}

		if vapp.VApp.Name != vapp.VApp.Children.VM[0].Name {
			return fmt.Errorf("VApp and VM names do not match. %s != %s",
				vapp.VApp.Name, vapp.VApp.Children.VM[0].Name)
		}

		status, _ := vapp.GetStatus()
		if status != "POWERED_OFF" {
			return fmt.Errorf("VApp is still powered on")
		}

		return nil
	}
}

const testAccCheckVcdVApp_basic = `
resource "vcd_network" "foonet" {
	name = "foonet"
	org = "%s"
	vdc = "%s"
	edge_gateway = "%s"
	gateway = "10.10.102.1"
	static_ip_pool {
		start_address = "10.10.102.2"
		end_address = "10.10.102.254"
	}
}

resource "vcd_network" "foonet3" {
	name = "foonet3"
	org = "%s"
	vdc = "%s"
	edge_gateway = "%s"
	gateway = "10.10.202.1"
	static_ip_pool {
		start_address = "10.10.202.2"
		end_address = "10.10.202.254"
	}
}

resource "vcd_vapp" "foobar" {
  org = "%s"
  vdc = "%s"
  name          = "foobar"
  template_name = "Skyscape_CentOS_6_4_x64_50GB_Small_v1.0.1"
  catalog_name  = "Skyscape Catalogue"
  network_name  = "${vcd_network.foonet.name}"
  memory        = 1024
  cpus          = 1
  ip            = "10.10.102.160"
}

resource "vcd_vapp" "foobar_allocated" {
  org = "%s"
  vdc = "%s"
  name          = "foobar-allocated"
  template_name = "Skyscape_CentOS_6_4_x64_50GB_Small_v1.0.1"
  catalog_name  = "Skyscape Catalogue"
  network_name  = "${vcd_network.foonet3.name}"
  memory        = 1024
  cpus          = 1
  ip            = "allocated"
}
`

const testAccCheckVcdVApp_powerOff = `
resource "vcd_network" "foonet2" {
	org = "%s"
	vdc = "%s"
	name = "foonet2"
	edge_gateway = "%s"
	gateway = "10.10.103.1"
	static_ip_pool {
		start_address = "10.10.103.2"
		end_address = "10.10.103.170"
	}

	dhcp_pool {
		start_address = "10.10.103.171"
		end_address = "10.10.103.254"
	}
}

resource "vcd_vapp" "foobar" {
  org = "%s"
  vdc = "%s"
  name          = "foobar"
  template_name = "Skyscape_CentOS_6_4_x64_50GB_Small_v1.0.1"
  catalog_name  = "Skyscape Catalogue"
  network_name  = "${vcd_network.foonet2.name}"
  memory        = 1024
  cpus          = 1
  ip            = "10.10.103.160"
  power_on      = false
}
`
