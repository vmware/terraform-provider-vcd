// +build org ALL functional

package vcd

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

const orgNameTestAccVcdOrg string = "TestAccVcdOrg"

func TestAccVcdOrgBasic(t *testing.T) {

	var params = StringMap{
		"OrgName":     orgNameTestAccVcdOrg,
		"FuncName":    "TestAccVcdOrgBasic",
		"FullName":    "Full " + orgNameTestAccVcdOrg,
		"Description": "Organization " + orgNameTestAccVcdOrg,
		"Tags":        "org",
	}

	if !usingSysAdmin() {
		t.Skip("TestAccVcdOrgBasic requires system admin privileges")
		return
	}

	configText := templateFill(testAccCheckVcdOrgBasic, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resourceName := "vcd_org." + orgNameTestAccVcdOrg
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOrgDestroy(orgNameTestAccVcdOrg),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdOrgExists("vcd_org."+orgNameTestAccVcdOrg),
					resource.TestCheckResourceAttr(
						resourceName, "name", orgNameTestAccVcdOrg),
					resource.TestCheckResourceAttr(
						resourceName, "full_name", params["FullName"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "description", params["Description"].(string)),
					resource.TestCheckResourceAttr(
						resourceName, "is_enabled", "true"),
				),
			},
		},
	})
}
func TestAccVcdOrgFull(t *testing.T) {

	if !usingSysAdmin() {
		t.Skip("TestAccVcdOrgFull requires system admin privileges")
		return
	}
	type testOrgData struct {
		name               string
		enabled            bool
		canPublishCatalogs bool
		deployedVmQuota    int
		storedVmQuota      int
	}
	var orgList = []testOrgData{
		{
			name:               "org1",
			enabled:            true,
			canPublishCatalogs: false,
			deployedVmQuota:    0,
			storedVmQuota:      0,
		},
		{
			name:               "org2",
			enabled:            false,
			canPublishCatalogs: true,
			deployedVmQuota:    1,
			storedVmQuota:      1,
		},
		{
			name:               "org3",
			enabled:            true,
			canPublishCatalogs: true,
			deployedVmQuota:    10,
			storedVmQuota:      10,
		},
		{
			name:               "org4",
			enabled:            false,
			canPublishCatalogs: false,
			deployedVmQuota:    100,
			storedVmQuota:      100,
		},
		{
			name:               "org5",
			enabled:            true,
			canPublishCatalogs: true,
			deployedVmQuota:    200,
			storedVmQuota:      200,
		},
	}
	willSkip := false

	for _, od := range orgList {

		var params = StringMap{
			"FuncName":           "TestAccVcdOrgFull" + "_" + od.name,
			"OrgName":            od.name,
			"FullName":           "Full " + od.name,
			"Description":        "Organization " + od.name,
			"CanPublishCatalogs": od.canPublishCatalogs,
			"DeployedVmQuota":    od.deployedVmQuota,
			"StoredVmQuota":      od.storedVmQuota,
			"IsEnabled":          od.enabled,
			"Tags":               "org",
		}

		configText := templateFill(testAccCheckVcdOrgFull, params)
		// Prepare update
		var updateParams = make(StringMap)

		for k, v := range params {
			updateParams[k] = v
		}
		updateParams["DeployedVmQuota"] = params["DeployedVmQuota"].(int) + 1
		updateParams["StoredVmQuota"] = params["StoredVmQuota"].(int) + 1
		updateParams["FuncName"] = params["FuncName"].(string) + "_updated"
		updateParams["FullName"] = params["FullName"].(string) + " updated"
		updateParams["Description"] = params["Description"].(string) + " updated"
		updateParams["CanPublishCatalogs"] = !params["CanPublishCatalogs"].(bool)
		updateParams["IsEnabled"] = !params["IsEnabled"].(bool)

		configTextUpdated := templateFill(testAccCheckVcdOrgFull, updateParams)
		if vcdShortTest {
			willSkip = true
			continue
		}
		fmt.Printf("org: %-10s - enabled %-5v - catalogs %-5v - quotas [%3d %3d]\n",
			od.name, od.enabled, od.canPublishCatalogs, od.deployedVmQuota, od.storedVmQuota)
		debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
		debugPrintf("#[DEBUG] CONFIGURATION: %s", configTextUpdated)

		resourceName := "vcd_org." + od.name
		resource.Test(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckOrgDestroy(od.name),
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: configText,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckVcdOrgExists("vcd_org."+od.name),
						resource.TestCheckResourceAttr(
							resourceName, "name", od.name),
						resource.TestCheckResourceAttr(
							resourceName, "full_name", params["FullName"].(string)),
						resource.TestCheckResourceAttr(
							resourceName, "description", params["Description"].(string)),
						resource.TestCheckResourceAttr(
							resourceName, "is_enabled", fmt.Sprintf("%v", od.enabled)),
						resource.TestCheckResourceAttr(
							resourceName, "can_publish_catalogs", fmt.Sprintf("%v", od.canPublishCatalogs)),
						resource.TestCheckResourceAttr(
							resourceName, "deployed_vm_quota", fmt.Sprintf("%d", od.deployedVmQuota)),
						resource.TestCheckResourceAttr(
							resourceName, "stored_vm_quota", fmt.Sprintf("%d", od.storedVmQuota)),
					),
				},
				resource.TestStep{
					Config: configTextUpdated,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(
							resourceName, "name", od.name),
						resource.TestCheckResourceAttr(
							resourceName, "full_name", updateParams["FullName"].(string)),
						resource.TestCheckResourceAttr(
							resourceName, "is_enabled", fmt.Sprintf("%v", !od.enabled)),
						resource.TestCheckResourceAttr(
							resourceName, "description", updateParams["Description"].(string)),
						resource.TestCheckResourceAttr(
							resourceName, "can_publish_catalogs", fmt.Sprintf("%v", !od.canPublishCatalogs)),
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
					ImportStateIdFunc: importStateIdByOrg(od.name),
					// These fields can't be retrieved from user data
					ImportStateVerifyIgnore: []string{"delete_force", "delete_recursive"},
				},
			},
		})
	}
	if willSkip {
		t.Skip(acceptanceTestsSkipped)
		return

	}
}

