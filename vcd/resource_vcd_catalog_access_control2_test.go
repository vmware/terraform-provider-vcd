//go:build catalog || functional || access_control || ALL

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"testing"
)

func TestAccVcdCatalogAccessControl2(t *testing.T) {
	preTestChecks(t)

	skipTestForServiceAccountAndApiToken(t)

	skipIfNotSysAdmin(t)

	const (
		catalog0   = "Test-Catalog-AC-0"
		catalog1   = "Test-Catalog-AC-1"
		catalog2   = "Test-Catalog-AC-2"
		catalog3   = "Test-Catalog-AC-3"
		catalog4   = "Test-Catalog-AC-4"
		catalog5   = "Test-Catalog-AC-5"
		catalog6   = "Test-Catalog-AC-6"
		acCatalog0 = "AC-Catalog0"
		acCatalog1 = "AC-Catalog1"
		acCatalog2 = "AC-Catalog2"
		acCatalog3 = "AC-Catalog3"
		acCatalog4 = "AC-Catalog4"
		acCatalog5 = "AC-Catalog5"
		acCatalog6 = "AC-Catalog6"
		userName1  = "test-user1"
		userName2  = "test-user2"
		userName3  = "test-user3"
		newOrg1    = "test-org1"
		newOrg2    = "test-org2"
		newOrg3    = "test-org3"
	)
	var params = StringMap{
		"Org1":                     testConfig.VCD.Org,
		"Org2":                     fmt.Sprintf("%s-1", testConfig.VCD.Org),
		"SharedToEveryone":         "true",
		"EveryoneAccessLevel":      fmt.Sprintf(`everyone_access_level = "%s"`, types.ControlAccessReadOnly),
		"AccessControlIdentifier0": acCatalog0,
		"AccessControlIdentifier1": acCatalog1,
		"AccessControlIdentifier2": acCatalog2,
		"AccessControlIdentifier3": acCatalog3,
		"AccessControlIdentifier4": acCatalog4,
		"AccessControlIdentifier5": acCatalog5,
		"AccessControlIdentifier6": acCatalog6,
		"CatalogName0":             catalog0,
		"CatalogName1":             catalog1,
		"CatalogName2":             catalog2,
		"CatalogName3":             catalog3,
		"CatalogName4":             catalog4,
		"CatalogName5":             catalog5,
		"CatalogName6":             catalog6,
		"UserName1":                userName1,
		"UserName2":                userName2,
		"UserName3":                userName3,
		"NewOrg1":                  newOrg1,
		"NewOrg2":                  newOrg2,
		"NewOrg3":                  newOrg3,
		"RoleName1":                govcd.OrgUserRoleOrganizationAdministrator,
		"RoleName2":                govcd.OrgUserRoleCatalogAuthor,
		"RoleName3":                govcd.OrgUserRoleCatalogAuthor,
		"AccessLevel1":             types.ControlAccessFullControl,
		"AccessLevel2":             types.ControlAccessReadWrite,
		"AccessLevel3":             types.ControlAccessReadOnly,
		"UserPassword":             "TO_BE_DISCARDED",
		"FuncName":                 t.Name(),
		"ProviderVcdSystem":        providerVcdSystem,
		"ProviderVcdOrg1":          providerVcdOrg1,
		"ProviderVcdOrg2":          providerVcdOrg2,
		"Tags":                     "catalog",
		"SkipNotice":               " ",
	}
	testParamsNotEmpty(t, params)

	configTextCreate := templateFill(testAccCatalogAccessControlCreation, params)
	params["FuncName"] = t.Name() + "-access"
	configTextAccess := templateFill(testAccCatalogAccessControlCreation+testAccCatalogAccessControlAccess, params)
	params["AccessLevel1"] = types.ControlAccessReadWrite
	params["SharedToEveryone"] = "false"
	params["EveryoneAccessLevel"] = ""
	params["FuncName"] = t.Name() + "-update"
	params["SkipNotice"] = "# skip-binary-test: only for updates"
	updateText := templateFill(testAccCatalogAccessControlCreation+testAccCatalogAccessControlAccess, params)
	params["FuncName"] = t.Name() + "-check"
	checkText := templateFill(testAccCatalogAccessControlCreation+testAccCatalogAccessControlAccess+testAccCatalogAccessControlCheck, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CREATION CONFIGURATION: %s", configTextCreate)
	debugPrintf("#[DEBUG] ACCESS CONTROL CONFIGURATION: %s", configTextAccess)
	debugPrintf("#[DEBUG] UPDATE CONFIGURATION: %s", updateText)
	debugPrintf("#[DEBUG] CHECK CONFIGURATION: %s", checkText)

	resourceAC0 := "vcd_catalog_access_control." + acCatalog0
	resourceAC1 := "vcd_catalog_access_control." + acCatalog1
	resourceAC2 := "vcd_catalog_access_control." + acCatalog2
	resourceAC3 := "vcd_catalog_access_control." + acCatalog3
	resourceAC4 := "vcd_catalog_access_control." + acCatalog4
	resource.Test(t, resource.TestCase{
		ProviderFactories: buildMultipleProviders(),
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckCatalogAccessControlDestroy(testConfig.VCD.Org, []string{
				catalog0, catalog1, catalog2, catalog3, catalog4, catalog5, catalog6}),
			testAccCheckOrgDestroy(newOrg1),
			testAccCheckOrgDestroy(newOrg2),
			testAccCheckOrgDestroy(newOrg3),
		),
		Steps: []resource.TestStep{
			// Test creation
			{
				Config: configTextCreate,
				Check: resource.ComposeTestCheckFunc(
					testCheckCatalogAndItemsExist(testConfig.VCD.Org, catalog0, false, false, 0, 0, 0),
					testCheckCatalogAndItemsExist(testConfig.VCD.Org, catalog1, false, false, 0, 0, 0),
					testCheckCatalogAndItemsExist(testConfig.VCD.Org, catalog2, false, false, 0, 0, 0),
					testCheckCatalogAndItemsExist(testConfig.VCD.Org, catalog3, false, false, 0, 0, 0),
					testCheckCatalogAndItemsExist(testConfig.VCD.Org, catalog4, false, false, 0, 0, 0),
					testCheckCatalogAndItemsExist(testConfig.VCD.Org, catalog5, false, false, 0, 0, 0),
					testCheckCatalogAndItemsExist(testConfig.VCD.Org, catalog6, false, false, 0, 0, 0)),
			},
			// Test access
			{
				Config: configTextAccess,
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
					resource.TestCheckTypeSetElemNestedAttrs(resourceAC2, "shared_with.*",
						map[string]string{
							"subject_name": userName1,
							"access_level": types.ControlAccessFullControl,
						}),
					resource.TestCheckResourceAttr(resourceAC2, "shared_with_everyone", "false"),
					resource.TestCheckResourceAttr(resourceAC2, "shared_with.#", "2"),

					resource.TestCheckTypeSetElemNestedAttrs(resourceAC2, "shared_with.*",
						map[string]string{
							"subject_name": userName2,
							"access_level": types.ControlAccessReadWrite,
						}),

					resource.TestCheckResourceAttr(resourceAC3, "shared_with_everyone", "false"),
					resource.TestCheckResourceAttr(resourceAC3, "shared_with.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceAC3, "shared_with.*",
						map[string]string{
							"subject_name": userName1,
							"access_level": types.ControlAccessFullControl,
						}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceAC3, "shared_with.*",
						map[string]string{
							"subject_name": userName2,
							"access_level": types.ControlAccessReadWrite,
						}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceAC3, "shared_with.*",
						map[string]string{
							"subject_name": userName3,
							"access_level": types.ControlAccessReadOnly,
						}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceAC4, "shared_with.*",
						map[string]string{
							"subject_name": userName2,
							"access_level": types.ControlAccessReadWrite,
						}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceAC4, "shared_with.*",
						map[string]string{
							"subject_name": fmt.Sprintf("%s-1", testConfig.VCD.Org),
							"access_level": types.ControlAccessReadOnly,
						}),
				),
			},
			// Test update
			{
				Config: updateText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogAccessControlExists(resourceAC0, testConfig.VCD.Org),
					resource.TestCheckResourceAttr(resourceAC0, "shared_with_everyone", "false"),
					resource.TestCheckResourceAttr(resourceAC1, "shared_with.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceAC1, "shared_with.*",
						map[string]string{
							"subject_name": userName1,
							"access_level": types.ControlAccessReadWrite,
						}),
					resource.TestCheckResourceAttr(resourceAC2, "shared_with.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceAC2, "shared_with.*",
						map[string]string{
							"subject_name": userName1,
							"access_level": types.ControlAccessReadWrite,
						}),
					resource.TestCheckResourceAttr(resourceAC3, "shared_with.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceAC3, "shared_with.*",
						map[string]string{
							"subject_name": userName1,
							"access_level": types.ControlAccessReadWrite,
						}),
				)},

			// Tests import by name
			{
				Config:            configTextAccess,
				ResourceName:      "vcd_catalog_access_control." + acCatalog1,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgObject(testConfig.VCD.Org, catalog1),
			},
			// Tests import by ID
			{
				Config:            configTextAccess,
				ResourceName:      "vcd_catalog_access_control." + acCatalog2,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateCatalogId(testConfig.VCD.Org, catalog2),
			},
		},
	})
	postTestChecks(t)
}

