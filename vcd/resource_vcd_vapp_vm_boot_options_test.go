//go:build vapp || vm || ALL || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdVAppVmBootOptions(t *testing.T) {
	preTestChecks(t)
	var (
		vapp     govcd.VApp
		vm       govcd.VM
		vappName string = t.Name()
		vmName   string = t.Name() + "VM"
	)

	if testConfig.VCD.ProviderVdc.StorageProfile == "" || testConfig.VCD.ProviderVdc.StorageProfile2 == "" {
		t.Skip("Both variables testConfig.VCD.ProviderVdc.StorageProfile and testConfig.VCD.ProviderVdc.StorageProfile2 must be set")
	}

	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.Nsxt.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"Catalog":     testSuiteCatalogName,
		"CatalogItem": testSuiteCatalogOVAItem,
		"VAppName":    vappName,
		"VMName":      vmName,
		"Tags":        "vapp vm",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name()
	configText1 := templateFill(testAccCheckVcdVAppVmBootOptions, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText1)

	params["FuncName"] = t.Name() + "-step1"
	configText2 := templateFill(testAccCheckVcdVAppVmBootOptionsStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText2)

	params["FuncName"] = t.Name() + "-step2"
	configText3 := templateFill(testAccCheckVcdVAppVmBootOptionsStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText3)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	vcd := createTemporaryVCDConnection(false)
	if vcd.Client.APIVCDMaxVersionIs("<37.1") {
		t.Skip("Most boot options are only available since 37.1")
	}

	resourceName := "vcd_vapp_vm." + vmName
	datasourceName := "data." + resourceName
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdVAppVmDestroy(vappName),
		Steps: []resource.TestStep{
			// Step 0 - create
			{
				// The `enter_bios_setup_on_next_boot flag is set to false on PowerOn, so it returns a non-empty plan`
				ExpectNonEmptyPlan: true,
				Config:             configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdNsxtVAppVmExists(vappName, vmName, "vcd_vapp_vm."+vmName, &vapp, &vm),
					resource.TestCheckResourceAttr(resourceName, "name", vmName),

					resource.TestCheckResourceAttr(resourceName, "os_type", "sles11_64Guest"),
					resource.TestCheckResourceAttr(resourceName, "hardware_version", "vmx-13"),
					resource.TestCheckResourceAttr(resourceName, "firmware", "efi"),

					resource.TestCheckResourceAttr(resourceName, "boot_options.0.efi_secure_boot", "true"),
					resource.TestCheckResourceAttr(resourceName, "boot_options.0.boot_delay", "2"),
					resource.TestCheckResourceAttr(resourceName, "boot_options.0.boot_retry_delay", "2"),
					resource.TestCheckResourceAttr(resourceName, "boot_options.0.boot_retry_enabled", "true"),

					resource.TestCheckResourceAttr(datasourceName, "firmware", "efi"),

					resource.TestCheckResourceAttr(datasourceName, "boot_options.0.efi_secure_boot", "true"),
					resource.TestCheckResourceAttr(datasourceName, "boot_options.0.boot_delay", "2"),
					resource.TestCheckResourceAttr(datasourceName, "boot_options.0.boot_retry_delay", "2"),
					resource.TestCheckResourceAttr(datasourceName, "boot_options.0.boot_retry_enabled", "true"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdNsxtVAppVmExists(vappName, vmName, "vcd_vapp_vm."+vmName, &vapp, &vm),
					resource.TestCheckResourceAttr(resourceName, "name", vmName),

					resource.TestCheckResourceAttr(resourceName, "os_type", "sles11_64Guest"),
					resource.TestCheckResourceAttr(resourceName, "hardware_version", "vmx-13"),
					resource.TestCheckResourceAttr(resourceName, "firmware", "bios"),

					resource.TestCheckResourceAttr(resourceName, "boot_options.0.efi_secure_boot", "false"),
					resource.TestCheckResourceAttr(resourceName, "boot_options.0.boot_delay", "1"),
					resource.TestCheckResourceAttr(resourceName, "boot_options.0.boot_retry_delay", "1"),
					resource.TestCheckResourceAttr(resourceName, "boot_options.0.boot_retry_enabled", "true"),

					resource.TestCheckResourceAttr(datasourceName, "firmware", "bios"),

					resource.TestCheckResourceAttr(datasourceName, "boot_options.0.efi_secure_boot", "false"),
					resource.TestCheckResourceAttr(datasourceName, "boot_options.0.boot_delay", "1"),
					resource.TestCheckResourceAttr(datasourceName, "boot_options.0.boot_retry_delay", "1"),
					resource.TestCheckResourceAttr(datasourceName, "boot_options.0.boot_retry_enabled", "true"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdNsxtVAppVmExists(vappName, vmName, "vcd_vapp_vm."+vmName, &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+vmName, "name", vmName),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+vmName, "os_type", "sles11_64Guest"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+vmName, "hardware_version", "vmx-13"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+vmName, "firmware", "efi"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+vmName, "boot_options.0.efi_secure_boot", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+vmName, "boot_options.0.boot_delay", "0"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+vmName, "boot_options.0.boot_retry_delay", "0"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+vmName, "boot_options.0.boot_retry_enabled", "false"),

					resource.TestCheckResourceAttr(datasourceName, "firmware", "efi"),

					resource.TestCheckResourceAttr(datasourceName, "boot_options.0.enter_bios_setup_on_next_boot", "false"),
					resource.TestCheckResourceAttr(datasourceName, "boot_options.0.efi_secure_boot", "true"),
					resource.TestCheckResourceAttr(datasourceName, "boot_options.0.boot_delay", "0"),
					resource.TestCheckResourceAttr(datasourceName, "boot_options.0.boot_retry_delay", "0"),
					resource.TestCheckResourceAttr(datasourceName, "boot_options.0.boot_retry_enabled", "false"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testSharedBootOptions = `
resource "vcd_vapp" "{{.VAppName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name     = "{{.VAppName}}"
  power_on = true
}

data "vcd_vapp_vm" "{{.VMName}}" {
  org       = vcd_vapp_vm.{{.VMName}}.org
  vdc       = vcd_vapp_vm.{{.VMName}}.vdc
  name      = vcd_vapp_vm.{{.VMName}}.name
  vapp_name = vcd_vapp_vm.{{.VMName}}.vapp_name
 }
`

const testAccCheckVcdVAppVmBootOptions = testSharedBootOptions + `
resource "vcd_vapp_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  power_on = true

  vapp_name     = vcd_vapp.{{.VAppName}}.name
  name          = "{{.VMName}}"
  computer_name = "compNameUp"

  memory        = 2048
  cpus          = 1

  os_type          = "sles11_64Guest"
  firmware         = "efi"
  hardware_version = "vmx-13"

  boot_options {
    efi_secure_boot = true
    boot_retry_delay = 2
    boot_retry_enabled = true
    boot_delay = 2
    enter_bios_setup_on_next_boot = true
  }
 }

`

const testAccCheckVcdVAppVmBootOptionsStep1 = testSharedBootOptions + `
resource "vcd_vapp_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  power_on = false

  vapp_name     = vcd_vapp.{{.VAppName}}.name
  name          = "{{.VMName}}"
  computer_name = "compNameUp"

  memory        = 2048
  cpus          = 1

  os_type          = "sles11_64Guest"
  firmware         = "bios"
  hardware_version = "vmx-13"

  boot_options {
    efi_secure_boot = false
    boot_retry_delay = 1
    boot_retry_enabled = true
    boot_delay = 1
    enter_bios_setup_on_next_boot = true
  }
 }
`

const testAccCheckVcdVAppVmBootOptionsStep2 = testSharedBootOptions + `
resource "vcd_vapp_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  power_on = true

  vapp_name     = vcd_vapp.{{.VAppName}}.name
  name          = "{{.VMName}}"
  computer_name = "compNameUp"

  memory        = 2048
  cpus          = 1

  os_type          = "sles11_64Guest"
  firmware         = "efi"
  hardware_version = "vmx-13"

  boot_options {
    efi_secure_boot = true
    boot_retry_delay = 0
    boot_retry_enabled = false
    boot_delay = 0
    enter_bios_setup_on_next_boot = false
  }
 }
`
