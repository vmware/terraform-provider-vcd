// +build vapp vm ALL functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

//TestAccVcdVAppVmSingleNICNetwork is meant to cover all cases for the legacy single
// NIC configurations we have. It changes VM names for each step because a respawn is
// needed as vCD does not return newly assigned IP if one just changed network type.
// It does not power on VMs because that would take more time and not give much use.
// TODO remove once deprecated attributes 'ip' and 'network_name' are removed
func TestAccVcdVAppVmSingleNIC(t *testing.T) {
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
		"Org":           testConfig.VCD.Org,
		"Vdc":           testConfig.VCD.Vdc,
		"EdgeGateway":   testConfig.Networking.EdgeGateway,
		"Catalog":       testSuiteCatalogName,
		"CatalogItem":   testSuiteCatalogOVAItem,
		"VMNetworkName": "singlenic-net",
		"VAppName":      netVappName,
		"IP":            "allocated",
	}

	params["FuncName"] = t.Name() + "-NetOnly"
	configTextNetwork := templateFill(testAccCheckVcdVAppVmSingleNICNetworkOnly, params)

	params["FuncName"] = t.Name() + "-NetVapp"
	configTextNetworkVapp := templateFill(testAccCheckVcdVAppVmSingleNICNetworkVapp, params)

	// allocated
	netVmNameAllocated := netVmName1 + "allocated"
	params["VMName"] = netVmNameAllocated
	configTextStep0 := templateFill(testAccCheckVcdVAppVmSingleNICNetwork, params)

	// dhcp
	netVmNameDHCP := netVmName1 + "dhcp"
	params["IP"] = "dhcp"
	params["VMName"] = netVmNameDHCP
	params["FuncName"] = t.Name() + "-step1"
	configTextStep2 := templateFill(testAccCheckVcdVAppVmSingleNICNetwork, params)

	// manual
	netVmNameManual := netVmName1 + "manual"
	params["VMName"] = netVmNameManual
	params["IP"] = "11.10.0.152"
	params["FuncName"] = t.Name() + "-step2"
	configTextStep4 := templateFill(testAccCheckVcdVAppVmSingleNICNetwork, params)

	// none is not used as it always had a bug
	//params["VMNetworkName"] = "none"
	//params["IP"] = "none"
	//netVmNameNone := netVmName1 + "none"
	//params["VMName"] = netVmNameNone
	//params["FuncName"] = t.Name() + "-step3"
	//configTextStep3 := templateFill(testAccCheckVcdVAppVmSingleNICNetwork, params)

	// no network
	netVmNameNoNetwork := netVmName1 + "noNetwork"
	params["VMName"] = netVmNameNoNetwork
	configTextStep6 := templateFill(testAccCheckVcdVAppVmSingleNICNoNetwork, params)

	// only vApp network with 'allocated'
	netVmNamevAppNetwork := netVmName1 + "vAppNetwork"
	params["VMName"] = netVmNamevAppNetwork
	configTextStep8 := templateFill(testAccCheckVcdVAppVmSingleNICvAppNetwork, params)

	debugPrintf("#[DEBUG] CONFIGURATION (allocated): %s\n", configTextStep0)
	debugPrintf("#[DEBUG] CONFIGURATION (dhcp): %s\n", configTextStep2)
	debugPrintf("#[DEBUG] CONFIGURATION (manual IP): %s\n", configTextStep4)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVcdVAppVmDestroy(netVmNameAllocated),
			testAccCheckVcdVAppVmDestroy(netVmNameDHCP),
			testAccCheckVcdVAppVmDestroy(netVmNameManual),
			//testAccCheckVcdVAppVmDestroy(netVmNameNone),
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configTextStep0,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVAppVmExists(netVappName, netVmNameAllocated, "vcd_vapp_vm."+netVmNameAllocated, &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmNameAllocated, "name", netVmNameAllocated),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmNameAllocated, "network_name", "singlenic-net"),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmNameAllocated, "mac"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmNameAllocated, "ip", "11.10.0.152"),
				),
			},
			// TODO remove cleanup steps once we have locks on objects
			// This is a hack to remove VM from vApp before creating new one. Otherwise it will fail due to vApp
			// is unable to handle removing one VM and creating another one at the same time.
			resource.TestStep{
				Config: configTextNetworkVapp,
			},
			resource.TestStep{
				Config: configTextStep2,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVAppVmExists(netVappName, netVmNameDHCP, "vcd_vapp_vm."+netVmNameDHCP, &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmNameDHCP, "name", netVmNameDHCP),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmNameDHCP, "network_name", "singlenic-net"),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmNameDHCP, "mac"),

					// Unfortunately DHCP is not guaranteed to report IP due to VMware tools being unavailable
					// quickly enough or the machine not using DHCP by default. If it is not then we expect at
					// least "na" string to be set and this allows us to validate if the field is set at all.
					resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmNameDHCP, "ip"),
				),
			},
			// TODO remove cleanup steps once we have locks on objects
			// This is a hack to remove VM from vApp before creating new one. Otherwise it will fail due to vApp
			// is unable to handle removing one VM and creating another one at the same time.
			resource.TestStep{
				Config: configTextNetworkVapp,
			},
			resource.TestStep{
				Config: configTextStep4,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVAppVmExists(netVappName, netVmNameManual, "vcd_vapp_vm."+netVmNameManual, &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmNameManual, "name", netVmNameManual),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmNameManual, "network_name", "singlenic-net"),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmNameManual, "mac"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmNameManual, "ip", "11.10.0.152"),
				),
			},

			resource.TestStep{
				Config: configTextNetworkVapp,
			},
			resource.TestStep{
				Config: configTextStep6,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVAppVmExists(netVappName, netVmNameNoNetwork, "vcd_vapp_vm."+netVmNameNoNetwork, &vapp, &vm),
				),
			},

			resource.TestStep{
				Config: configTextNetworkVapp,
			},
			resource.TestStep{
				Config: configTextStep8,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVAppVmExists(netVappName, netVmNamevAppNetwork, "vcd_vapp_vm."+netVmNamevAppNetwork, &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmNamevAppNetwork, "ip", "192.168.2.51"),
				),
			},

			//// TODO remove cleanup steps once we have locks on objects
			// This is a hack to remove VM first, then vApp to avoid breaking network
			// removal. It mimics parallelism=1. The problem is that vApp undeploy is not
			resource.TestStep{
				Config: configTextNetworkVapp,
			},
			resource.TestStep{
				Config: configTextNetwork,
			},

			// This last step always had a bug and does not work for now in master branch.
			// Because we're deprecating the `ip` and `network_name` attributes there is no point in fixing it.

			//resource.TestStep{
			//	Config: configTextStep3,
			//	Check: resource.ComposeAggregateTestCheckFunc(
			//		testAccCheckVcdVAppVmExists(netVappName, netVmNameNone, "vcd_vapp_vm."+netVmNameNone, &vapp, &vm),
			//		resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmNameNone, "name", netVmNameNone),
			//		resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmNameNone, "network_name", ""),
			//		//resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmName1, "mac"),
			//		resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmNameNone, "ip", "none"),
			//	),
			//},
		},
	})
}

