// +build vapp vm ALL functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccVcdVmInternalDisk(t *testing.T) {
	// Thus it won't run in the short test
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	// In general VM internal disks works with Org users, but since we need to create VDC with disabled fast provisioning value, we have to be sys admins
	if !usingSysAdmin() {
		t.Skip("VM internal disks tests requires system admin privileges")
		return
	}

	/*	adminVdc, err := getAdminVdc()
		if err != nil {
			t.Skip(fmt.Sprintf("error retrieving if VDC is thing provisioned %s", err))
			return
		}

		vapp, err := getAvailableVapp()
		if err != nil {
			t.Skip("No suitable vApp found for this test")
			return
		}
		var vm *govcd.VM

		if vapp.VApp.Children != nil && len(vapp.VApp.Children.VM) > 0 {
			vm, err = vapp.GetVMById(vapp.VApp.Children.VM[0].ID, false)
			if err != nil {
				t.Skip(fmt.Sprintf("error retrieving VM %s", vapp.VApp.Children.VM[0].Name))
				return
			}
		}
		if vm == nil {
			t.Skip(fmt.Sprintf("No VM available in vApp %s", vapp.VApp.Name))
			return
		}

		// for test to run correctly VM has to be power on
		task, err := vm.PowerOn()
		if err != nil {
			t.Skip(fmt.Sprintf("Error powering up vm %s", err))
			return
		}
		_ = task.WaitTaskCompletion()*/
	internalDiskSize := 20000

	storageProfile := testConfig.VCD.ProviderVdc.StorageProfile
	/*	if *adminVdc.UsesFastProvisioning {
		// to avoid `Cannot use multiple storage profiles in a fast-provisioned VDC` we need to reuse VM storage profile
		storageProfile = vm.VM.StorageProfile.Name
	}*/

	diskResourceName := "disk1"
	diskSize := "13333"
	biggerDiskSize := "14333"
	busType := "sata"
	busNumber := "1"
	unitNumber := "0"
	allowReboot := true

	vappName := "TestInternalDiskVapp"
	vmName := "TestInternalDiskVm"
	vdcName := "ForInternalDiskTest"
	var params = StringMap{
		"Org": testConfig.VCD.Org,
		//"VDC": testConfig.VCD.Vdc,
		//		"VappName":           vapp.VApp.Name,
		//		"VmName":             vm.VM.Name,
		"FuncName":           "TestVappVmDS",
		"Tags":               "vm",
		"DiskResourceName":   diskResourceName,
		"Size":               diskSize,
		"SizeBigger":         biggerDiskSize,
		"BusType":            busType,
		"BusNumber":          busNumber,
		"UnitNumber":         unitNumber,
		"StorageProfileName": storageProfile,
		"AllowReboot":        allowReboot,

		"VdcName":                   vdcName,
		"OrgName":                   testConfig.VCD.Org,
		"AllocationModel":           "ReservationPool",
		"ProviderVdc":               testConfig.VCD.ProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.ProviderVdc.NetworkPool,
		"Allocated":                 "1024",
		"Reserved":                  "1024",
		"Limit":                     "1024",
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		// because vDC ignores empty values and use default
		"MemoryGuaranteed": "1",
		"CpuGuaranteed":    "1",

		"Catalog":      testSuiteCatalogName,
		"CatalogItem":  testSuiteCatalogOVAItem,
		"VappName":     vappName,
		"VmName":       vmName,
		"ComputerName": vmName + "Unique",

		"InternalDiskSize": internalDiskSize,
	}
	params["FuncName"] = t.Name() + "-IdeCreate"
	configTextIde := templateFill(sourceTestVmInternalDiskIde, params)
	params["FuncName"] = t.Name() + "-CreateALl"
	configText := templateFill(sourceTestVmInternalDisk, params)
	params["FuncName"] = t.Name() + "-Update1"
	configText_update1 := templateFill(sourceTestVmInternalDisk_Update1, params)
	params["FuncName"] = t.Name() + "-Update2"
	//configText_update2 := templateFill(sourceTestVmInternalDisk_Update2, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText+configText_update1)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config:      configTextIde,
				ExpectError: regexp.MustCompile(`.*The attempted operation cannot be performed in the current state \(Powered on\).*`),
				Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "size_in_mb", diskSize),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "bus_type", "ide"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "bus_number", "0"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "unit_number", "0"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "storage_profile", storageProfile),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "allow_vm_reboot", "false"),
					testCheckInternalDiskNonStringOutputs(internalDiskSize),
				),
			},
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "size_in_mb", diskSize),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "bus_type", busType),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "bus_number", busNumber),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "unit_number", unitNumber),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "storage_profile", storageProfile),
					//resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "thin_provisioned", strconv.FormatBool(*adminVdc.IsThinProvision)),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "thin_provisioned", "true"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "iops", "0"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "size_in_mb", diskSize),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "bus_type", "ide"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "bus_number", "0"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "unit_number", "0"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "storage_profile", storageProfile),
					//resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "thin_provisioned", strconv.FormatBool(*adminVdc.IsThinProvision)),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "thin_provisioned", "true"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "iops", "0"),
				),
			},
			resource.TestStep{
				Config: configText_update1,
				//ExpectError:        regexp.MustCompile(`.*You must power off the virtual machine.*to change its hard disks, bus, or unit numbers.*`),
				//ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "size_in_mb", biggerDiskSize),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "bus_type", busType),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "bus_number", busNumber),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "unit_number", unitNumber),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "storage_profile", storageProfile),
					//resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "thin_provisioned", strconv.FormatBool(*adminVdc.IsThinProvision)),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "thin_provisioned", "true"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "iops", "0"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "allow_vm_reboot", "false"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "bus_type", "ide"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "bus_number", "0"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "unit_number", "0"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "storage_profile", storageProfile),
					//resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "thin_provisioned", strconv.FormatBool(*adminVdc.IsThinProvision)),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "thin_provisioned", "true"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "size_in_mb", biggerDiskSize),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "iops", "0"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "allow_vm_reboot", "true"),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_vm_internal_disk." + diskResourceName + "-import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdVmObject(testConfig.VCD.Org, vdcName, vappName, vmName, "3000"),
				// These fields can't be retrieved
				ImportStateVerifyIgnore: []string{"org", "vdc", "allow_vm_reboot", "thin_provisioned"},
			},
			/*resource.TestStep{
				Config: configText_update2,
				Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "size_in_mb", diskSize),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "bus_type", "ide"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "bus_number", "1"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "unit_number", "0"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "storage_profile", testConfig.VCD.ProviderVdc.StorageProfile),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "allow_vm_reboot", "true"),
				),
			},*/
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

