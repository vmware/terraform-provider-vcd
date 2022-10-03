//go:build vapp || vm || ALL || functional
// +build vapp vm ALL functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Terraform codebase for VM management is very complicated and is backed by 4 types of VM:
//  * `types.InstantiateVmTemplateParams` (Standalone VM from template)
//  * `types.ReComposeVAppParams` (vApp VM from template)
//  * `types.RecomposeVAppParamsForEmptyVm` (Empty vApp VM)
//  * `types.CreateVmParams` (Empty Standalone VM)
//
// Each of these 4 types have different fields for creation (just like UI differs), but the
// expectation for the user is to get a VM with all configuration available in HCL, no matter the type.
//
// As a result, the architecture of VM creation is such, that it uses above defined types to create
// VMs with minimal configuration and then perform additions API calls. There are still risks that
// some VMs get less configured than others. To overcome this risk, there is a new set of tests.
// Each of these tests aim to ensure that exactly the same configuration is achieved.

// TestAccVcdVAppVm_4types attempts to test minimal create configuration for all 4 types of VMs
func TestAccVcdVAppVm_4types(t *testing.T) {
	preTestChecks(t)

	var params = StringMap{
		"TestName":    t.Name(),
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.Nsxt.Vdc,
		"Catalog":     testConfig.VCD.Catalog.NsxtBackedCatalogName,
		"CatalogItem": testConfig.VCD.Catalog.NsxtCatalogItem,

		// "EdgeGateway":                  testConfig.Networking.EdgeGateway,
		// "NetworkName":                  "TestAccVcdVAppVmNetHwVirt",
		// "VappName":                     vappNameHwVirt,
		// "VmName":                       vmNameHwVirt,
		// "ExposeHardwareVirtualization": "false",
		"Tags": "vapp vm",
	}
	testParamsNotEmpty(t, params)

	configTextStep1 := templateFill(testAccVcdVAppVm_4types_Step1, params)

	// params["ExposeHardwareVirtualization"] = "true"
	// params["FuncName"] = t.Name() + "-step1"
	// configTextStep1 := templateFill(testAccCheckVcdVAppVm_hardwareVirtualization, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextStep1)
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		// CheckDestroy:      testAccCheckVcdVAppVmDestroy(vappNameHwVirt),
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
		// testAccCheckVcdVAppVmDestroy(vappNameHwVirt),
		),
		Steps: []resource.TestStep{
			{
				Config: configTextStep1,
				Check: resource.ComposeTestCheckFunc(

					// vApp checks
					resource.TestCheckResourceAttr("vcd_vapp.template-vm", "name", t.Name()+"-template-vm"),
					resource.TestCheckResourceAttr("vcd_vapp.template-vm", "description", "vApp for Template VM description"),

					resource.TestCheckResourceAttr("vcd_vapp.empty-vm", "name", t.Name()+"-empty-vm"),
					resource.TestCheckResourceAttr("vcd_vapp.empty-vm", "description", "vApp for Empty VM description"),

					// Template vApp VM checks
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "vm_type", "vcd_vapp_vm"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "name", t.Name()+"-template-vapp-vm"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "description", t.Name()+"-template-vapp-vm"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "cpu_hot_add_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "memory_hot_add_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "expose_hardware_virtualization", "false"),

					// Empty vApp VM checks
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "vm_type", "vcd_vapp_vm"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "name", t.Name()+"-empty-vapp-vm"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "description", t.Name()+"-empty-vapp-vm"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "computer_name", "vapp-vm"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "cpus", "1"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "memory", "1024"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "os_type", "sles10_64Guest"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "hardware_version", "vmx-14"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "cpu_hot_add_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "memory_hot_add_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "expose_hardware_virtualization", "false"),

					// Standalone template VM checks
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "vm_type", "vcd_vapp_vm"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "name", t.Name()+"-template-standalone-vm"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "description", t.Name()+"-template-standalone-vm"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "cpu_hot_add_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "memory_hot_add_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "expose_hardware_virtualization", "false"),

					// Standalone empty VM checks
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "vm_type", "vcd_vapp_vm"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "name", t.Name()+"-empty-standalone-vm"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "description", t.Name()+"-standalone"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "cpus", "1"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "memory", "1024"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "os_type", "sles10_64Guest"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "hardware_version", "vmx-14"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "cpu_hot_add_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "memory_hot_add_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "expose_hardware_virtualization", "false"),
				),
			},
			// {
			// 	Config: configTextStep1,
			// 	Check: resource.ComposeTestCheckFunc(
			// 		testAccCheckVcdVAppVmExists(vappNameHwVirt, vmNameHwVirt, "vcd_vapp_vm."+vmNameHwVirt, &vapp, &vm),
			// 		resource.TestCheckResourceAttr(
			// 			"vcd_vapp_vm."+vmNameHwVirt, "name", vmNameHwVirt),
			// 		resource.TestCheckResourceAttr(
			// 			"vcd_vapp_vm."+vmNameHwVirt, "expose_hardware_virtualization", "true"),
			// 	),
			// },
		},
	})
	postTestChecks(t)
}

