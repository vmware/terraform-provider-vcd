package vcd

import (
	// "fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdVAppVm_HardwareVirtualization(t *testing.T) {
	vappNameHwVirt := "TestAccVcdVAppHwVirt"
	vmNameHwVirt := "TestAccVcdVAppHwVirt"
	var vapp govcd.VApp
	var vm govcd.VM

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"NetworkName": "TestAccVcdVAppVmNetHwVirt",
		"Catalog":     testSuiteCatalogName,
		"CatalogItem": testSuiteCatalogOVAItem,
		"VappName":    vappNameHwVirt,
		"VmName":      vmNameHwVirt,
	}

	configTextStep0 := templateFill(testAccCheckVcdVAppVm_hardwareVirtualization, params)

	params["ExposeHardwareVirtualization"] = true
	params["FuncName"] = t.Name() + "step1"
	configTextStep1 := templateFill(testAccCheckVcdVAppVm_hardwareVirtualization, params)

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextStep0)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdVAppVmDestroy(vappNameHwVirt),
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
}

const testAccCheckVcdVAppVm_hardwareVirtualization = `
resource "vcd_network_routed" "{{.NetworkName}}" {
  name         = "{{.NetworkName}}"
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "10.10.104.1"

  static_ip_pool {
    start_address = "10.10.104.2"
    end_address   = "10.10.104.254"
  }
}

resource "vcd_vapp" "{{.VappName}}" {
  name = "{{.VappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
}

resource "vcd_vapp_vm" "{{.VmName}}" {
  org           					= "{{.Org}}"
  vdc           					= "{{.Vdc}}"
  vapp_name     					= "${vcd_vapp.{{.VappName}}.name}"
  network_name  					= "${vcd_network_routed.{{.NetworkName}}.name}"
  name          					= "{{.VmName}}"
  catalog_name  					= "{{.Catalog}}"
  template_name 					= "{{.CatalogItem}}"
  memory        					= 384
  cpus          					= 2
  cpu_cores     					= 1
  ip            					= "10.10.104.161"
  expose_hardware_virtualization	= "{{.ExposeHardwareVirtualization}}"

  depends_on    = ["vcd_vapp.{{.VappName}}", "vcd_network_routed.{{.NetworkName}}"]
}
`
