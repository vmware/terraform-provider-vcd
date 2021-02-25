// +build vm standaloneVm ALL functional
// +build !skipStandaloneVm

package vcd

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	testingTags["standaloneVm"] = "resource_vcd_vapp_vm_test.go"
}

func TestAccVcdStandaloneVmTemplate(t *testing.T) {
	// making sure the VM name is unique
	var standaloneVmName = fmt.Sprintf("%s-%d", t.Name(), os.Getpid())
	var diskResourceName = fmt.Sprintf("%s_disk", t.Name())
	var diskName = fmt.Sprintf("%s-disk", t.Name())

	orgName := testConfig.VCD.Org
	vdcName := testConfig.VCD.Vdc
	var params = StringMap{
		"Org":                orgName,
		"Vdc":                vdcName,
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
		"Tags":               "vm standaloneVm",
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
		CheckDestroy:      testAccCheckVcdStandaloneVmDestroy(standaloneVmName, orgName, vdcName),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdStandaloneVmExists(standaloneVmName, "vcd_vm."+standaloneVmName, orgName, vdcName),
					resource.TestCheckResourceAttr(
						"vcd_vm."+standaloneVmName, "vm_type", string(standaloneVmType)),
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
	// making sure the VM name is unique
	standaloneVmName := fmt.Sprintf("%s-%d", t.Name(), os.Getpid())

	if testConfig.Media.MediaName == "" {
		fmt.Println("Warning: `MediaName` is not configured: boot image won't be tested.")
	}

	orgName := testConfig.VCD.Org
	vdcName := testConfig.VCD.Vdc
	var params = StringMap{
		"Org":         orgName,
		"Vdc":         vdcName,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"Catalog":     testSuiteCatalogName,
		"CatalogItem": testSuiteCatalogOVAItem,
		"VMName":      standaloneVmName,
		"Tags":        "vm standaloneVm",
		"Media":       testConfig.Media.MediaName,
	}

	// Create objects for testing field values across update steps
	nic0Mac := testCachedFieldValue{}
	nic1Mac := testCachedFieldValue{}

	configTextVM := templateFill(testAccCheckVcdStandaloneEmptyVm, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVM)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdStandaloneVmDestroy(standaloneVmName, orgName, vdcName),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configTextVM,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdStandaloneVmExists(standaloneVmName, "vcd_vm."+standaloneVmName, orgName, vdcName),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "name", standaloneVmName),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.name", "multinic-net2"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.ip", "12.10.0.152"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.adapter_type", "PCNet32"),
					resource.TestCheckResourceAttrSet("vcd_vm."+standaloneVmName, "network.0.mac"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.connected", "true"),
					nic0Mac.cacheTestResourceFieldValue("vcd_vm."+standaloneVmName, "network.0.mac"),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.name", "multinic-net"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.is_primary", "true"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.ip_allocation_mode", "DHCP"),
					// resource.TestCheckResourceAttrSet("vcd_vm."+standaloneVmName, "network.1.ip"), // We cannot guarantee DHCP
					resource.TestCheckResourceAttrSet("vcd_vm."+standaloneVmName, "network.1.mac"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.connected", "true"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.adapter_type", "VMXNET3"),
					nic1Mac.cacheTestResourceFieldValue("vcd_vm."+standaloneVmName, "network.1.mac"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "os_type", "sles11_64Guest"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "hardware_version", "vmx-13"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "expose_hardware_virtualization", "true"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "computer_name", "compName"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "description", "test empty standalone VM"),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "cpu_hot_add_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "memory_hot_add_enabled", "true"),
				),
			},
		},
	})
}

func testAccCheckVcdStandaloneVmExists(vmName, node, orgName, vdcName string) resource.TestCheckFunc {
	if orgName == "" {
		orgName = testConfig.VCD.Org
	}
	if vdcName == "" {
		vdcName = testConfig.VCD.Vdc
	}
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[node]
		if !ok {
			return fmt.Errorf("not found: %s", node)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no VM ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)
		_, vdc, err := conn.GetOrgAndVdc(orgName, vdcName)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, vdcName, orgName, err)
		}

		_, err = vdc.QueryVmByName(vmName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckVcdStandaloneVmDestroy(vmName string, orgName string, vdcName string) resource.TestCheckFunc {
	if orgName == "" {
		orgName = testConfig.VCD.Org
	}
	if vdcName == "" {
		vdcName = testConfig.VCD.Vdc
	}
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "vcd_vm" {
				continue
			}
			_, vdc, err := conn.GetOrgAndVdc(orgName, vdcName)
			if err != nil {
				return fmt.Errorf(errorRetrievingVdcFromOrg, vdcName, orgName, err)
			}

			_, err = vdc.QueryVmByName(vmName)

			if err == nil {
				return fmt.Errorf("VM still exist")
			}

			return nil
		}

		return nil
	}
}

const testAccCheckVcdStandaloneVm_basic = `
data "vcd_edgegateway" "existing" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  name = "{{.EdgeGateway}}"
}

resource "vcd_network_routed_v2" "{{.NetworkName}}" {
  name            = "{{.NetworkName}}"
  org             = "{{.Org}}"
  vdc             = "{{.Vdc}}"
  edge_gateway_id = data.vcd_edgegateway.existing.id
  gateway         = "10.10.102.1"
  prefix_length   = 24

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
    name               = vcd_network_routed_v2.{{.NetworkName}}.name
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
data "vcd_edgegateway" "existing" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  name = "{{.EdgeGateway}}"
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

resource "vcd_network_routed_v2" "net2" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name            = "multinic-net2"
  edge_gateway_id = data.vcd_edgegateway.existing.id
  gateway         = "12.10.0.1"
  prefix_length   = 24

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
    name               = vcd_network_routed_v2.net2.name
    ip_allocation_mode = "POOL"
    is_primary         = false
	adapter_type       = "PCNet32"
  }

  network {
    type               = "org"
    name               = vcd_network_routed.net.name
    ip_allocation_mode = "DHCP"
    is_primary         = true
  }
}
`
