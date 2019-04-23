package vcd

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdVAppVmNetwork(t *testing.T) {
	var (
		vapp        govcd.VApp
		vm          govcd.VM
		netVappName string = "TestAccVcdVAppNetwork"
		netVmName1  string = "TestAccVcdVAppVmNetwork"
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
	}

	configText := templateFill(testAccCheckVcdVAppVmNetwork, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdVAppVmDestroy(netVappName),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppVmExists(netVappName, netVmName1, "vcd_vapp_vm."+netVmName1, &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "name", netVmName1),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "networks.0.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "networks.0.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmName1, "networks.0.mac"),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmName1, "networks.0.ip"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "networks.1.is_primary", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "networks.1.ip_allocation_mode", "DHCP"),
					// resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmName1, "networks.1.ip"),		// We cannot guarantee DHCP
					resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmName1, "networks.1.mac"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "networks.2.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "networks.2.ip_allocation_mode", "MANUAL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "networks.2.ip", "11.10.0.170"),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmName1, "networks.2.mac"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "networks.3.ip", ""),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "networks.3.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "networks.3.ip_allocation_mode", "NONE"),
				),
			},
		},
	})
}

const testAccCheckVcdVAppVmNetwork = `
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
  
  resource "vcd_network_routed" "net2" {
	org = "{{.Org}}"
	vdc = "{{.Vdc}}"
  
	name         = "multinic-net2"
	edge_gateway = "{{.EdgeGateway}}"
	gateway      = "12.10.0.1"
  
	dhcp_pool {
	  start_address = "12.10.0.2"
	  end_address   = "12.10.0.100"
	}
  
	static_ip_pool {
	  start_address = "12.10.0.152"
	  end_address   = "12.10.0.254"
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
	# network_name = "${vcd_network_routed.net.name}"
  
	networks = [{
	  orgnetwork = "${vcd_network_routed.net.name}"
	  ip_allocation_mode = "POOL"
	  is_primary         = false
	},
	  {
		orgnetwork = "${vcd_network_routed.net.name}"
		ip_allocation_mode = "DHCP"
		is_primary         = true
	  },
	  {
		orgnetwork         = "${vcd_network_routed.net.name}"
		ip                 = "11.10.0.170"
		ip_allocation_mode = "MANUAL"
		is_primary         = false
	  },
	  {
		orgnetwork = "${vcd_network_routed.net2.name}"
		ip_allocation_mode = "NONE"
	  },
	]
  }  
`
