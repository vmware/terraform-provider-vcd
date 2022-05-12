//go:build vdc || ALL || functional
// +build vdc ALL functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"strings"
	"testing"
)

func TestAccVcdOrgVdcAccessControl(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
	}

	userName1 := strings.ToLower(t.Name())
	userName2 := strings.ToLower(t.Name()) + "2"
	accessControlName := "test-access-control"

	var params = StringMap{
		"Org":               testConfig.VCD.Org,
		"Vdc":               testConfig.VCD.Vdc,
		"AccessControlName": accessControlName,
		"UserName":          userName1,
		"UserName2":         userName2,
		"PasswordFile":      orgUserPasswordFile,
		"RoleName":          govcd.OrgUserRoleOrganizationAdministrator,
	}

	params["FuncName"] = t.Name() + "step1"
	configText := templateFill(testAccCheckVcdAccessControlStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	params["FuncName"] = t.Name() + "step2"
	configTex2 := templateFill(testAccCheckVcdAccessControlStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVDCControlAccessDestroy(),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					assertVdcAccessControlIsSharedWithEverybody(),
				),
			},
			{
				Config: configTex2,
				Check: resource.ComposeTestCheckFunc(
					assertVdcAccessControlIsSharedWithSpecificUser(userName1),
					assertVdcAccessControlIsSharedWithSpecificUser(userName2),
				),
			},
			{
				ResourceName:      fmt.Sprintf("vcd_org_vdc_access_control.%s", accessControlName),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     testConfig.VCD.Org + "." + testConfig.VCD.Vdc,
			},
		},
	})
}

const testAccCheckVcdAccessControlStep1 = `
resource "vcd_org_vdc_access_control" "{{.AccessControlName}}" {
  org                   = "{{.Org}}"
  vdc                   = "{{.Vdc}}"
  shared_with_everyone  = true
  everyone_access_level = "ReadOnly"
}
`

const testAccCheckVcdAccessControlStep2 = `
resource "vcd_org_user" "{{.UserName}}" {
  org            = "{{.Org}}"
  name           = "{{.UserName}}"
  password_file  = "{{.PasswordFile}}"
  role           = "{{.RoleName}}"
  take_ownership = true
}

resource "vcd_org_user" "{{.UserName2}}" {
  org            = "{{.Org}}"
  name           = "{{.UserName2}}"
  password_file  = "{{.PasswordFile}}"
  role           = "{{.RoleName}}"
  take_ownership = true
}

resource "vcd_org_vdc_access_control" "{{.AccessControlName}}" {
  org                   = "{{.Org}}"
  vdc                   = "{{.Vdc}}"
  shared_with_everyone  = false

  shared_with {
    user_id             = vcd_org_user.{{.UserName}}.id
    access_level        = "ReadOnly"
  }

  shared_with {
    user_id             = vcd_org_user.{{.UserName2}}.id
    access_level        = "ReadOnly"
  }
}
`

func testAccCheckVDCControlAccessDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
		if err != nil {
			return fmt.Errorf("error retrieving Org %s - %s", testConfig.VCD.Org, err)
		}

		controlAccessParams, err := vdc.GetControlAccess(true)
		if err != nil {
			return fmt.Errorf("error retrieving VDC controll access parameters - %s", err)
		}

		if controlAccessParams.IsSharedToEveryone || controlAccessParams.AccessSettings != nil {
			return fmt.Errorf("expected to have VDC sharing settings set to none and got something else")
		}
		return nil
	}
}

func assertVdcAccessControlIsSharedWithEverybody() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
		if err != nil {
			return fmt.Errorf("error retrieving Org %s - %s", testConfig.VCD.Org, err)
		}

		controlAccessParams, err := vdc.GetControlAccess(true)
		if err != nil {
			return fmt.Errorf("error retrieving VDC controll access parameters - %s", err)
		}

		if !controlAccessParams.IsSharedToEveryone {
			fmt.Errorf("this VDC was expected to be shared with everyone but it is not")
		}

		return nil
	}
}

func assertVdcAccessControlIsSharedWithSpecificUser(userName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
		if err != nil {
			return fmt.Errorf("error retrieving Org %s - %s", testConfig.VCD.Org, err)
		}

		controlAccessParams, err := vdc.GetControlAccess(true)
		if err != nil {
			return fmt.Errorf("error retrieving VDC controll access parameters - %s", err)
		}

		if controlAccessParams.AccessSettings == nil {
			return fmt.Errorf("there are not users configured for sharing in this VDC and they were expected to be")
		}

		for _, accessControlEntry := range controlAccessParams.AccessSettings.AccessSetting {
			if accessControlEntry.Subject.Name == userName {
				return nil
			}
		}

		return fmt.Errorf("userName %s wasn't found in VDC %s and it was expected to be", userName, vdc.Vdc.Name)
	}
}
