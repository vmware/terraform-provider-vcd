//go:build (vm || standaloneVm || ALL || functional) && !skipStandaloneVm
// +build vm standaloneVm ALL functional
// +build !skipStandaloneVm

package vcd

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdVmAdvancedComputeProperties(t *testing.T) {
	preTestChecks(t)

	var standaloneVmName = fmt.Sprintf("%s-%d", t.Name(), os.Getpid())
	var emptyVmName = fmt.Sprintf("%s%s-%d", t.Name(), "_empty", os.Getpid())
	var params = StringMap{
		"Org":                       testConfig.VCD.Org,
		"Vdc":                       testConfig.VCD.Vdc,
		"Catalog":                   testSuiteCatalogName,
		"CatalogItem":               testSuiteCatalogOVAItem,
		"VmName":                    standaloneVmName,
		"EmptyVmName":               emptyVmName,
		"Tags":                      "vm standaloneVm",
		"MemoryPriorityType":        "CUSTOM",
		"MemoryShares":              "480",
		"MemoryReservation":         "8",
		"MemoryLimit":               "48",
		"CpuPriorityType":           "CUSTOM",
		"CpuShares":                 "512",
		"CpuReservation":            "200",
		"CpuLimit":                  "1000",
		"MemoryPriorityTypeUpdate":  "CUSTOM",
		"MemorySharesUpdate":        "240",
		"MemoryReservationUpdate":   "4",
		"MemoryLimitUpdate":         "24",
		"CpuPriorityTypeUpdate":     "CUSTOM",
		"CpuSharesUpdate":           "256",
		"CpuReservationUpdate":      "100",
		"CpuLimitUpdate":            "500",
		"MemoryPriorityTypeUpdate2": "NORMAL",
		"MemoryReservationUpdate2":  "4",
		"MemoryLimitUpdate2":        "24",
		"CpuPriorityTypeUpdate2":    "NORMAL",
		"CpuReservationUpdate2":     "100",
		"CpuLimitUpdate2":           "500",
	}

	configText := templateFill(testAccCheckVcdVm_advancedCompute, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)

	params["FuncName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccCheckVcdVm_advancedComputeUpdate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText1)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccCheckVcdVm_advancedComputeUpdate2, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdStandaloneVmDestroy(standaloneVmName, "", ""),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdStandaloneVmExists(standaloneVmName, "vcd_vm."+standaloneVmName, "", ""),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "name", standaloneVmName),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "memory_priority", params["MemoryPriorityType"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "memory_shares", params["MemoryShares"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "memory_reservation", params["MemoryReservation"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "memory_limit", params["MemoryLimit"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "cpu_priority", params["CpuPriorityTypeUpdate"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "cpu_shares", params["CpuShares"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "cpu_reservation", params["CpuReservation"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "cpu_limit", params["CpuLimit"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+emptyVmName, "name", emptyVmName),
					resource.TestCheckResourceAttr("vcd_vm."+emptyVmName, "memory_priority", params["MemoryPriorityType"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+emptyVmName, "memory_shares", params["MemoryShares"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+emptyVmName, "memory_reservation", params["MemoryReservation"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+emptyVmName, "memory_limit", params["MemoryLimit"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+emptyVmName, "cpu_priority", params["CpuPriorityTypeUpdate"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+emptyVmName, "cpu_shares", params["CpuShares"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+emptyVmName, "cpu_reservation", params["CpuReservation"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+emptyVmName, "cpu_limit", params["CpuLimit"].(string)),
				),
			},
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "name", standaloneVmName),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "memory_priority", params["MemoryPriorityTypeUpdate"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "memory_shares", params["MemorySharesUpdate"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "memory_reservation", params["MemoryReservationUpdate"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "memory_limit", params["MemoryLimitUpdate"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "cpu_priority", params["CpuPriorityTypeUpdate"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "cpu_shares", params["CpuSharesUpdate"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "cpu_reservation", params["CpuReservationUpdate"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "cpu_limit", params["CpuLimitUpdate"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+emptyVmName, "name", emptyVmName),
					resource.TestCheckResourceAttr("vcd_vm."+emptyVmName, "memory_priority", params["MemoryPriorityTypeUpdate"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+emptyVmName, "memory_shares", params["MemorySharesUpdate"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+emptyVmName, "memory_reservation", params["MemoryReservationUpdate"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+emptyVmName, "memory_limit", params["MemoryLimitUpdate"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+emptyVmName, "cpu_priority", params["CpuPriorityTypeUpdate"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+emptyVmName, "cpu_shares", params["CpuSharesUpdate"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+emptyVmName, "cpu_reservation", params["CpuReservationUpdate"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+emptyVmName, "cpu_limit", params["CpuLimitUpdate"].(string)),
				),
			},
			resource.TestStep{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "name", standaloneVmName),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "memory_priority", params["MemoryPriorityTypeUpdate2"].(string)),
					resource.TestMatchResourceAttr("vcd_vm."+standaloneVmName, "memory_shares", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "memory_reservation", params["MemoryReservationUpdate2"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "memory_limit", params["MemoryLimitUpdate2"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "cpu_priority", params["CpuPriorityTypeUpdate2"].(string)),
					resource.TestMatchResourceAttr("vcd_vm."+standaloneVmName, "cpu_shares", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "cpu_reservation", params["CpuReservationUpdate2"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "cpu_limit", params["CpuLimitUpdate2"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+emptyVmName, "name", emptyVmName),
					resource.TestCheckResourceAttr("vcd_vm."+emptyVmName, "memory_priority", params["MemoryPriorityTypeUpdate2"].(string)),
					resource.TestMatchResourceAttr("vcd_vm."+emptyVmName, "memory_shares", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr("vcd_vm."+emptyVmName, "memory_reservation", params["MemoryReservationUpdate2"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+emptyVmName, "memory_limit", params["MemoryLimitUpdate2"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+emptyVmName, "cpu_priority", params["CpuPriorityTypeUpdate2"].(string)),
					resource.TestMatchResourceAttr("vcd_vm."+emptyVmName, "cpu_shares", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr("vcd_vm."+emptyVmName, "cpu_reservation", params["CpuReservationUpdate2"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+emptyVmName, "cpu_limit", params["CpuLimitUpdate2"].(string)),
					resource.TestCheckResourceAttr("vcd_vm."+emptyVmName, "memory", "256"),
					resource.TestCheckResourceAttr("vcd_vm."+emptyVmName, "cpus", "1"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdVm_advancedCompute = `
resource "vcd_vm" "{{.VmName}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  name          = "{{.VmName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1

  memory_priority = "{{.MemoryPriorityType}}"
  memory_shares        = "{{.MemoryShares}}"
  memory_reservation   = "{{.MemoryReservation}}"
  memory_limit         = "{{.MemoryLimit}}"

  cpu_priority = "{{.CpuPriorityType}}"
  cpu_shares        = "{{.CpuShares}}"
  cpu_reservation   = "{{.CpuReservation}}"
  cpu_limit         = "{{.CpuLimit}}"  
}

resource "vcd_vm" "{{.EmptyVmName}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  name          = "{{.EmptyVmName}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1
  
  os_type          = "sles10_64Guest"
  hardware_version = "vmx-11"
  computer_name    = "whaetev2"

  memory_priority = "{{.MemoryPriorityType}}"
  memory_shares = "{{.MemoryShares}}"
  memory_reservation = "{{.MemoryReservation}}"
  memory_limit = "{{.MemoryLimit}}"

  cpu_priority = "{{.CpuPriorityType}}"
  cpu_shares = "{{.CpuShares}}"
  cpu_reservation = "{{.CpuReservation}}"
  cpu_limit = "{{.CpuLimit}}"  
}
`
const testAccCheckVcdVm_advancedComputeUpdate = `
# skip-binary-test: only for updates

resource "vcd_vm" "{{.VmName}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  name          = "{{.VmName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1

  memory_priority = "{{.MemoryPriorityTypeUpdate}}"
  memory_shares = "{{.MemorySharesUpdate}}"
  memory_reservation = "{{.MemoryReservationUpdate}}"
  memory_limit = "{{.MemoryLimitUpdate}}"

  cpu_priority = "{{.CpuPriorityTypeUpdate}}"
  cpu_shares = "{{.CpuSharesUpdate}}"
  cpu_reservation = "{{.CpuReservationUpdate}}"
  cpu_limit = "{{.CpuLimitUpdate}}"  
}

resource "vcd_vm" "{{.EmptyVmName}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  name          = "{{.EmptyVmName}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1

  os_type          = "sles10_64Guest"
  hardware_version = "vmx-11"
  computer_name    = "whaetev2"

  memory_priority = "{{.MemoryPriorityTypeUpdate}}"
  memory_shares        = "{{.MemorySharesUpdate}}"
  memory_reservation   = "{{.MemoryReservationUpdate}}"
  memory_limit         = "{{.MemoryLimitUpdate}}"

  cpu_priority = "{{.CpuPriorityTypeUpdate}}"
  cpu_shares        = "{{.CpuSharesUpdate}}"
  cpu_reservation   = "{{.CpuReservationUpdate}}"
  cpu_limit         = "{{.CpuLimitUpdate}}"  
}
`

// only when priority type is `CUSTOM`, then cpu_shares configuration is accepted
const testAccCheckVcdVm_advancedComputeUpdate2 = `
# skip-binary-test: only for updates

resource "vcd_vm" "{{.VmName}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  name          = "{{.VmName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1

  memory_priority = "{{.MemoryPriorityTypeUpdate2}}"
  memory_reservation   = "{{.MemoryReservationUpdate2}}"
  memory_limit         = "{{.MemoryLimitUpdate2}}"

  cpu_priority    = "{{.CpuPriorityTypeUpdate2}}"
  cpu_reservation      = "{{.CpuReservationUpdate2}}"
  cpu_limit            = "{{.CpuLimitUpdate2}}"  
}

resource "vcd_vm" "{{.EmptyVmName}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  name          = "{{.EmptyVmName}}"
  memory        = 256
  cpus          = 1
  cpu_cores     = 1

  os_type          = "sles10_64Guest"
  hardware_version = "vmx-11"
  computer_name    = "whaetev2"

  memory_priority = "{{.MemoryPriorityTypeUpdate2}}"
  memory_reservation   = "{{.MemoryReservationUpdate2}}"
  memory_limit         = "{{.MemoryLimitUpdate2}}"

  cpu_priority    = "{{.CpuPriorityTypeUpdate2}}"
  cpu_reservation      = "{{.CpuReservationUpdate2}}"
  cpu_limit            = "{{.CpuLimitUpdate2}}"  
}
`
