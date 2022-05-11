//go:build vdc || ALL || functional
// +build vdc ALL functional

package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"strings"
	"testing"
)

func TestAccVcdOrgVdcAccessControl(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
	}

	accessControlName := "test-access-control"

	var params = StringMap{
		"Org":               testConfig.VCD.Org,
		"Vdc":               testConfig.VCD.Vdc,
		"AccessControlName": accessControlName,
		"UserName":          strings.ToLower(t.Name()),
		"PasswordFile":      orgUserPasswordFile,
		"RoleName":          govcd.OrgUserRoleOrganizationAdministrator,
	}

	params["FuncName"] = t.Name() + "step1"
	configText := templateFill(testAccCheckVcdAccessControlStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	params["FuncName"] = t.Name() + "step2"
	configTex2 := templateFill(testAccCheckVcdAccessControlStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		// CheckDestroy:      testAccCheckCatalogDestroy,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check:  resource.ComposeTestCheckFunc(),
			},
			{
				Config: configTex2,
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

const testAccCheckVcdAccessControlStep1 = `
resource "vcd_org_vdc_access_control" "{{.AccessControlName}}" {
  org                   = "{{.Org}}"
  vdc                   = "{{.Vdc}}"
  shared_with_everyone  = true
  everyone_access_level = "ReadOnly"
}
`

const testAccCheckVcdAccessControlStep2 = `
resource "vcd_org_user" "{{.UserName}}" {
  org            = "{{.Org}}"
  name           = "{{.UserName}}"
  password_file  = "{{.PasswordFile}}"
  role           = "{{.RoleName}}"
  take_ownership = true
}

resource "vcd_org_vdc_access_control" "{{.AccessControlName}}" {
  org                   = "{{.Org}}"
  vdc                   = "{{.Vdc}}"
  shared_with_everyone  = false
  shared_with {
    user_id             = vcd_org_user.{{.UserName}}.id
    access_level        = "ReadOnly"
  }
}
`