/*func getAdminVdc() (*types.AdminVdc, error) {
	vcdClient, err := getTestVCDFromJson(testConfig)
	if err != nil {
		return nil, fmt.Errorf("error getting client configuration: %s", err)
	}
	err = ProviderAuthenticate(vcdClient, testConfig.Provider.User, testConfig.Provider.Password, testConfig.Provider.Token, testConfig.Provider.SysOrg)
	if err != nil {
		return nil, fmt.Errorf("authentication error: %s", err)
	}
	org, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
	if err != nil {
		return nil, fmt.Errorf("org not found : %s", err)
	}
	adminVdc, err := org.GetAdminVDCByName(testConfig.VCD.Vdc, false)
	if err != nil {
		return nil, fmt.Errorf("vdc not found : %s", err)
	}
	return adminVdc.AdminVdc, nil
}*/

// we need VDC with disabled fast provisioning to edit disks
const sourceTestVmInternalDiskOrgVdcAndVM = `
resource "vcd_org_vdc" "{{.VdcName}}" {
  org  = "{{.OrgName}}"
  name = "{{.VdcName}}" 

  allocation_model = "{{.AllocationModel}}"
  network_pool_name     = "{{.NetworkPool}}"
  provider_vdc_name     = "{{.ProviderVdc}}"

  compute_capacity {
    cpu {
      allocated = "{{.Allocated}}"
      limit     = "{{.Limit}}"
    }

    memory {
      allocated = "{{.Allocated}}"
      limit     = "{{.Limit}}"
    }
  }

  storage_profile {
    name = "{{.ProviderVdcStorageProfile}}"
    enabled  = true
    limit    = 102400
    default  = true
  }

  enabled                  = true
  enable_thin_provisioning = true
  enable_fast_provisioning = false
  delete_force             = true
  delete_recursive         = true
}

resource "vcd_vapp" "{{.VappName}}" {
  org              = "{{.Org}}"
  vdc              =  vcd_org_vdc.{{.VdcName}}.name
  name = "{{.VappName}}"
}

resource "vcd_vapp_vm" "{{.VmName}}" {
  org              = "{{.Org}}"
  vdc              =  vcd_org_vdc.{{.VdcName}}.name
  vapp_name     = vcd_vapp.{{.VappName}}.name
  name          = "{{.VmName}}"
  computer_name = "{{.ComputerName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 1024
  cpus          = 1
  cpu_cores     = 1

  override_template_disk {
    bus_type         = "paravirtual"
    size_in_mb       = "{{.InternalDiskSize}}"
    bus_number       = 0
    unit_number      = 0
    iops             = 0
    thin_provisioned = true
    storage_profile  = "{{.StorageProfileName}}"
  }
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

const sourceTestVmInternalDiskIde = sourceTestVmInternalDiskOrgVdcAndVM + `
resource "vcd_vm_internal_disk" "imported" {
  vapp_name   = "vApp_system_1"
  vm_name     = "TerraformDisk1"
  bus_type    = "paravirtual"
  size_in_mb  = "22384"
  bus_number  = 0
  unit_number = 0
  #storage_profile = "Development"
  #allow_vm_reboot = true
  #depends_on   = ["vcd_vapp_vm.Override3Disks3"]
}
`

const sourceTestVmInternalDisk = sourceTestVmInternalDiskOrgVdcAndVM + `
resource "vcd_vm_internal_disk" "{{.DiskResourceName}}_ide" {
  org             = "{{.Org}}"
  vdc             =  vcd_org_vdc.{{.VdcName}}.name
  vapp_name       = vcd_vapp.{{.VappName}}.name
  vm_name         = vcd_vapp_vm.{{.VmName}}.name
  bus_type        = "ide"
  size_in_mb      = "{{.Size}}"
  bus_number      = "0"
  unit_number     = "0"
  storage_profile = "{{.StorageProfileName}}"
  allow_vm_reboot = "true" 
}

