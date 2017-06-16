package vcd

import (
	"fmt"

	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	govcd "github.com/ukcloud/govcloudair"
)

func TestAccVcdVApp_PowerOff(t *testing.T) {
	var vapp govcd.VApp

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdVAppDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckVcdVApp_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppExists("vcd_vapp.foobar", &vapp),
					resource.TestCheckResourceAttr(
						"vcd_vapp.foobar", "name", "foobar"),
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

		resp, err := conn.OrgVdc.FindVAppByName(rs.Primary.ID)
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

		_, err := conn.OrgVdc.FindVAppByName(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("VPCs still exist")
		}

		return nil
	}

	return nil
}

const testAccCheckVcdVApp_basic = `
resource "vcd_vapp" "foobar" {
  name = "foobar"
}
`
