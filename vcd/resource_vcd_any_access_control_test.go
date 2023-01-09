//go:build access_control || vdc || catalog || vapp || ALL || functional
// +build access_control vdc catalog vapp ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

// TestAccVcdAnyAccessControlGroups checks that vcd_org_vdc_access_control, vcd_vapp_access_control,
// and vcd_catalog_access_control resources are capable of handling group IDs correctly.
func TestAccVcdAnyAccessControlGroups(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"Org":          testConfig.VCD.Org,
		"Vdc":          testConfig.Nsxt.Vdc,
		"Catalog":      testConfig.VCD.Catalog.NsxtBackedCatalogName,
		"TestName":     t.Name(),
		"LdapServerIp": testConfig.Networking.LdapServer,
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	configText := templateFill(testAccVcdAnyAccessControlGroupsLdapStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckVDCControlAccessDestroy(),
			testAccCheckVappAccessControlDestroy(testConfig.VCD.Org, testConfig.Nsxt.Vdc, []string{t.Name()}),
		),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_org_vdc_access_control.test", "id"),
				),
			},
		},
	})
}

const testAccVcdAnyAccessControlGroupsLdap = `
data "vcd_org" "test-org" {
  name = "{{.Org}}"
}

resource "vcd_org_ldap" "test-config" {
  org_id    = data.vcd_org.test-org.id
  ldap_mode = "CUSTOM"
  custom_settings {
    server                  = "{{.LdapServerIp}}"
    port                    = 389
    is_ssl                  = false
    username                = "cn=admin,dc=planetexpress,dc=com"
    password                = "GoodNewsEveryone"
    authentication_method   = "SIMPLE"
    base_distinguished_name = "dc=planetexpress,dc=com"
    connector_type          = "OPEN_LDAP"
    user_attributes {
      object_class                = "inetOrgPerson"
      unique_identifier           = "uid"
      display_name                = "cn"
      username                    = "uid"
      given_name                  = "givenName"
      surname                     = "sn"
      telephone                   = "telephoneNumber"
      group_membership_identifier = "dn"
      email                       = "mail"
    }
    group_attributes {
      name                        = "cn"
      object_class                = "group"
      membership                  = "member"
      unique_identifier           = "cn"
      group_membership_identifier = "dn"
    }
  }
  lifecycle {
    # password value does not get returned by GET
    ignore_changes = [custom_settings[0].password]
  }
}

resource "vcd_org_group" "admin_staff" {
  org = "{{.Org}}"

  provider_type = "INTEGRATED" # LDAP
  name          = "admin_staff"
  role          = "Organization Administrator"

  depends_on = [vcd_org_ldap.test-config]
}

resource "vcd_org_group" "ship_crew" {
  org = "{{.Org}}"

  provider_type = "INTEGRATED" # LDAP
  name          = "ship_crew"
  role          = "Organization Administrator"

  depends_on = [vcd_org_ldap.test-config]
}
`

const testAccVcdAnyAccessControlGroupsLdapStep1 = testAccVcdAnyAccessControlGroupsLdap + `
resource "vcd_org_vdc_access_control" "test" {
  org                   = "{{.Org}}"
  vdc                   = "{{.Vdc}}"
  shared_with_everyone  = false

  shared_with {
	group_id     = vcd_org_group.admin_staff.id
    access_level = "ReadOnly"
  }

  shared_with {
    group_id     = vcd_org_group.ship_crew.id
    access_level = "ReadOnly"
  }
}

data "vcd_catalog" "test" {
  org  = "{{.Org}}"
  name = "{{.Catalog}}"
}

resource "vcd_catalog_access_control" "AC-users-and-orgs" {
  catalog_id = data.vcd_catalog.test.id

  shared_with_everyone = false

  shared_with {
    group_id     = vcd_org_group.admin_staff.id
    access_level = "FullControl"
  }

  shared_with {
    group_id     = vcd_org_group.ship_crew.id
    access_level = "Change"
  }
}

resource "vcd_vapp" "test" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name = "{{.TestName}}"
}

resource "vcd_vapp_access_control" "test" {
  vapp_id = vcd_vapp.test.id

  shared_with_everyone = false

  shared_with {
    group_id     = vcd_org_group.admin_staff.id
    access_level = "FullControl"
  }
  shared_with {
    group_id     = vcd_org_group.ship_crew.id
    access_level = "Change"
  }
}
`

func testAccCheckVDCControlAccessDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.Nsxt.Vdc)
		if err != nil {
			return fmt.Errorf("error retrieving Org %s - %s", testConfig.VCD.Org, err)
		}

		controlAccessParams, err := vdc.GetControlAccess(true)
		if err != nil {
			return fmt.Errorf("error retrieving VDC controll access parameters - %s", err)
		}

		if controlAccessParams.IsSharedToEveryone || controlAccessParams.AccessSettings != nil {
			spew.Dump(controlAccessParams)
			return fmt.Errorf("expected to have VDC sharing settings set to none and got something else: %v", controlAccessParams)
		}
		return nil
	}
}

func testAccCheckVappAccessControlDestroy(orgName, vdcName string, vappNames []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*VCDClient)
		org, err := conn.VCDClient.GetAdminOrgByName(orgName)
		if err != nil {
			return fmt.Errorf("error: could not find Org: %s", err)
		}
		vdc, err := org.GetVDCByName(vdcName, false)
		if err != nil {
			return fmt.Errorf("error: could not find VDC: %s", err)
		}
		var destroyed int
		var existing []string
		for _, vappName := range vappNames {

			_, err = vdc.GetVAppByName(vappName, false)
			if err != nil && govcd.IsNotFound(err) {
				// The vApp was removed
				destroyed++
			}
			if err == nil {
				existing = append(existing, vappName)
			}
		}

		if destroyed == len(vappNames) {
			return nil
		}
		return fmt.Errorf("vapps %v not deleted yet", existing)
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
