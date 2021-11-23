//go:build role || ALL || functional
// +build role ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdGlobalRole(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip("TestAccVcdGlobalRole requires system admin privileges")
		return
	}
	skipTestForApiToken(t)
	var globalRoleName = t.Name()
	var globalRoleUpdateName = t.Name() + "-update"
	var globalRoleDescription = "A long description containing some text."
	var globalRoleUpdateDescription = "A shorter description."

	var params = StringMap{
		"Tenant":                      testConfig.VCD.Org,
		"GlobalRoleName":              globalRoleName,
		"GlobalRoleUpdateName":        globalRoleUpdateName,
		"GlobalRoleDescription":       globalRoleDescription,
		"GlobalRoleUpdateDescription": globalRoleUpdateDescription,
		"FuncName":                    globalRoleName,
		"Tags":                        "role",
	}
	configText := templateFill(testAccGlobalRole, params)

	params["FuncName"] = globalRoleUpdateName
	params["GlobalRoleDescription"] = globalRoleUpdateDescription
	configTextUpdate := templateFill(testAccGlobalRoleUpdate, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION basic: %s\n", configText)
	debugPrintf("#[DEBUG] CONFIGURATION update: %s\n", configTextUpdate)

	resourceDef := "vcd_global_role." + globalRoleName
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGlobalRoleDestroy(resourceDef),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalRoleExists(resourceDef),
					resource.TestCheckResourceAttr(resourceDef, "name", globalRoleName),
					resource.TestCheckResourceAttr(resourceDef, "description", globalRoleDescription),
					resource.TestCheckResourceAttr(resourceDef, "publish_to_all_tenants", "false"),
					resource.TestCheckResourceAttr(resourceDef, "rights.#", "6"),
					resource.TestCheckResourceAttr(resourceDef, "tenants.#", "1"),
				),
			},
			resource.TestStep{
				Config: configTextUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalRoleExists(resourceDef),
					resource.TestCheckResourceAttr(resourceDef, "name", globalRoleUpdateName),
					resource.TestCheckResourceAttr(resourceDef, "description", globalRoleUpdateDescription),
					resource.TestCheckResourceAttr(resourceDef, "publish_to_all_tenants", "true"),
					resource.TestCheckResourceAttr(resourceDef, "rights.#", "5"),
				),
			},
			resource.TestStep{
				ResourceName:      resourceDef,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdTopHierarchy(globalRoleUpdateName),
			},
		},
	})
	postTestChecks(t)
}

func testAccCheckGlobalRoleExists(identifier string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[identifier]
		if !ok {
			return fmt.Errorf("not found: %s", identifier)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no role ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		_, err := conn.Client.GetGlobalRoleById(rs.Primary.ID)
		return err
	}
}

func testAccCheckGlobalRoleDestroy(identifier string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[identifier]
		if !ok {
			return fmt.Errorf("not found: %s", identifier)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no role ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		_, err := conn.Client.GetGlobalRoleById(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("%s not deleted yet", identifier)
		}
		return nil

	}
}

const testAccGlobalRole = `
resource "vcd_global_role" "{{.GlobalRoleName}}" {
  name        = "{{.GlobalRoleName}}"
  description = "{{.GlobalRoleDescription}}"
  rights = [
    "Catalog: Add vApp from My Cloud",
    "Catalog: Edit Properties",
    "Catalog: View Private and Shared Catalogs",
    "Organization vDC Compute Policy: View",
    "vApp Template / Media: Edit",
    "vApp Template / Media: View",
  ]
  publish_to_all_tenants = false
  tenants                = [ "{{.Tenant}}" ]
}
`

const testAccGlobalRoleUpdate = `
resource "vcd_global_role" "{{.GlobalRoleName}}" {
  name        = "{{.GlobalRoleUpdateName}}"
  description = "{{.GlobalRoleUpdateDescription}}"
  rights = [
    # "Catalog: Add vApp from My Cloud",
    "Catalog: Edit Properties",
    "Catalog: View Private and Shared Catalogs",
    "Organization vDC Compute Policy: View",
    "vApp Template / Media: Edit",
    "vApp Template / Media: View",
  ]
  publish_to_all_tenants = true
}
`
