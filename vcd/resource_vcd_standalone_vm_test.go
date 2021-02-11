// +build vm ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	testingTags["vm"] = "resource_vcd_vm_test.go"
}

func TestAccVcdStandaloneVm_Basic(t *testing.T) {
	var standaloneVmName = "TestStandaloneVmTemplate"
	//var vapp govcd.VApp
	//var vm govcd.VM
	var diskResourceName = "TestAccVcdVAppVm_Basic_1"
	var diskName = "TestAccVcdIndependentDiskBasic"

	var params = StringMap{
		"Org":                testConfig.VCD.Org,
		"Vdc":                testConfig.VCD.Vdc,
		"EdgeGateway":        testConfig.Networking.EdgeGateway,
		"NetworkName":        "TestAccVcdVAppVmNet",
		"Catalog":            testSuiteCatalogName,
		"CatalogItem":        testSuiteCatalogOVAItem,
		"VmName":             standaloneVmName,
		"ComputerName":       standaloneVmName + "-unique",
		"diskName":           diskName,
		"size":               "5",
		"busType":            "SCSI",
		"busSubType":         "lsilogicsas",
		"storageProfileName": "*",
		"diskResourceName":   diskResourceName,
		"Tags":               "vm",
	}

	configText := templateFill(testAccCheckVcdStandaloneVm_basic, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		//CheckDestroy:      testAccCheckVcdVAppVmDestroy(vappName2),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					//testAccCheckVcdVAppVmExists(vappName2, standaloneVmName, "vcd_vm."+standaloneVmName, &vapp, &vm),
					resource.TestCheckResourceAttr(
						"vcd_vm."+standaloneVmName, "name", standaloneVmName),
					resource.TestCheckResourceAttr(
						"vcd_vm."+standaloneVmName, "description", "test standalone VM"),
					resource.TestCheckResourceAttr(
						"vcd_vm."+standaloneVmName, "computer_name", standaloneVmName+"-unique"),
					resource.TestCheckResourceAttr(
						"vcd_vm."+standaloneVmName, "network.0.ip", "10.10.102.161"),
					resource.TestCheckResourceAttr(
						"vcd_vm."+standaloneVmName, "power_on", "true"),
					resource.TestCheckResourceAttr(
						"vcd_vm."+standaloneVmName, "metadata.vm_metadata", "VM Metadata."),
					resource.TestCheckOutput("disk", diskName),
					resource.TestCheckOutput("disk_bus_number", "1"),
					resource.TestCheckOutput("disk_unit_number", "0"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_vm."+standaloneVmName, "disk.*", map[string]string{
						"size_in_mb": "5",
					}),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_vm." + standaloneVmName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgVdcObject(testConfig, standaloneVmName),
				// These fields can't be retrieved from user data
				ImportStateVerifyIgnore: []string{"template_name", "catalog_name",
					"accept_all_eulas", "power_on", "computer_name", "prevent_update_power_off"},
			},
		},
	})
}

