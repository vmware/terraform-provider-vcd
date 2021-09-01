//go:build vapp || functional || access_control || ALL
// +build vapp functional access_control ALL

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func TestAccVcdVappAccessControl(t *testing.T) {
	preTestChecks(t)

	if testConfig.VCD.Org == "" {
		t.Skip("[TestAccVcdVappAccessControl] no Org found in configuration")
	}
	if testConfig.VCD.Vdc == "" {
		t.Skip("[TestAccVcdVappAccessControl] no VDC found in configuration")
	}

	var params = StringMap{
		"Org":                      testConfig.VCD.Org,
		"Vdc":                      testConfig.VCD.Vdc,
		"SharedToEveryone":         "true",
		"EveryoneAccessLevel":      fmt.Sprintf(`everyone_access_level = "%s"`, types.ControlAccessReadWrite),
		"AccessControlIdentifier0": "AC-Vapp0",
		"AccessControlIdentifier1": "AC-Vapp1",
		"AccessControlIdentifier2": "AC-Vapp2",
		"AccessControlIdentifier3": "AC-Vapp3",
		"VappName0":                "Vapp-AC-0",
		"VappName1":                "Vapp-AC-1",
		"VappName2":                "Vapp-AC-2",
		"VappName3":                "Vapp-AC-3",
		"UserName1":                "ac-user1",
		"UserName2":                "ac-user2",
		"UserName3":                "ac-user3",
		"RoleName1":                govcd.OrgUserRoleOrganizationAdministrator,
		"RoleName2":                govcd.OrgUserRoleVappAuthor,
		"RoleName3":                govcd.OrgUserRoleCatalogAuthor,
		"AccessLevel1":             types.ControlAccessFullControl,
		"AccessLevel2":             types.ControlAccessReadWrite,
		"AccessLevel3":             types.ControlAccessReadOnly,
		"UserPassword":             "TO_BE_DISCARDED",
		"FuncName":                 "TestAccVcdVappAccessControl",
		"Tags":                     "vapp",
		"SkipNotice":               "",
	}

	configText := templateFill(testAccVappAccessControl, params)
	params["AccessLevel1"] = types.ControlAccessReadWrite
	params["SharedToEveryone"] = "false"
	params["EveryoneAccessLevel"] = ""
	params["FuncName"] = "TestAccVcdVappAccessControl-update"
	params["SkipNotice"] = "# skip-binary-test: only for updates"
	updateText := templateFill(testAccVappAccessControl, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CREATION CONFIGURATION: %s", configText)
	debugPrintf("#[DEBUG] UPDATE CONFIGURATION: %s", updateText)

	resourceAC0 := "vcd_vapp_access_control.AC-Vapp0"
	resourceAC1 := "vcd_vapp_access_control.AC-Vapp1"
	resourceAC2 := "vcd_vapp_access_control.AC-Vapp2"
	resourceAC3 := "vcd_vapp_access_control.AC-Vapp3"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVappAccessControlDestroy(testConfig.VCD.Org, testConfig.VCD.Vdc, []string{"Vapp-AC-0", "Vapp-AC-1", "Vapp-AC-2", "Vapp-AC-3"}),
		Steps: []resource.TestStep{
			// Test creation
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVappAccessControlExists(resourceAC0, testConfig.VCD.Org, testConfig.VCD.Vdc),
					testAccCheckVcdVappAccessControlExists(resourceAC1, testConfig.VCD.Org, testConfig.VCD.Vdc),
					testAccCheckVcdVappAccessControlExists(resourceAC2, testConfig.VCD.Org, testConfig.VCD.Vdc),
					testAccCheckVcdVappAccessControlExists(resourceAC3, testConfig.VCD.Org, testConfig.VCD.Vdc),
					resource.TestCheckResourceAttr(resourceAC0, "shared_with_everyone", "true"),
					resource.TestCheckResourceAttr(resourceAC0, "everyone_access_level", "Change"),
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
				),
			},
			resource.TestStep{
				Config: updateText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVappAccessControlExists(resourceAC0, testConfig.VCD.Org, testConfig.VCD.Vdc),
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
			resource.TestStep{
				Config:            configText,
				ResourceName:      "vcd_vapp_access_control.AC-Vapp1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgVdcObject(testConfig, "Vapp-AC-1"),
			},
			// Tests import by ID
			resource.TestStep{
				Config:            configText,
				ResourceName:      "vcd_vapp_access_control.AC-Vapp2",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdViaResource("vcd_vapp_access_control.AC-Vapp2"),
			},
		},
	})
	postTestChecks(t)
}

func testAccCheckVcdVappAccessControlExists(resourceName string, orgName, vdcName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no vApp ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		org, err := conn.VCDClient.GetAdminOrgByName(orgName)
		if err != nil {
			return fmt.Errorf("error: could not find Org: %s", err)
		}
		vdc, err := org.GetVDCByName(vdcName, false)
		if err != nil {
			return fmt.Errorf("error: could not find VDC: %s", err)
		}

		vapp, err := vdc.GetVAppById(rs.Primary.ID, false)
		if err != nil {
			return fmt.Errorf("error: could not find vApp %s: %s", rs.Primary.ID, err)
		}
		_, err = vapp.GetAccessControl(tenantContext)
		if err != nil {
			return fmt.Errorf("error: could not get access control for vApp %s: %s", vapp.VApp.Name, err)
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

const testAccVappAccessControl = `
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

resource "vcd_vapp" "{{.VappName0}}" {
  name = "{{.VappName0}}"
}

resource "vcd_vapp" "{{.VappName1}}" {
  name = "{{.VappName1}}"
}

resource "vcd_vapp" "{{.VappName2}}" {
  name = "{{.VappName2}}"
}

resource "vcd_vapp" "{{.VappName3}}" {
  name = "{{.VappName3}}"
}

resource "vcd_vapp_access_control" "{{.AccessControlIdentifier0}}" {

  org      = "{{.Org}}"
  vdc      = "{{.Vdc}}"
  vapp_id  = vcd_vapp.{{.VappName0}}.id

  shared_with_everyone    = {{.SharedToEveryone}}
  {{.EveryoneAccessLevel}}
}

resource "vcd_vapp_access_control" "{{.AccessControlIdentifier1}}" {

  org      = "{{.Org}}"
  vdc      = "{{.Vdc}}"
  vapp_id  = vcd_vapp.{{.VappName1}}.id

  shared_with_everyone    = false

  shared_with {
    user_id      = vcd_org_user.{{.UserName1}}.id
    access_level = "{{.AccessLevel1}}"
  }
}

resource "vcd_vapp_access_control" "{{.AccessControlIdentifier2}}" {

  org      = "{{.Org}}"
  vdc      = "{{.Vdc}}"
  vapp_id  = vcd_vapp.{{.VappName2}}.id

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

resource "vcd_vapp_access_control" "{{.AccessControlIdentifier3}}" {

  org      = "{{.Org}}"
  vdc      = "{{.Vdc}}"
  vapp_id  = vcd_vapp.{{.VappName3}}.id

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
`
