// +build vapp ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

// TestAccVcdVappDS tests a vApp data source if a vApp is found in the VDC
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
data "vcd_vapp_vm" "{{.VmName}}" {
  name             = "{{.VmName}}"
  org              = "{{.Org}}"
  vdc              = "{{.VDC}}"
  vapp_name        = "{{.VappName}}"
}

output "name" {
  value = data.vcd_vapp_vm.{{.VmName}}.name
}

output "description" {
  value = data.vcd_vapp_vm.{{.VmName}}.description
}
output "storage_profile" {
  value = data.vcd_vapp_vm.{{.VmName}}.storage_profile
}

output "href" {
  value = data.vcd_vapp_vm.{{.VmName}}.href
}
`
