//go:build vapp || vm || ALL || functional
// +build vapp vm ALL functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdVAppVmProperties(t *testing.T) {
	preTestChecks(t)
	var vapp govcd.VApp
	var vm govcd.VM

	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"Catalog":     testSuiteCatalogName,
		"CatalogItem": testSuiteCatalogOVAItem,
		"VappName":    vappName2,
		"VmName":      vmName,
		"Tags":        "vapp vm",
	}

	configText := templateFill(testAccCheckVcdVAppVm_properties, params)

	params["FuncName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccCheckVcdVAppVm_propertiesUpdate, params)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccCheckVcdVAppVm_propertiesRemove, params)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccCheckVcdVAppVm_propertiesRemove, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)

	deleteVapp := func() {
		t.Log("Deleting vApp for next step")
		err := vm.Delete()
		if err != nil {
			t.Errorf("error manually deleting VM: %s", err)
			t.FailNow()
		}
		//ignore error
		task, _ := vapp.Undeploy()
		if err != nil {
			err = task.WaitTaskCompletion()
			if err != nil {
				t.Errorf("error manually undeploy vApp: %s", err)
				t.FailNow()
			}
		}
		task, err = vapp.Delete()
		if err != nil {
			t.Errorf("error manually deleting vApp: %s", err)
			t.FailNow()
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			t.Errorf("error manually deleting vApp: %s", err)
			t.FailNow()
		}
		t.Log("Deleting vApp successful")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testParamsNotEmpty(t, params) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdVAppVmDestroy(vappName2),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppVmExists(vappName2, vmName, "vcd_vapp_vm."+vmName, &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+vmName, "name", vmName),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+vmName, `guest_properties.guest.hostname`, "test-host"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+vmName, `guest_properties.guest.another.subkey`, "another-value"),
				),
			},
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppVmExists(vappName2, vmName, "vcd_vapp_vm."+vmName, &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+vmName, "name", vmName),
					resource.TestCheckNoResourceAttr("vcd_vapp_vm."+vmName, `guest_properties.guest.hostname`),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+vmName, `guest_properties.guest.another.subkey`, "new-value"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+vmName, `guest_properties.guest.third.subkey`, "third-value"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppVmExists(vappName2, vmName, "vcd_vapp_vm."+vmName, &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+vmName, "name", vmName),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+vmName, `guest_properties.%`, "0"),
				),
			},
			// Validates that if vApp is missing, resource can be recreated and no error is thrown. Covers issue #611
			{
				Config:             configText3,
				PreConfig:          deleteVapp,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdVAppVm_properties = `
resource "vcd_vapp" "{{.VappName}}" {
  name = "{{.VappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
}

resource "vcd_vapp_vm" "{{.VmName}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  vapp_name     = vcd_vapp.{{.VappName}}.name
  name          = "{{.VmName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1

  guest_properties = {
	"guest.hostname"       = "test-host"
	"guest.another.subkey" = "another-value"
  }
}
`

const testAccCheckVcdVAppVm_propertiesUpdate = `
resource "vcd_vapp" "{{.VappName}}" {
  name = "{{.VappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
}

resource "vcd_vapp_vm" "{{.VmName}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  vapp_name     = vcd_vapp.{{.VappName}}.name
  name          = "{{.VmName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1

  guest_properties = {
	"guest.another.subkey" = "new-value"
	"guest.third.subkey"   = "third-value"
  }
}
`

const testAccCheckVcdVAppVm_propertiesRemove = `
resource "vcd_vapp" "{{.VappName}}" {
  name = "{{.VappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
}

resource "vcd_vapp_vm" "{{.VmName}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  vapp_name     = vcd_vapp.{{.VappName}}.name
  name          = "{{.VmName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1
}
`