const testAccVcdVAppVm_4types_Step1 = `
resource "vcd_vapp" "template-vm" {
  org         = "{{.Org}}"
  vdc         = "{{.Vdc}}"
  name        = "{{.TestName}}-template-vm"
  description = "vApp for Template VM description"
}

resource "vcd_vapp" "empty-vm" {
  org         = "{{.Org}}"
  vdc         = "{{.Vdc}}"
  name        = "{{.TestName}}-empty-vm"
  description = "vApp for Empty VM description"
}

resource "vcd_vapp_vm" "template-vm" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"

  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  
  vapp_name   = vcd_vapp.template-vm.name
  name        = "{{.TestName}}-template-vapp-vm"
  description = "{{.TestName}}-template-vapp-vm"
}

resource "vcd_vapp_vm" "empty-vm" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  
  vapp_name     = vcd_vapp.empty-vm.name
  name          = "{{.TestName}}-empty-vapp-vm"
  description   = "{{.TestName}}-empty-vapp-vm"
  computer_name = "vapp-vm"

  cpus   = 1
  memory = 1024

  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"
}

resource "vcd_vm" "template-vm" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"

  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  
  name        = "{{.TestName}}-template-standalone-vm"
  description = "{{.TestName}}-template-standalone-vm"
}

resource "vcd_vm" "empty-vm" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"

  name          = "{{.TestName}}-empty-standalone-vm"
  description   = "{{.TestName}}-standalone"
  computer_name = "standalone"

  cpus   = 1
  memory = 1024

  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"
}
`