func importStateIdByOrg(objectName string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		return objectName, nil
	}
}

func testAccCheckVcdOrgExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Org ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		orgName := rs.Primary.Attributes["name"]

		_, err := conn.VCDClient.GetAdminOrgByName(orgName)
		if err != nil {
			return fmt.Errorf("error: could not find Org: %v", err)
		}

		return nil
	}
}

func testAccCheckOrgDestroy(orgName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)
		var org *govcd.AdminOrg
		var err error
		for N := 0; N < 30; N++ {
			org, err = conn.GetAdminOrgByName(orgName)
			if err != nil && org == nil {
				break
			}
			time.Sleep(time.Second)
		}
		if err != govcd.ErrorEntityNotFound {
			return fmt.Errorf("org %s was not destroyed", orgName)
		}
		if org != nil {
			return fmt.Errorf("org %s was found", orgName)
		}
		return nil
	}
}

func init() {
	testingTags["org"] = "resource_vcd_org_test.go"
}

const testAccCheckVcdOrgBasic = `
resource "vcd_org" "{{.OrgName}}" {
  name              = "{{.OrgName}}"
  full_name         = "{{.FullName}}"
  description       = "{{.Description}}"
  delete_force      = "true"
  delete_recursive  = "true"
}
`

const testAccCheckVcdOrgFull = `
resource "vcd_org" "{{.OrgName}}" {
  name                 = "{{.OrgName}}"
  full_name            = "{{.FullName}}"
  description          = "{{.Description}}"
  can_publish_catalogs = "{{.CanPublishCatalogs}}"
  deployed_vm_quota    = {{.DeployedVmQuota}}
  stored_vm_quota      = {{.StoredVmQuota}}
  is_enabled           = "{{.IsEnabled}}"
  delete_force         = "true"
  delete_recursive     = "true"
}
`
