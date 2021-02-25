// +build standaloneVm vm ALL functional
// +build !skipStandaloneVm

package vcd

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdStandaloneHotUpdateVm(t *testing.T) {
	var standaloneVmName = fmt.Sprintf("%s-%d", t.Name(), os.Getpid())

	if testConfig.Media.MediaName == "" {
		fmt.Println("Warning: `MediaName` is not configured: boot image won't be tested.")
	}

	if testConfig.VCD.ProviderVdc.StorageProfile == "" || testConfig.VCD.ProviderVdc.StorageProfile2 == "" {
		t.Skip("Both variables testConfig.VCD.ProviderVdc.StorageProfile and testConfig.VCD.ProviderVdc.StorageProfile2 must be set")
	}

	var params = StringMap{
		"Org":             testConfig.VCD.Org,
		"Vdc":             testConfig.VCD.Vdc,
		"EdgeGateway":     testConfig.Networking.EdgeGateway,
		"Catalog":         testSuiteCatalogName,
		"CatalogItem":     testSuiteCatalogOVAItem,
		"VMName":          standaloneVmName,
		"Tags":            "standaloneVm vm",
		"Media":           testConfig.Media.MediaName,
		"StorageProfile":  testConfig.VCD.ProviderVdc.StorageProfile,
		"StorageProfile2": testConfig.VCD.ProviderVdc.StorageProfile2,
	}

	configTextVM := templateFill(testAccCheckVcdHotUpdateVm, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVM)

	params["FuncName"] = t.Name() + "-step1"
	configTextVMUpdateStep1 := templateFill(testAccCheckVcdHotUpdateVmStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVMUpdateStep1)

	params["FuncName"] = t.Name() + "-step2"
	configTextVMUpdateStep2 := templateFill(testAccCheckVcdHotUpdateVmStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVMUpdateStep2)

	params["FuncName"] = t.Name() + "-step3"
	configTextVMUpdateStep3 := templateFill(testAccCheckVcdHotUpdateVmStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVMUpdateStep3)

	params["FuncName"] = t.Name() + "-step4"
	configTextVMUpdateStep4 := templateFill(testAccCheckVcdHotUpdateVmStep4, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVMUpdateStep4)

	params["FuncName"] = t.Name() + "-step5"
	configTextVMUpdateStep5 := templateFill(testAccCheckVcdHotUpdateVmStep5, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVMUpdateStep5)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdStandaloneVmDestroy(standaloneVmName, "", ""),
		Steps: []resource.TestStep{
			// Step 0 - create
			resource.TestStep{
				Config: configTextVM,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdStandaloneVmExists(standaloneVmName, "vcd_vm."+standaloneVmName, "", ""),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "name", standaloneVmName),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "cpu_hot_add_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "memory_hot_add_enabled", "true"),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "memory", "2048"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "cpus", "1"),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.ip_allocation_mode", "NONE"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.connected", "false"),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.is_primary", "true"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.name", "multinic-net"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.ip_allocation_mode", "DHCP"),
					resource.TestCheckResourceAttrSet("vcd_vm."+standaloneVmName, "network.1.mac"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.connected", "true"),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "metadata.mediaItem_metadata", "data 1"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "metadata.mediaItem_metadata2", "data 2"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "metadata.mediaItem_metadata3", "data 3"),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, `guest_properties.guest.hostname`, "test-host"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, `guest_properties.guest.another.subkey`, "another-value"),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, `storage_profile`, params["StorageProfile"].(string)),
				),
			},
			// Step 1 - update - network changes
			resource.TestStep{
				Config: configTextVMUpdateStep1,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdStandaloneVmExists(standaloneVmName, "vcd_vm."+standaloneVmName, "", ""),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "name", standaloneVmName),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "cpu_hot_add_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "memory_hot_add_enabled", "true"),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "memory", "3072"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "cpus", "3"),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.name", "multinic-net"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.ip_allocation_mode", "DHCP"),
					resource.TestCheckResourceAttrSet("vcd_vm."+standaloneVmName, "network.0.mac"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.connected", "true"),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.is_primary", "true"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.ip_allocation_mode", "NONE"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.connected", "false"),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "metadata.mediaItem_metadata", "data 1"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "metadata.mediaItem_metadata2", "data 3"),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, `guest_properties.guest.hostname`, "test-host2"),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, `storage_profile`, params["StorageProfile2"].(string)),
				),
			},
			// Step 2 - update
			resource.TestStep{
				Config:      configTextVMUpdateStep2,
				ExpectError: regexp.MustCompile(`update stopped: VM needs to power off to change properties.*`),
			},
			// Step 3 - update - add new network section
			resource.TestStep{
				Config: configTextVMUpdateStep3,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdStandaloneVmExists(standaloneVmName, "vcd_vm."+standaloneVmName, "", ""),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "name", standaloneVmName),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "cpu_hot_add_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "memory_hot_add_enabled", "true"),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "memory", "3072"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "cpus", "3"),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.name", "multinic-net"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.ip_allocation_mode", "DHCP"),
					resource.TestCheckResourceAttrSet("vcd_vm."+standaloneVmName, "network.0.mac"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.connected", "true"),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.is_primary", "true"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.ip_allocation_mode", "NONE"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.connected", "false"),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.2.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.2.ip_allocation_mode", "NONE"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.2.connected", "false"),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, `storage_profile`, params["StorageProfile2"].(string)),
				),
			},
			// Step 4 - update - network changes
			resource.TestStep{
				Config: configTextVMUpdateStep5,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdStandaloneVmExists(standaloneVmName, "vcd_vm."+standaloneVmName, "", ""),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "name", standaloneVmName),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "cpu_hot_add_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "memory_hot_add_enabled", "true"),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "memory", "3072"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "cpus", "3"),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.ip_allocation_mode", "NONE"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.connected", "false"),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.name", "multinic-net"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.is_primary", "true"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.ip_allocation_mode", "DHCP"),
					resource.TestCheckResourceAttrSet("vcd_vm."+standaloneVmName, "network.1.mac"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.connected", "true"),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, `storage_profile`, params["StorageProfile2"].(string)),
				),
			},
		},
	})
}

