// +build vapp vm ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdVAppEmptyVm(t *testing.T) {
	var (
		vapp        govcd.VApp
		vm          govcd.VM
		netVappName string = t.Name()
		netVmName1  string = t.Name() + "VM"
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
		"VAppName":    netVappName,
		"VMName":      netVmName1,
		"Tags":        "vapp vm",
		"Media":       testConfig.Media.MediaName,
	}

	// Create objects for testing field values across update steps
	nic0Mac := testCachedFieldValue{}
	nic1Mac := testCachedFieldValue{}
	nic2Mac := testCachedFieldValue{}

	configTextVM := templateFill(testAccCheckVcdVAppEmptyVm, params)

	params["FuncName"] = t.Name() + "-step1"
	configTextVMUpdateStep1 := templateFill(testAccCheckVcdVAppEmptyVmStep1, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVM)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdVAppVmDestroy(netVappName),
		Steps: []resource.TestStep{
			// Step 0 - Create with variations of all possible NICs
			resource.TestStep{
				Config: configTextVM,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVAppVmExists(netVappName, netVmName1, "vcd_vapp_vm."+netVmName1, &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "name", netVmName1),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.name", "multinic-net"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.ip", "11.10.0.152"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.adapter_type", "PCNet32"),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmName1, "network.0.mac"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.connected", "true"),
					nic0Mac.cacheTestResourceFieldValue("vcd_vapp_vm."+netVmName1, "network.0.mac"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.1.name", "multinic-net"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.1.type", "org"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.1.is_primary", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.1.ip_allocation_mode", "DHCP"),
					// resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmName1, "network.1.ip"), // We cannot guarantee DHCP
					resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmName1, "network.1.mac"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.1.connected", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.1.adapter_type", "VMXNET3"),
					nic1Mac.cacheTestResourceFieldValue("vcd_vapp_vm."+netVmName1, "network.1.mac"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.2.name", "multinic-net"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.2.type", "org"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.2.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.2.ip_allocation_mode", "MANUAL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.2.ip", "11.10.0.170"),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmName1, "network.2.mac"),
					// Adapter type is set to "E1000"
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.2.adapter_type", "E1000"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.2.connected", "true"),
					nic2Mac.cacheTestResourceFieldValue("vcd_vapp_vm."+netVmName1, "network.2.mac"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.3.name", "multinic-net2"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.3.type", "org"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.3.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.3.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.3.ip", "12.10.0.152"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.3.connected", "true"),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmName1, "network.3.mac"),
					// Adapter type is set to "E1000E"
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.3.adapter_type", "E1000E"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.3.mac", "00:00:00:11:11:11"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.4.name", ""),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.4.type", "none"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.4.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.4.ip_allocation_mode", "NONE"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.4.ip", ""),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmName1, "network.4.mac"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.4.connected", "false"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.5.name", ""),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.5.type", "none"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.5.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.5.ip_allocation_mode", "NONE"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.5.ip", ""),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmName1, "network.5.mac"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.5.connected", "false"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.6.name", "vapp-net"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.6.type", "vapp"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.6.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.6.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.6.ip", "192.168.2.51"),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmName1, "network.6.mac"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.6.connected", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.6.adapter_type", "VMXNET3"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.7.name", "vapp-routed-net"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.7.type", "vapp"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.7.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.7.ip_allocation_mode", "MANUAL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.7.connected", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.7.ip", "192.168.2.2"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.8.name", "multinic-net"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.8.type", "org"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.8.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.8.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.8.connected", "true"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.9.name", "multinic-net2"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.9.type", "org"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.9.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.9.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.9.connected", "true"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "os_type", "sles11_64Guest"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "hardware_version", "vmx-13"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "expose_hardware_virtualization", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "computer_name", "compName"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "description", "test empty VM"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "cpu_hot_add_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "memory_hot_add_enabled", "true"),
				),
			},
			// Step 1 - update
			resource.TestStep{
				Config: configTextVMUpdateStep1,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVAppVmExists(netVappName, netVmName1, "vcd_vapp_vm."+netVmName1, &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "name", netVmName1),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.name", "multinic-net"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.ip", "11.10.0.152"),
					// Ensure that the MAC address (and the NIC itself) stays the same after update procedure
					resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmName1, "network.0.mac"),
					// Ensuring adapter type stays intact after update
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.adapter_type", "PCNet32"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.connected", "true"),
					nic0Mac.testCheckCachedResourceFieldValue("vcd_vapp_vm."+netVmName1, "network.0.mac"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.1.name", "multinic-net"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.1.type", "org"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.1.is_primary", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.1.ip_allocation_mode", "DHCP"),
					//resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmName1, "network.1.ip"), // We cannot guarantee DHCP
					resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmName1, "network.1.mac"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.1.connected", "true"),
					// Ensuring adapter type stays intact after update
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.1.adapter_type", "VMXNET3"),
					nic1Mac.testCheckCachedResourceFieldValue("vcd_vapp_vm."+netVmName1, "network.1.mac"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.2.name", "multinic-net"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.2.type", "org"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.2.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.2.ip_allocation_mode", "MANUAL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.2.ip", "11.10.0.170"),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmName1, "network.2.mac"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.2.connected", "true"),
					// Ensuring adapter type stays intact after update
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.2.adapter_type", "E1000"),
					nic2Mac.testCheckCachedResourceFieldValue("vcd_vapp_vm."+netVmName1, "network.2.mac"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.3.name", "multinic-net2"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.3.type", "org"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.3.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.3.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.3.ip", "12.10.0.152"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.3.mac", "00:00:00:11:11:11"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.3.connected", "true"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "os_type", "rhel4Guest"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "hardware_version", "vmx-14"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "expose_hardware_virtualization", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "computer_name", "compNameUp"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "description", "test empty VM updated"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "cpu_hot_add_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "memory_hot_add_enabled", "false"),
				),
			},
		},
	})
}

