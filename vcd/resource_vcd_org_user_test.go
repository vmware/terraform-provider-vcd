// +build user functional ALL

package vcd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

type userTestData struct {
	name       string // name of the user. Note: only lowercase letters allowed
	roleName   string // the role this user is created with
	secondRole string // The role to which we change using Update()
}

var orgUserPasswordText = "CHANGE-ME"
var orgUserPasswordFile = "org_user_pwd.txt"

func prepareUserData(t *testing.T) []userTestData {

	if !fileExists(orgUserPasswordFile) {

		// if the password file does not exist, we create and fill a new one
		password := []byte(orgUserPasswordText)
		file, err := os.Create(orgUserPasswordFile)
		if err != nil {
			t.Skip(fmt.Sprintf("error creating file %s: %s", orgUserPasswordFile, err))
		}
		writer := bufio.NewWriter(file)
		count, err := writer.Write(password)
		if err != nil || count == 0 {
			t.Skip(fmt.Sprintf("error writing to file %s (written bytes %d): %s", orgUserPasswordFile, count, err))
		}
		err = writer.Flush()
		if err != nil {
			t.Skip(fmt.Sprintf("error flushing file %s: %s", orgUserPasswordFile, err))
		}
		_ = file.Close()
	}
	return []userTestData{
		{
			name:       "test_user_admin",
			roleName:   govcd.OrgUserRoleOrganizationAdministrator,
			secondRole: govcd.OrgUserRoleVappAuthor,
		},
		{
			name:       "test_user_vapp_author",
			roleName:   govcd.OrgUserRoleVappAuthor,
			secondRole: govcd.OrgUserRoleVappUser,
		},
		{
			name:       "test_user_vapp_user",
			roleName:   govcd.OrgUserRoleVappUser,
			secondRole: govcd.OrgUserRoleConsoleAccessOnly,
		},
		{
			name:       "test_user_console_access",
			roleName:   govcd.OrgUserRoleConsoleAccessOnly,
			secondRole: govcd.OrgUserRoleCatalogAuthor,
		},
		{
			name:       "test_user_catalog_author",
			roleName:   govcd.OrgUserRoleCatalogAuthor,
			secondRole: govcd.OrgUserRoleOrganizationAdministrator,
		},
	}
}

func TestAccVcdOrgUserBasic(t *testing.T) {

	userData := prepareUserData(t)
	willSkipTests := false

	for _, ud := range userData {

		var params = StringMap{
			"Org":          testConfig.VCD.Org,
			"UserName":     ud.name,
			"PasswordFile": orgUserPasswordFile,
			"RoleName":     ud.roleName,
			"Tags":         "user",
			"FuncName":     "TestUser_" + ud.name + "_basic",
		}
		configText := templateFill(testAccOrgUserBasic, params)
		if vcdShortTest {
			willSkipTests = true
		} else {
			fmt.Printf("%s (%s)\n", ud.name, ud.roleName)
			debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
			resource.Test(t, resource.TestCase{
				PreCheck:     func() { testAccPreCheck(t) },
				Providers:    testAccProviders,
				CheckDestroy: nil,
				Steps: []resource.TestStep{
					resource.TestStep{
						Config: configText,
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(
								"vcd_org_user."+ud.name, "name", ud.name),
							resource.TestCheckResourceAttr(
								"vcd_org_user."+ud.name, "role", ud.roleName),
							// The following values are set by default
							resource.TestCheckResourceAttr(
								"vcd_org_user."+ud.name, "provider_type", govcd.OrgUserProviderIntegrated),
							resource.TestCheckResourceAttr(
								"vcd_org_user."+ud.name, "is_enabled", "false"),
							resource.TestCheckResourceAttr(
								"vcd_org_user."+ud.name, "deployed_vm_quota", "10"),
							resource.TestCheckResourceAttr(
								"vcd_org_user."+ud.name, "stored_vm_quota", "10"),
						),
					},
				},
			})
		}
	}
	if willSkipTests {
		t.Skip(acceptanceTestsSkipped)
		return
	}
}

