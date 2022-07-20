//go:build vapp || vm || ALL || functional
// +build vapp vm ALL functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func init() {
	testingTags["vm"] = "resource_vcd_vapp_vm_test.go"
}

var vappName2 string = "TestAccVcdVAppVmVapp"
var vmName string = "TestAccVcdVAppVmVm"

func TestAccVcdVAppVm_Basic(t *testing.T) {
	preTestChecks(t)
	var vapp govcd.VApp
	var vm govcd.VM
	var diskResourceName = "TestAccVcdVAppVm_Basic_1"
	var diskName = "TestAccVcdIndependentDiskBasic"

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
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccCheckVcdVAppVm_basic, params)
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
					resource.TestCheckOutput("disk_bus_number", "1"),
					resource.TestCheckOutput("disk_unit_number", "0"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_vapp_vm."+vmName, "disk.*", map[string]string{
						"size_in_mb": "5",
					}),
				),
			},
			{
				ResourceName:      "vcd_vapp_vm." + vmName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdVappObject(testConfig, vappName2, vmName),
				// These fields can't be retrieved from user data
				ImportStateVerifyIgnore: []string{"template_name", "catalog_name",
					"accept_all_eulas", "power_on", "computer_name", "prevent_update_power_off"},
			},
		},
	})
	postTestChecks(t)
}

func TestAccVcdVAppVm_Clone(t *testing.T) {
	preTestChecks(t)
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
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccCheckVcdVAppVm_clone, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	vm1 := "vcd_vapp_vm." + vmName
	vm2 := "vcd_vapp_vm." + vmName + "-clone"

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdVAppVmDestroy(vappName2),
		Steps: []resource.TestStep{
			{
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
	postTestChecks(t)
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
  size_in_mb      = "{{.size}}"
  bus_type        = "{{.busType}}"
  bus_sub_type    = "{{.busSubType}}"
  storage_profile = "{{.storageProfileName}}"
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
    type               = "org"
    name               = vcd_vapp_org_network.vappNetwork1.org_network_name
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
  value = tolist(vcd_vapp_vm.{{.VmName}}.disk)[0].name
}
output "disk_bus_number" {
  value = tolist(vcd_vapp_vm.{{.VmName}}.disk)[0].bus_number
}
output "disk_unit_number" {
  value = tolist(vcd_vapp_vm.{{.VmName}}.disk)[0].unit_number
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
}

resource "vcd_vapp_org_network" "vappNetwork1" {
  org                = "{{.Org}}"
  vdc                = "{{.Vdc}}"
  vapp_name          = vcd_vapp.{{.VappName}}.name
  org_network_name   = vcd_network_routed.{{.NetworkName}}.name 
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
    type               = "org"
    name               = vcd_vapp_org_network.vappNetwork1.org_network_name
    ip_allocation_mode = "MANUAL"
    ip                 = "{{.IP}}"
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

// TestAccVcdVAppVm_SizingPolicy checks that the VApp VM is created correctly when a sizing policy with limits
// is used in a Flex VDC and it didn't specify any CPU or Memory.
func TestAccVcdVAppVm_SizingPolicy(t *testing.T) {
	preTestChecks(t)

	var params = StringMap{
		"Org":                  testConfig.VCD.Org,
		"Vdc":                  testConfig.VCD.Vdc,
		"Name":                 "TestAccVcdVAppVm_SizingPolicy",
		"SizingPolicyCpus":     "2",
		"SizingPolicyCpuCores": "1",
		"SizingPolicyMemory":   "512",
		"VappName":             vappName2,
		"Catalog":              testSuiteCatalogName,
		"CatalogItem":          testSuiteCatalogOVAItem,
		"ProviderVdc":          testConfig.VCD.ProviderVdc.Name,
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccVcdVAppVm_SizingPolicy, params)
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
					resource.TestCheckResourceAttr(
						"vcd_vapp_vm.test-vm", "name", "TestAccVcdVAppVm_SizingPolicy"),
					resource.TestMatchResourceAttr(
						"vcd_vapp_vm.test-vm", "sizing_policy_id", regexp.MustCompile("urn:vcloud:vdcComputePolicy:[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$")),
					resource.TestCheckResourceAttr(
						"vcd_vapp_vm.test-vm", "cpus", params["SizingPolicyCpus"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_vm.test-vm", "cpu_cores", params["SizingPolicyCpuCores"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_vm.test-vm", "memory", params["SizingPolicyMemory"].(string)),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdVAppVm_SizingPolicy = `
resource "vcd_vm_sizing_policy" "test-sizing-policy" {
  name        = "{{.Name}}"
  description = "{{.Name}}"

  cpu {
    shares                = "886"
    limit_in_mhz          = "1400"
    count                 = "{{.SizingPolicyCpus}}"
    speed_in_mhz          = "1000"
    cores_per_socket      = "{{.SizingPolicyCpuCores}}"
    reservation_guarantee = "0.55"
  }

  memory {
    shares                = "885"
    size_in_mb            = "{{.SizingPolicyMemory}}"
    limit_in_mb           = "1024"
    reservation_guarantee = "0.3"
  }
}

resource "vcd_org_vdc" "test-vdc" {
  name        = "{{.Name}}"
  org         = "{{.Org}}"
  description = "{{.Name}}"
  default_vm_sizing_policy_id = vcd_vm_sizing_policy.test-sizing-policy.id
  vm_sizing_policy_ids        = [vcd_vm_sizing_policy.test-sizing-policy.id]
  
  allocation_model = "Flex"
  provider_vdc_name = "{{.ProviderVdc}}"
  compute_capacity {
    cpu {
      allocated = 2048
    }

    memory {
      allocated = 2048
    }
  }
  storage_profile {
    name    = "*"
    limit   = 20240
    default = true
  }
  enabled                  = true
  enable_thin_provisioning = true
  enable_fast_provisioning = true
  delete_force             = true
  delete_recursive         = true
  elasticity = true
  include_vm_memory_overhead = false
}


resource "vcd_vapp" "test-vapp" {
  name = "{{.VappName}}"
  org  = "{{.Org}}"
  vdc  = vcd_org_vdc.test-vdc.name
}

resource "vcd_vapp_vm" "test-vm" {
  org           = "{{.Org}}"
  vdc           = vcd_org_vdc.test-vdc.name
  vapp_name     = vcd_vapp.test-vapp.name
  name          = "{{.Name}}"
  computer_name = "foo"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  sizing_policy_id = vcd_vm_sizing_policy.test-sizing-policy.id
}
`
