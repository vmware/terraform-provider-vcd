//go:build vapp || vm || ALL || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	testingTags["vm"] = "resource_vcd_vapp_vm_copy_test.go"
}

func TestAccVcdVAppVmCopyBasic(t *testing.T) {
	preTestChecks(t)

	var params = StringMap{
		"TestName":      t.Name(),
		"Org":           testConfig.VCD.Org,
		"Vdc":           testConfig.Nsxt.Vdc,
		"RoutedNetwork": testConfig.Nsxt.RoutedNetwork,

		"VappName":     t.Name() + "-dest-vapp",
		"ComputerName": "ComputerName",

		"Tags": "vapp vm",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccVcdVAppVmCopyBasic, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdNsxtVAppVmDestroy(params["VappName"].(string)),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_vm.source", "name", t.Name()+"-source"),
					resource.TestCheckResourceAttr("vcd_vm.source", "power_on", "false"),
					resource.TestCheckResourceAttr("vcd_vm.source", "network.#", "0"),

					resource.TestCheckResourceAttr("vcd_vapp.destination", "power_on", "false"),

					resource.TestCheckResourceAttr("vcd_vapp_vm.copy", "power_on", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.copy", "network.#", "1"),

					resource.TestCheckResourceAttr("vcd_vm.copy", "power_on", "false"),
					resource.TestCheckResourceAttr("vcd_vm.copy", "network.#", "1"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdVAppVmCopyBasic = `
resource "vcd_vm" "source" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  power_on = false

  name        = "{{.TestName}}-source"
  memory      = 512
  cpus        = 2
  cpu_cores   = 1

  os_type          = "sles11_64Guest"
  hardware_version = "vmx-13"
  computer_name    = "compName"
}

resource "vcd_vapp" "destination" {
  name     = "{{.VappName}}"
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

  network {
    type               = "org"
    name               = "{{.RoutedNetwork}}"
    ip_allocation_mode = "POOL"
  }
}
`

// TestAccVcdVAppVm_BasicCopySameVdc_ShrinkNics checks if having less networks on
// VM copies than in the source VM does not break the code.
func TestAccVcdVAppVmCopyShrinkNics(t *testing.T) {
	preTestChecks(t)

	var params = StringMap{
		"TestName":        t.Name(),
		"Org":             testConfig.VCD.Org,
		"Vdc":             testConfig.Nsxt.Vdc,
		"RoutedNetwork":   testConfig.Nsxt.RoutedNetwork,
		"IsolatedNetwork": testConfig.Nsxt.IsolatedNetwork,
		"EdgeGateway":     testConfig.Nsxt.EdgeGateway,

		"VappName":     vappName2,
		"VmName":       vmName,
		"ComputerName": vmName + "-unique",

		"Tags": "vapp vm",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccVcdVAppVmCopyShrinkNics, params)
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
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_vm.source", "name", t.Name()+"-source"),
					resource.TestCheckResourceAttr("vcd_vm.source", "power_on", "true"),
					resource.TestCheckResourceAttr("vcd_vm.source", "network.#", "2"),

					resource.TestCheckResourceAttr("vcd_vapp.destination", "power_on", "true"),

					resource.TestCheckResourceAttr("vcd_vapp_vm.copy", "power_on", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.copy", "network.#", "1"),

					resource.TestCheckResourceAttr("vcd_vm.copy", "power_on", "true"),
					resource.TestCheckResourceAttr("vcd_vm.copy", "network.#", "1"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdVAppVmCopyShrinkNics = `
data "vcd_org_vdc" "main" {
  org  = "{{.Org}}"
  name = "{{.Vdc}}"
}

data "vcd_nsxt_edgegateway" "main" {
  org  = "{{.Org}}"
  name = "{{.EdgeGateway}}"
}

data "vcd_network_isolated_v2" "net" {
  org      = "{{.Org}}"
  owner_id = data.vcd_org_vdc.main.id
  name     = "{{.IsolatedNetwork}}"
}

data "vcd_network_routed_v2" "net" {
  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.main.id
  name            = "{{.RoutedNetwork}}"
}

resource "vcd_vm" "source" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  power_on = true

  description = "test empty VM"
  name        = "{{.TestName}}-source"
  memory      = 512
  cpus        = 2
  cpu_cores   = 1

  os_type          = "sles11_64Guest"
  hardware_version = "vmx-13"
  computer_name    = "compName"

  network {
    type               = "org"
    name               = data.vcd_network_isolated_v2.net.name
    ip_allocation_mode = "POOL"
  }

  network {
    type               = "org"
    name               = data.vcd_network_routed_v2.net.name
    ip_allocation_mode = "POOL"
  }
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

resource "vcd_vapp_vm" "copy" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  vapp_name     = vcd_vapp.destination.name
  name          = "{{.TestName}}-dest"
  computer_name = "{{.ComputerName}}"

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

  copy_from_vm_id  = vcd_vm.source.id
  power_on         = true
  memory           = 1024
  cpus             = 2
  cpu_cores        = 1

  network {
    type               = "org"
    name               = data.vcd_network_isolated_v2.net.name
    ip_allocation_mode = "POOL"
  }
}
`

func TestAccVcdVAppVmCopyDifferentVdc(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"TestName":       t.Name(),
		"Org":            testConfig.VCD.Org,
		"Vdc":            testConfig.Nsxt.Vdc,
		"DestinationVdc": t.Name() + "-vdc",
		"ProviderVdc":    testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":    testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"StorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile2,

		"VappName":     vappName2,
		"VmName":       vmName,
		"ComputerName": vmName + "-unique",

		"Tags": "vapp vm",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccVcdVAppVmCopyDifferentVdc, params)
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
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_vapp_vm.source", "vdc", params["Vdc"].(string)),
					resource.TestCheckResourceAttr("vcd_vapp_vm.source", "power_on", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.source", "network.#", "0"),

					resource.TestCheckResourceAttr("vcd_vapp.destination", "power_on", "true"),
					resource.TestCheckResourceAttr("vcd_vapp.destination", "vdc", params["DestinationVdc"].(string)),

					resource.TestCheckResourceAttr("vcd_vapp_vm.copy", "power_on", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.copy", "network.#", "1"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.copy", "vdc", params["DestinationVdc"].(string)),

					resource.TestCheckResourceAttr("vcd_vm.copy", "power_on", "true"),
					resource.TestCheckResourceAttr("vcd_vm.copy", "network.#", "0"),
					resource.TestCheckResourceAttr("vcd_vm.copy", "vdc", params["DestinationVdc"].(string)),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdVAppVmCopyDifferentVdc = `
resource "vcd_vm_sizing_policy" "minSize" {
  name = "min-size"
}

resource "vcd_org_vdc" "sizing-policy" {
  org  = "{{.Org}}"
  name = "{{.TestName}}-vdc"

  allocation_model  = "Flex"
  network_pool_name = "{{.NetworkPool}}"
  provider_vdc_name = "{{.ProviderVdc}}"

  compute_capacity {
    cpu {
      allocated = "0"
      limit     = "24000"
    }

    memory {
      allocated = "0"
      limit     = "24000"
    }
  }

  storage_profile {
    name     = "{{.StorageProfile}}"
    enabled  = true
    limit    = 90240
    default  = true
  }

  enabled                    = true
  enable_thin_provisioning   = true
  enable_fast_provisioning   = true
  delete_force               = true
  delete_recursive           = true
  elasticity                 = true
  include_vm_memory_overhead = false
  network_quota              = 100
  default_compute_policy_id   = vcd_vm_sizing_policy.minSize.id
  vm_sizing_policy_ids        = [vcd_vm_sizing_policy.minSize.id]
}

resource "vcd_vapp" "source" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  name = "{{.TestName}}"
}

resource "vcd_vapp_vm" "source" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  # You cannot remove NICs from an active virtual machine on which no operating system is installed.
  power_on = true

  description = "test empty VM"
  vapp_name   = vcd_vapp.source.name
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
  vdc      = vcd_org_vdc.sizing-policy.name
  power_on = true

  depends_on = [vcd_org_vdc.sizing-policy]
}

resource "vcd_vapp_network" "vappNet" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.sizing-policy.name

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
  name = "{{.Vdc}}"
}

resource "vcd_vapp_vm" "copy" {
  org           = "{{.Org}}"
  vdc           = vcd_org_vdc.sizing-policy.name
  vapp_name     = vcd_vapp.destination.name
  name          = "{{.TestName}}-dest"
  computer_name = "{{.ComputerName}}"

  copy_from_vm_id  = vcd_vapp_vm.source.id
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
  vdc           = vcd_org_vdc.sizing-policy.name
  name          = "{{.TestName}}-dest"
  computer_name = "{{.ComputerName}}"

  copy_from_vm_id  = vcd_vapp_vm.source.id
  power_on         = true
  memory           = 1024
  cpus             = 2
  cpu_cores        = 1

  sizing_policy_id = vcd_vm_sizing_policy.minSize.id
}
`
