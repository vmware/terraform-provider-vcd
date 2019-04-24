package vcd

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdVAppVmSingleNICNetwork(t *testing.T) {
	var (
		vapp        govcd.VApp
		vm          govcd.VM
		netVappName string = "TestAccVcdVAppNetwork"
		netVmName1  string = t.Name()
	)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"Catalog":     testSuiteCatalogName,
		"CatalogItem": testSuiteCatalogOVAItem,
		"VAppName":    netVappName,
		"VMName":      netVmName1,
		"IP":		   "allocated",
		//"IP":		   "dhcp",
		//"IP":		   "none",
		//"IP":		   "1.1.1.1",
	}

	configText := templateFill(testAccCheckVcdVAppVmSingleNICNetwork, params)

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdVAppVmDestroy(netVappName),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVAppVmExists(netVappName, netVmName1, "vcd_vapp_vm."+netVmName1, &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "name", netVmName1),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network_name", "singlenic-net"),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmName1, "mac"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "ip", "11.10.0.152"),

					//resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network_name", "singlenic-net"),


					//resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "networks.0.is_primary", "false"),
					//resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "networks.0.ip_allocation_mode", "POOL"),
					//resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmName1, "networks.0.mac"),
					//resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmName1, "networks.0.ip"),
					//
					//resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "networks.1.is_primary", "true"),
					//resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "networks.1.ip_allocation_mode", "DHCP"),
					//// resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmName1, "networks.1.ip"),		// We cannot guarantee DHCP
					//resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmName1, "networks.1.mac"),
					//
					//resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "networks.2.is_primary", "false"),
					//resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "networks.2.ip_allocation_mode", "MANUAL"),
					//resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "networks.2.ip", "11.10.0.170"),
					//resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmName1, "networks.2.mac"),
					//
					//resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "networks.3.ip", ""),
					//resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "networks.3.is_primary", "false"),
					//resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "networks.3.ip_allocation_mode", "NONE"),
				),
			},
		},
	})
}

const testAccCheckVcdVAppVmSingleNICNetwork = `
resource "vcd_network_routed" "net" {
	org = "{{.Org}}"
	vdc = "{{.Vdc}}"
  
	name         = "singlenic-net"
	edge_gateway = "{{.EdgeGateway}}"
	gateway      = "11.10.0.1"
  
	dhcp_pool {
	  start_address = "11.10.0.2"
	  end_address   = "11.10.0.100"
	}
  
	static_ip_pool {
	  start_address = "11.10.0.152"
	  end_address   = "11.10.0.152"
	}
  }
  resource "vcd_vapp" "{{.VAppName}}" {
	org = "{{.Org}}"
	vdc = "{{.Vdc}}"
  
	name = "{{.VAppName}}"
  }
  
  resource "vcd_vapp_vm" "{{.VMName}}" {
	org = "{{.Org}}"
	vdc = "{{.Vdc}}"
  
	vapp_name     = "${vcd_vapp.{{.VAppName}}.name}"
	name          = "{{.VMName}}"
	catalog_name  = "{{.Catalog}}"
	template_name = "{{.CatalogItem}}"
	memory        = 512
	cpus          = 2
	cpu_cores     = 1
  
	# ip = "dhcp"
	# ip = "allocated"
	# ip = "11.10.0.155"
	ip = "{{.IP}}"
	network_name = "${vcd_network_routed.net.name}"

  }  
`
