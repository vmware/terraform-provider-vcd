//go:build catalog || functional || access_control || ALL
// +build catalog functional access_control ALL

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func TestAccVcdCatalogAccessControl(t *testing.T) {
	preTestChecks(t)

	skipTestForApiToken(t)

	if !usingSysAdmin() {
		t.Skipf("%s requires system admin privileges", t.Name())
		return
	}

	var params = StringMap{
		"Org":                      testConfig.VCD.Org,
		"Org2":                     fmt.Sprintf("%s-1", testConfig.VCD.Org),
		"SharedToEveryone":         "true",
		"EveryoneAccessLevel":      fmt.Sprintf(`everyone_access_level = "%s"`, types.ControlAccessReadOnly),
		"AccessControlIdentifier0": "AC-Catalog0",
		"AccessControlIdentifier1": "AC-Catalog1",
		"AccessControlIdentifier2": "AC-Catalog2",
		"AccessControlIdentifier3": "AC-Catalog3",
		"AccessControlIdentifier4": "AC-Catalog4",
		"CatalogName0":             "Catalog-AC-0",
		"CatalogName1":             "Catalog-AC-1",
		"CatalogName2":             "Catalog-AC-2",
		"CatalogName3":             "Catalog-AC-3",
		"CatalogName4":             "Catalog-AC-4",
		"UserName1":                "ac-user1",
		"UserName2":                "ac-user2",
		"UserName3":                "ac-user3",
		"RoleName1":                govcd.OrgUserRoleOrganizationAdministrator,
		"RoleName2":                govcd.OrgUserRoleCatalogAuthor,
		"RoleName3":                govcd.OrgUserRoleCatalogAuthor,
		"AccessLevel1":             types.ControlAccessFullControl,
		"AccessLevel2":             types.ControlAccessReadWrite,
		"AccessLevel3":             types.ControlAccessReadOnly,
		"UserPassword":             "TO_BE_DISCARDED",
		"FuncName":                 t.Name(),
		"Tags":                     "catalog",
		"SkipNotice":               " ",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccCatalogAccessControl, params)
	params["AccessLevel1"] = types.ControlAccessReadWrite
	params["SharedToEveryone"] = "false"
	params["EveryoneAccessLevel"] = ""
	params["FuncName"] = t.Name() + "-update"
	params["SkipNotice"] = "# skip-binary-test: only for updates"
	updateText := templateFill(testAccCatalogAccessControl, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CREATION CONFIGURATION: %s", configText)
	debugPrintf("#[DEBUG] UPDATE CONFIGURATION: %s", updateText)

	resourceAC0 := "vcd_catalog_access_control.AC-Catalog0"
	resourceAC1 := "vcd_catalog_access_control.AC-Catalog1"
	resourceAC2 := "vcd_catalog_access_control.AC-Catalog2"
	resourceAC3 := "vcd_catalog_access_control.AC-Catalog3"
	resourceAC4 := "vcd_catalog_access_control.AC-Catalog4"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckCatalogAccessControlDestroy(testConfig.VCD.Org, []string{"Catalog-AC-0", "Catalog-AC-1", "Catalog-AC-2", "Catalog-AC-3"}),
		Steps: []resource.TestStep{
			// Test creation
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogAccessControlExists(resourceAC0, testConfig.VCD.Org),
					testAccCheckVcdCatalogAccessControlExists(resourceAC1, testConfig.VCD.Org),
					testAccCheckVcdCatalogAccessControlExists(resourceAC2, testConfig.VCD.Org),
					testAccCheckVcdCatalogAccessControlExists(resourceAC3, testConfig.VCD.Org),
					resource.TestCheckResourceAttr(resourceAC0, "shared_with_everyone", "true"),
					resource.TestCheckResourceAttr(resourceAC0, "everyone_access_level", types.ControlAccessReadOnly),
					resource.TestCheckResourceAttr(resourceAC0, "shared_with.#", "0"),

					resource.TestCheckResourceAttr(resourceAC1, "shared_with_everyone", "false"),
					resource.TestCheckResourceAttr(resourceAC1, "shared_with.#", "1"),
					testAccFindValuesInSet(resourceAC2, "shared_with", map[string]string{
						"subject_name": "ac-user1",
						"access_level": types.ControlAccessFullControl,
					}),

					resource.TestCheckResourceAttr(resourceAC2, "shared_with_everyone", "false"),
					resource.TestCheckResourceAttr(resourceAC2, "shared_with.#", "2"),
					testAccFindValuesInSet(resourceAC2, "shared_with", map[string]string{
						"subject_name": "ac-user1",
						"access_level": types.ControlAccessFullControl,
					}),
					testAccFindValuesInSet(resourceAC2, "shared_with", map[string]string{
						"subject_name": "ac-user2",
						"access_level": types.ControlAccessReadWrite,
					}),

					resource.TestCheckResourceAttr(resourceAC3, "shared_with_everyone", "false"),
					resource.TestCheckResourceAttr(resourceAC3, "shared_with.#", "3"),
					testAccFindValuesInSet(resourceAC3, "shared_with", map[string]string{
						"subject_name": "ac-user1",
						"access_level": types.ControlAccessFullControl,
					}),
					testAccFindValuesInSet(resourceAC3, "shared_with", map[string]string{
						"subject_name": "ac-user2",
						"access_level": types.ControlAccessReadWrite,
					}),
					testAccFindValuesInSet(resourceAC3, "shared_with", map[string]string{
						"subject_name": "ac-user3",
						"access_level": types.ControlAccessReadOnly,
					}),
					testAccFindValuesInSet(resourceAC4, "shared_with", map[string]string{
						"subject_name": "ac-user2",
						"access_level": types.ControlAccessReadWrite,
					}),
					testAccFindValuesInSet(resourceAC4, "shared_with", map[string]string{
						"subject_name": fmt.Sprintf("%s-1", testConfig.VCD.Org),
						"access_level": types.ControlAccessReadOnly,
					}),
				),
			},
			{
				Config: updateText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogAccessControlExists(resourceAC0, testConfig.VCD.Org),
					resource.TestCheckResourceAttr(resourceAC0, "shared_with_everyone", "false"),
					resource.TestCheckResourceAttr(resourceAC1, "shared_with.#", "1"),
					testAccFindValuesInSet(resourceAC1, "shared_with", map[string]string{
						"subject_name": "ac-user1",
						"access_level": types.ControlAccessReadWrite,
					}),
					resource.TestCheckResourceAttr(resourceAC2, "shared_with.#", "2"),
					testAccFindValuesInSet(resourceAC2, "shared_with", map[string]string{
						"subject_name": "ac-user1",
						"access_level": types.ControlAccessReadWrite,
					}),
					resource.TestCheckResourceAttr(resourceAC3, "shared_with.#", "3"),
					testAccFindValuesInSet(resourceAC3, "shared_with", map[string]string{
						"subject_name": "ac-user1",
						"access_level": types.ControlAccessReadWrite,
					}),
				)},

			// Tests import by name
			{
				Config:            configText,
				ResourceName:      "vcd_catalog_access_control.AC-Catalog1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgObject(testConfig.VCD.Org, "Catalog-AC-1"),
			},
			// Tests import by ID
			{
				Config:            configText,
				ResourceName:      "vcd_catalog_access_control.AC-Catalog2",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateCatalogIdViaResource("vcd_catalog_access_control.AC-Catalog2"),
			},
		},
	})
	postTestChecks(t)
}

