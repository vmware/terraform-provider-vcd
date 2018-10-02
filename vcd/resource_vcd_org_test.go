package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/go-vcloud-director/govcd"
)

var orgNameTestAccVcdOrgBasic string = "TestAccVcdOrgBasic"

func TestAccVcdOrgBasic(t *testing.T) {

	var e govcd.Org
	var params = StringMap{
		"OrgName": orgNameTestAccVcdOrgBasic,
	}

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	configText := templateFill(testAccCheckVcdOrg_basic, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOrgDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdOrgExists("vcd_org."+orgNameTestAccVcdOrgBasic, &e),
					resource.TestCheckResourceAttr(
						"vcd_org."+orgNameTestAccVcdOrgBasic, "name", orgNameTestAccVcdOrgBasic),
					resource.TestCheckResourceAttr(
						"vcd_org."+orgNameTestAccVcdOrgBasic, "full_name", orgNameTestAccVcdOrgBasic),
					resource.TestCheckResourceAttr(
						"vcd_org."+orgNameTestAccVcdOrgBasic, "is_enabled", "true"),
				),
			},
		},
	})
}

func testAccCheckVcdOrgExists(n string, org *govcd.Org) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Org ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		orgName := rs.Primary.Attributes["name"]

		new_org, err := govcd.GetOrgByName(conn.VCDClient, orgName)
		if err != nil {
			return fmt.Errorf("error: could not find Org: %v", err)
		}
		org = &new_org
		return nil
	}
}

func testAccCheckOrgDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_org" && rs.Primary.Attributes["name"] != orgNameTestAccVcdOrgBasic {
			continue
		}

		org, err := govcd.GetOrgByName(conn.VCDClient, rs.Primary.Attributes["name"])
		if org != (govcd.Org{}) || err != nil {
			return fmt.Errorf("org with name %s was found", rs.Primary.Attributes["name"])
		}

	}

	return nil
}

const testAccCheckVcdOrg_basic = `
resource "vcd_org" "{{.OrgName}}" {
  name       = "{{.OrgName}}"
  full_name  = "{{.OrgName}}"
  is_enabled = "true"
  force      = "true"
  recursive  = "true"
}
`
