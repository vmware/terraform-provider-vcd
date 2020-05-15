// +build user functional ALL

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdOrgGroup(t *testing.T) {

	role1 := govcd.OrgUserRoleOrganizationAdministrator
	role2 := govcd.OrgUserRoleDeferToIdentityProvider

	var params = StringMap{
		"Org":       testConfig.VCD.Org,
		"GroupName": "TestAccVcdOrgGroup",
		"RoleName":  role1,
		"Tags":      "user",
	}

	configText := templateFill(testAccOrgGroup, params)

	params["FuncName"] = t.Name() + "-Step1"
	params["RoleName"] = role2
	configText2 := templateFill(testAccOrgGroup, params)

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckVcdGroupDestroy(params["GroupName"].(string)),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_org_group.group", "id", regexp.MustCompile(`^urn:vcloud:group:`)),
					resource.TestCheckResourceAttr("vcd_org_group.group", "name", "TestAccVcdOrgGroup"),
					resource.TestCheckResourceAttr("vcd_org_group.group", "role", role1),
					resource.TestCheckResourceAttr("vcd_org_group.group", "description", ""),
				),
			},
			resource.TestStep{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_org_group.group", "id", regexp.MustCompile(`^urn:vcloud:group:`)),
					resource.TestCheckResourceAttr("vcd_org_group.group", "name", "TestAccVcdOrgGroup"),
					resource.TestCheckResourceAttr("vcd_org_group.group", "role", role2),
					resource.TestCheckResourceAttr("vcd_org_group.group", "description", ""),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_org_group.group-import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgObject(testConfig, params["GroupName"].(string)),
			},
		},
	})
}

func testAccCheckVcdGroupDestroy(groupName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)
		adminOrg, err := conn.GetAdminOrg(testConfig.VCD.Org)
		if err != nil {
			return err
		}
		group, err := adminOrg.GetGroupByName(groupName, false)
		if err != govcd.ErrorEntityNotFound {
			return fmt.Errorf("group %s was not destroyed", groupName)
		}
		if group != nil {
			return fmt.Errorf("group %s was found in %s ", groupName, adminOrg.AdminOrg.Name)
		}
		return nil
	}
}

const testAccOrgGroup = `
resource "vcd_org_group" "group" {
  name = "{{.GroupName}}"
  role = "{{.RoleName}}"
}
`
