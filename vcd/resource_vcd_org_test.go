package vcd

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/go-vcloud-director/govcd"
)

const orgNameTestAccVcdOrgBasic string = "TestAccVcdOrgBasic"

func TestAccVcdOrgBasic(t *testing.T) {

	var e govcd.Org
	var params = StringMap{
		"OrgName": orgNameTestAccVcdOrgBasic,
	}

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	if !usingSysAdmin() {
		t.Skip("TestAccVcdOrgBasic requires system admin privileges")
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
						"vcd_org."+orgNameTestAccVcdOrgBasic, "description", orgNameTestAccVcdOrgBasic),
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

		deleted := false
		var org govcd.Org
		var err error
		for N := 0; N < 10; N++ {
			org, err = govcd.GetOrgByName(conn.VCDClient, rs.Primary.Attributes["name"])
			deleted = (org == (govcd.Org{})) && (err != nil)
			if deleted {
				break
			}
			time.Sleep(time.Second)
		}

		if !deleted {
			return fmt.Errorf("org with name %s was found (%#v %v)", rs.Primary.Attributes["name"], org, err)
		}

	}

	return nil
}

const testAccCheckVcdOrg_basic = `
resource "vcd_org" "{{.OrgName}}" {
  name              = "{{.OrgName}}"
  full_name         = "{{.OrgName}}"
  description       = "{{.OrgName}}"
  is_enabled        = "true"
  delete_force      = "true"
  delete_recursive  = "true"
}
`