resource "vcd_vm_internal_disk" "{{.DiskResourceName}}" {
  org             = "{{.Org}}"
  vdc             =  vcd_org_vdc.{{.VdcName}}.name
  vapp_name       = vcd_vapp.{{.VappName}}.name
  vm_name         = vcd_vapp_vm.{{.VmName}}.name
  bus_type        = "{{.BusType}}"
  size_in_mb      = "{{.Size}}"
  bus_number      = "{{.BusNumber}}"
  unit_number     = "{{.UnitNumber}}"
  storage_profile = "{{.StorageProfileName}}"
  allow_vm_reboot = "false"
}
`

const sourceTestVmInternalDisk_Update1 = sourceTestVmInternalDiskOrgVdcAndVM + `
resource "vcd_vm_internal_disk" "{{.DiskResourceName}}" {
  org             = "{{.Org}}"
  vdc             =  vcd_org_vdc.{{.VdcName}}.name
  vapp_name       = vcd_vapp.{{.VappName}}.name
  vm_name         = vcd_vapp_vm.{{.VmName}}.name
  bus_type        = "{{.BusType}}"
  size_in_mb      = "{{.SizeBigger}}"
  bus_number      = "{{.BusNumber}}"
  unit_number     = "{{.UnitNumber}}"
  storage_profile = "{{.StorageProfileName}}"
  allow_vm_reboot = "false"
}

resource "vcd_vm_internal_disk" "{{.DiskResourceName}}_ide" {
  org             = "{{.Org}}"
  vdc             =  vcd_org_vdc.{{.VdcName}}.name
  vapp_name       = vcd_vapp.{{.VappName}}.name
  vm_name         = vcd_vapp_vm.{{.VmName}}.name
  bus_type        = "ide"
  size_in_mb      = "{{.SizeBigger}}"
  bus_number      = "0"
  unit_number     = "0"
  storage_profile = "{{.StorageProfileName}}"
  allow_vm_reboot = "true"
}
`

/*const sourceTestVmInternalDisk = sataDisk + `
resource "vcd_vm_internal_disk" "{{.DiskResourceName}}_ide" {
  org              = "{{.Org}}"
  vdc              = "{{.VDC}}"
  vapp_name     = "{{.VappName}}"
  vm_name     = "{{.VmName}}"
  bus_type = "ide"
  size_in_mb = "{{.Size}}"
  bus_number = "0"
  unit_number = "1"
  storage_profile = "{{.StorageProfileName}}"
  allow_vm_reboot = "true"
}
`

const sourceTestVmInternalDisk_Update1 = sataDisk + `
resource "vcd_vm_internal_disk" "{{.DiskResourceName}}_ide" {
  org              = "{{.Org}}"
  vdc              = "{{.VDC}}"
  vapp_name     = "{{.VappName}}"
  vm_name     = "{{.VmName}}"
  bus_type = "ide"
  size_in_mb = "{{.Size}}"
  bus_number = "1"
  unit_number = "0"
  storage_profile = "{{.StorageProfileName}}"
  allow_vm_reboot = "false"
}
`

const sourceTestVmInternalDisk_Update2 = sataDisk + `
resource "vcd_vm_internal_disk" "{{.DiskResourceName}}_ide" {
  org              = "{{.Org}}"
  vdc              = "{{.VDC}}"
  vapp_name     = "{{.VappName}}"
  vm_name     = "{{.VmName}}"
  bus_type = "ide"
  size_in_mb = "{{.Size}}"
  bus_number = "1"
  unit_number = "0"
  storage_profile = "{{.StorageProfileName}}"
  allow_vm_reboot = "true"
}
`*/
