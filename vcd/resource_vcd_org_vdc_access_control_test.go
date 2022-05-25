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

	orgUserPasswordFile := "org_user_pwd.txt"
	userName1 := strings.ToLower(t.Name())
	userName2 := strings.ToLower(t.Name()) + "2"
	accessControlName := "test-access-control"
	newVdcName := t.Name()

	var params = StringMap{
		"Org":                       testConfig.VCD.Org,
		"Vdc":                       testConfig.VCD.Vdc,
		"AccessControlName":         accessControlName,
		"AccessControlName2":        accessControlName + "2",
		"UserName":                  userName1,
		"UserName2":                 userName2,
		"PasswordFile":              orgUserPasswordFile,
		"RoleName":                  govcd.OrgUserRoleOrganizationAdministrator,
		"NewVdcName":                newVdcName,
		"NetworkPool":               testConfig.VCD.ProviderVdc.NetworkPool,
		"ProviderVdc":               testConfig.VCD.ProviderVdc.Name,
		"Allocated":                 "1024",
		"Limit":                     "1024",
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
	}

	params["FuncName"] = t.Name() + "step1"
	configText := templateFill(testAccCheckVcdAccessControlStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	params["FuncName"] = t.Name() + "step2"
	configTex2 := templateFill(testAccCheckVcdAccessControlStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	params["FuncName"] = t.Name() + "step3"
	configText3 := templateFill(testAccCheckVcdAccessControlStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	params["FuncName"] = t.Name() + "step4"
	configText4 := templateFill(testAccCheckVcdAccessControlStep4, params)
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
					assertVdcAccessControlIsSharedWithEverybody(testConfig.VCD.Vdc),
				),
			},
			{
				Config: configTex2,
				Check: resource.ComposeTestCheckFunc(
					assertVdcAccessControlIsSharedWithSpecificUser(userName1, testConfig.VCD.Vdc),
					assertVdcAccessControlIsSharedWithSpecificUser(userName2, testConfig.VCD.Vdc),
				),
			},
			{
				ResourceName:      fmt.Sprintf("vcd_org_vdc_access_control.%s", accessControlName),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     testConfig.VCD.Org + "." + testConfig.VCD.Vdc,
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					assertVdcAccessControlIsSharedWithEverybody(newVdcName),
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeTestCheckFunc(
					assertVdcAccessControlIsSharedWithSpecificUser(userName1, newVdcName),
					assertVdcAccessControlIsSharedWithSpecificUser(userName2, newVdcName),
				),
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

const testAccCheckVcdAccessControlStep3 = `
resource "vcd_org_vdc" "{{.NewVdcName}}" {
  name = "{{.NewVdcName}}"
  org  = "{{.Org}}"

  allocation_model  = "Flex"
  network_pool_name = "{{.NetworkPool}}"
  provider_vdc_name = "{{.ProviderVdc}}"

  compute_capacity {
    cpu {
      allocated = "{{.Allocated}}"
      limit     = "{{.Limit}}"
    }

    memory {
      allocated = "{{.Allocated}}"
      limit     = "{{.Limit}}"
    }
  }

  storage_profile {
    name    = "{{.ProviderVdcStorageProfile}}"
    enabled = true
    limit   = 10240
    default = true
  }

  enabled                    = true
  enable_thin_provisioning   = true
  enable_fast_provisioning   = true
  delete_force               = true
  delete_recursive           = true
  elasticity                 = false
  include_vm_memory_overhead = false
}

resource "vcd_org_vdc_access_control" "{{.AccessControlName2}}" {
  org                   = "{{.Org}}"
  vdc                   = vcd_org_vdc.{{.NewVdcName}}.name
  shared_with_everyone  = true
  everyone_access_level = "ReadOnly"

  depends_on            = [ vcd_org_vdc.{{.NewVdcName}} ]
}
`

const testAccCheckVcdAccessControlStep4 = `
resource "vcd_org_vdc" "{{.NewVdcName}}" {
  name = "{{.NewVdcName}}"
  org  = "{{.Org}}"

  allocation_model  = "Flex"
  network_pool_name = "{{.NetworkPool}}"
  provider_vdc_name = "{{.ProviderVdc}}"

  compute_capacity {
    cpu {
      allocated = "{{.Allocated}}"
      limit     = "{{.Limit}}"
    }

    memory {
      allocated = "{{.Allocated}}"
      limit     = "{{.Limit}}"
    }
  }

  storage_profile {
    name    = "{{.ProviderVdcStorageProfile}}"
    enabled = true
    limit   = 10240
    default = true
  }

  enabled                    = true
  enable_thin_provisioning   = true
  enable_fast_provisioning   = true
  delete_force               = true
  delete_recursive           = true
  elasticity                 = false
  include_vm_memory_overhead = false
}

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

resource "vcd_org_vdc_access_control" "{{.AccessControlName2}}" {
  org                   = "{{.Org}}"
  vdc                   = vcd_org_vdc.{{.NewVdcName}}.name
  shared_with_everyone  = false
  shared_with {
    user_id             = vcd_org_user.{{.UserName}}.id
    access_level        = "ReadOnly"
  }
  shared_with {
    user_id             = vcd_org_user.{{.UserName2}}.id
    access_level        = "ReadOnly"
  }
  
  depends_on            = [ vcd_org_vdc.{{.NewVdcName}} ]
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

func assertVdcAccessControlIsSharedWithEverybody(vdcName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, vdcName)
		if err != nil {
			return fmt.Errorf("error retrieving Org %s - %s", testConfig.VCD.Org, err)
		}

		controlAccessParams, err := vdc.GetControlAccess(true)
		if err != nil {
			return fmt.Errorf("error retrieving VDC controll access parameters - %s", err)
		}

		if !controlAccessParams.IsSharedToEveryone {
			return fmt.Errorf("this VDC was expected to be shared with everyone but it is not")
		}

		return nil
	}
}

func assertVdcAccessControlIsSharedWithSpecificUser(userName string, vdcName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, vdcName)
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