const testAccCheckVcdVAppEmpty = `
resource "vcd_vapp" "{{.VAppName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name       = "{{.VAppName}}"
  depends_on = ["vcd_network_routed.net", "vcd_network_routed.net2"]
}
`

const testAccCheckVcdVAppEmptyVmNetworkShared = `
resource "vcd_vapp_network" "vappIsolatedNet" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name       = "vapp-net"
  vapp_name  = vcd_vapp.{{.VAppName}}.name
  gateway    = "192.168.2.1"
  netmask    = "255.255.255.0"
  dns1       = "192.168.2.1"
  dns2       = "192.168.2.2"
  dns_suffix = "mybiz.biz"

  static_ip_pool {
    start_address = "192.168.2.51"
    end_address   = "192.168.2.100"
  }
}

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

resource "vcd_vapp_org_network" "vappAttachedNet" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name        = vcd_vapp.{{.VAppName}}.name
  org_network_name = vcd_network_routed.net.name
}

resource "vcd_vapp_org_network" "vappAttachedRoutedNet2" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name        = vcd_vapp.{{.VAppName}}.name
  org_network_name = vcd_network_routed.net2.name
  is_fenced        = true
}

resource "vcd_vapp_network" "vappRoutedNet" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name             = "vapp-routed-net"
  vapp_name        = vcd_vapp.{{.VAppName}}.name
  gateway          = "192.168.2.1"
  netmask          = "255.255.255.0"
  org_network_name = vcd_network_routed.net.name
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

const testAccCheckVcdVAppEmptyVm = testAccCheckVcdVAppEmpty + testAccCheckVcdVAppEmptyVmNetworkShared + `
resource "vcd_vapp_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  # You cannot remove NICs from an active virtual machine on which no operating system is installed.
  power_on = false

  vapp_name     = vcd_vapp.{{.VAppName}}.name
  description   = "test empty VM"
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
    name               = vcd_vapp_org_network.vappAttachedNet.org_network_name
    ip_allocation_mode = "POOL"
    is_primary         = false
	adapter_type       = "PCNet32"
  }

  network {
    type               = "org"
    name               = vcd_vapp_org_network.vappAttachedNet.org_network_name
    ip_allocation_mode = "DHCP"
    is_primary         = true
  }

  network {
    type               = "org"
    name               = vcd_vapp_org_network.vappAttachedNet.org_network_name
    ip                 = "11.10.0.170"
    ip_allocation_mode = "MANUAL"
    is_primary         = false
    adapter_type       = "e1000"
  }

  network {
    type               = "org"
    name               = vcd_vapp_org_network.vappAttachedRoutedNet2.org_network_name
    ip_allocation_mode = "POOL"
    is_primary         = false
    adapter_type       = "e1000e"
	mac                = "00:00:00:11:11:11"
  }

  network {
    type               = "none"
    ip_allocation_mode = "NONE"
    ip                 = ""
    name               = ""
    connected          = false
  }

  network {
    type               = "none"
    ip_allocation_mode = "NONE"
    connected          = false
  }

  network {
    type               = "vapp"
    name               = vcd_vapp_network.vappIsolatedNet.name
    ip_allocation_mode = "POOL"
    adapter_type       = "VMXNET3"
  }

  network {
    type               = "vapp"
    name               = vcd_vapp_network.vappRoutedNet.name
    ip_allocation_mode = "MANUAL"
    ip                 = "192.168.2.2"
  }

  network {
    type               = "org"
    name               = vcd_vapp_org_network.vappAttachedNet.org_network_name
    ip_allocation_mode = "POOL"
  }

  network {
    type               = "org"
    name              = vcd_vapp_org_network.vappAttachedRoutedNet2.org_network_name
    ip_allocation_mode = "POOL"
  }
 }
