//go:build vdc || catalog || vapp || ALL || functional
// +build vdc catalog vapp ALL functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdAnyAccessControlGroups checks that vcd_org_vdc_access_control resource is capable of
// correctly handling group IDs correctly.
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
		CheckDestroy:      testAccCheckVDCControlAccessDestroy(),
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
