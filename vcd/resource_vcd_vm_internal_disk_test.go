// +build vapp vm ALL functional

package vcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdVmInternalDisk(t *testing.T) {
	// This test requires access to the vCD before filling templates
	// Thus it won't run in the short test
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	adminVdc, err := getAdminVdc()
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

	storageProfile := testConfig.VCD.ProviderVdc.StorageProfile
	if *adminVdc.UsesFastProvisioning {
		// to avoid `Cannot use multiple storage profiles in a fast-provisioned VDC` we need to reuse VM storage profile
		storageProfile = vm.VM.StorageProfile.Name
	}

	diskResourceName := "disk1"
	diskSize := "13333"
	biggerDiskSize := "14333"
	busType := "sata"
	busNumber := "1"
	unitNumber := "0"
	allowReboot := true

	var params = StringMap{
		"Org":                testConfig.VCD.Org,
		"VDC":                testConfig.VCD.Vdc,
		"VappName":           vapp.VApp.Name,
		"VmName":             vm.VM.Name,
		"FuncName":           "TestVappVmDS",
		"Tags":               "vm",
		"DiskResourceName":   diskResourceName,
		"Size":               diskSize,
		"SizeBigger":         biggerDiskSize,
		"BusType":            busType,
		"BusNumber":          busNumber,
		"UnitNumber":         unitNumber,
		"StorageProfileName": testConfig.VCD.ProviderVdc.StorageProfile,
		"AllowReboot":        allowReboot,
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
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "unit_number", "1"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "storage_profile", storageProfile),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "allow_vm_reboot", "false"),
				),
			},
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "size_in_mb", diskSize),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "bus_type", busType),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "bus_number", busNumber),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "unit_number", unitNumber),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "storage_profile", storageProfile),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "thin_provisioned", strconv.FormatBool(*adminVdc.IsThinProvision)),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "iops", "0"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "size_in_mb", diskSize),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "bus_type", "ide"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "bus_number", "0"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "unit_number", "1"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "storage_profile", storageProfile),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "thin_provisioned", strconv.FormatBool(*adminVdc.IsThinProvision)),
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
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "thin_provisioned", strconv.FormatBool(*adminVdc.IsThinProvision)),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "iops", "0"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "allow_vm_reboot", "false"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "bus_type", "ide"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "bus_number", "0"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "unit_number", "1"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "storage_profile", storageProfile),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "thin_provisioned", strconv.FormatBool(*adminVdc.IsThinProvision)),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "size_in_mb", biggerDiskSize),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "iops", "0"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "allow_vm_reboot", "true"),
				),
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

func getAdminVdc() (*types.AdminVdc, error) {
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
}

const sourceTestVmInternalDiskIde = `
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
  allow_vm_reboot = "false"
}
`

const sourceTestVmInternalDisk = `
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

resource "vcd_vm_internal_disk" "{{.DiskResourceName}}" {
  org              = "{{.Org}}"
  vdc              = "{{.VDC}}"
  vapp_name     = "{{.VappName}}"
  vm_name     = "{{.VmName}}"
  bus_type = "{{.BusType}}"
  size_in_mb = "{{.Size}}"
  bus_number = "{{.BusNumber}}"
  unit_number = "{{.UnitNumber}}"
  storage_profile = "{{.StorageProfileName}}"
  allow_vm_reboot = "false"
}
`

const sourceTestVmInternalDisk_Update1 = `
resource "vcd_vm_internal_disk" "{{.DiskResourceName}}" {
  org              = "{{.Org}}"
  vdc              = "{{.VDC}}"
  vapp_name     = "{{.VappName}}"
  vm_name     = "{{.VmName}}"
  bus_type = "{{.BusType}}"
  size_in_mb = "{{.SizeBigger}}"
  bus_number = "{{.BusNumber}}"
  unit_number = "{{.UnitNumber}}"
  storage_profile = "{{.StorageProfileName}}"
  allow_vm_reboot = "false"
}

resource "vcd_vm_internal_disk" "{{.DiskResourceName}}_ide" {
  org              = "{{.Org}}"
  vdc              = "{{.VDC}}"
  vapp_name     = "{{.VappName}}"
  vm_name     = "{{.VmName}}"
  bus_type = "ide"
  size_in_mb = "{{.SizeBigger}}"
  bus_number = "0"
  unit_number = "1"
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
