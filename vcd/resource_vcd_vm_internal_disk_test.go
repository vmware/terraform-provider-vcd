// +build vapp vm ALL functional

package vcd

import (
	"fmt"
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

	diskResourceName := "disk1"
	diskSize := "13333"
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
		"BusType":            busType,
		"BusNumber":          busNumber,
		"UnitNumber":         unitNumber,
		"StorageProfileName": testConfig.VCD.ProviderVdc.StorageProfile,
		"AllowReboot":        allowReboot,
	}
	configText := templateFill(sourceTestVmInternalDisk, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "size_in_mb", diskSize),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "bus_type", busType),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "bus_number", busNumber),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "unit_number", unitNumber),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "storage_profile", testConfig.VCD.ProviderVdc.StorageProfile),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "allow_vm_reboot", strconv.FormatBool(allowReboot)),
				),
			},
		},
	})
}

const sourceTestVmInternalDisk = `
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
  allow_vm_reboot = "{{.AllowReboot}}"
}
`