`

const testAccCheckVcdVAppEmptyVmStep1 = testAccCheckVcdVAppEmptyVmNetworkShared + `
# skip-binary-test: only for updates
resource "vcd_vapp" "{{.VAppName}}" {
	org = "{{.Org}}"
	vdc = "{{.Vdc}}"

	name       = "{{.VAppName}}"
	depends_on = ["vcd_network_routed.net", "vcd_network_routed.net2"]
}

resource "vcd_vapp_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name     = vcd_vapp.{{.VAppName}}.name
  name          = "{{.VMName}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1
  description   = "test empty VM updated"

  os_type                        = "rhel4Guest"
  hardware_version               = "vmx-14"
  catalog_name                   = ""
  boot_image                     = ""
  expose_hardware_virtualization = false
  computer_name                  = "compNameUp"

  network {
    type               = "org"
    name               = vcd_vapp_org_network.vappAttachedNet.org_network_name
    ip_allocation_mode = "POOL"
    is_primary         = false
  }

  network {
    type               = "org"
    name               = vcd_vapp_org_network.vappAttachedNet.org_network_name
    ip_allocation_mode = "DHCP"
    is_primary         = true
  }

  network {
    type               = "org"
    name               = vcd_vapp_org_network.vappAttachedNet.org_network_name
    ip                 = "11.10.0.170"
    ip_allocation_mode = "MANUAL"
    is_primary         = false
  }

  network {
    type               = "org"
    name               = vcd_vapp_org_network.vappAttachedRoutedNet2.org_network_name
    ip_allocation_mode = "POOL"
    is_primary         = false
	mac                = "00:00:00:11:11:11"
  }

  network {
    type               = "vapp"
    name               = vcd_vapp_network.vappRoutedNet.name
    ip_allocation_mode = "MANUAL"
    ip                 = "192.168.2.2"
  }

  network {
    type               = "org"
    name               = vcd_vapp_org_network.vappAttachedNet.org_network_name
    ip_allocation_mode = "POOL"
  }

  network {
    type               = "org"
    name              = vcd_vapp_org_network.vappAttachedRoutedNet2.org_network_name
    ip_allocation_mode = "POOL"
  } 
}
`
