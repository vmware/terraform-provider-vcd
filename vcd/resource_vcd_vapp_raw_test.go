package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/go-vcloud-director/govcd"
	"log"
	"os"
	"testing"
)

func TestAccVcdVAppRaw_Basic(t *testing.T) {
	var vapp govcd.VApp

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"NetworkName": "TestAccVcdVAppRawNet",
		"Catalog":     testConfig.VCD.Catalog.Name,
		"CatalogItem": testConfig.VCD.Catalog.Catalogitem,
		"VappName":    "TestAccVcdVAppRawVapp",
		"VmName":      "TestAccVcdVAppRawVm",
	}
	configText := templateFill(testAccCheckVcdVAppRaw_basic, params)
	if os.Getenv("GOVCD_DEBUG") != "" {
		log.Printf("#[DEBUG] CONFIGURATION: %s\n", configText)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdVAppRawDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppRawExists(fmt.Sprintf("vcd_vapp.%s", params["VappName"].(string)), &vapp),
					resource.TestCheckResourceAttr(
						fmt.Sprintf("vcd_vapp.%s", params["VappName"].(string)), "name", params["VappName"].(string)),
				),
			},
		},
	})
}

func testAccCheckVcdVAppRawExists(n string, vapp *govcd.VApp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no VAPP ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)
		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.VCD.Vdc, testConfig.VCD.Org, err)
		}

		resp, err := vdc.FindVAppByName(rs.Primary.ID)
		if err != nil {
			return err
		}

		*vapp = resp

		return nil
	}
}

func testAccCheckVcdVAppRawDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_vapp" {
			continue
		}
		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.VCD.Vdc, testConfig.VCD.Org, err)
		}

		_, err = vdc.FindVAppByName(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("VPCs still exist")
		}

		return nil
	}

	return nil
}

const testAccCheckVcdVAppRaw_basic = `
resource "vcd_network" "{{.NetworkName}}" {
  name         = "{{.NetworkName}}"
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "10.10.102.1"

  static_ip_pool {
    start_address = "10.10.102.2"
    end_address   = "10.10.102.254"
  }
}

resource "vcd_vapp" "{{.VappName}}" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  name = "{{.VappName}}"
}

resource "vcd_vapp_vm" "{{.VmName}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  vapp_name     = "${vcd_vapp.{{.VappName}}.name}"
  name          = "{{.VmName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 1024
  cpus          = 1

  network_name = "${vcd_network.{{.NetworkName}}.name}"
  ip           = "10.10.102.161"
  depends_on   = ["vcd_vapp.{{.VappName}}"]
}
`
