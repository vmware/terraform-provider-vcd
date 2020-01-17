// +build vapp vm ALL functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdVAppVmDhcpWait(t *testing.T) {
	var (
		vapp        govcd.VApp
		vm          govcd.VM
		netVappName string = t.Name()
		netVmName1  string = t.Name() + "VM"
	)

	var params = StringMap{
		"Org":             testConfig.VCD.Org,
		"Vdc":             testConfig.VCD.Vdc,
		"EdgeGateway":     testConfig.Networking.EdgeGateway,
		"Catalog":         testSuiteCatalogName,
		"CatalogItem":     testSuiteCatalogOVAItem,
		"VAppName":        netVappName,
		"VMName":          netVmName1,
		"Tags":            "vapp vm",
		"DhcpWaitSeconds": 180,
	}

	configTextVM := templateFill(testAccCheckVcdVAppVmDhcpWait, params)

	params["FuncName"] = t.Name() + "-step1"
	params["DhcpWaitSeconds"] = 200
	configTextVMDhcpWaitUpdateStep1 := templateFill(testAccCheckVcdVAppVmDhcpWait, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVM)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdVAppVmDestroy(netVappName),
		Steps: []resource.TestStep{
			// Step 0 - Create with variations of all possible NICs
			resource.TestStep{
				Config: configTextVM,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVAppVmExists(netVappName, netVmName1, "vcd_vapp_vm."+netVmName1, &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "name", netVmName1),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.name", "multinic-net"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.is_primary", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.ip_allocation_mode", "DHCP"),
					resource.TestMatchResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.ip", regexp.MustCompile(`^11.10.0.\d{1,3}$`)),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmName1, "network.0.mac"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network_dhcp_wait_seconds", "180"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.1.ip_allocation_mode", "NONE"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.1.is_primary", "false"),
				),
			},
			resource.TestStep{
				Config: configTextVMDhcpWaitUpdateStep1,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVAppVmExists(netVappName, netVmName1, "vcd_vapp_vm."+netVmName1, &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "name", netVmName1),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.name", "multinic-net"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.is_primary", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.ip_allocation_mode", "DHCP"),
					resource.TestMatchResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.ip", regexp.MustCompile(`^11.10.0.\d{1,3}$`)),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmName1, "network.0.mac"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network_dhcp_wait_seconds", "200"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.1.ip_allocation_mode", "NONE"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.1.is_primary", "false"),
				),
			},
		},
	})
}

const testAccCheckVcdVAppVmDhcpWaitShared = `
resource "vcd_vapp" "{{.VAppName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name       = "{{.VAppName}}"
  depends_on = [vcd_network_routed.net]
}

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

const testAccCheckVcdVAppVmDhcpWait = testAccCheckVcdVAppVmDhcpWaitShared + `
resource "vcd_vapp_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name     = vcd_vapp.{{.VAppName}}.name
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
  }
}
`
