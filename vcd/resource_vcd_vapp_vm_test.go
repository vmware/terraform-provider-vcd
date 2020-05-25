// +build vapp vm ALL functional

package vcd

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func init() {
	testingTags["vm"] = "resource_vcd_vapp_vm_test.go"
}

var vappName2 string = "TestAccVcdVAppVmVapp"
var vmName string = "TestAccVcdVAppVmVm"

func TestAccVcdVAppVm_Basic(t *testing.T) {
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
					resource.TestCheckOutput("disk_bus_number", "1"),
					resource.TestCheckOutput("disk_unit_number", "0"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+vmName, "disk.3908069514.size_in_mb", "5"),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_vapp_vm." + vmName + "-import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdVappObject(testConfig, vappName2, vmName),
				// These fields can't be retrieved from user data
				ImportStateVerifyIgnore: []string{"template_name", "catalog_name", "network_name",
					"initscript", "accept_all_eulas", "power_on", "computer_name"},
			},
		},
	})
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

// TestAccVcdVappVm_NicIndex aims to replicate an issue when a VM is created outside of Terraform
// and NICs are not indexed starting with 0. It is not possible to create a VM with network cards
// starting with index other than 0 in terraform-provider-vcd although it can happen when the VM is
// imported from other systems.
//
// This test creates a VM using go-vcloud-director SDK with such vCD NIC indexes:
// * vCD NIC index 1 (is_primary=false) - Terraform network block index 0 (is_primary=false)
// * vCD NIC index 2 (is_primary=true) - Terraform network block index 1 (is_primary=true)
// Before Issue 458 was resolved, terraform-provider-vcd would report is_primary=false for both NICs.
// (GitHub issue: https://github.com/terraform-providers/terraform-provider-vcd/issues/458)
func TestAccVcdVappVm_NicIndex(t *testing.T) {
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	vapp, _, err := testAccVcdVappVmNicIndexcreateVappVm(t.Name()+"-vApp", t.Name())
	if err != nil {
		t.Errorf("error creating VM using SDK: %s", err)
	}
	// The VM was created therefore it must be deleted afterwards
	defer func() {
		if vapp != nil {
			task, err := vapp.Delete()
			if err != nil {
				t.Errorf("error deleting vApp after test: %s", err)
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				t.Errorf("error waiting for vApp deletion: %s", err)
			}
		}
	}()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		// The step below imports an existing VM (created via go-vcloud-director SDK) with odd NIC
		// indexing and uses ImportStateCheck function to check that NIC priority is correctly set
		Steps: []resource.TestStep{
			resource.TestStep{
				ResourceName:      "vcd_vapp_vm." + vmName + "-import",
				ImportState:       true,
				ImportStateIdFunc: importStateIdVappObject(testConfig, t.Name()+"-vApp", t.Name()),
				ImportStateCheck:  validateNicPriority,
			},
		},
	})
}

// validateNicPriority checks if NIC with vCD index 2 is reported as primary nic after import
func validateNicPriority(state []*terraform.InstanceState) error {

	if os.Getenv("GOVCD_DEBUG") != "" {
		stateString := dumpResourceWithNonEmptyValues(state[0], true)
		fmt.Println(stateString)
	}

	// This imported VM has no NIC with index 0, but has NICs with indexes 1 and 2. However
	// Terraform starts indexing them from 0 therefore vCD NIC indexes do not coincide with
	// Terraform network block index:
	// VCD NIC index 1 = Terraform NIC index 0
	// VCD NIC index 2 = Terraform NIC index 1

	nicIndex1Primary := state[0].Attributes["network.0.is_primary"]
	nicIndex2Primary := state[0].Attributes["network.1.is_primary"]

	// nicIndex1Primary should be false, because in vCD VM NIC 1 is not configured as primary
	// nicIndex2Primary should be true, because in vCD VM has NIC 2 configured as primary
	if nicIndex1Primary != "false" {
		return fmt.Errorf("expected nic with vCD index 1 (terraform index 0) to have "+
			"is_primary=false, got: %s", nicIndex1Primary)
	}

	if nicIndex2Primary != "true" {
		return fmt.Errorf("expected nic with vCD index 2 (terraform index 1) to have "+
			"is_primary=true, got: %s", nicIndex2Primary)
	}

	return nil
}