func testAccCheckVcdCatalogAccessControlExists(resourceName string, orgName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Catalog ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		org, err := conn.VCDClient.GetAdminOrgByName(orgName)
		if err != nil {
			return fmt.Errorf("error: could not find Org: %s", err)
		}

		catalog, err := org.GetCatalogById(rs.Primary.ID, false)
		if err != nil {
			return fmt.Errorf("error: could not find catalog %s: %s", rs.Primary.ID, err)
		}
		_, err = catalog.GetAccessControl(tenantContext)
		if err != nil {
			return fmt.Errorf("error: could not get access control for catalog %s: %s", catalog.Catalog.Name, err)
		}

		return nil
	}
}

func testAccCheckCatalogAccessControlDestroy(orgName string, catalogNames []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*VCDClient)
		org, err := conn.VCDClient.GetAdminOrgByName(orgName)
		if err != nil {
			return fmt.Errorf("error: could not find Org: %s", err)
		}
		var destroyed int
		var existing []string
		for _, catalogName := range catalogNames {

			_, err = org.GetCatalogByName(catalogName, false)
			if err != nil && govcd.IsNotFound(err) {
				// The catalog was removed
				destroyed++
			}
			if err == nil {
				existing = append(existing, catalogName)
			}
		}

		if destroyed == len(catalogNames) {
			return nil
		}
		return fmt.Errorf("catalogs %v not deleted yet", existing)
	}
}

