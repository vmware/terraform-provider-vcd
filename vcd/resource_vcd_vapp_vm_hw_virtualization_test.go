//go:build vapp || vm || ALL || functional
// +build vapp vm ALL functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdVAppVm_HardwareVirtualization(t *testing.T) {
	preTestChecks(t)
	vappNameHwVirt := "TestAccVcdVAppHwVirt"
	vmNameHwVirt := "TestAccVcdVAppHwVirt"
	var vapp govcd.VApp
	var vm govcd.VM

	var params = StringMap{
		"Org":                          testConfig.VCD.Org,
		"Vdc":                          testConfig.VCD.Vdc,
		"EdgeGateway":                  testConfig.Networking.EdgeGateway,
		"NetworkName":                  "TestAccVcdVAppVmNetHwVirt",
		"Catalog":                      testSuiteCatalogName,
		"CatalogItem":                  testSuiteCatalogOVAItem,
		"VappName":                     vappNameHwVirt,
		"VmName":                       vmNameHwVirt,
		"ExposeHardwareVirtualization": "false",
		"Tags":                         "vapp vm",
	}

	configTextStep0 := templateFill(testAccCheckVcdVAppVm_hardwareVirtualization, params)

	params["ExposeHardwareVirtualization"] = "true"
	params["FuncName"] = t.Name() + "-step1"
	configTextStep1 := templateFill(testAccCheckVcdVAppVm_hardwareVirtualization, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextStep0)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdVAppVmDestroy(vappNameHwVirt),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configTextStep0,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppVmExists(vappNameHwVirt, vmNameHwVirt, "vcd_vapp_vm."+vmNameHwVirt, &vapp, &vm),
					resource.TestCheckResourceAttr(
						"vcd_vapp_vm."+vmNameHwVirt, "name", vmNameHwVirt),
					resource.TestCheckResourceAttr(
						"vcd_vapp_vm."+vmNameHwVirt, "expose_hardware_virtualization", "false"),
				),
			},
			resource.TestStep{
				Config: configTextStep1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppVmExists(vappNameHwVirt, vmNameHwVirt, "vcd_vapp_vm."+vmNameHwVirt, &vapp, &vm),
					resource.TestCheckResourceAttr(
						"vcd_vapp_vm."+vmNameHwVirt, "name", vmNameHwVirt),
					resource.TestCheckResourceAttr(
						"vcd_vapp_vm."+vmNameHwVirt, "expose_hardware_virtualization", "true"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdVAppVm_hardwareVirtualization = `
resource "vcd_vapp" "{{.VappName}}" {
  name = "{{.VappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
}

resource "vcd_vapp_vm" "{{.VmName}}" {
  org                            = "{{.Org}}"
  vdc                            = "{{.Vdc}}"
  vapp_name                      = vcd_vapp.{{.VappName}}.name
  name                           = "{{.VmName}}"
  catalog_name                   = "{{.Catalog}}"
  template_name                  = "{{.CatalogItem}}"
  memory                         = 384
  cpus                           = 2
  cpu_cores                      = 1
  expose_hardware_virtualization = {{.ExposeHardwareVirtualization}}
}
`