// testAccVcdVappVmNicIndexcreateVappVm creates vApp and empty VM with NICs 1 and 2 (no NIC 0)
func testAccVcdVappVmNicIndexcreateVappVm(vappName, vmName string) (*govcd.VApp, *govcd.VM, error) {
	vcdClient := createTemporaryVCDConnection()
	org, err := vcdClient.GetOrgByName(testConfig.VCD.Org)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting org: %s", err)
	}
	vdc, err := org.GetVDCByName(testConfig.VCD.Vdc, false)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting vdc: %s", err)
	}

	err = vdc.ComposeRawVApp(vappName)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating vApp: %s", err)
	}

	vapp, err := vdc.GetVAppByName(vappName, true)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to find vApp by name %s: %s", vappName, err)
	}
	// must wait until the vApp exits
	err = vapp.BlockWhileStatus("UNRESOLVED", vcdClient.Client.MaxRetryTimeout)
	if err != nil {
		return vapp, nil, fmt.Errorf("error waiting for created test vApp to have working state: %s", err)
	}

	desiredNetConfig := &types.NetworkConnectionSection{}
	desiredNetConfig.PrimaryNetworkConnectionIndex = 2
	desiredNetConfig.NetworkConnection = append(desiredNetConfig.NetworkConnection,
		&types.NetworkConnection{
			IsConnected:             true,
			IPAddressAllocationMode: types.IPAllocationModeNone,
			Network:                 types.NoneNetwork,
			NetworkConnectionIndex:  1,
		},
		&types.NetworkConnection{
			IsConnected:             true,
			IPAddressAllocationMode: types.IPAllocationModeNone,
			Network:                 types.NoneNetwork,
			NetworkConnectionIndex:  2,
		})

	newDisk := types.DiskSettings{
		AdapterType:       "5",
		SizeMb:            int64(16384),
		BusNumber:         0,
		UnitNumber:        0,
		ThinProvisioned:   takeBoolPointer(true),
		OverrideVmDefault: true}
	requestDetails := &types.RecomposeVAppParamsForEmptyVm{
		CreateItem: &types.CreateItem{
			Name:                      vmName,
			NetworkConnectionSection:  desiredNetConfig,
			Description:               "created by test",
			GuestCustomizationSection: nil,
			VmSpecSection: &types.VmSpecSection{
				Modified:          takeBoolPointer(true),
				Info:              "Virtual Machine specification",
				OsType:            "debian10Guest",
				NumCpus:           takeIntPointer(2),
				NumCoresPerSocket: takeIntPointer(1),
				CpuResourceMhz:    &types.CpuResourceMhz{Configured: 1},
				MemoryResourceMb:  &types.MemoryResourceMb{Configured: 512},
				MediaSection:      nil,
				DiskSection:       &types.DiskSection{DiskSettings: []*types.DiskSettings{&newDisk}},
				HardwareVersion:   &types.HardwareVersion{Value: "vmx-13"}, // need support older version vCD
				VirtualCpuType:    "VM32",
			},
		},
		AllEULAsAccepted: true,
	}

	vm, err := vapp.AddEmptyVm(requestDetails)
	if err != nil {
		return vapp, nil, fmt.Errorf("error creating empty VM: %s", err)
	}

	return vapp, vm, nil
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

// dumpResourceWithNonEmptyValues helps to print all attributes in InstanceState
// skipEmptyValues allows to hide attributes with empty values
func dumpResourceWithNonEmptyValues(res *terraform.InstanceState, skipEmptyValues bool) string {
	var buf bytes.Buffer
	attributes := res.Attributes
	attrKeys := make([]string, 0, len(attributes))
	for ak, av := range attributes {
		if ak == "id" {
			continue
		}
		// Skip attributes with empty values
		if skipEmptyValues && av == "" {
			continue
		}

		attrKeys = append(attrKeys, ak)
	}
	sort.Strings(attrKeys)

	for _, ak := range attrKeys {
		av := attributes[ak]
		buf.WriteString(fmt.Sprintf("%s = %s\n", ak, av))
	}
	return buf.String()
}
