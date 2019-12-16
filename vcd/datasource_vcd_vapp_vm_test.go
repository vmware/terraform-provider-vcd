// +build vm ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

// TestAccVcdVappDS tests a VM data source if a vApp + VM is found in the VDC
func TestAccVcdVappVmDS(t *testing.T) {
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

	var params = StringMap{
		"Org":      testConfig.VCD.Org,
		"VDC":      testConfig.VCD.Vdc,
		"VappName": vapp.VApp.Name,
		"VmName":   vm.VM.Name,
		"FuncName": "TestVappVmDS",
		"Tags":     "vm",
	}
	configText := templateFill(datasourceTestVappVm, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("name", vm.VM.Name),
					resource.TestCheckOutput("storage_profile", vm.VM.StorageProfile.Name),
					resource.TestCheckOutput("description", vm.VM.Description),
					resource.TestCheckOutput("href", vm.VM.HREF),
				),
			},
		},
	})
}

const datasourceTestVappVm = `
data "vcd_vapp_vm" "vm-ds" {
  name             = "{{.VmName}}"
  org              = "{{.Org}}"
  vdc              = "{{.VDC}}"
  vapp_name        = "{{.VappName}}"
}

output "name" {
  value = data.vcd_vapp_vm.vm-ds.name
}

output "description" {
  value = data.vcd_vapp_vm.vm-ds.description
}
output "storage_profile" {
  value = data.vcd_vapp_vm.vm-ds.storage_profile
}

output "href" {
  value = data.vcd_vapp_vm.vm-ds.href
}
`
