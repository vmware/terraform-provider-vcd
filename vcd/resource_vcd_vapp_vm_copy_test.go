//go:build vapp || vm || ALL || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	testingTags["vm"] = "resource_vcd_vapp_vm_copy_test.go"
}

func TestAccVcdVAppVm_BasicCopySameVdc(t *testing.T) {
	preTestChecks(t)

	var params = StringMap{
		"TestName": t.Name(),
		"Org":      testConfig.VCD.Org,
		"Vdc":      testConfig.Nsxt.Vdc,

		"VappName":     vappName2,
		"VmName":       vmName,
		"ComputerName": vmName + "-unique",

		"Tags": "vapp vm",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccCheckVcdVAppVm_basicCopy, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdVAppVmDestroy(vappName2),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check:  resource.ComposeTestCheckFunc(
				// testAccCheckVcdVAppVmExists(vappName2, vmName, "vcd_vapp_vm."+vmName, &vapp, &vm),
				// resource.TestCheckResourceAttr("vcd_vapp_vm."+vmName, "name", vmName),
				// resource.TestCheckResourceAttr("vcd_vapp_vm."+vmName, "computer_name", vmName+"-unique"),
				// resource.TestCheckResourceAttr("vcd_vapp_vm."+vmName, "power_on", "false"),
				// resource.TestCheckResourceAttr("vcd_vapp_vm."+vmName, "metadata.vm_metadata", "VM Metadata."),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdVAppVm_basicCopy = `
resource "vcd_vm" "source" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  # You cannot remove NICs from an active virtual machine on which no operating system is installed.
  power_on = false

  description = "test empty VM"
  name        = "{{.TestName}}-source"
  memory      = 512
  cpus        = 2
  cpu_cores   = 1

  os_type          = "sles11_64Guest"
  hardware_version = "vmx-13"
  computer_name    = "compName"
}

resource "vcd_vapp" "destination" {
  name     = "{{.VappName}}-dest"
  org      = "{{.Org}}"
  vdc      = "{{.Vdc}}"
  power_on = false
}

resource "vcd_vapp_network" "vappNet" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name                   = "{{.TestName}}-vapp-network"
  vapp_name              = vcd_vapp.destination.name
  gateway                = "192.168.2.1"
  prefix_length          = "24"
  reboot_vapp_on_removal = true

  static_ip_pool {
    start_address = "192.168.2.51"
    end_address   = "192.168.2.100"
  }
}

resource "vcd_vapp_vm" "copy" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  vapp_name     = vcd_vapp.destination.name
  name          = "{{.TestName}}-dest"
  computer_name = "{{.ComputerName}}"

  copy_from_vm_id  = vcd_vm.source.id
  power_on         = false
  memory           = 1024
  cpus             = 2
  cpu_cores        = 1

  network {
    type               = "vapp"
    name               = vcd_vapp_network.vappNet.name
    ip_allocation_mode = "POOL"
  }
}

resource "vcd_vm" "copy" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  name          = "{{.TestName}}-dest"
  computer_name = "{{.ComputerName}}"

  copy_from_vm_id  = vcd_vm.source.id
  power_on         = false
  memory           = 1024
  cpus             = 2
  cpu_cores        = 1

  #network {
  #  type               = "vapp"
  #  name               = vcd_vapp_network.vappNet.name
  #  ip_allocation_mode = "POOL"
  #}
}
`

func TestAccVcdVAppVm_BasicCopyDifferentVdc(t *testing.T) {
	preTestChecks(t)

	var params = StringMap{
		"TestName":  t.Name(),
		"Org":       testConfig.VCD.Org,
		"Vdc":       testConfig.Nsxt.Vdc,
		"SourceVdc": testConfig.Nsxt.Vdc + "-group-member-0",

		"VappName":     vappName2,
		"VmName":       vmName,
		"ComputerName": vmName + "-unique",

		"Tags": "vapp vm",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccCheckVcdVAppVm_basicCopy_Vdcs, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdVAppVmDestroy(vappName2),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check:  resource.ComposeTestCheckFunc(
				// testAccCheckVcdVAppVmExists(vappName2, vmName, "vcd_vapp_vm."+vmName, &vapp, &vm),
				// resource.TestCheckResourceAttr("vcd_vapp_vm."+vmName, "name", vmName),
				// resource.TestCheckResourceAttr("vcd_vapp_vm."+vmName, "computer_name", vmName+"-unique"),
				// resource.TestCheckResourceAttr("vcd_vapp_vm."+vmName, "power_on", "false"),
				// resource.TestCheckResourceAttr("vcd_vapp_vm."+vmName, "metadata.vm_metadata", "VM Metadata."),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdVAppVm_basicCopy_Vdcs = `
resource "vcd_vm" "source" {
  org = "{{.Org}}"
  vdc = "{{.SourceVdc}}"

  # You cannot remove NICs from an active virtual machine on which no operating system is installed.
  power_on = true

  description = "test empty VM"
  name        = "{{.TestName}}-source"
  memory      = 512
  cpus        = 2
  cpu_cores   = 1

  os_type          = "sles11_64Guest"
  hardware_version = "vmx-13"
  computer_name    = "compName"
}

resource "vcd_vapp" "destination" {
  name     = "{{.VappName}}-dest"
  org      = "{{.Org}}"
  vdc      = "{{.Vdc}}"
  power_on = true
}

resource "vcd_vapp_network" "vappNet" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name                   = "{{.TestName}}-vapp-network"
  vapp_name              = vcd_vapp.destination.name
  gateway                = "192.168.2.1"
  prefix_length          = "24"
  reboot_vapp_on_removal = true

  static_ip_pool {
    start_address = "192.168.2.51"
    end_address   = "192.168.2.100"
  }
}

data "vcd_org_vdc" "source" {
  org  = "{{.Org}}"
  name = "{{.SourceVdc}}"
}

resource "vcd_vapp_vm" "copy" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  vapp_name     = vcd_vapp.destination.name
  name          = "{{.TestName}}-dest"
  computer_name = "{{.ComputerName}}"

  copy_from_vdc_id = data.vcd_org_vdc.source.id
  copy_from_vm_id  = vcd_vm.source.id
  power_on         = true
  memory           = 1024
  cpus             = 2
  cpu_cores        = 1

  network {
    type               = "vapp"
    name               = vcd_vapp_network.vappNet.name
    ip_allocation_mode = "POOL"
  }
}

resource "vcd_vm" "copy" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  name          = "{{.TestName}}-dest"
  computer_name = "{{.ComputerName}}"

  copy_from_vdc_id = data.vcd_org_vdc.source.id
  copy_from_vm_id  = vcd_vm.source.id
  power_on         = true
  memory           = 1024
  cpus             = 2
  cpu_cores        = 1

  #network {
  #  type               = "vapp"
  #  name               = vcd_vapp_network.vappNet.name
  #  ip_allocation_mode = "POOL"
  #}
}
`
