// +build vapp vm ALL functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

var vappName2 string = "TestAccVcdVAppVmVapp"
var vmName string = "TestAccVcdVAppVmVm"

func TestAccVcdVAppVm_Basic(t *testing.T) {
	var vapp govcd.VApp
	var vm govcd.VM
	var diskResourceName = "TestAccVcdVAppVm_Basic_1"
	var diskName = "TestAccVcdIndependentDiskBasic"
	var internalDiskSize = 20000

	var params = StringMap{
		"Org":                testConfig.VCD.Org,
		"Vdc":                testConfig.VCD.Vdc,
		"EdgeGateway":        testConfig.Networking.EdgeGateway,
		"NetworkName":        "TestAccVcdVAppVmNet",
		"Catalog":            testSuiteCatalogName,
		"CatalogItem":        testSuiteCatalogOVAItem,
		"VappName":           vappName2,
		"VmName":             vmName,
		"ComputerName":       vmName + "-unique",
		"diskName":           diskName,
		"size":               "5",
		"busType":            "SCSI",
		"busSubType":         "lsilogicsas",
		"storageProfileName": "*",
		"diskResourceName":   diskResourceName,
		"Tags":               "vapp vm",
		"InternalDiskSize":   internalDiskSize,
	}

	configText := templateFill(testAccCheckVcdVAppVm_basic, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdVAppVmDestroy(vappName2),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppVmExists(vappName2, vmName, "vcd_vapp_vm."+vmName, &vapp, &vm),
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
					resource.TestCheckOutput("disk", diskName),
					testCheckInternalDiskNonStringOutputs(internalDiskSize),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_vapp_vm." + vmName + "-import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdVappObject(testConfig, vappName2, vmName),
				// These fields can't be retrieved from user data
				ImportStateVerifyIgnore: []string{"template_name", "catalog_name", "network_name",
					"initscript", "accept_all_eulas", "power_on", "computer_name", "override_template_disk"},
			},
		},
	})
}

func testCheckInternalDiskNonStringOutputs(internalDiskSize int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		outputs := s.RootModule().Outputs

		if outputs["internal_disk_size"].Value != internalDiskSize {
			return fmt.Errorf("internal disk size value didn't match")
		}

		if outputs["internal_disk_iops"].Value != 0 {
			return fmt.Errorf("internal disk iops value didn't match")
		}

		if outputs["internal_disk_bus_type"].Value != "paravirtual" {
			return fmt.Errorf("internal disk bus type value didn't match")
		}

		if outputs["internal_disk_bus_number"].Value != 0 {
			return fmt.Errorf("internal disk bus number value didn't match")
		}

		if outputs["internal_disk_unit_number"].Value != 0 {
			return fmt.Errorf("internal disk unit number value didn't match")
		}

		if outputs["internal_disk_thin_provisioned"].Value != true {
			return fmt.Errorf("internal disk thin provisioned value didn't match")
		}

		if outputs["internal_disk_storage_profile"].Value != "*" {
			return fmt.Errorf("internal disk storage profile value didn't match")
		}

		return nil
	}
}

func TestAccVcdVAppVm_Clone(t *testing.T) {
	var vapp govcd.VApp
	var vm govcd.VM

	var params = StringMap{
		"Org":                testConfig.VCD.Org,
		"Vdc":                testConfig.VCD.Vdc,
		"EdgeGateway":        testConfig.Networking.EdgeGateway,
		"NetworkName":        "TestAccVcdVAppVmNet",
		"Catalog":            testSuiteCatalogName,
		"CatalogItem":        testSuiteCatalogOVAItem,
		"VappName":           vappName2,
		"VmName":             vmName,
		"VmName2":            vmName + "-clone",
		"ComputerName":       vmName + "-unique",
		"size":               "5",
		"busType":            "SCSI",
		"busSubType":         "lsilogicsas",
		"storageProfileName": "*",
		"IP":                 "10.10.102.161",
		"IP2":                "10.10.102.162",
		"Tags":               "vapp vm",
	}

	configText := templateFill(testAccCheckVcdVAppVm_clone, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	vm1 := "vcd_vapp_vm." + vmName
	vm2 := "vcd_vapp_vm." + vmName + "-clone"

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdVAppVmDestroy(vappName2),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppVmExists(vappName2, vmName, "vcd_vapp_vm."+vmName, &vapp, &vm),
					resource.TestCheckResourceAttr(
						vm1, "name", vmName),
					resource.TestCheckResourceAttr(
						vm1, "computer_name", params["ComputerName"].(string)),
					resource.TestCheckResourceAttr(
						vm1, "network.0.ip", params["IP"].(string)),
					resource.TestCheckResourceAttr(
						vm1, "power_on", "true"),
					resource.TestCheckResourceAttr(
						vm1, "metadata.vm_metadata", "VM Metadata."),
					resource.TestCheckResourceAttr(
						vm2, "network.0.ip", params["IP2"].(string)),
					resource.TestCheckResourceAttrPair(
						vm1, "vapp_name", vm2, "vapp_name"),
					resource.TestCheckResourceAttrPair(
						vm1, "metadata", vm2, "metadata"),
					resource.TestCheckResourceAttrPair(
						vm1, "network.0.name", vm2, "network.0.name"),
					resource.TestCheckResourceAttrPair(
						vm1, "network.0.type", vm2, "network.0.type"),
					resource.TestCheckResourceAttrPair(
						vm1, "network.0.ip_allocation_mode", vm2, "network.0.ip_allocation_mode"),
				),
			},
		},
	})
}

