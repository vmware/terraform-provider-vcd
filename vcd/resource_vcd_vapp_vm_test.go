//go:build vapp || vm || ALL || functional

package vcd

import (
	"fmt"
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
						"vcd_vapp_vm."+vmName, "power_on", "false"),
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
				ImportStateIdFunc: importStateIdVappObject(vappName2, vmName, testConfig.VCD.Vdc),
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
						vm1, "power_on", "false"),
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
  name     = "{{.VappName}}"
  org      = "{{.Org}}"
  vdc      = "{{.Vdc}}"
  power_on = false
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
  power_on      = false
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

  power_on = false
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
  power_on      = false
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
  power_on      = false
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

// TestAccVcdVmAndVAppVmWithIds tests that vApp VMs and standalone VMs can be created using the
// `vapp_template_id` and `boot_image_id` attributes.
// TODO: Ideally, we should refactor all tests to use `vapp_template_id` and `boot_image_id` and create a test
// for the deprecated fields to avoid regressions.
func TestAccVcdVmAndVAppVmWithIds(t *testing.T) {
	preTestChecks(t)

	var params = StringMap{
		"TestName":     t.Name(),
		"Org":          testConfig.VCD.Org,
		"Catalog":      testConfig.VCD.Catalog.NsxtBackedCatalogName,
		"VAppTemplate": testConfig.VCD.Catalog.NsxtCatalogItem,
		"MediaItem":    testConfig.Media.NsxtBackedMediaName,
		"Vdc":          testConfig.Nsxt.Vdc,
		"FuncName":     t.Name(),
		"Tags":         "vapp standaloneVm vm",
	}
	testParamsNotEmpty(t, params)

	configTextStep1 := templateFill(testAccVcdVmAndVappVmWithIds, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextStep1)
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVcdNsxtVAppVmDestroy(t.Name()+"-template-vm"),
			testAccCheckVcdNsxtVAppVmDestroy(t.Name()+"-boot-image-vm"),
			testAccCheckVcdStandaloneVmDestroy(t.Name()+"-template-standalone-vm", testConfig.VCD.Org, testConfig.Nsxt.Vdc),
			testAccCheckVcdStandaloneVmDestroy(t.Name()+"-boot-image-standalone-vm", testConfig.VCD.Org, testConfig.Nsxt.Vdc),
		),
		Steps: []resource.TestStep{
			{
				Config: configTextStep1,
				Check: resource.ComposeAggregateTestCheckFunc(
					// The checks here are pretty basic as we just want to check that VMs get correctly created
					resource.TestCheckResourceAttr("vcd_vapp.template-vm", "name", t.Name()+"-template-vm"),
					resource.TestCheckResourceAttr("vcd_vapp.boot-image-vm", "name", t.Name()+"-boot-image-vm"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "vm_type", "vcd_vapp_vm"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.boot-image-vm", "vm_type", "vcd_vapp_vm"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "vm_type", "vcd_vm"),
					resource.TestCheckResourceAttr("vcd_vm.boot-image-vm", "vm_type", "vcd_vm"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdVmAndVappVmWithIds = `
data "vcd_catalog" "{{.Catalog}}" {
  org  = "{{.Org}}"
  name = "{{.Catalog}}"
}

data "vcd_catalog_vapp_template" "{{.VAppTemplate}}" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.{{.Catalog}}.id
  name       = "{{.VAppTemplate}}"
}

data "vcd_catalog_media" "{{.MediaItem}}" {
  org        = "{{.Org}}"
  catalog    = data.vcd_catalog.{{.Catalog}}.name
  name       = "{{.MediaItem}}"
}

resource "vcd_vapp" "template-vm" {
  org         = "{{.Org}}"
  vdc         = "{{.Vdc}}"
  name        = "{{.TestName}}-template-vm"
  description = "vApp for Template VM description"
}

resource "vcd_vapp" "boot-image-vm" {
  org         = "{{.Org}}"
  vdc         = "{{.Vdc}}"
  name        = "{{.TestName}}-boot-image-vm"
  description = "vApp for Boot Image VM description"
}

resource "vcd_vapp_vm" "template-vm" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_template_id  = data.vcd_catalog_vapp_template.{{.VAppTemplate}}.id
  
  vapp_name   = vcd_vapp.template-vm.name
  name        = "{{.TestName}}-template-vapp-vm"
  description = "{{.TestName}}-template-vapp-vm"
}

resource "vcd_vapp_vm" "boot-image-vm" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"
  
  vapp_name     = vcd_vapp.boot-image-vm.name
  name          = "{{.TestName}}-boot-image-vapp-vm"
  description   = "{{.TestName}}-boot-image-vapp-vm"
  computer_name = "vapp-vm"

  boot_image_id = data.vcd_catalog_media.{{.MediaItem}}.id

  cpus   = 1
  memory = 1024

  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"
}

resource "vcd_vm" "template-vm" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_template_id  = data.vcd_catalog_vapp_template.{{.VAppTemplate}}.id
  
  name        = "{{.TestName}}-template-standalone-vm"
  description = "{{.TestName}}-template-standalone-vm"
}

resource "vcd_vm" "boot-image-vm" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name          = "{{.TestName}}-boot-image-standalone-vm"
  description   = "{{.TestName}}-boot-image-standalone-vm"
  computer_name = "standalone"

  boot_image_id = data.vcd_catalog_media.{{.MediaItem}}.id

  cpus   = 1
  memory = 1024

  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"
}
`

// TestAccVcdVAppVmMetadata tests metadata CRUD on vApp VMs
func TestAccVcdVAppVmMetadata(t *testing.T) {
	testMetadataEntryCRUD(t,
		testAccCheckVcdVAppVmMetadata, "vcd_vapp_vm.test-vapp-vm",
		testAccCheckVcdVAppVmMetadataDatasource, "data.vcd_vapp_vm.test-vapp-vm-ds",
		StringMap{
			"Catalog": testConfig.VCD.Catalog.NsxtBackedCatalogName,
			"Media":   testConfig.Media.NsxtBackedMediaName,
		})
}

const testAccCheckVcdVAppVmMetadata = `
resource "vcd_vapp" "test-vapp" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  name = "{{.Name}}"
}

resource "vcd_vapp_vm" "test-vapp-vm" {
  org              = vcd_vapp.test-vapp.org
  vdc              = vcd_vapp.test-vapp.vdc
  name             = vcd_vapp.test-vapp.name
  vapp_name        = vcd_vapp.test-vapp.name
  computer_name    = "dummy"
  memory           = 2048
  cpus             = 2
  cpu_cores        = 1
  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"
  catalog_name     = "{{.Catalog}}"
  boot_image       = "{{.Media}}"
  {{.Metadata}}
}
`

const testAccCheckVcdVAppVmMetadataDatasource = `
data "vcd_vapp_vm" "test-vapp-vm-ds" {
  org       = vcd_vapp_vm.test-vapp-vm.org
  vdc       = vcd_vapp_vm.test-vapp-vm.vdc
  vapp_name = vcd_vapp_vm.test-vapp-vm.name
  name      = vcd_vapp_vm.test-vapp-vm.name
}
`

func TestAccVcdVAppVmMetadataIgnore(t *testing.T) {
	skipIfNotSysAdmin(t)

	getObjectById := func(vcdClient *VCDClient, id string) (metadataCompatible, error) {
		adminOrg, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Org '%s': %s", testConfig.VCD.Org, err)
		}
		vdc, err := adminOrg.GetVDCByName(testConfig.Nsxt.Vdc, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve VDC '%s': %s", testConfig.Nsxt.Vdc, err)
		}
		vApp, err := vdc.GetVAppByName(t.Name(), true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve vApp '%s': %s", t.Name(), err)
		}
		vm, err := vApp.GetVMById(id, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve VM '%s': %s", id, err)
		}
		return vm, nil
	}

	testMetadataEntryIgnore(t,
		testAccCheckVcdVAppVmMetadata, "vcd_vapp_vm.test-vapp-vm",
		testAccCheckVcdVAppVmMetadataDatasource, "data.vcd_vapp_vm.test-vapp-vm-ds",
		getObjectById, StringMap{
			"Catalog": testConfig.VCD.Catalog.NsxtBackedCatalogName,
			"Media":   testConfig.Media.NsxtBackedMediaName,
		})
}

// TestAccVcdVmMetadata tests metadata CRUD on VMs
func TestAccVcdVmMetadata(t *testing.T) {
	testMetadataEntryCRUD(t,
		testAccCheckVcdVmMetadata, "vcd_vm.test-vm",
		testAccCheckVcdVmMetadataDatasource, "data.vcd_vm.test-vm-ds",
		StringMap{
			"Catalog": testConfig.VCD.Catalog.NsxtBackedCatalogName,
			"Media":   testConfig.Media.NsxtBackedMediaName,
		})
}

const testAccCheckVcdVmMetadata = `
resource "vcd_vm" "test-vm" {
  org              = "{{.Org}}"
  vdc              = "{{.Vdc}}"
  name             = "{{.Name}}"
  computer_name    = "dummy"
  memory           = 2048
  cpus             = 2
  cpu_cores        = 1
  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"
  catalog_name     = "{{.Catalog}}"
  boot_image       = "{{.Media}}"
  {{.Metadata}}
}
`

const testAccCheckVcdVmMetadataDatasource = `
data "vcd_vm" "test-vm-ds" {
  org       = vcd_vm.test-vm.org
  vdc       = vcd_vm.test-vm.vdc
  name      = vcd_vm.test-vm.name
}
`

func TestAccVcdVmMetadataIgnore(t *testing.T) {
	skipIfNotSysAdmin(t)

	getObjectById := func(vcdClient *VCDClient, id string) (metadataCompatible, error) {
		org, err := vcdClient.GetOrgByName(testConfig.VCD.Org)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Org '%s': %s", testConfig.VCD.Org, err)
		}
		vm, err := org.QueryVmById(id)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve VM '%s': %s", id, err)
		}
		return vm, nil
	}

	testMetadataEntryIgnore(t,
		testAccCheckVcdVmMetadata, "vcd_vm.test-vm",
		testAccCheckVcdVmMetadataDatasource, "data.vcd_vm.test-vm-ds",
		getObjectById, StringMap{
			"Catalog": testConfig.VCD.Catalog.NsxtBackedCatalogName,
			"Media":   testConfig.Media.NsxtBackedMediaName,
		})
}