// * storage_profile
// * cpu_hot_add_enabled
// * memory_hot_add_enabled
// * computer_name
// * expose_hardware_virtualization
// * metadata
// * guest_properties
func TestAccVcdVAppVm_4types_storage_profile(t *testing.T) {
	preTestChecks(t)

	var params = StringMap{
		"TestName":       t.Name(),
		"Org":            testConfig.VCD.Org,
		"Vdc":            testConfig.Nsxt.Vdc,
		"Catalog":        testConfig.VCD.Catalog.NsxtBackedCatalogName,
		"CatalogItem":    testConfig.VCD.Catalog.NsxtCatalogItem,
		"StorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile2,

		// "EdgeGateway":                  testConfig.Networking.EdgeGateway,
		// "NetworkName":                  "TestAccVcdVAppVmNetHwVirt",
		// "VappName":                     vappNameHwVirt,
		// "VmName":                       vmNameHwVirt,
		// "ExposeHardwareVirtualization": "false",
		"Tags": "vapp vm",
	}
	testParamsNotEmpty(t, params)

	configTextStep1 := templateFill(testAccVcdVAppVm_4types_storage_profile_Step1, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextStep1)
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		// CheckDestroy:      testAccCheckVcdVAppVmDestroy(vappNameHwVirt),
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
		// testAccCheckVcdVAppVmDestroy(vappNameHwVirt),
		),
		Steps: []resource.TestStep{
			{
				Config: configTextStep1,
				Check: resource.ComposeTestCheckFunc(

					// vApp checks
					resource.TestCheckResourceAttr("vcd_vapp.template-vm", "name", t.Name()+"-template-vm"),
					resource.TestCheckResourceAttr("vcd_vapp.template-vm", "description", "vApp for Template VM description"),
					resource.TestCheckResourceAttr("vcd_vapp.empty-vm", "name", t.Name()+"-empty-vm"),
					resource.TestCheckResourceAttr("vcd_vapp.empty-vm", "description", "vApp for Empty VM description"),

					// Template vApp VM checks
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "vm_type", "vcd_vapp_vm"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "name", t.Name()+"-template-vapp-vm"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "description", t.Name()+"-template-vapp-vm"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "storage_profile", params["StorageProfile"].(string)),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "computer_name", "comp-name"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "cpu_hot_add_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "memory_hot_add_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "expose_hardware_virtualization", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "metadata.vm1", "VM Metadata"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "metadata.vm2", "VM Metadata2"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", `guest_properties.guest.hostname`, "test-host"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", `guest_properties.guest.another.subkey`, "another-value"),

					// Empty vApp VM checks
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "vm_type", "vcd_vapp_vm"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "name", t.Name()+"-empty-vapp-vm"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "description", t.Name()+"-empty-vapp-vm"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "computer_name", "comp-name"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "cpus", "1"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "memory", "1024"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "os_type", "rhel8_64Guest"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "hardware_version", "vmx-17"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "storage_profile", params["StorageProfile"].(string)),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "computer_name", "comp-name"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "cpu_hot_add_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "memory_hot_add_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "expose_hardware_virtualization", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "metadata.vm1", "VM Metadata"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "metadata.vm2", "VM Metadata2"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", `guest_properties.guest.hostname`, "test-host"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", `guest_properties.guest.another.subkey`, "another-value"),

					// Standalone template VM checks
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "vm_type", "vcd_vapp_vm"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "name", t.Name()+"-template-standalone-vm"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "description", t.Name()+"-template-standalone-vm"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "storage_profile", params["StorageProfile"].(string)),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "computer_name", "comp-name"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "cpu_hot_add_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "memory_hot_add_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "expose_hardware_virtualization", "true"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "metadata.vm1", "VM Metadata"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "metadata.vm2", "VM Metadata2"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", `guest_properties.guest.hostname`, "test-host"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", `guest_properties.guest.another.subkey`, "another-value"),

					// Standalone empty VM checks
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "vm_type", "vcd_vapp_vm"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "name", t.Name()+"-empty-standalone-vm"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "description", t.Name()+"-standalone"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "cpus", "1"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "memory", "1024"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "os_type", "rhel8_64Guest"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "hardware_version", "vmx-17"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "storage_profile", params["StorageProfile"].(string)),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "computer_name", "comp-name"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "cpu_hot_add_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "memory_hot_add_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "expose_hardware_virtualization", "true"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "metadata.vm1", "VM Metadata"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "metadata.vm2", "VM Metadata2"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", `guest_properties.guest.hostname`, "test-host"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", `guest_properties.guest.another.subkey`, "another-value"),
				),
			},
			// {
			// 	Config: configTextStep1,
			// 	Check: resource.ComposeTestCheckFunc(
			// 		testAccCheckVcdVAppVmExists(vappNameHwVirt, vmNameHwVirt, "vcd_vapp_vm."+vmNameHwVirt, &vapp, &vm),
			// 		resource.TestCheckResourceAttr(
			// 			"vcd_vapp_vm."+vmNameHwVirt, "name", vmNameHwVirt),
			// 		resource.TestCheckResourceAttr(
			// 			"vcd_vapp_vm."+vmNameHwVirt, "expose_hardware_virtualization", "true"),
			// 	),
			// },
		},
	})
	postTestChecks(t)
}

