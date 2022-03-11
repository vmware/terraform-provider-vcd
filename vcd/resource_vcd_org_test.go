//go:build org || ALL || functional
// +build org ALL functional

package vcd

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

const orgNameTestAccVcdOrg string = "TestAccVcdOrg"

func TestAccVcdOrgBasic(t *testing.T) {
	preTestChecks(t)

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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckOrgDestroy(orgNameTestAccVcdOrg),
		Steps: []resource.TestStep{
			{
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
					// Testing defaults lease values is not reliable, as such values vary for different vCD versions
				),
			},
		},
	})
	postTestChecks(t)
}
func TestAccVcdOrgFull(t *testing.T) {
	preTestChecks(t)

	if !usingSysAdmin() {
		t.Skip("TestAccVcdOrgFull requires system admin privileges")
		return
	}
	type testOrgData struct {
		name                  string
		enabled               bool
		canPublishCatalogs    bool
		deployedVmQuota       int
		storedVmQuota         int
		runtimeLease          int
		powerOffonLeaseExp    bool
		vappStorageLease      int
		vappDeleteOnLeaseExp  bool
		templStorageLease     int
		templDeleteOnLeaseExp bool
	}
	var orgList = []testOrgData{
		{
			name:                  "org1",
			enabled:               true,
			canPublishCatalogs:    false,
			deployedVmQuota:       0,
			storedVmQuota:         0,
			runtimeLease:          0, // never expires
			powerOffonLeaseExp:    true,
			vappStorageLease:      0, // never expires
			templDeleteOnLeaseExp: true,
			templStorageLease:     0, // never expires
			vappDeleteOnLeaseExp:  true,
		},
		{
			name:                  "org2",
			enabled:               false,
			canPublishCatalogs:    true,
			deployedVmQuota:       1,
			storedVmQuota:         1,
			runtimeLease:          3600, // 1 hour
			powerOffonLeaseExp:    true,
			vappStorageLease:      3600, // 1 hour
			templDeleteOnLeaseExp: true,
			templStorageLease:     3600, // 1 hour
			vappDeleteOnLeaseExp:  true,
		},
		{
			name:                  "org3",
			enabled:               true,
			canPublishCatalogs:    true,
			deployedVmQuota:       10,
			storedVmQuota:         10,
			runtimeLease:          3600 * 24, // 1 day
			powerOffonLeaseExp:    false,
			vappStorageLease:      3600 * 24 * 30, // 1 month
			templDeleteOnLeaseExp: false,
			templStorageLease:     3600 * 24 * 365, // 1 year
			vappDeleteOnLeaseExp:  false,
		},
		{
			name:                  "org4",
			enabled:               false,
			canPublishCatalogs:    false,
			deployedVmQuota:       100,
			storedVmQuota:         100,
			runtimeLease:          3600 * 24 * 15, // 15 days
			powerOffonLeaseExp:    false,
			vappStorageLease:      3600 * 24 * 15, // 15 days
			templDeleteOnLeaseExp: false,
			templStorageLease:     3600 * 24 * 15, // 15 days
			vappDeleteOnLeaseExp:  false,
		},
		{
			name:                  "org5",
			enabled:               true,
			canPublishCatalogs:    true,
			deployedVmQuota:       200,
			storedVmQuota:         200,
			runtimeLease:          3600 * 24 * 7, // 7 days (the default)
			powerOffonLeaseExp:    false,
			vappStorageLease:      3600 * 24 * 14, // 14 days (the default)
			templDeleteOnLeaseExp: false,
			templStorageLease:     3600 * 24 * 30, // 30 days (the default)
			vappDeleteOnLeaseExp:  false,
		},
	}
	willSkip := false

	for _, od := range orgList {

		var params = StringMap{
			"FuncName":              "TestAccVcdOrgFull" + "_" + od.name,
			"OrgName":               od.name,
			"FullName":              "Full " + od.name,
			"Description":           "Organization " + od.name,
			"CanPublishCatalogs":    od.canPublishCatalogs,
			"DeployedVmQuota":       od.deployedVmQuota,
			"StoredVmQuota":         od.storedVmQuota,
			"IsEnabled":             od.enabled,
			"RuntimeLease":          od.runtimeLease,
			"PowerOffOnLeaseExp":    od.powerOffonLeaseExp,
			"VappStorageLease":      od.vappStorageLease,
			"VappDeleteOnLeaseExp":  od.vappDeleteOnLeaseExp,
			"TemplStorageLease":     od.templStorageLease,
			"TemplDeleteOnLeaseExp": od.templDeleteOnLeaseExp,
			"Tags":                  "org",
			"MetadataKey":        	"key1",
			"MetadataValue":        "value1",
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
		updateParams["MetadataKey"] = "key3"
		updateParams["MetadataValue"] = "value3"

		configTextUpdated := templateFill(testAccCheckVcdOrgFull, updateParams)
		if vcdShortTest {
			willSkip = true
			continue
		}
		fmt.Printf("org: %-5s - enabled %-5v - catalogs %-5v - quotas [%3d %3d] - vapp {%10d %5v %10d %5v} - tmpl {%10d %5v}\n",
			od.name, od.enabled, od.canPublishCatalogs, od.deployedVmQuota, od.storedVmQuota,
			od.runtimeLease, od.powerOffonLeaseExp, od.vappStorageLease, od.vappDeleteOnLeaseExp,
			od.templStorageLease, od.templDeleteOnLeaseExp)
		debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
		debugPrintf("#[DEBUG] CONFIGURATION: %s", configTextUpdated)

		resourceName := "vcd_org." + od.name
		resource.Test(t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProviders,
			CheckDestroy:      testAccCheckOrgDestroy(od.name),
			Steps: []resource.TestStep{
				{
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
						resource.TestCheckResourceAttr(
							resourceName, "vapp_lease.0.maximum_runtime_lease_in_sec", fmt.Sprintf("%d", od.runtimeLease)),
						resource.TestCheckResourceAttr(
							resourceName, "vapp_lease.0.power_off_on_runtime_lease_expiration", fmt.Sprintf("%v", od.powerOffonLeaseExp)),
						resource.TestCheckResourceAttr(
							resourceName, "vapp_lease.0.maximum_storage_lease_in_sec", fmt.Sprintf("%d", od.vappStorageLease)),
						resource.TestCheckResourceAttr(
							resourceName, "vapp_lease.0.delete_on_storage_lease_expiration", fmt.Sprintf("%v", od.vappDeleteOnLeaseExp)),
						resource.TestCheckResourceAttr(
							resourceName, "vapp_template_lease.0.maximum_storage_lease_in_sec", fmt.Sprintf("%d", od.templStorageLease)),
						resource.TestCheckResourceAttr(
							resourceName, "vapp_template_lease.0.delete_on_storage_lease_expiration", fmt.Sprintf("%v", od.templDeleteOnLeaseExp)),
						resource.TestCheckResourceAttr(
							resourceName, "metadata.key1", "value1"),
						resource.TestCheckResourceAttr(
							resourceName, "metadata.key2", "value2"),
					),
				},
				{
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
						resource.TestCheckResourceAttr(
							resourceName, "metadata.key3", "value3"),
						resource.TestCheckResourceAttr(
							resourceName, "metadata.key2", "value2"),
						stateDumper(),
						sleepTester(),
					),
				},
				{
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,
					ImportStateIdFunc: importStateIdTopHierarchy(od.name),
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
	postTestChecks(t)
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
  vapp_lease {
    maximum_runtime_lease_in_sec          = {{.RuntimeLease}}
    power_off_on_runtime_lease_expiration = {{.PowerOffOnLeaseExp}}
    maximum_storage_lease_in_sec          = {{.VappStorageLease}}
    delete_on_storage_lease_expiration    = {{.VappDeleteOnLeaseExp}}
  }
  vapp_template_lease {
    maximum_storage_lease_in_sec          = {{.TemplStorageLease}}
    delete_on_storage_lease_expiration    = {{.TemplDeleteOnLeaseExp}}
  }
  metadata = {
    {{.MetadataKey}} = "{{.MetadataValue}}"
	key2 = "value2"
  }
}
`

func sleepTester() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fmt.Println("sleeping")
		time.Sleep(1 * time.Minute)
		return nil
	}
}

func stateDumper() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		spew.Dump(s)
		return nil
	}
}
