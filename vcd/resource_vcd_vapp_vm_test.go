package vcd

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	govcd "github.com/vmware/go-vcloud-director/govcd"
)

func TestAccVcdVAppVm_Basic(t *testing.T) {
	var vapp govcd.VApp
	var vm govcd.VM

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdVAppVmDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckVcdVAppVm_basic, testOrg, testVDC, os.Getenv("VCD_EDGE_GATEWAY"), testOrg, testVDC, testOrg, testVDC),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppVmExists("vcd_vapp_vm.moo", &vapp, &vm),
					resource.TestCheckResourceAttr(
						"vcd_vapp_vm.moo", "name", "moo"),
					resource.TestCheckResourceAttr(
						"vcd_vapp_vm.moo", "ip", "10.10.102.161"),
					resource.TestCheckResourceAttr(
						"vcd_vapp_vm.moo", "power_on", "true"),
				),
			},
		},
	})
}

func testAccCheckVcdVAppVmExists(n string, vapp *govcd.VApp, vm *govcd.VM) resource.TestCheckFunc {
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
		vapp, err := vdc.FindVAppByName("foobar")

		resp, err := vdc.FindVMByName(vapp, "moo")

		if err != nil {
			return err
		}

		*vm = resp

		return nil
	}
}

func testAccCheckVcdVAppVmDestroy(s *terraform.State) error {
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
		_, err = vdc.FindVAppByName("foobar")

		if err == nil {
			return fmt.Errorf("VPCs still exist")
		}

		return nil
	}

	return nil
}

const testAccCheckVcdVAppVm_basic = `
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

resource "vcd_vapp" "foobar" {
  name          = "foobar"
  org = "%s"
  vdc = "%s"
  template_name = "Skyscape_CentOS_6_4_x64_50GB_Small_v1.0.1"
  catalog_name  = "Skyscape Catalogue"
  network_name  = "${vcd_network.foonet.name}"
  memory        = 1024
  cpus          = 1
  ip            = "10.10.102.160"
}

resource "vcd_vapp_vm" "moo" {
  org = "%s"
  vdc = "%s"
  vapp_name     = "${vcd_vapp.foobar.name}"
  name          = "moo"
  catalog_name  = "Skyscape Catalogue"
  template_name = "Skyscape_CentOS_6_4_x64_50GB_Small_v1.0.1"
  memory        = 1024
  cpus          = 1
  ip            = "10.10.102.161"
}
`