const testAccCatalogAccessControl = `
{{.SkipNotice}}
resource "vcd_org_user" "{{.UserName1}}" {
  org            = "{{.Org}}"
  name           = "{{.UserName1}}"
  password       = "{{.UserPassword}}"
  role           = "{{.RoleName1}}"
  take_ownership = false
}

resource "vcd_org_user" "{{.UserName2}}" {
  org            = "{{.Org}}"
  name           = "{{.UserName2}}"
  password       = "{{.UserPassword}}"
  role           = "{{.RoleName2}}"
  take_ownership = false
}

resource "vcd_org_user" "{{.UserName3}}" {
  org            = "{{.Org}}"
  name           = "{{.UserName3}}"
  password       = "{{.UserPassword}}"
  role           = "{{.RoleName3}}"
  take_ownership = false
}

resource "vcd_catalog" "{{.CatalogName0}}" {
  name             = "{{.CatalogName0}}"
  delete_force     = true
  delete_recursive = true
}

resource "vcd_catalog" "{{.CatalogName1}}" {
  name             = "{{.CatalogName1}}"
  delete_force     = true
  delete_recursive = true
}

resource "vcd_catalog" "{{.CatalogName2}}" {
  name             = "{{.CatalogName2}}"
  delete_force     = true
  delete_recursive = true
}

resource "vcd_catalog" "{{.CatalogName3}}" {
  name             = "{{.CatalogName3}}"
  delete_force     = true
  delete_recursive = true
}

resource "vcd_catalog" "{{.CatalogName4}}" {
  name             = "{{.CatalogName4}}"
  delete_force     = true
  delete_recursive = true
}

data "vcd_org" "other-org" {
  name = "{{.Org2}}"
}

resource "vcd_catalog_access_control" "{{.AccessControlIdentifier0}}" {
  org         = "{{.Org}}"
  catalog_id  = vcd_catalog.{{.CatalogName0}}.id

  shared_with_everyone    = {{.SharedToEveryone}}
  {{.EveryoneAccessLevel}}
}

resource "vcd_catalog_access_control" "{{.AccessControlIdentifier1}}" {
  org         = "{{.Org}}"
  catalog_id  = vcd_catalog.{{.CatalogName1}}.id

  shared_with_everyone = false

  shared_with {
    user_id      = vcd_org_user.{{.UserName1}}.id
    access_level = "{{.AccessLevel1}}"
  }
}

resource "vcd_catalog_access_control" "{{.AccessControlIdentifier2}}" {
  org         = "{{.Org}}"
  catalog_id  = vcd_catalog.{{.CatalogName2}}.id

  shared_with_everyone    = false

  shared_with {
    user_id      = vcd_org_user.{{.UserName1}}.id
    access_level = "{{.AccessLevel1}}"
  }
  shared_with {
    user_id      = vcd_org_user.{{.UserName2}}.id
    access_level = "{{.AccessLevel2}}"
  }
}

resource "vcd_catalog_access_control" "{{.AccessControlIdentifier3}}" {
  org         = "{{.Org}}"
  catalog_id  = vcd_catalog.{{.CatalogName3}}.id

  shared_with_everyone = false

  shared_with {
    user_id      = vcd_org_user.{{.UserName1}}.id
    access_level = "{{.AccessLevel1}}"
  }
  shared_with {
    user_id      = vcd_org_user.{{.UserName2}}.id
    access_level = "{{.AccessLevel2}}"
  }
  shared_with {
    user_id      = vcd_org_user.{{.UserName3}}.id
    access_level = "{{.AccessLevel3}}"
  }
}

resource "vcd_catalog_access_control" "{{.AccessControlIdentifier4}}" {
  org         = "{{.Org}}"
  catalog_id  = vcd_catalog.{{.CatalogName4}}.id

  shared_with_everyone = false

  shared_with {
    org_id       = data.vcd_org.other-org.id
    access_level = "{{.AccessLevel3}}"
  }
  shared_with {
    user_id      = vcd_org_user.{{.UserName2}}.id
    access_level = "{{.AccessLevel2}}"
  }
}
`
