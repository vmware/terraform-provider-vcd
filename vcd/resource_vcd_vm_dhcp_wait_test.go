// +build standaloneVm vm ALL functional
// +build !skipStandaloneVm

package vcd

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdStandaloneVmDhcpWait(t *testing.T) {
	preTestChecks(t)
	var standaloneVmName = fmt.Sprintf("%s-%d", t.Name(), os.Getpid())

	var params = StringMap{
		"Org":             testConfig.VCD.Org,
		"Vdc":             testConfig.VCD.Vdc,
		"EdgeGateway":     testConfig.Networking.EdgeGateway,
		"Catalog":         testSuiteCatalogName,
		"CatalogItem":     testSuiteCatalogOVAItem,
		"VMName":          standaloneVmName,
		"Tags":            "standaloneVm vm",
		"DhcpWaitSeconds": 300,
	}

	configTextVM := templateFill(testAccCheckVcdVmDhcpWait, params)

	params["FuncName"] = t.Name() + "-step1"
	params["DhcpWaitSeconds"] = 310
	configTextVMDhcpWaitUpdateStep1 := templateFill(testAccCheckVcdVmDhcpWait, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	reIp := regexp.MustCompile(`^11.10.0.\d{1,3}$`)
	skipEnvVar := "VCD_SKIP_DHCP_CHECK"
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVM)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdStandaloneVmDestroy(standaloneVmName, "", ""),
		Steps: []resource.TestStep{
			// Step 0 - Create with variations of all possible NICs
			resource.TestStep{
				Config: configTextVM,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdStandaloneVmExists(standaloneVmName, "vcd_vm."+standaloneVmName, "", ""),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "name", standaloneVmName),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.name", "multinic-net"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.is_primary", "true"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.ip_allocation_mode", "DHCP"),
					skipOnEnvVariable(skipEnvVar, "1", "IP regexp "+reIp.String(),
						resource.TestMatchResourceAttr("vcd_vm."+standaloneVmName, "network.0.ip", reIp)),
					resource.TestCheckResourceAttrSet("vcd_vm."+standaloneVmName, "network.0.mac"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.connected", "true"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network_dhcp_wait_seconds", "300"),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.ip_allocation_mode", "NONE"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.connected", "false"),

					// Check data source
					skipOnEnvVariable(skipEnvVar, "1", "IP regexp "+reIp.String(),
						resource.TestMatchResourceAttr("data.vcd_vm.ds", "network.0.ip", reIp)),

					skipOnEnvVariable(skipEnvVar, "1", "comparing IPs",
						resource.TestCheckResourceAttrPair("vcd_vm."+standaloneVmName, "network.0.ip", "data.vcd_vm.ds", "network.0.ip")),
					resource.TestCheckResourceAttr("data.vcd_vm.ds", "network_dhcp_wait_seconds", "300"),
				),
			},
			resource.TestStep{
				Config: configTextVMDhcpWaitUpdateStep1,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdStandaloneVmExists(standaloneVmName, "vcd_vm."+standaloneVmName, "", ""),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "name", standaloneVmName),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.name", "multinic-net"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.is_primary", "true"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.ip_allocation_mode", "DHCP"),
					skipOnEnvVariable(skipEnvVar, "1", "IP regexp "+reIp.String(),
						resource.TestMatchResourceAttr("vcd_vm."+standaloneVmName, "network.0.ip", reIp)),
					resource.TestCheckResourceAttrSet("vcd_vm."+standaloneVmName, "network.0.mac"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network_dhcp_wait_seconds", "310"),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.ip_allocation_mode", "NONE"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.connected", "false"),

					// Check data source
					skipOnEnvVariable(skipEnvVar, "1", "IP regexp "+reIp.String(),
						resource.TestMatchResourceAttr("data.vcd_vm.ds", "network.0.ip", reIp)),
					skipOnEnvVariable(skipEnvVar, "1", "comparing IPs",
						resource.TestCheckResourceAttrPair("vcd_vm."+standaloneVmName, "network.0.ip", "data.vcd_vm.ds", "network.0.ip")),
					resource.TestCheckResourceAttr("data.vcd_vm.ds", "network_dhcp_wait_seconds", "310"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdVmDhcpWaitShared = `
resource "vcd_network_routed" "net" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name         = "multinic-net"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "11.10.0.1"

  dhcp_pool {
    start_address = "11.10.0.2"
    end_address   = "11.10.0.100"
  }

  static_ip_pool {
    start_address = "11.10.0.152"
    end_address   = "11.10.0.254"
  }
}
`

const testAccCheckVcdVmDhcpWait = testAccCheckVcdVmDhcpWaitShared + `
resource "vcd_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name          = "{{.VMName}}"
  computer_name = "dhcp-vm"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1

  network_dhcp_wait_seconds = {{.DhcpWaitSeconds}}
  network {
    type               = "org"
    name               = vcd_network_routed.net.name
    ip_allocation_mode = "DHCP"
    is_primary         = true
  }
 
  network {
    type               = "none"
    ip_allocation_mode = "NONE"
    connected          = "false"
  }
}

data "vcd_vm" "ds" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name                      = vcd_vm.{{.VMName}}.name
  network_dhcp_wait_seconds = {{.DhcpWaitSeconds}}
  depends_on                = [vcd_vm.{{.VMName}}]
}
`
