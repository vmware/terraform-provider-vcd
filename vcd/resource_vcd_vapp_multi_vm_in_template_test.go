// +build vapp vm ALL functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdVAppMultiVmInTemplate(t *testing.T) {
	var vapp govcd.VApp
	var vm govcd.VM
	vappName := t.Name()
	vmName := t.Name() + "VM"
	vmName2 := t.Name() + "VM2"
	var params = StringMap{
		"Org":                testConfig.VCD.Org,
		"Vdc":                testConfig.VCD.Vdc,
		"EdgeGateway":        testConfig.Networking.EdgeGateway,
		"NetworkName":        "TestAccVcdVAppVmNet",
		"Catalog":            testConfig.VCD.Catalog.Name,
		"CatalogItemMultiVm": testConfig.VCD.Catalog.CatalogItemWithMultiVms,
		"VmNameInTemplate":   testConfig.VCD.Catalog.VmName1,
		"VmNameInTemplate2":  testConfig.VCD.Catalog.VmName2,
		"VappName":           vappName,
		"VmName":             vmName,
		"VmName2":            vmName2,
		"ComputerName":       vmName + "-unique",
		"Tags":               "vapp vm",
	}

	configText := templateFill(testAccCheckVcdVAppVmMultiVmInTemplate, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	if testConfig.VCD.Catalog.CatalogItemWithMultiVms == "" || testConfig.VCD.Catalog.VmName1 == "" || testConfig.VCD.Catalog.VmName2 == "" {
		t.Skip("Variables catalogItemWithMultiVms, vmName1, vmName2  must be set to run multi VM in vApp template tests")
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdVAppVmDestroy(vappName),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppVmExists(vappName, vmName, "vcd_vapp_vm."+vmName, &vapp, &vm),
					resource.TestCheckResourceAttr(
						"vcd_vapp_vm."+vmName, "name", vmName),
					resource.TestCheckResourceAttr(
						"vcd_vapp_vm."+vmName, "computer_name", vmName+"-unique"),
					resource.TestCheckResourceAttr(
						"vcd_vapp_vm."+vmName, "network.0.ip", "10.10.102.161"),
					resource.TestCheckResourceAttr(
						"vcd_vapp_vm."+vmName, "power_on", "true"),
					resource.TestCheckResourceAttr(
						"vcd_vapp_vm."+vmName, "metadata.vm_metadata", "VM Metadata."),
					resource.TestCheckResourceAttr(
						"vcd_vapp_vm."+vmName2, "name", vmName2),
					resource.TestCheckResourceAttr(
						"vcd_vapp_vm."+vmName2, "computer_name", vmName+"-unique"),
					resource.TestCheckResourceAttr(
						"vcd_vapp_vm."+vmName2, "network.0.ip", "10.10.102.162"),
					resource.TestCheckResourceAttr(
						"vcd_vapp_vm."+vmName2, "power_on", "true"),
					resource.TestCheckResourceAttr(
						"vcd_vapp_vm."+vmName2, "metadata.vm_metadata", "VM Metadata."),
				),
			},
		},
	})
}

const testAccCheckVcdVAppVmMultiVmInTemplate = `
resource "vcd_network_routed" "{{.NetworkName}}" {
  name         = "{{.NetworkName}}"
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "10.10.102.1"

  static_ip_pool {
    start_address = "10.10.102.2"
    end_address   = "10.10.102.254"
  }
}

resource "vcd_vapp" "{{.VappName}}" {
  name = "{{.VappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
}

resource "vcd_vapp_org_network" "vappNetwork1" {
  org                = "{{.Org}}"
  vdc                = "{{.Vdc}}"
  vapp_name          = vcd_vapp.{{.VappName}}.name
  org_network_name   = vcd_network_routed.{{.NetworkName}}.name 
}

resource "vcd_vapp_vm" "{{.VmName}}" {
  org         		  = "{{.Org}}"
  vdc         		  = "{{.Vdc}}"
  vapp_name 	      = vcd_vapp.{{.VappName}}.name
  name		          = "{{.VmName}}"
  computer_name 	  = "{{.ComputerName}}"
  catalog_name		  = "{{.Catalog}}"
  template_name		  = "{{.CatalogItemMultiVm}}"
  vm_name_in_template = "{{.VmNameInTemplate}}"
  memory        	  = 1024
  cpus                = 2
  cpu_cores           = 1

  metadata = {
    vm_metadata = "VM Metadata."
  }

  network {
    type               = "org"
    name               = vcd_vapp_org_network.vappNetwork1.org_network_name
    ip_allocation_mode = "MANUAL"
    ip                 = "10.10.102.161"
  }
}

resource "vcd_vapp_vm" "{{.VmName2}}" {
  org         		  = "{{.Org}}"
  vdc         		  = "{{.Vdc}}"
  vapp_name 	      = vcd_vapp.{{.VappName}}.name
  name		          = "{{.VmName2}}"
  computer_name 	  = "{{.ComputerName}}"
  catalog_name		  = "{{.Catalog}}"
  template_name		  = "{{.CatalogItemMultiVm}}"
  vm_name_in_template = "{{.VmNameInTemplate2}}"
  memory        	  = 1024
  cpus                = 2
  cpu_cores           = 1

  metadata = {
    vm_metadata = "VM Metadata."
  }

  network {
    type               = "org"
    name               = vcd_vapp_org_network.vappNetwork1.org_network_name
    ip_allocation_mode = "MANUAL"
    ip                 = "10.10.102.162"
  }
}
`
