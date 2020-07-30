// +build vapp vm ALL functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"regexp"
	"strings"
	"testing"
)

func TestAccVcdVAppHotUpdateVm(t *testing.T) {
	var (
		vapp        govcd.VApp
		vm          govcd.VM
		hotVappName string = t.Name()
		hotVmName1  string = t.Name() + "VM"
	)

	if testConfig.Media.MediaName == "" {
		fmt.Println("Warning: `MediaName` is not configured: boot image won't be tested.")
	}

	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"Catalog":     testSuiteCatalogName,
		"CatalogItem": testSuiteCatalogOVAItem,
		"VAppName":    hotVappName,
		"VMName":      hotVmName1,
		"Tags":        "vapp vm",
		"Media":       testConfig.Media.MediaName,
	}

	configTextVM := templateFill(testAccCheckVcdVAppHotUpdateVm, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVM)

	params["FuncName"] = t.Name() + "-step1"
	configTextVMUpdateStep1 := templateFill(testAccCheckVcdVAppHotUpdateVmStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVMUpdateStep1)

	params["FuncName"] = t.Name() + "-step2"
	configTextVMUpdateStep2 := templateFill(testAccCheckVcdVAppHotUpdateVmStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVMUpdateStep2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdVAppVmDestroy(hotVappName),
		Steps: []resource.TestStep{
			// Step 0 - Create with variations of all possible NICs
			resource.TestStep{
				Config: configTextVM,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVAppVmExists(hotVappName, hotVmName1, "vcd_vapp_vm."+hotVmName1, &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "name", hotVmName1),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "cpu_hot_add_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "memory_hot_add_enabled", "true"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "memory", "2048"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "cpus", "1"),
				),
			},
			// Step 1 - update
			resource.TestStep{
				Config: configTextVMUpdateStep1,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVAppVmExists(hotVappName, hotVmName1, "vcd_vapp_vm."+hotVmName1, &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "name", hotVmName1),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "cpu_hot_add_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "memory_hot_add_enabled", "true"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "memory", "3072"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "cpus", "3"),
					testAccCheckVcdVmNotRestarted("vcd_vapp_vm."+hotVmName1, hotVappName, hotVmName1),
				),
			},
			resource.TestStep{
				Config:      configTextVMUpdateStep2,
				ExpectError: regexp.MustCompile(`update stopped: VM needs to reboot to change properties.*`),
			},
		},
	})
}

func testAccCheckVcdVmNotRestarted(n string, vappName, vmName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no vApp ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)
		org, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.VCD.Vdc, testConfig.VCD.Org, err)
		}

		vapp, err := vdc.GetVAppByName(vappName, false)
		if err != nil {
			return err
		}

		vm, err := vapp.GetVMByName(vmName, false)
		if err != nil {
			return err
		}

		tasks, err := org.GetTaskList()
		if err != nil {
			return err
		}

		for _, task := range tasks.Task {
			if strings.Contains(task.Operation, "Stopped") && task.Owner.ID == vm.VM.ID {
				return fmt.Errorf("found task which stopped VM")
			}
		}

		return nil
	}
}

const testAccCheckVcdVAppHotUpdateVm = `
resource "vcd_vapp" "{{.VAppName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name       = "{{.VAppName}}"
}

resource "vcd_vapp_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  power_on = true

  vapp_name     = vcd_vapp.{{.VAppName}}.name
  name          = "{{.VMName}}"
  computer_name = "compNameUp"

  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"

  memory        = 2048
  cpus          = 1

  cpu_hot_add_enabled    = true
  memory_hot_add_enabled = true

 }
`

const testAccCheckVcdVAppHotUpdateVmStep1 = `
# skip-binary-test: only for updates
resource "vcd_vapp" "{{.VAppName}}" {
	org = "{{.Org}}"
	vdc = "{{.Vdc}}"

	name       = "{{.VAppName}}"
}

resource "vcd_vapp_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name     = vcd_vapp.{{.VAppName}}.name
  computer_name = "compNameUp"
  name          = "{{.VMName}}"

  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
 
  memory        = 3072
  cpus          = 3

  cpu_hot_add_enabled    = true
  memory_hot_add_enabled = true
}
`

const testAccCheckVcdVAppHotUpdateVmStep2 = `
# skip-binary-test: only for updates
resource "vcd_vapp" "{{.VAppName}}" {
	org = "{{.Org}}"
	vdc = "{{.Vdc}}"

	name       = "{{.VAppName}}"
}

resource "vcd_vapp_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name     = vcd_vapp.{{.VAppName}}.name
  computer_name = "compNameUp"
  name          = "{{.VMName}}"

  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
 
  memory        = 3072
  cpus          = 3

  cpu_hot_add_enabled    = false
  memory_hot_add_enabled = true

  prevent_reboot = true
}
`
