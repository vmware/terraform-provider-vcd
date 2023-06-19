//go:build api || functional || ALL

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccServiceAccount_SysOrg(t *testing.T) {
	preTestChecks(t)
	skipTestForApiToken(t)

	params := StringMap{
		"SaName":                 t.Name(),
		"Org":                    testConfig.Provider.SysOrg,
		"RoleName":               "System Administrator",
		"SoftwareId":             "12345678-1234-1234-1234-1234567890ab",
		"SoftwareVersion":        "1.0.0",
		"Uri":                    "example.com",
		"FileName":               "sa_api_token.json",
		"SoftwareIdUpdated":      "87654321-4321-4321-4321-ba0987654321",
		"SoftwareVersionUpdated": "2.3.4",
		"UriUpdated":             "test.xyz",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccServiceAccount_SysOrg, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccServiceAccount_SysOrg_Active, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	t.Cleanup(deleteApiTokenFile(t.Name()))
	resourceName := "vcd_service_account.sysadmin"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceAccountDestroy(params["Org"].(string), params["SaName"].(string)),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceName, "software_id", params["SoftwareId"].(string)),
					resource.TestCheckResourceAttr(resourceName, "software_version", params["SoftwareVersion"].(string)),
					resource.TestCheckResourceAttr(resourceName, "uri", params["Uri"].(string)),
					resource.TestCheckResourceAttr(resourceName, "active", "false"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceName, "software_id", params["SoftwareIdUpdated"].(string)),
					resource.TestCheckResourceAttr(resourceName, "software_version", params["SoftwareVersionUpdated"].(string)),
					resource.TestCheckResourceAttr(resourceName, "uri", params["UriUpdated"].(string)),
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
					testCheckFileExists(params["FileName"].(string)),
				),
			},
		},
	})
}

const testAccServiceAccount_SysOrg_Role = `
data "vcd_role" "sys_admin" {
  org  = "{{.Org}}"
  name = "{{.RoleName}}"
}
`

const testAccServiceAccount_SysOrg = testAccServiceAccount_SysOrg_Role + `
resource "vcd_service_account" "sysadmin" {
  name = "{{.SaName}}"
  org  = "{{.Org}}"

  role             = data.vcd_role.sys_admin.id
  software_id      = "{{.SoftwareId}}"
  software_version = "{{.SoftwareVersion}}"
  uri              = "{{.Uri}}"

  active = false
}
`
const testAccServiceAccount_SysOrg_Active = testAccServiceAccount_SysOrg_Role + `
resource "vcd_service_account" "sysadmin" {
  name = "{{.SaName}}"
  org  = "{{.Org}}"

  role             = data.vcd_role.sys_admin.id
  software_id      = "{{.SoftwareIdUpdated}}"
  software_version = "{{.SoftwareVersionUpdated}}"
  uri              = "{{.UriUpdated}}"

  active   = true
  file_name = "{{.FileName}}"
}
`

func TestAccServiceAccount_Org(t *testing.T) {
	preTestChecks(t)
	skipTestForApiToken(t)

	params := StringMap{
		"SaName":                 t.Name(),
		"Org":                    testConfig.VCD.Org,
		"SoftwareId":             "12345678-1234-1234-1234-1234567890ab",
		"SoftwareVersion":        "1.0.0",
		"Uri":                    "example.com",
		"FileName":               "sa_api_token.json",
		"SoftwareIdUpdated":      "87654321-4321-4321-4321-ba0987654321",
		"SoftwareVersionUpdated": "2.3.4",
		"UriUpdated":             "test.xyz",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccServiceAccount_Org, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccServiceAccount_OrgDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "step3"
	configText3 := templateFill(testAccServiceAccount_Org_Active, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	t.Cleanup(deleteApiTokenFile(params["FileName"].(string)))
	resourceName := "vcd_service_account.org_user"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceAccountDestroy(params["Org"].(string), params["SaName"].(string)),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceName, "software_id", params["SoftwareId"].(string)),
					resource.TestCheckResourceAttr(resourceName, "software_version", params["SoftwareVersion"].(string)),
					resource.TestCheckResourceAttr(resourceName, "uri", params["Uri"].(string)),
					resource.TestCheckResourceAttr(resourceName, "active", "false"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("data.vcd_service_account.org_user_ds", resourceName, []string{"file_name", "allow_token_file", "%"}),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceName, "software_id", params["SoftwareIdUpdated"].(string)),
					resource.TestCheckResourceAttr(resourceName, "software_version", params["SoftwareVersionUpdated"].(string)),
					resource.TestCheckResourceAttr(resourceName, "uri", params["UriUpdated"].(string)),
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
					testCheckFileExists(params["FileName"].(string)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdOrgObject(params["Org"].(string), params["SaName"].(string)),
				ImportStateVerifyIgnore: []string{"org"},
			},
		},
	})
}

const testAccServiceAccount_Org_Roles = `
data "vcd_role"	"vapp_author" {
  org  = "{{.Org}}"
  name = "vApp Author"
}

data "vcd_role"	"catalog_author" {
  org  = "{{.Org}}"
  name = "Catalog Author"
}
`

const testAccServiceAccount_Org = testAccServiceAccount_Org_Roles + `
resource "vcd_service_account" "org_user" {
  name = "{{.SaName}}"
  org  = "{{.Org}}"

  role             = data.vcd_role.vapp_author.id
  software_id      = "{{.SoftwareId}}"
  software_version = "{{.SoftwareVersion}}"
  uri              = "{{.Uri}}"

  active = false
}

`

const testAccServiceAccount_OrgDS = testAccServiceAccount_Org + `
data "vcd_service_account" "org_user_ds" {
  org  = "{{.Org}}"
  name = "{{.SaName}}"		
}		
`

const testAccServiceAccount_Org_Active = testAccServiceAccount_Org_Roles + `
resource "vcd_service_account" "org_user" {
  name = "{{.SaName}}"
  org  = "{{.Org}}"

  role             = data.vcd_role.catalog_author.id
  software_id      = "{{.SoftwareIdUpdated}}"
  software_version = "{{.SoftwareVersionUpdated}}"
  uri              = "{{.UriUpdated}}"

  active   = true
  file_name = "{{.FileName}}"
}
`

func testAccCheckServiceAccountDestroy(orgName, saName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		org, err := conn.GetOrgByName(orgName)
		if err != nil {
			return fmt.Errorf(errorRetrievingOrg, err)
		}

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "vcd_service_account" || rs.Primary.Attributes["name"] != saName {
				continue
			}

			_, err := org.GetServiceAccountById(rs.Primary.ID)
			if err == nil {
				return fmt.Errorf("error: service account still exists post-destroy")
			}

			return nil
		}

		return nil
	}
}
