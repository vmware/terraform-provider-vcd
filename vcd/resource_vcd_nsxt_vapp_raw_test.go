// +build vapp vm nsxt ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdNsxtVAppRaw_Basic(t *testing.T) {

	if testConfig.Nsxt.Vdc == "" || testConfig.Nsxt.EdgeGateway == "" {
		t.Skip("Either NSXT VDC or edge gateway not defined")
		return
	}
	var vapp govcd.VApp

	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.Nsxt.Vdc,
		"EdgeGateway": testConfig.Nsxt.EdgeGateway,
		"NetworkName": "TestAccVcdNsxtVAppRawNet",
		"Catalog":     testSuiteCatalogName,
		"CatalogItem": testSuiteCatalogOVAItem,
		"VappName":    "TestAccVcdNsxtVAppRawVapp",
		"VmName1":      "TestAccVcdNsxtVAppRawVm1",
		"VmName2":     "TestAccVcdNsxtVAppRawVm2",
		"Media":       testConfig.Media.MediaName,
		"Tags":        "vapp vm nsxt",
	}
	configText := templateFill(testAccCheckVcdNsxtVAppRaw_basic, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdNsxtVAppRawDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdNsxtVAppRawExists(fmt.Sprintf("vcd_vapp.%s", params["VappName"].(string)), &vapp),
					resource.TestCheckResourceAttr(
						fmt.Sprintf("vcd_vapp.%s", params["VappName"].(string)), "name", params["VappName"].(string)),
					resource.TestCheckResourceAttr(
						fmt.Sprintf("vcd_vapp_vm.%s", params["VmName1"].(string)), "name", params["VmName1"].(string)),
					resource.TestCheckResourceAttr(
						fmt.Sprintf("vcd_vapp_vm.%s", params["VmName2"].(string)), "name", params["VmName2"].(string)),
				),
			},
		},
	})
}

func testAccCheckVcdNsxtVAppRawExists(n string, vapp *govcd.VApp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no vApp ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)
		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.Nsxt.Vdc)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.Nsxt.Vdc, testConfig.VCD.Org, err)
		}

		newVapp, err := vdc.GetVAppByNameOrId(rs.Primary.ID, false)
		if err != nil {
			return err
		}

		*vapp = *newVapp

		return nil
	}
}

func testAccCheckVcdNsxtVAppRawDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_vapp" {
			continue
		}
		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.Nsxt.Vdc)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.Nsxt.Vdc, testConfig.VCD.Org, err)
		}

		_, err = vdc.GetVAppByNameOrId(rs.Primary.ID, false)

		if err == nil {
			return fmt.Errorf("VPCs still exist")
		}

		return nil
	}

	return nil
}

const testAccCheckVcdNsxtVAppRaw_basic = `
data "vcd_nsxt_edgegateway" "existing" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  name = "{{.EdgeGateway}}"
}

resource "vcd_network_routed_v2" "{{.NetworkName}}" {
  name            = "{{.NetworkName}}"
  org             = "{{.Org}}"
  vdc             = "{{.Vdc}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  gateway         = "10.10.102.1"
  prefix_length   = 24

  static_ip_pool {
    start_address = "10.10.102.2"
    end_address   = "10.10.102.254"
  }
}

resource "vcd_vapp" "{{.VappName}}" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  name = "{{.VappName}}"
  depends_on   = [vcd_network_routed_v2.{{.NetworkName}}]
}

resource "vcd_vapp_org_network" "vappNetwork1" {
  org                = "{{.Org}}"
  vdc                = "{{.Vdc}}"
  vapp_name          = vcd_vapp.{{.VappName}}.name
  org_network_name   = vcd_network_routed_v2.{{.NetworkName}}.name 
}

resource "vcd_vapp_vm" "{{.VmName1}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  vapp_name     = vcd_vapp.{{.VappName}}.name
  name          = "{{.VmName1}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 1024
  cpus          = 1

  network {
    type               = "org"
    name               = vcd_vapp_org_network.vappNetwork1.org_network_name
    ip_allocation_mode = "POOL"
  }
}

resource "vcd_vapp_vm" "{{.VmName2}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  power_on = false

  vapp_name     = vcd_vapp.{{.VappName}}.name
  description   = "test empty VM"
  name          = "{{.VmName2}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1 
  
  os_type                        = "sles11_64Guest"
  hardware_version               = "vmx-13"
  catalog_name                   = "{{.Catalog}}"
  boot_image                     = "{{.Media}}"
  expose_hardware_virtualization = true
  computer_name                  = "compName"

  cpu_hot_add_enabled    = true
  memory_hot_add_enabled = true

  network {
    type               = "org"
    name               = vcd_vapp_org_network.vappNetwork1.org_network_name
    ip_allocation_mode = "POOL"
    is_primary         = true
	adapter_type       = "PCNet32"
  }
}
`
