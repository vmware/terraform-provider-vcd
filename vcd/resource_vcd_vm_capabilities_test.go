//go:build (vm || standaloneVm || ALL || functional) && !skipStandaloneVm
// +build vm standaloneVm ALL functional
// +build !skipStandaloneVm

package vcd

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdStandaloneVmCapabilities(t *testing.T) {
	preTestChecks(t)

	var standaloneVmName = fmt.Sprintf("%s-%d", t.Name(), os.Getpid())
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"Catalog":     testSuiteCatalogName,
		"CatalogItem": testSuiteCatalogOVAItem,
		"VmName":      standaloneVmName,
		"Tags":        "vm standaloneVm",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccCheckVcdVm_capabilities, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)

	params["FuncName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccCheckVcdVm_capabilitiesUpdate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText1)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdStandaloneVmDestroy(standaloneVmName, "", ""),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdStandaloneVmExists(standaloneVmName, "vcd_vm."+standaloneVmName, "", ""),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "name", standaloneVmName),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "cpu_hot_add_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "memory_hot_add_enabled", "true"),
				),
			},
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdStandaloneVmExists(standaloneVmName, "vcd_vm."+standaloneVmName, "", ""),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "name", standaloneVmName),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "cpu_hot_add_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "memory_hot_add_enabled", "false"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdVm_capabilities = `
resource "vcd_vm" "{{.VmName}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  name          = "{{.VmName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1

  cpu_hot_add_enabled    = true
  memory_hot_add_enabled = true
  
}
`

const testAccCheckVcdVm_capabilitiesUpdate = `
# skip-binary-test: only for updates

resource "vcd_vm" "{{.VmName}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  name          = "{{.VmName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1
}
`