func init() {
	testingTags["vm"] = "resource_vcd_vapp_vm_test.go"
}

const testAccCheckVcdVAppVm_basic = `
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

resource "vcd_independent_disk" "{{.diskResourceName}}" {
  org             = "{{.Org}}"
  vdc             = "{{.Vdc}}"
  name            = "{{.diskName}}"
  size            = "{{.size}}"
  bus_type        = "{{.busType}}"
  bus_sub_type    = "{{.busSubType}}"
  storage_profile = "{{.storageProfileName}}"
}

resource "vcd_vapp" "{{.VappName}}" {
  name = "{{.VappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  depends_on = ["vcd_network_routed.{{.NetworkName}}"]
}

resource "vcd_vapp_vm" "{{.VmName}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  vapp_name     = vcd_vapp.{{.VappName}}.name
  name          = "{{.VmName}}"
  computer_name = "{{.ComputerName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 1024
  cpus          = 2
  cpu_cores     = 1

  metadata = {
    vm_metadata = "VM Metadata."
  }
 
  override_template_disk {
    bus_type         = "paravirtual"
    size_in_mb       = "{{.InternalDiskSize}}"
    bus_number       = 0
    unit_number      = 0
    iops             = 0
    thin_provisioned = true
    storage_profile  = "{{.storageProfileName}}"
  }
 
  network {
    name               = vcd_network_routed.{{.NetworkName}}.name
    ip                 = "10.10.102.161"
    type               = "org"
    ip_allocation_mode = "MANUAL"
  }

  disk {
    name        = vcd_independent_disk.{{.diskResourceName}}.name
    bus_number  = 1
    unit_number = 0
  }
}

output "disk" {
  value = tolist(vcd_vapp_vm.{{.VmName}}.disk)[0].name
}

output "internal_disk_size" {
  value = vcd_vapp_vm.{{.VmName}}.internal_disk[0].size_in_mb
  depends_on = [vcd_vapp_vm.{{.VmName}}]
}

output "internal_disk_iops" {
  value = vcd_vapp_vm.{{.VmName}}.internal_disk[0].iops
  depends_on = [vcd_vapp_vm.{{.VmName}}]
}

output "internal_disk_bus_type" {
  value = vcd_vapp_vm.{{.VmName}}.internal_disk[0].bus_type
  depends_on = [vcd_vapp_vm.{{.VmName}}]
}

output "internal_disk_bus_number" {
  value = vcd_vapp_vm.{{.VmName}}.internal_disk[0].bus_number
  depends_on = [vcd_vapp_vm.{{.VmName}}]
}

output "internal_disk_unit_number" {
  value = vcd_vapp_vm.{{.VmName}}.internal_disk[0].unit_number
  depends_on = [vcd_vapp_vm.{{.VmName}}]
}

output "internal_disk_thin_provisioned" {
  value = vcd_vapp_vm.{{.VmName}}.internal_disk[0].thin_provisioned
  depends_on = [vcd_vapp_vm.{{.VmName}}]
}

output "internal_disk_storage_profile" {
  value = vcd_vapp_vm.{{.VmName}}.internal_disk[0].storage_profile
  depends_on = [vcd_vapp_vm.{{.VmName}}]
}
`

const testAccCheckVcdVAppVm_clone = `
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
  depends_on = ["vcd_network_routed.{{.NetworkName}}"]
}

resource "vcd_vapp_vm" "{{.VmName}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  vapp_name     = vcd_vapp.{{.VappName}}.name
  name          = "{{.VmName}}"
  computer_name = "{{.ComputerName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 1024
  cpus          = 2
  cpu_cores     = 1

  metadata = {
    vm_metadata = "VM Metadata."
  }

  network {
    name               = vcd_network_routed.{{.NetworkName}}.name
    ip                 = "{{.IP}}"
    type               = "org"
    ip_allocation_mode = "MANUAL"
  }
}

resource "vcd_vapp_vm" "{{.VmName2}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  vapp_name     = vcd_vapp_vm.{{.VmName}}.vapp_name
  name          = "{{.VmName2}}"
  computer_name = vcd_vapp_vm.{{.VmName}}.computer_name
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 1024
  cpus          = 2
  cpu_cores     = 1

  metadata = vcd_vapp_vm.{{.VmName}}.metadata

  network {
    name               = vcd_vapp_vm.{{.VmName}}.network.0.name
    ip                 = "{{.IP2}}"
    type               = vcd_vapp_vm.{{.VmName}}.network.0.type
    ip_allocation_mode = vcd_vapp_vm.{{.VmName}}.network.0.ip_allocation_mode
  }
}

`