func TestAccVcdOrgUserFull(t *testing.T) {

	userData := prepareUserData(t)
	willSkipTests := false

	storedQuota := 10
	deployedQuota := 10
	for _, ud := range userData {

		storedQuota += 2
		deployedQuota += 3
		var params = StringMap{
			"Org":             testConfig.VCD.Org,
			"UserName":        ud.name,
			"OrgUserPassword": orgUserPasswordText,
			"RoleName":        ud.roleName,
			"ProviderType":    govcd.OrgUserProviderIntegrated,
			"StoredVmQuota":   storedQuota,
			"DeployedVmQuota": deployedQuota,
			"FullName":        strings.ReplaceAll(ud.name, "_", " "),
			"Description":     "Org user " + ud.name,
			"EmailAddress":    ud.name + "@test.company.org",
			"IsEnabled":       true,
			"IM":              "@" + ud.name,
			"Tags":            "user",
			"FuncName":        "TestUser_" + ud.name + "_full",
		}
		configText := templateFill(testAccOrgUserFull, params)

		// Prepare update
		var updateParams = make(StringMap)

		for k, v := range params {
			updateParams[k] = v
		}
		updateParams["FullName"] = strings.ReplaceAll(ud.name, "_", "==")
		updateParams["Description"] = "Org User updated " + ud.name
		updateParams["DeployedVmQuota"] = params["DeployedVmQuota"].(int) + 1
		updateParams["StoredVmQuota"] = params["StoredVmQuota"].(int) + 1
		updateParams["RoleName"] = ud.secondRole
		updateParams["FuncName"] = params["FuncName"].(string) + "_updated"

		configTextUpdated := templateFill(testAccOrgUserFull, updateParams)

		if vcdShortTest {
			willSkipTests = true
		} else {
			fmt.Printf("%s (%s)\n", ud.name, ud.roleName)
			debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
			debugPrintf("#[DEBUG] UPDATED CONFIGURATION: %s", configTextUpdated)
			resourceName := "vcd_org_user." + ud.name
			resource.Test(t, resource.TestCase{
				PreCheck:     func() { testAccPreCheck(t) },
				Providers:    testAccProviders,
				CheckDestroy: testAccCheckVcdUserDestroy(ud.name),
				Steps: []resource.TestStep{
					resource.TestStep{
						Config: configText,
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(
								resourceName, "name", ud.name),
							resource.TestCheckResourceAttr(
								resourceName, "role", ud.roleName),
							resource.TestCheckResourceAttr(
								resourceName, "provider_type", params["ProviderType"].(string)),
							resource.TestCheckResourceAttr(
								resourceName, "is_enabled", fmt.Sprintf("%v", params["IsEnabled"].(bool))),
							resource.TestCheckResourceAttr(
								resourceName, "email_address", params["EmailAddress"].(string)),
							resource.TestCheckResourceAttr(
								resourceName, "full_name", params["FullName"].(string)),
							resource.TestCheckResourceAttr(
								resourceName, "description", params["Description"].(string)),
							resource.TestCheckResourceAttr(
								resourceName, "instant_messaging", params["IM"].(string)),
							resource.TestCheckResourceAttr(
								resourceName, "deployed_vm_quota", fmt.Sprintf("%d", params["DeployedVmQuota"].(int))),
							resource.TestCheckResourceAttr(
								resourceName, "stored_vm_quota", fmt.Sprintf("%d", params["StoredVmQuota"].(int))),
						),
					},
					resource.TestStep{
						Config: configTextUpdated,
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(
								resourceName, "role", ud.secondRole),
							resource.TestCheckResourceAttr(
								resourceName, "full_name", updateParams["FullName"].(string)),
							resource.TestCheckResourceAttr(
								resourceName, "description", updateParams["Description"].(string)),
							resource.TestCheckResourceAttr(
								resourceName, "deployed_vm_quota", fmt.Sprintf("%d", updateParams["DeployedVmQuota"].(int))),
							resource.TestCheckResourceAttr(
								resourceName, "stored_vm_quota", fmt.Sprintf("%d", updateParams["StoredVmQuota"].(int))),
						),
					},
					resource.TestStep{
						ResourceName:      resourceName + "-import",
						ImportState:       true,
						ImportStateVerify: true,
						ImportStateIdFunc: importStateIdByOrgUser(testConfig, ud.name),
						// These fields can't be retrieved from user data
						ImportStateVerifyIgnore: []string{"take_ownership", "password", "password_file"},
					},
				},
			})
		}
	}
	if willSkipTests {
		t.Skip(acceptanceTestsSkipped)
		return
	}
}

func testAccCheckVcdUserDestroy(userName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)
		adminOrg, err := conn.GetAdminOrg(testConfig.VCD.Org)
		if err != nil {
			return err
		}
		user, err := adminOrg.FetchUserByName(userName, false)
		if err != govcd.ErrorEntityNotFound {
			return fmt.Errorf("user %s was not destroyed", userName)
		}
		if user != nil {
			return fmt.Errorf("user %s was found in %s ", userName, adminOrg.AdminOrg.Name)
		}
		return nil
	}
}

func importStateIdByOrgUser(vcd TestConfig, objectName string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		importId := testConfig.VCD.Org + "." + objectName
		if testConfig.VCD.Org == "" || objectName == "" {
			return "", fmt.Errorf("missing information to generate import path: %s", importId)
		}

		return importId, nil
	}
}

func init() {
	testingTags["user"] = "resource_vcd_org_user_test.go"
}

const testAccOrgUserBasic = `
# skip-binary-test: depends on external file
resource "vcd_org_user" "{{.UserName}}" {
  org            = "{{.Org}}"
  name           = "{{.UserName}}"
  password_file  = "{{.PasswordFile}}"
  role           = "{{.RoleName}}"
  take_ownership = true
}
`

const testAccOrgUserFull = `
resource "vcd_org_user" "{{.UserName}}" {
  org               = "{{.Org}}"
  name              = "{{.UserName}}"
  password          = "{{.OrgUserPassword}}"
  full_name         = "{{.FullName}}"
  description       = "{{.Description}}"
  role              = "{{.RoleName}}"
  is_enabled        = {{.IsEnabled}}
  take_ownership    = true
  provider_type     = "{{.ProviderType}}"
  stored_vm_quota   = {{.StoredVmQuota}}
  deployed_vm_quota = {{.DeployedVmQuota}}
  instant_messaging = "{{.IM}}"
  email_address     = "{{.EmailAddress}}"
}
`