const testStandaloneSharedHotUpdate = `
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

`

const testAccCheckVcdHotUpdateVm = testStandaloneSharedHotUpdate + `
resource "vcd_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  power_on = true

  name          = "{{.VMName}}"
  computer_name = "compNameUp"

  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"

  memory        = 2048
  cpus          = 1

  cpu_hot_add_enabled    = true
  memory_hot_add_enabled = true

  network {
    type               = "none"
    ip_allocation_mode = "NONE"
    connected          = false
  }
 
  network {
    type               = "org"
    name               = vcd_network_routed.net.name
    ip_allocation_mode = "DHCP"
    is_primary         = true
  }

  metadata = {
    mediaItem_metadata = "data 1"
    mediaItem_metadata2 = "data 2"
    mediaItem_metadata3 = "data 3"
  }

  guest_properties = {
	"guest.hostname"       = "test-host"
	"guest.another.subkey" = "another-value"
  }

  storage_profile = "{{.StorageProfile}}"
 }
`

const testAccCheckVcdHotUpdateVmStep1 = `# skip-binary-test: only for updates
` + testStandaloneSharedHotUpdate + `
resource "vcd_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  computer_name = "compNameUp"
  name          = "{{.VMName}}"

  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
 
  memory        = 3072
  cpus          = 3

  cpu_hot_add_enabled    = true
  memory_hot_add_enabled = true

  network {
    type               = "org"
    name               = vcd_network_routed.net.name
    ip_allocation_mode = "DHCP"
  }
 
  network {
    type               = "none"
    ip_allocation_mode = "NONE"
    connected          = false
    is_primary         = true
  }

  metadata = {
    mediaItem_metadata = "data 1"
    mediaItem_metadata2 = "data 3"
  }

  guest_properties = {
	"guest.hostname"       = "test-host2"
  }

  storage_profile = "{{.StorageProfile2}}"
}
`

const testAccCheckVcdHotUpdateVmStep2 = `# skip-binary-test: only for updates
` + testStandaloneSharedHotUpdate + `
resource "vcd_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  computer_name = "compNameUp"
  name          = "{{.VMName}}"

  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
 
  memory        = 3072
  cpus          = 3

  cpu_hot_add_enabled    = false
  memory_hot_add_enabled = true

  prevent_update_power_off = true

  network {
    type               = "org"
    name               = vcd_network_routed.net.name
    ip_allocation_mode = "DHCP"
  }
 
  network {
    type               = "none"
    ip_allocation_mode = "NONE"
    connected          = false
    is_primary         = true
  }

  storage_profile = "{{.StorageProfile2}}"
}
`
const testAccCheckVcdHotUpdateVmStep3 = `# skip-binary-test: only for updates
` + testStandaloneSharedHotUpdate + `
resource "vcd_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  computer_name = "compNameUp"
  name          = "{{.VMName}}"

  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
 
  memory        = 3072
  cpus          = 3

  cpu_hot_add_enabled    = true
  memory_hot_add_enabled = true

  prevent_update_power_off = true

  network {
    type               = "org"
    name               = vcd_network_routed.net.name
    ip_allocation_mode = "DHCP"
  }
 
  network {
    type               = "none"
    ip_allocation_mode = "NONE"
    connected          = false
    is_primary         = true
  }

  network {
    type               = "none"
    ip_allocation_mode = "NONE"
    connected          = false
  }

  storage_profile = "{{.StorageProfile2}}"
}
`

const testAccCheckVcdHotUpdateVmStep4 = `# skip-binary-test: only for updates
` + testStandaloneSharedHotUpdate + `
resource "vcd_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  computer_name = "compNameUp"
  name          = "{{.VMName}}"

  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
 
  memory        = 3072
  cpus          = 3

  cpu_hot_add_enabled    = true
  memory_hot_add_enabled = true

  prevent_update_power_off = true

  network {
    type               = "none"
    ip_allocation_mode = "NONE"
    connected          = false
    is_primary         = true
  }

  network {
    type               = "none"
    ip_allocation_mode = "NONE"
    connected          = false
    is_primary         = false
  }

  storage_profile = "{{.StorageProfile2}}"
}
`

const testAccCheckVcdHotUpdateVmStep5 = `# skip-binary-test: only for updates
` + testStandaloneSharedHotUpdate + `
resource "vcd_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  computer_name = "compNameUp"
  name          = "{{.VMName}}"

  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
 
  memory        = 3072
  cpus          = 3

  cpu_hot_add_enabled    = true
  memory_hot_add_enabled = true

  prevent_update_power_off = false

  network {
    type               = "none"
    ip_allocation_mode = "NONE"
    connected          = false
    is_primary         = false
  }

  network {
    type               = "org"
    name               = vcd_network_routed.net.name
    ip_allocation_mode = "DHCP"
    is_primary         = true
  }
 
  storage_profile = "{{.StorageProfile2}}"
}
`
