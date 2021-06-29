// +build role ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdRightsBundle(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip("TestAccVcdRightsBundle requires system admin privileges")
		return
	}
	var rightsBundleName = t.Name()
	var rightsBundleUpdateName = t.Name() + "-update"
	var rightsBundleDescription = "A long description containing some text."
	var rightsBundleUpdateDescription = "A shorter description."

	var params = StringMap{
		"Tenant":                        testConfig.VCD.Org,
		"RightsBundleName":              rightsBundleName,
		"RightsBundleUpdateName":        rightsBundleUpdateName,
		"RightsBundleDescription":       rightsBundleDescription,
		"RightsBundleUpdateDescription": rightsBundleUpdateDescription,
		"FuncName":                      rightsBundleName,
		"Tags":                          "role",
	}
	configText := templateFill(testAccRightsBundle, params)

	params["FuncName"] = rightsBundleUpdateName
	params["RightsBundleDescription"] = rightsBundleUpdateDescription
	configTextUpdate := templateFill(testAccRightsBundleUpdate, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION basic: %s\n", configText)
	debugPrintf("#[DEBUG] CONFIGURATION update: %s\n", configTextUpdate)

	resourceDef := "vcd_rights_bundle." + rightsBundleName
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRightsBundleDestroy(resourceDef),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRightsBundleExists(resourceDef),
					resource.TestCheckResourceAttr(resourceDef, "name", rightsBundleName),
					resource.TestCheckResourceAttr(resourceDef, "description", rightsBundleDescription),
					resource.TestCheckResourceAttr(resourceDef, "publish_to_all_tenants", "false"),
					resource.TestCheckResourceAttr(resourceDef, "rights.#", "6"),
					resource.TestCheckResourceAttr(resourceDef, "tenants.#", "1"),
				),
			},
			resource.TestStep{
				Config: configTextUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRightsBundleExists(resourceDef),
					resource.TestCheckResourceAttr(resourceDef, "name", rightsBundleUpdateName),
					resource.TestCheckResourceAttr(resourceDef, "description", rightsBundleUpdateDescription),
					resource.TestCheckResourceAttr(resourceDef, "publish_to_all_tenants", "true"),
					resource.TestCheckResourceAttr(resourceDef, "rights.#", "5"),
				),
			},
			resource.TestStep{
				ResourceName:      resourceDef,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdTopHierarchy(rightsBundleUpdateName),
			},
		},
	})
	postTestChecks(t)
}

func testAccCheckRightsBundleExists(identifier string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[identifier]
		if !ok {
			return fmt.Errorf("not found: %s", identifier)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no role ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		_, err := conn.Client.GetRightsBundleById(rs.Primary.ID)
		return err
	}
}

func testAccCheckRightsBundleDestroy(identifier string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[identifier]
		if !ok {
			return fmt.Errorf("not found: %s", identifier)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no role ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		_, err := conn.Client.GetRightsBundleById(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("%s not deleted yet", identifier)
		}
		return nil

	}
}

const testAccRightsBundle = `
resource "vcd_rights_bundle" "{{.RightsBundleName}}" {
  name        = "{{.RightsBundleName}}"
  description = "{{.RightsBundleDescription}}"
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

const testAccRightsBundleUpdate = `
resource "vcd_rights_bundle" "{{.RightsBundleName}}" {
  name        = "{{.RightsBundleUpdateName}}"
  description = "{{.RightsBundleUpdateDescription}}"
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
