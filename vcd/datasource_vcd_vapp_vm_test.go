//go:build vm || ALL || functional
// +build vm ALL functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdVappDS tests a VM data source if a vApp + VM is found in the VDC
func TestAccVcdVappVmDS(t *testing.T) {
	preTestChecks(t)
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"VDC":         testConfig.VCD.Vdc,
		"Catalog":     testSuiteCatalogName,
		"CatalogItem": testSuiteCatalogOVAItem,
		"FuncName":    "TestVappVmDS",
		"Tags":        "vm",
	}
	configText := templateFill(datasourceTestVappVm, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vcd_vapp_vm.vm-ds", "name", "web1"),
					resource.TestCheckResourceAttr("data.vcd_vapp_vm.vm-ds", "storage_profile", "*"),
					resource.TestCheckResourceAttrSet("data.vcd_vapp_vm.vm-ds", "href"),
					resource.TestCheckResourceAttrSet("data.vcd_vapp_vm.vm-ds", "description"),
					resource.TestCheckResourceAttr("data.vcd_vapp_vm.vm-ds", "customization.0.enabled", "true"),
					resource.TestCheckResourceAttr("data.vcd_vapp_vm.vm-ds", "customization.0.change_sid", "false"),
					resource.TestCheckResourceAttr("data.vcd_vapp_vm.vm-ds", "customization.0.join_domain", "false"),
					resource.TestCheckResourceAttr("data.vcd_vapp_vm.vm-ds", "customization.0.admin_password", ""),
					resource.TestCheckResourceAttr("data.vcd_vapp_vm.vm-ds", "customization.0.number_of_auto_logons", "0"),
				),
			},
		},
	})
	postTestChecks(t)
}

const datasourceTestVappVm = `
resource "vcd_vapp" "web" {
  name = "web"
}

resource "vcd_vapp_vm" "web1" {
  vapp_name     = vcd_vapp.web.name
  name          = "web1"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 1024
  cpus          = 2
  cpu_cores     = 1
  power_on      = false

  storage_profile = "*"

  customization {
    enabled               = true
    change_sid            = false
    join_domain           = false
    number_of_auto_logons = 0
  }
}

data "vcd_vapp_vm" "vm-ds" {
  name             = vcd_vapp_vm.web1.name
  vapp_name        = vcd_vapp.web.name
  org              = "{{.Org}}"
  vdc              = "{{.VDC}}"
  depends_on       = [vcd_vapp_vm.web1]
}
`
