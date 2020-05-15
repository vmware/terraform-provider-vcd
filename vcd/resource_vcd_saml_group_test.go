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
	var params = StringMap{
		"Org":       testConfig.VCD.Org,
		"GroupName": "TestAccVcdOrgGroup",
		"RoleName":  govcd.OrgUserRoleDeferToIdentityProvider,
		"Tags":      "user",
	}

	configText := templateFill(testAccOrgGroup, params)
	resource.ParallelTest(t, resource.TestCase{
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckVcdGroupDestroy(params["GroupName"].(string)),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_org_saml_group.group", "id", regexp.MustCompile(`^urn:vcloud:group:`)),
					resource.TestCheckResourceAttr("vcd_org_saml_group.group", "name", "TestAccVcdOrgGroup"),
					resource.TestCheckResourceAttr("vcd_org_saml_group.group", "description", ""),
					// When rule_tag is not specified - we expect it to be the same as ID
					// resource.TestCheckResourceAttrPair("vcd_org_saml_group.group", "name", "vcd_nsxv_snat.test", "id"),
				),
			},
			// resource.TestStep{ // Step 2 - resource import
			// 	ResourceName:      "vcd_nsxv_snat.imported",
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// 	ImportStateIdFunc: importStateIdByResourceName("vcd_nsxv_snat.test"),
			// },
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
resource "vcd_org_saml_group" "group" {
  org            = "{{.Org}}"
  name           = "{{.GroupName}}"
  role           = "{{.RoleName}}"
}
`
