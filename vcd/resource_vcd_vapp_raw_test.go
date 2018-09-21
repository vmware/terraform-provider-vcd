package vcd

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	govcd "github.com/vmware/go-vcloud-director/govcd"
	"os"
	"testing"
)

func TestAccVcdVAppRaw_Basic(t *testing.T) {
	var vapp govcd.VApp

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdVAppRawDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckVcdVAppRaw_basic, os.Getenv("VCD_EDGE_GATEWAY")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppRawExists("vcd_vapp.foobar", &vapp),
					resource.TestCheckResourceAttr(
						"vcd_vapp.foobar", "name", "foobar"),
				),
			},
		},
	})
}

func testAccCheckVcdVAppRawExists(n string, vapp *govcd.VApp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VAPP ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		resp, err := conn.OrgVdc.FindVAppByName(rs.Primary.ID)
		if err != nil {
			return err
		}

		*vapp = resp

		return nil
	}
}

func testAccCheckVcdVAppRawDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_vapp" {
			continue
		}

		_, err := conn.OrgVdc.FindVAppByName(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("VPCs still exist")
		}

		return nil
	}

	return nil
}

const testAccCheckVcdVAppRaw_basic = `
resource "vcd_network" "foonet" {
	name = "foonet"
	edge_gateway = "%s"
	gateway = "10.10.102.1"
	static_ip_pool {
		start_address = "10.10.102.2"
		end_address = "10.10.102.254"
	}
}

resource "vcd_vapp" "foobar" {
  name = "foobar"
}

resource "vcd_vapp_vm" "moo" {
  vapp_name     = "${vcd_vapp.foobar.name}"
  name          = "moo"
  catalog_name  = "Skyscape Catalogue"
  template_name = "Skyscape_CentOS_6_4_x64_50GB_Small_v1.0.1"
  memory        = 1024
  cpus          = 1

  network_name  = "${vcd_network.foonet.name}"
  ip            = "10.10.102.161"
}
`
