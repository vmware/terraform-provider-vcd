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

func TestAccVcdStandaloneVmShrinkCpu(t *testing.T) {
	preTestChecks(t)

	var standaloneVmName = fmt.Sprintf("%s-shrink-%d", t.Name(), os.Getpid())
	var params = StringMap{
		"Org":     testConfig.VCD.Org,
		"Vdc":     testConfig.VCD.Vdc,
		"Catalog": testSuiteCatalogName,
		"OvaPath": testConfig.Ova.OvaPath,
		"VmName":  standaloneVmName,
		"Tags":    "vm standaloneVm",
	}

	configText := templateFill(testAccCheckVcdVmShrinkCpu, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdStandaloneVmDestroy(standaloneVmName, "", ""),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdStandaloneVmExists(standaloneVmName, "vcd_vm.shrink-vm", "", ""),
					resource.TestCheckResourceAttr("vcd_vm.shrink-vm", "name", standaloneVmName),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdVmShrinkCpu = `
resource "vcd_catalog_item" "fourcpu4cores" {
  org     = "{{.Org}}"
  catalog = "{{.Catalog}}"

  name                 = "4cpu-4cores"
  ova_path             = "{{.OvaPath}}"
  show_upload_progress = true
}

resource "vcd_vm" "shrink-vm" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  name          = "{{.VmName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = vcd_catalog_item.fourcpu4cores.name

  memory    = 512
  cpus      = 2
  cpu_cores = 1

  depends_on = [vcd_catalog_item.fourcpu4cores]
}

`