const testAccCatalogAccessControlCreation = `
{{.SkipNotice}}

resource "vcd_org" "{{.NewOrg1}}" {
  provider         = {{.ProviderVcdSystem}}
  name             = "{{.NewOrg1}}"
  full_name        = "{{.NewOrg1}}"
  delete_force     = true
  delete_recursive = true
}

resource "vcd_org" "{{.NewOrg2}}" {
  provider         = {{.ProviderVcdSystem}}
  name             = "{{.NewOrg2}}"
  full_name        = "{{.NewOrg2}}"
  delete_force     = true
  delete_recursive = true
}

resource "vcd_org" "{{.NewOrg3}}" {
  provider         = {{.ProviderVcdSystem}}
  name             = "{{.NewOrg3}}"
  full_name        = "{{.NewOrg3}}"
  delete_force     = true
  delete_recursive = true
}

resource "vcd_org_user" "{{.UserName1}}" {
  provider       = {{.ProviderVcdOrg1}}
  org            = "{{.Org1}}"
  name           = "{{.UserName1}}"
  password       = "{{.UserPassword}}"
  role           = "{{.RoleName1}}"
  take_ownership = false
}

resource "vcd_org_user" "{{.UserName2}}" {
  provider       = {{.ProviderVcdOrg1}}
  org            = "{{.Org1}}"
  name           = "{{.UserName2}}"
  password       = "{{.UserPassword}}"
  role           = "{{.RoleName2}}"
  take_ownership = false
}

resource "vcd_org_user" "{{.UserName3}}" {
  provider       = {{.ProviderVcdOrg1}}
  org            = "{{.Org1}}"
  name           = "{{.UserName3}}"
  password       = "{{.UserPassword}}"
  role           = "{{.RoleName3}}"
  take_ownership = false
}

resource "vcd_catalog" "{{.CatalogName0}}" {
  provider         = {{.ProviderVcdOrg1}}
  name             = "{{.CatalogName0}}"
  delete_force     = true
  delete_recursive = true
}

resource "vcd_catalog" "{{.CatalogName1}}" {
  provider         = {{.ProviderVcdOrg1}}
  name             = "{{.CatalogName1}}"
  delete_force     = true
  delete_recursive = true
}

resource "vcd_catalog" "{{.CatalogName2}}" {
  provider         = {{.ProviderVcdOrg1}}
  name             = "{{.CatalogName2}}"
  delete_force     = true
  delete_recursive = true
}

resource "vcd_catalog" "{{.CatalogName3}}" {
  provider         = {{.ProviderVcdOrg1}}
  name             = "{{.CatalogName3}}"
  delete_force     = true
  delete_recursive = true
}

resource "vcd_catalog" "{{.CatalogName4}}" {
  provider         = {{.ProviderVcdOrg1}}
  name             = "{{.CatalogName4}}"
  delete_force     = true
  delete_recursive = true
}

resource "vcd_catalog" "{{.CatalogName5}}" {
  provider         = {{.ProviderVcdOrg1}}
  name             = "{{.CatalogName5}}"
  delete_force     = true
  delete_recursive = true
}

resource "vcd_catalog" "{{.CatalogName6}}" {
  provider         = {{.ProviderVcdOrg1}}
  name             = "{{.CatalogName6}}"
  delete_force     = true
  delete_recursive = true
}

data "vcd_org" "other-org" {
  provider         = {{.ProviderVcdSystem}}
  name = "{{.Org2}}"
}
`

