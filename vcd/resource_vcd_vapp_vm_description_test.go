//go:build vapp || vm || ALL || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdVAppVmDescriptionHotUpdate checks that changing only the description of a powered on VM
// without CPU and memory hot add does not power it off. It needs neither networks nor a second
// storage profile.
func TestAccVcdVAppVmDescriptionHotUpdate(t *testing.T) {
	preTestChecks(t)
	vappName := t.Name()
	vmName := t.Name() + "VM"
	resourceName := "vcd_vapp_vm.vm"
	params := StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"Catalog":     testSuiteCatalogName,
		"CatalogItem": testSuiteCatalogOVAItem,
		"VAppName":    vappName,
		"VMName":      vmName,
		"Description": "initial description",
		"Tags":        "vapp vm",
	}
	testParamsNotEmpty(t, params)
	configText := templateFill(testAccVcdVAppVmDescriptionHotUpdate, params)
	params["FuncName"] = t.Name() + "-step1"
	params["Description"] = "updated description"
	configTextUpdate := templateFill(testAccVcdVAppVmDescriptionHotUpdate, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdVAppVmDestroy(vappName),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "initial description"),
					resource.TestCheckResourceAttr(resourceName, "cpu_hot_add_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "memory_hot_add_enabled", "false"),
				),
			},
			{
				Config: configTextUpdate,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "updated description"),
					resource.TestCheckResourceAttr(resourceName, "power_on", "true"),
					resource.TestCheckResourceAttr(resourceName, "cpu_hot_add_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "memory_hot_add_enabled", "false"),
					testAccCheckVcdVmNotRestarted(resourceName, vappName, vmName),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdVAppVmDescriptionHotUpdate = `
# skip-binary-test: update-only description regression test
resource "vcd_vapp" "description" {
  org      = "{{.Org}}"
  vdc      = "{{.Vdc}}"
  name     = "{{.VAppName}}"
  power_on = false
}

resource "vcd_vapp_vm" "vm" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  vapp_name     = vcd_vapp.description.name
  name          = "{{.VMName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  description   = "{{.Description}}"
  power_on      = true

  cpu_hot_add_enabled      = false
  memory_hot_add_enabled   = false
  prevent_update_power_off = true
}
`
