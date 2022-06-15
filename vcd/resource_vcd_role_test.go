//go:build role || ALL || functional
// +build role ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdRole(t *testing.T) {
	preTestChecks(t)
	skipTestForApiToken(t)
	var roleName = t.Name()
	var roleUpdateName = t.Name() + "-update"
	var roleDescription = "A long description containing some text."
	var roleUpdateDescription = "A shorter description."

	var params = StringMap{
		"Org":                   testConfig.VCD.Org,
		"RoleName":              roleName,
		"RoleUpdateName":        roleUpdateName,
		"RoleDescription":       roleDescription,
		"RoleUpdateDescription": roleUpdateDescription,
		"FuncName":              roleName,
		"Tags":                  "role",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccRole, params)

	params["FuncName"] = roleUpdateName
	params["roleDescription"] = roleUpdateDescription
	configTextUpdate := templateFill(testAccRoleUpdate, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION basic: %s\n", configText)
	debugPrintf("#[DEBUG] CONFIGURATION update: %s\n", configTextUpdate)

	resourceDef := "vcd_role." + roleName
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRoleDestroy(resourceDef),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(resourceDef),
					resource.TestCheckResourceAttr(resourceDef, "name", roleName),
					resource.TestCheckResourceAttr(resourceDef, "rights.#", "6"),
				),
			},
			{
				Config: configTextUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(resourceDef),
					resource.TestCheckResourceAttr(resourceDef, "name", roleUpdateName),
					resource.TestCheckResourceAttr(resourceDef, "description", roleUpdateDescription),
					resource.TestCheckResourceAttr(resourceDef, "rights.#", "5"),
				),
			},
			{
				ResourceName:      resourceDef,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgObject(testConfig, roleUpdateName),
			},
		},
	})
	postTestChecks(t)
}

func testAccCheckRoleExists(identifier string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[identifier]
		if !ok {
			return fmt.Errorf("not found: %s", identifier)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no role ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		org, err := conn.GetAdminOrg("")
		if err != nil {
			return err
		}
		_, err = org.GetRoleById(rs.Primary.ID)
		return err
	}
}

func testAccCheckRoleDestroy(identifier string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[identifier]
		if !ok {
			return fmt.Errorf("not found: %s", identifier)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no role ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		org, err := conn.GetAdminOrg("")
		if err != nil {
			return err
		}
		_, err = org.GetRoleById(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("%s not deleted yet", identifier)
		}
		return nil

	}
}

const testAccRole = `
resource "vcd_role" "{{.RoleName}}" {
  org         = "{{.Org}}"
  name        = "{{.RoleName}}"
  description = "{{.RoleDescription}}"
  rights = [
    "Catalog: Add vApp from My Cloud",
    "Catalog: Edit Properties",
    "Catalog: View Private and Shared Catalogs",
    "Organization vDC Compute Policy: View",
    "vApp Template / Media: Edit",
    "vApp Template / Media: View",
  ]
}
`

const testAccRoleUpdate = `
resource "vcd_role" "{{.RoleName}}" {
  org         = "{{.Org}}"
  name        = "{{.RoleUpdateName}}"
  description = "{{.RoleUpdateDescription}}"
  rights = [
    # "Catalog: Add vApp from My Cloud",
    "Catalog: Edit Properties",
    "Catalog: View Private and Shared Catalogs",
    "Organization vDC Compute Policy: View",
    "vApp Template / Media: Edit",
    "vApp Template / Media: View",
  ]
}
`
