// +build org ALL functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

// Cloning an organization using an existing organization as data source
func TestAccVcdDatasourceOrg(t *testing.T) {

	if !usingSysAdmin() {
		t.Skip("TestAccVcdDatasourceOrg requires system admin privileges")
		return
	}
	type testOrgData struct {
		name               string
		enabled            bool
		canPublishCatalogs bool
		deployedVmQuota    int
		storedVmQuota      int
	}

	orgName1 := testConfig.VCD.Org
	orgName2 := orgName1 + "-clone"
	var params = StringMap{
		"FuncName": "TestAccVcdDatasourceOrg",
		"OrgName1": orgName1,
		"OrgName2": orgName2,
		"Tags":     "org",
	}

	configText := templateFill(testAccCheckVcdDatasourceOrg, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	datasource1 := "data.vcd_org." + orgName1
	resourceName2 := "vcd_org." + orgName2
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOrgDestroy(orgName2),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdOrgExists(resourceName2),
					resource.TestCheckResourceAttr(
						resourceName2, "name", orgName2),
					resource.TestCheckResourceAttrPair(
						datasource1, "deployed_vm_quota", resourceName2, "deployed_vm_quota"),
					resource.TestCheckResourceAttrPair(
						datasource1, "stored_vm_quota", resourceName2, "stored_vm_quota"),
					resource.TestCheckResourceAttrPair(
						datasource1, "full_name", resourceName2, "full_name"),
					resource.TestCheckResourceAttrPair(
						datasource1, "deployed_vm_quota", resourceName2, "deployed_vm_quota"),
					resource.TestCheckResourceAttrPair(
						datasource1, "is_enabled", resourceName2, "is_enabled"),
					resource.TestCheckResourceAttrPair(
						datasource1, "can_publish_catalogs", resourceName2, "can_publish_catalogs"),
				),
			},
		},
	})
}

const testAccCheckVcdDatasourceOrg = `
data "vcd_org" "{{.OrgName1}}" {
  name = "{{.OrgName1}}"
}

resource "vcd_org" "{{.OrgName2}}" {
  name                 = "{{.OrgName2}}"
  full_name            = "${data.vcd_org.{{.OrgName1}}.full_name}"
  can_publish_catalogs = "${data.vcd_org.{{.OrgName1}}.can_publish_catalogs}"
  deployed_vm_quota    = "${data.vcd_org.{{.OrgName1}}.deployed_vm_quota}"
  stored_vm_quota      = "${data.vcd_org.{{.OrgName1}}.stored_vm_quota}"
  is_enabled           = "${data.vcd_org.{{.OrgName1}}.is_enabled}"
  delete_force         = "true"
  delete_recursive     = "true"
}
`
