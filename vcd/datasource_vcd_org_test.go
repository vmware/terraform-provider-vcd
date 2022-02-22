//go:build org || ALL || functional
// +build org ALL functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Cloning an organization using an existing organization as data source
func TestAccVcdDatasourceOrg(t *testing.T) {
	preTestChecks(t)

	if !usingSysAdmin() {
		t.Skip("TestAccVcdDatasourceOrg requires system admin privileges")
		return
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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckOrgDestroy(orgName2),
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
					resource.TestCheckResourceAttrPair(
						datasource1, "delay_after_power_on_seconds", resourceName2, "delay_after_power_on_seconds"),
					resource.TestCheckResourceAttrPair(
						datasource1, "vapp_lease.0.maximum_runtime_lease_in_sec",
						resourceName2, "vapp_lease.0.maximum_runtime_lease_in_sec"),
					resource.TestCheckResourceAttrPair(
						datasource1, "vapp_lease.0.maximum_storage_lease_in_sec",
						resourceName2, "vapp_lease.0.maximum_storage_lease_in_sec"),
					resource.TestCheckResourceAttrPair(
						datasource1, "vapp_lease.0.power_off_on_runtime_lease_expiration",
						resourceName2, "vapp_lease.0.power_off_on_runtime_lease_expiration"),
					resource.TestCheckResourceAttrPair(
						datasource1, "vapp_lease.0.delete_on_storage_lease_expiration",
						resourceName2, "vapp_lease.0.delete_on_storage_lease_expiration"),
					resource.TestCheckResourceAttrPair(
						datasource1, "vapp_template_lease.0.maximum_storage_lease_in_sec",
						resourceName2, "vapp_template_lease.0.maximum_storage_lease_in_sec"),
					resource.TestCheckResourceAttrPair(
						datasource1, "vapp_template_lease.0.delete_on_storage_lease_expiration",
						resourceName2, "vapp_template_lease.0.delete_on_storage_lease_expiration"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdDatasourceOrg = `
data "vcd_org" "{{.OrgName1}}" {
  name = "{{.OrgName1}}"
}

resource "vcd_org" "{{.OrgName2}}" {
  name                         = "{{.OrgName2}}"
  full_name                    = data.vcd_org.{{.OrgName1}}.full_name
  can_publish_catalogs         = data.vcd_org.{{.OrgName1}}.can_publish_catalogs
  deployed_vm_quota            = data.vcd_org.{{.OrgName1}}.deployed_vm_quota
  stored_vm_quota              = data.vcd_org.{{.OrgName1}}.stored_vm_quota
  is_enabled                   = data.vcd_org.{{.OrgName1}}.is_enabled
  delay_after_power_on_seconds = data.vcd_org.{{.OrgName1}}.delay_after_power_on_seconds
  delete_force                 = "true"
  delete_recursive             = "true"
  vapp_lease {
    maximum_runtime_lease_in_sec          = data.vcd_org.{{.OrgName1}}.vapp_lease.0.maximum_runtime_lease_in_sec
    power_off_on_runtime_lease_expiration = data.vcd_org.{{.OrgName1}}.vapp_lease.0.power_off_on_runtime_lease_expiration
    maximum_storage_lease_in_sec          = data.vcd_org.{{.OrgName1}}.vapp_lease.0.maximum_storage_lease_in_sec
    delete_on_storage_lease_expiration    = data.vcd_org.{{.OrgName1}}.vapp_lease.0.delete_on_storage_lease_expiration
  }
  vapp_template_lease {
    maximum_storage_lease_in_sec          = data.vcd_org.{{.OrgName1}}.vapp_template_lease.0.maximum_storage_lease_in_sec
    delete_on_storage_lease_expiration    = data.vcd_org.{{.OrgName1}}.vapp_template_lease.0.delete_on_storage_lease_expiration
  }
}
`