const testAccCatalogAccessControlAccess = ` 
resource "vcd_catalog_access_control" "{{.AccessControlIdentifier0}}" {
  provider    = {{.ProviderVcdOrg1}}
  org         = "{{.Org1}}"
  catalog_id  = vcd_catalog.{{.CatalogName0}}.id

  shared_with_everyone    = {{.SharedToEveryone}}
  {{.EveryoneAccessLevel}}
}

resource "vcd_catalog_access_control" "{{.AccessControlIdentifier1}}" {
  provider    = {{.ProviderVcdOrg1}}
  org         = "{{.Org1}}"
  catalog_id  = vcd_catalog.{{.CatalogName1}}.id

  shared_with_everyone = false

  shared_with {
    user_id      = vcd_org_user.{{.UserName1}}.id
    access_level = "{{.AccessLevel1}}"
  }
}

resource "vcd_catalog_access_control" "{{.AccessControlIdentifier2}}" {
  provider    = {{.ProviderVcdOrg1}}
  org         = "{{.Org1}}"
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
  provider    = {{.ProviderVcdOrg1}}
  org         = "{{.Org1}}"
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
  provider    = {{.ProviderVcdSystem}}
  org         = "{{.Org1}}"
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

resource "vcd_catalog_access_control" "{{.AccessControlIdentifier5}}" {
  provider    = {{.ProviderVcdOrg1}}
  org         = "{{.Org1}}"
  catalog_id  = vcd_catalog.{{.CatalogName5}}.id

  shared_with_everyone             = false
  read_only_shared_with_all_orgs = true

  shared_with {
    user_id       = vcd_org_user.{{.UserName3}}.id
    access_level = "{{.AccessLevel3}}"
  }
  shared_with {
    user_id      = vcd_org_user.{{.UserName2}}.id
    access_level = "{{.AccessLevel2}}"
  }
}

resource "vcd_catalog_access_control" "{{.AccessControlIdentifier6}}" {
  provider    = {{.ProviderVcdOrg1}}
  org         = "{{.Org1}}"
  catalog_id  = vcd_catalog.{{.CatalogName6}}.id

  shared_with_everyone           = false
  read_only_shared_with_all_orgs = true
}
`

const testAccCatalogAccessControlCheck = `
#data "vcd_catalog" "{{.CatalogName1}}" {
#  provider = {{.ProviderVcdOrg2}}
#  org      = "{{.Org1}}"
#  name     = "{{.CatalogName1}}"
#}

data "vcd_catalog" "{{.CatalogName5}}" {
  provider = {{.ProviderVcdOrg2}}
  org      = "{{.Org1}}"
  name     = "{{.CatalogName5}}"
}

data "vcd_catalog" "{{.CatalogName6}}" {
  provider = {{.ProviderVcdOrg2}}
  org      = "{{.Org1}}"
  name     = "{{.CatalogName6}}"
}
`
