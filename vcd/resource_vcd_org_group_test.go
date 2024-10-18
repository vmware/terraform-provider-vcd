//go:build user || ldap || functional || ALL

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v3/govcd"
)

// TestAccVcdOrgGroup tests INTEGRATED (LDAP) group management in Terraform.
// In step 0 it configures the LDAP identity provider using vcd_org_ldap
// LDAP configuration will be removed automatically after test run
//
// Note: This test requires an existing LDAP server and its IP set in testConfig.Networking.LdapServer
func TestAccVcdOrgGroup(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)
	skipTestForServiceAccountAndApiToken(t)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	if testConfig.Networking.LdapServer == "" {
		t.Skip("TestAccVcdOrgGroup requires a working LDAP server (set the IP in testConfig.Networking.LdapServer)")
		return
	}

	testParamsNotEmpty(t, StringMap{"Networking.ExternalNetwork": testConfig.Networking.ExternalNetwork,
		"VCD.Catalog.CatalogItem": testConfig.VCD.Catalog.CatalogItem, "VCD.Catalog.Name": testConfig.VCD.Catalog.Name})

	role1 := govcd.OrgUserRoleOrganizationAdministrator
	role2 := govcd.OrgUserRoleVappAuthor

	var params = StringMap{
		"OrgName":      testConfig.VCD.Org,
		"LdapServerIp": testConfig.Networking.LdapServer,
		"ProviderType": "INTEGRATED",
		"RoleName":     role1,
		"Tags":         "user",
		"FuncName":     t.Name() + "-Step0",
		"Description":  "Description1",
	}
	testParamsNotEmpty(t, params)

	ldapSetupConfig := templateFill(testAccOrgLdap, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 0 (LDAP server configuration): %s", ldapSetupConfig)

	params["FuncName"] = t.Name() + "-Step1"
	groupConfigText := templateFill(testAccOrgLdap+testAccOrgGroup, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", groupConfigText)

	params["FuncName"] = t.Name() + "-Step2"
	params["RoleName"] = role2
	params["Description"] = "Description2"
	groupConfigText2 := templateFill(testAccOrgLdap+testAccOrgGroup, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", groupConfigText2)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// groupIdRegex is reused a few times in tests to match IDs
	groupIdRegex := regexp.MustCompile(`^urn:vcloud:group:`)
	ldapResourceDef := "vcd_org_ldap." + testConfig.VCD.Org
	// Note: don't run this test in parallel, as it would clash with TestAccVcdOrgLdap
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVcdGroupDestroy("admin_staff"),
			testAccCheckVcdGroupDestroy("ship_crew"),
			testAccCheckOrgLdapDestroy(ldapResourceDef),
		),
		// CheckDestroy: testAccCheckVcdGroupDestroy(params["GroupName"].(string)),
		Steps: []resource.TestStep{
			// Step 0 - uses an existing LDAP server to set up the identity provider
			{
				Config: ldapSetupConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrgLdapExists(ldapResourceDef),
					resource.TestCheckResourceAttr(ldapResourceDef, "ldap_mode", "CUSTOM"),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.server", testConfig.Networking.LdapServer),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.authentication_method", "SIMPLE"),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.connector_type", "OPEN_LDAP"),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.user_attributes.0.object_class", "inetOrgPerson"),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.group_attributes.0.object_class", "group"),
				),
			},
			{
				// Step 1 - tests org group and users
				Config: groupConfigText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_org_group.group1", "id", groupIdRegex),
					resource.TestCheckResourceAttr("vcd_org_group.group1", "name", "ship_crew"),
					resource.TestCheckResourceAttr("vcd_org_group.group1", "role", role1),
					resource.TestCheckResourceAttr("vcd_org_group.group1", "description", "Description1"),
					resource.TestMatchResourceAttr("vcd_org_group.group2", "id", groupIdRegex),
					resource.TestCheckResourceAttr("vcd_org_group.group2", "name", "admin_staff"),
					resource.TestCheckResourceAttr("vcd_org_group.group2", "role", role1),
					resource.TestCheckResourceAttr("vcd_org_group.group2", "description", "Description1"),
					// This check should belong to vcd_org_user tests, but here is simpler and quicker
					resource.TestCheckResourceAttr("vcd_org_user.user1", "group_names.0", "ship_crew"),
				),
			},
			{
				// Step 2 - tests org group and users
				Config: groupConfigText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_org_group.group1", "id", groupIdRegex),
					resource.TestCheckResourceAttr("vcd_org_group.group1", "name", "ship_crew"),
					resource.TestCheckResourceAttr("vcd_org_group.group1", "role", role2),
					resource.TestCheckResourceAttr("vcd_org_group.group1", "description", "Description2"),
					// We check the user_names set here because it's populated when the group state is refreshed. In previous step,
					// it would be nil as it didn't have users.
					resource.TestCheckResourceAttr("vcd_org_group.group1", "user_names.0", "fry"),
					resource.TestMatchResourceAttr("vcd_org_group.group2", "id", groupIdRegex),
					resource.TestCheckResourceAttr("vcd_org_group.group2", "name", "admin_staff"),
					resource.TestCheckResourceAttr("vcd_org_group.group2", "role", role2),
					resource.TestCheckResourceAttr("vcd_org_group.group2", "description", "Description2"),
				),
			},
			{
				Config: groupConfigText2 + testAccVcdOrgGroupDS, // Datasource check
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_org_group.sourced_group1", "vcd_org_group.group1", nil),
					resourceFieldsEqual("data.vcd_org_group.sourced_group2", "vcd_org_group.group2", nil),
					resourceFieldsEqual("data.vcd_org_user.sourced_user1", "vcd_org_user.user1", []string{"%", "user_id"}),
				),
			},
			{
				ResourceName:      "vcd_org_group.group1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgObject(testConfig.VCD.Org, "ship_crew"),
			},
		},
	})
	postTestChecks(t)
}

// testAccCheckVcdGroupDestroy verifies if Org Group with given name does not exist in vCD
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
resource "vcd_org_group" "group1" {
  provider_type = "INTEGRATED"
  name          = "ship_crew"
  role          = "{{.RoleName}}"
  description   = "{{.Description}}"
  depends_on = [
    vcd_org_ldap.{{.OrgName}}
  ]
}

resource "vcd_org_group" "group2" {
  provider_type = "INTEGRATED"
  name          = "admin_staff"
  role          = "{{.RoleName}}"
  description   = "{{.Description}}"
  depends_on = [
    vcd_org_ldap.{{.OrgName}}
  ]
}

resource "vcd_org_user" "user1" {
  provider_type  = "INTEGRATED"
  name           = "fry"
  role           = "Organization Administrator"
  is_external    = true

  depends_on = [
    vcd_org_group.group1,
  ]
}
`

const testAccVcdOrgGroupDS = `
# skip-binary-test: Terraform resource cannot have resource and datasource in the same file

data "vcd_org_group" "sourced_group1" {
  name = vcd_org_group.group1.name
}

data "vcd_org_group" "sourced_group2" {
  name = vcd_org_group.group2.name
}

data "vcd_org_user" "sourced_user1" {
  name = "fry"
}
`