func TestAccVcdStandaloneEmptyVm(t *testing.T) {
	var (
		//vapp        govcd.VApp
		//vm          govcd.VM
		netVmName1 string = t.Name() + "VM"
	)

	if testConfig.Media.MediaName == "" {
		fmt.Println("Warning: `MediaName` is not configured: boot image won't be tested.")
	}

	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"Catalog":     testSuiteCatalogName,
		"CatalogItem": testSuiteCatalogOVAItem,
		"VAppName":    "",
		"VMName":      netVmName1,
		"Tags":        "vm",
		"Media":       testConfig.Media.MediaName,
	}

	// Create objects for testing field values across update steps
	nic0Mac := testCachedFieldValue{}
	nic1Mac := testCachedFieldValue{}
	//nic2Mac := testCachedFieldValue{}

	configTextVM := templateFill(testAccCheckVcdStandaloneEmptyVm, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVM)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		//CheckDestroy:      testAccCheckVcdVAppVmDestroy(netVappName),
		Steps: []resource.TestStep{
			// Step 0 - Create with variations of all possible NICs
			resource.TestStep{
				Config: configTextVM,
				Check: resource.ComposeAggregateTestCheckFunc(
					//testAccCheckVcdVAppVmExists(netVappName, netVmName1, "vcd_vm."+netVmName1, &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "name", netVmName1),

					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.0.name", "multinic-net"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.0.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.0.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.0.ip", "11.10.0.152"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.0.adapter_type", "PCNet32"),
					resource.TestCheckResourceAttrSet("vcd_vm."+netVmName1, "network.0.mac"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.0.connected", "true"),
					nic0Mac.cacheTestResourceFieldValue("vcd_vm."+netVmName1, "network.0.mac"),

					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.1.name", "multinic-net2"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.1.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.1.is_primary", "true"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.1.ip_allocation_mode", "DHCP"),
					// resource.TestCheckResourceAttrSet("vcd_vm."+netVmName1, "network.1.ip"), // We cannot guarantee DHCP
					resource.TestCheckResourceAttrSet("vcd_vm."+netVmName1, "network.1.mac"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.1.connected", "true"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.1.adapter_type", "VMXNET3"),
					nic1Mac.cacheTestResourceFieldValue("vcd_vm."+netVmName1, "network.1.mac"),

					/*
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.2.name", "multinic-net"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.2.type", "org"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.2.is_primary", "false"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.2.ip_allocation_mode", "MANUAL"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.2.ip", "11.10.0.170"),
						resource.TestCheckResourceAttrSet("vcd_vm."+netVmName1, "network.2.mac"),
						// Adapter type is set to "E1000"
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.2.adapter_type", "E1000"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.2.connected", "true"),
						nic2Mac.cacheTestResourceFieldValue("vcd_vm."+netVmName1, "network.2.mac"),

						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.3.name", "multinic-net2"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.3.type", "org"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.3.is_primary", "false"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.3.ip_allocation_mode", "POOL"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.3.ip", "12.10.0.152"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.3.connected", "true"),
						resource.TestCheckResourceAttrSet("vcd_vm."+netVmName1, "network.3.mac"),
						// Adapter type is set to "E1000E"
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.3.adapter_type", "E1000E"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.3.mac", "00:00:00:11:11:11"),

						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.4.name", ""),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.4.type", "none"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.4.is_primary", "false"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.4.ip_allocation_mode", "NONE"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.4.ip", ""),
						resource.TestCheckResourceAttrSet("vcd_vm."+netVmName1, "network.4.mac"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.4.connected", "false"),

						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.5.name", ""),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.5.type", "none"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.5.is_primary", "false"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.5.ip_allocation_mode", "NONE"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.5.ip", ""),
						resource.TestCheckResourceAttrSet("vcd_vm."+netVmName1, "network.5.mac"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.5.connected", "false"),

						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.6.name", "vapp-net"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.6.type", "vapp"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.6.is_primary", "false"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.6.ip_allocation_mode", "POOL"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.6.ip", "192.168.2.51"),
						resource.TestCheckResourceAttrSet("vcd_vm."+netVmName1, "network.6.mac"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.6.connected", "true"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.6.adapter_type", "VMXNET3"),

						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.7.name", "vapp-routed-net"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.7.type", "vapp"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.7.is_primary", "false"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.7.ip_allocation_mode", "MANUAL"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.7.connected", "true"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.7.ip", "192.168.2.2"),

						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.8.name", "multinic-net"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.8.type", "org"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.8.is_primary", "false"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.8.ip_allocation_mode", "POOL"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.8.connected", "true"),

						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.9.name", "multinic-net2"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.9.type", "org"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.9.is_primary", "false"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.9.ip_allocation_mode", "POOL"),
						resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "network.9.connected", "true"),
					*/

					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "os_type", "sles11_64Guest"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "hardware_version", "vmx-13"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "expose_hardware_virtualization", "true"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "computer_name", "compName"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "description", "test empty standalone VM"),

					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "cpu_hot_add_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "memory_hot_add_enabled", "true"),
				),
			},
		},
	})
}

const testAccCheckVcdStandaloneVm_basic = `
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
  size_in_mb      = "{{.size}}"
  bus_type        = "{{.busType}}"
  bus_sub_type    = "{{.busSubType}}"
  storage_profile = "{{.storageProfileName}}"
}

resource "vcd_vm" "{{.VmName}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  name          = "{{.VmName}}"
  computer_name = "{{.ComputerName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  description   = "test standalone VM"
  memory        = 1024
  cpus          = 2
  cpu_cores     = 1

  metadata = {
    vm_metadata = "VM Metadata."
  }

  network {
    type               = "org"
    name               = vcd_network_routed.{{.NetworkName}}.name
    ip_allocation_mode = "MANUAL"
    ip                 = "10.10.102.161"
  }

  disk {
    name        = vcd_independent_disk.{{.diskResourceName}}.name
    bus_number  = 1
    unit_number = 0
  }
}

output "disk" {
  value = tolist(vcd_vm.{{.VmName}}.disk)[0].name
}
output "disk_bus_number" {
  value = tolist(vcd_vm.{{.VmName}}.disk)[0].bus_number
}
output "disk_unit_number" {
  value = tolist(vcd_vm.{{.VmName}}.disk)[0].unit_number
}
output "vm" {
  value = vcd_vm.{{.VmName}}
}
`

const testAccCheckVcdStandaloneEmptyVmNetworkShared = `
resource "vcd_network_routed" "net" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name         = "multinic-net"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "11.10.0.1"

  dhcp_pool {
    start_address = "11.10.0.2"
    end_address   = "11.10.0.100"
  }

  static_ip_pool {
    start_address = "11.10.0.152"
    end_address   = "11.10.0.254"
  }
}

resource "vcd_network_routed" "net2" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name         = "multinic-net2"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "12.10.0.1"

  static_ip_pool {
    start_address = "12.10.0.152"
    end_address   = "12.10.0.254"
  }
}
`

const testAccCheckVcdStandaloneEmptyVm = testAccCheckVcdStandaloneEmptyVmNetworkShared + `
resource "vcd_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  # You cannot remove NICs from an active virtual machine on which no operating system is installed.
  power_on = false

  description   = "test empty standalone VM"
  name          = "{{.VMName}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1 
  
  os_type                        = "sles11_64Guest"
  hardware_version               = "vmx-13"
  catalog_name                   = "{{.Catalog}}"
  boot_image                     = "{{.Media}}"
  expose_hardware_virtualization = true
  computer_name                  = "compName"

  cpu_hot_add_enabled    = true
  memory_hot_add_enabled = true

  network {
    type               = "org"
    name               = vcd_network_routed.net.name
    ip_allocation_mode = "POOL"
    is_primary         = false
	adapter_type       = "PCNet32"
  }

  network {
    type               = "org"
    name               = vcd_network_routed.net2.name
    ip_allocation_mode = "DHCP"
    is_primary         = true
  }

 }
`