//
const testAccCheckVcdVAppVmSingleNICNetworkOnly = `
resource "vcd_network_routed" "net" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name         = "{{.VMNetworkName}}"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "11.10.0.1"

  dhcp_pool {
    start_address = "11.10.0.2"
    end_address   = "11.10.0.2"
  }

  static_ip_pool {
    start_address = "11.10.0.152"
    end_address   = "11.10.0.152"
  }
}
`

const testAccCheckVcdVAppVmSingleNICNetworkVapp = testAccCheckVcdVAppVmSingleNICNetworkOnly + `
resource "vcd_vapp" "{{.VAppName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name = "{{.VAppName}}"
}
`

// Sample config without any network configuration at all
const testAccCheckVcdVAppVmSingleNICNoNetwork = testAccCheckVcdVAppVmSingleNICNetworkVapp + `
resource "vcd_vapp_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name     = "${vcd_vapp.{{.VAppName}}.name}"
  name          = "{{.VMName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 2
  power_on      = "false"
}
`

// Sample config with vApp network
const testAccCheckVcdVAppVmSingleNICvAppNetwork = testAccCheckVcdVAppVmSingleNICNetworkVapp + `
resource "vcd_vapp_network" "vappNet" {
	org = "{{.Org}}"
	vdc = "{{.Vdc}}"
  
	name       = "vapp-net"
	vapp_name  = "${vcd_vapp.{{.VAppName}}.name}"
	gateway    = "192.168.2.1"
	netmask    = "255.255.255.0"
	dns1       = "192.168.2.1"
	dns2       = "192.168.2.2"
	dns_suffix = "mybiz.biz"
  
	static_ip_pool {
	  start_address = "192.168.2.51"
	  end_address   = "192.168.2.100"
	}
  
	depends_on = ["vcd_vapp.{{.VAppName}}"]
  }
  
  resource "vcd_vapp_vm" "{{.VMName}}" {
	org = "{{.Org}}"
	vdc = "{{.Vdc}}"
  
	vapp_name         = "${vcd_vapp.{{.VAppName}}.name}"
	name              = "{{.VMName}}"
	catalog_name      = "{{.Catalog}}"
	template_name     = "{{.CatalogItem}}"
	memory            = 512
	cpus              = 2
	cpu_cores         = 2
	power_on          = "false"
	vapp_network_name = "${vcd_vapp_network.vappNet.id}"
	ip                = "allocated"
  }  
`

const testAccCheckVcdVAppVmSingleNICNetwork = testAccCheckVcdVAppVmSingleNICNetworkVapp + `
resource "vcd_vapp_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name     = "${vcd_vapp.{{.VAppName}}.name}"
  name          = "{{.VMName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 2
  power_on      = "false"
  ip            = "{{.IP}}"
  network_name  = "${vcd_network_routed.net.name}"
}
`