const testAccVcdVAppVm_4types_storage_profile_Step1 = `
data "vcd_storage_profile" "nsxt-vdc" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  name = "{{.StorageProfile}}"
}

resource "vcd_vapp" "template-vm" {
  org         = "{{.Org}}"
  vdc         = "{{.Vdc}}"
  name        = "{{.TestName}}-template-vm"
  description = "vApp for Template VM description"
}

resource "vcd_vapp" "empty-vm" {
  org         = "{{.Org}}"
  vdc         = "{{.Vdc}}"
  name        = "{{.TestName}}-empty-vm"
  description = "vApp for Empty VM description"
}

resource "vcd_vapp_vm" "template-vm" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"

  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  computer_name = "comp-name"
  
  vapp_name   = vcd_vapp.template-vm.name
  name        = "{{.TestName}}-template-vapp-vm"
  description = "{{.TestName}}-template-vapp-vm"

  cpu_hot_add_enabled            = true
  memory_hot_add_enabled         = true
  expose_hardware_virtualization = true

  storage_profile = data.vcd_storage_profile.nsxt-vdc.name

  metadata = {
    "vm1" = "VM Metadata"
    "vm2" = "VM Metadata2"
  }

  guest_properties = {
	"guest.hostname"       = "test-host"
	"guest.another.subkey" = "another-value"
  }
}

resource "vcd_vapp_vm" "empty-vm" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  
  vapp_name     = vcd_vapp.empty-vm.name
  name          = "{{.TestName}}-empty-vapp-vm"
  description   = "{{.TestName}}-empty-vapp-vm"
  computer_name = "comp-name"

  cpus   = 1
  memory = 1024

  cpu_hot_add_enabled            = true
  memory_hot_add_enabled         = true
  expose_hardware_virtualization = true

  os_type          = "rhel8_64Guest"
  hardware_version = "vmx-17"

  storage_profile = data.vcd_storage_profile.nsxt-vdc.name

  metadata = {
    "vm1" = "VM Metadata"
    "vm2" = "VM Metadata2"
  }

  guest_properties = {
	"guest.hostname"       = "test-host"
	"guest.another.subkey" = "another-value"
  }
}

resource "vcd_vm" "template-vm" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"

  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  computer_name = "comp-name"
  
  name        = "{{.TestName}}-template-standalone-vm"
  description = "{{.TestName}}-template-standalone-vm"

  cpu_hot_add_enabled            = true
  memory_hot_add_enabled         = true
  expose_hardware_virtualization = true

  storage_profile = data.vcd_storage_profile.nsxt-vdc.name

  metadata = {
    "vm1" = "VM Metadata"
    "vm2" = "VM Metadata2"
  }

  guest_properties = {
	"guest.hostname"       = "test-host"
	"guest.another.subkey" = "another-value"
  }
}

resource "vcd_vm" "empty-vm" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"

  name          = "{{.TestName}}-empty-standalone-vm"
  description   = "{{.TestName}}-standalone"
  computer_name = "comp-name"

  cpus   = 1
  memory = 1024

  cpu_hot_add_enabled            = true
  memory_hot_add_enabled         = true
  expose_hardware_virtualization = true

  os_type          = "rhel8_64Guest"
  hardware_version = "vmx-17"

  storage_profile = data.vcd_storage_profile.nsxt-vdc.name

  metadata = {
    "vm1" = "VM Metadata"
    "vm2" = "VM Metadata2"
  }

  guest_properties = {
	"guest.hostname"       = "test-host"
	"guest.another.subkey" = "another-value"
  }
}
`
