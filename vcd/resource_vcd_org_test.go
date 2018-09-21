package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	govcd "github.com/vmware/go-vcloud-director/govcd"
)

func TestAccVcdOrgBasic(t *testing.T) {

	var e govcd.Org

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOrgDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckVcdOrg_basic),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdOrgExists("vcd_org.test1", &e),
					resource.TestCheckResourceAttr(
						"vcd_org.test1", "name", "test1"),
					resource.TestCheckResourceAttr(
						"vcd_org.test1", "full_name", "test1"),
					resource.TestCheckResourceAttr(
						"vcd_org.test1", "is_enabled", "true"),
				),
			},
		},
	})
}

func testAccCheckVcdOrgExists(n string, org *govcd.Org) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ORG ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		orgName := rs.Primary.Attributes["name"]

		new_org, err := govcd.GetOrgByName(conn.VCDClient, orgName)
		if err != nil {
			return fmt.Errorf("error could not find org: %v", err)
		}
		org = &new_org
		return nil
	}
}

func testAccCheckOrgDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_org" && rs.Primary.Attributes["name"] != "test1" {
			continue
		}

		org, err := govcd.GetOrgByName(conn.VCDClient, rs.Primary.Attributes["name"])
		if org != (govcd.Org{}) || err != nil {
			return fmt.Errorf("Org with name %s was found", rs.Primary.Attributes["name"])
		}

	}

	return nil
}

const testAccCheckVcdOrg_basic = `
resource "vcd_org" "test1"{
  name = "test1"
  full_name = "test1"
  is_enabled = "true"
  force = "true"
  recursive = "true"
}
`
