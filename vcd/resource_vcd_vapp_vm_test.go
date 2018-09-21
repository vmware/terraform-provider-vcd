package vcd

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	govcd "github.com/vmware/go-vcloud-director/govcd"
)

var vappName2 string = "TestAccVcdVAppVmVapp"
var vmName string = "TestAccVcdVAppVmVm"

func TestAccVcdVAppVm_Basic(t *testing.T) {
	var vapp govcd.VApp
	var vm govcd.VM

	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"NetworkName": "TestAccVcdVAppVmNet",
		"Catalog":     testConfig.VCD.Catalog.Name,
		"CatalogItem": testConfig.VCD.Catalog.Catalogitem,
		"VappName":    vappName2,
		"VmName":      vmName,
	}

	configText := templateFill(testAccCheckVcdVAppVm_basic, params)
	if os.Getenv("GOVCD_DEBUG") != "" {
		log.Printf("#[DEBUG] CONFIGURATION: %s\n", configText)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdVAppVmDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppVmExists("vcd_vapp_vm."+vmName, &vapp, &vm),
					resource.TestCheckResourceAttr(
						"vcd_vapp_vm."+vmName, "name", vmName),
					resource.TestCheckResourceAttr(
						"vcd_vapp_vm."+vmName, "ip", "10.10.102.161"),
					resource.TestCheckResourceAttr(
						"vcd_vapp_vm."+vmName, "power_on", "true"),
				),
			},
		},
	})
}

func testAccCheckVcdVAppVmExists(n string, vapp *govcd.VApp, vm *govcd.VM) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VAPP ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)
		org, err := govcd.GetOrgByName(conn.VCDClient, testConfig.VCD.Org)
		if err != nil || org == (govcd.Org{}) {
			return fmt.Errorf("Could not find test Org")
		}
		vdc, err := org.GetVdcByName(testConfig.VCD.Vdc)
		if err != nil || vdc == (govcd.Vdc{}) {
			return fmt.Errorf("Could not find test Vdc")
		}
		vapp, err := vdc.FindVAppByName(vappName2)

		resp, err := vdc.FindVMByName(vapp, vmName)

		if err != nil {
			return err
		}

		*vm = resp

		return nil
	}
}

func testAccCheckVcdVAppVmDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_vapp" {
			continue
		}
		org, err := govcd.GetOrgByName(conn.VCDClient, testConfig.VCD.Org)
		if err != nil || org == (govcd.Org{}) {
			return fmt.Errorf("Could not find test Org")
		}
		vdc, err := org.GetVdcByName(testConfig.VCD.Vdc)
		if err != nil || vdc == (govcd.Vdc{}) {
			return fmt.Errorf("Could not find test Vdc")
		}
		_, err = vdc.FindVAppByName(vappName2)

		if err == nil {
			return fmt.Errorf("VPCs still exist")
		}

		return nil
	}

	return nil
}

const testAccCheckVcdVAppVm_basic = `
resource "vcd_network" "{{.NetworkName}}" {
	name          = "{{.NetworkName}}"
	org           = "{{.Org}}"
	vdc           = "{{.Vdc}}"
	edge_gateway  = "{{.EdgeGateway}}"
	gateway 	  = "10.10.102.1"
	static_ip_pool {
		start_address = "10.10.102.2"
		end_address   = "10.10.102.254"
	}
}

resource "vcd_vapp" "{{.VappName}}" {
  name          = "{{.VappName}}"
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
}

resource "vcd_vapp_vm" "{{.VmName}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  vapp_name     = "${vcd_vapp.{{.VappName}}.name}"
  network_name  = "${vcd_network.{{.NetworkName}}.name}"
  name          = "{{.VmName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 1024
  cpus          = 1
  ip            = "10.10.102.161"
  depends_on    = ["vcd_vapp.{{.VappName}}"]
}
`
