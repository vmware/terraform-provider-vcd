package vcd

import (
	// "fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	// "github.com/hashicorp/terraform/terraform"
	// "github.com/vmware/go-vcloud-director/v2/govcd"
)

var vappNameHardwareVirtualization string = "TestAccVcdVAppHwVirt"
var vmNameHardwareVirtualization string = "TestAccVcdVAppHwVirt"

func TestAccVcdVAppVm_HardwareVirtualization(t *testing.T) {
	// var vapp govcd.VApp
	// var vm govcd.VM

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
		"VappName":    vappNameHardwareVirtualization,
		"VmName":      vmNameHardwareVirtualization,
	}

	configTextStep0 := templateFill(testAccCheckVcdVAppVm_hardwareVirtualization, params)

	params["ExposeHardwareVirtualization"] = true
	params["FuncName"] = "step1"
	configTextStep1 := templateFill(testAccCheckVcdVAppVm_hardwareVirtualization, params)

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextStep0)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdVAppVmDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configTextStep0,
				Check: resource.ComposeTestCheckFunc(
					// Need to pre a more generic check than this
					// testAccCheckVcdVAppVmExists("vcd_vapp_vm."+vmNameHardwareVirtualization, &vapp, &vm),
					resource.TestCheckResourceAttr(
						"vcd_vapp_vm."+vmNameHardwareVirtualization, "name", vmNameHardwareVirtualization),
					resource.TestCheckResourceAttr(
						"vcd_vapp_vm."+vmNameHardwareVirtualization, "expose_hardware_virtualization", "false"),
				),
			},
			resource.TestStep{
				Config: configTextStep1,
				Check: resource.ComposeTestCheckFunc(
					// Need to pre a more generic check than this
					// testAccCheckVcdVAppVmExists("vcd_vapp_vm."+vmNameHardwareVirtualization, &vapp, &vm),
					resource.TestCheckResourceAttr(
						"vcd_vapp_vm."+vmNameHardwareVirtualization, "name", vmNameHardwareVirtualization),
					resource.TestCheckResourceAttr(
						"vcd_vapp_vm."+vmNameHardwareVirtualization, "expose_hardware_virtualization", "true"),
				),
			},
		},
	})
}

// func testAccCheckVcdVAppVmExists(n string, vapp *govcd.VApp, vm *govcd.VM) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		rs, ok := s.RootModule().Resources[n]
// 		if !ok {
// 			return fmt.Errorf("not found: %s", n)
// 		}

// 		if rs.Primary.ID == "" {
// 			return fmt.Errorf("no VAPP ID is set")
// 		}

// 		conn := testAccProvider.Meta().(*VCDClient)
// 		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
// 		if err != nil {
// 			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.VCD.Vdc, testConfig.VCD.Org, err)
// 		}

// 		vapp, err := vdc.FindVAppByName(vappName2)
// 		if err != nil {
// 			return err
// 		}

// 		resp, err := vdc.FindVMByName(vapp, vmName)

// 		if err != nil {
// 			return err
// 		}

// 		*vm = resp

// 		return nil
// 	}
// }

// func testAccCheckVcdVAppVmDestroy(s *terraform.State) error {
// 	conn := testAccProvider.Meta().(*VCDClient)

// 	for _, rs := range s.RootModule().Resources {
// 		if rs.Type != "vcd_vapp" {
// 			continue
// 		}
// 		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
// 		if err != nil {
// 			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.VCD.Vdc, testConfig.VCD.Org, err)
// 		}

// 		_, err = vdc.FindVAppByName(vappName2)

// 		if err == nil {
// 			return fmt.Errorf("VPCs still exist")
// 		}

// 		return nil
// 	}

// 	return nil
// }

const testAccCheckVcdVAppVm_hardwareVirtualization = `
resource "vcd_network_routed" "{{.NetworkName}}" {
  name         = "{{.NetworkName}}"
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "10.10.103.1"

  static_ip_pool {
    start_address = "10.10.103.2"
    end_address   = "10.10.103.254"
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
  ip            					= "10.10.103.161"
  expose_hardware_virtualization	= "{{.ExposeHardwareVirtualization}}"

  depends_on    = ["vcd_vapp.{{.VappName}}", "vcd_network_routed.{{.NetworkName}}"]
}
`
