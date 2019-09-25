// +build catalog ALL functional

package vcd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

var TestAccVcdMediaInsert = "TestAccVcdMediaInsertBasic"
var vappNameForInsert = "TestAccVcdVAppForInsert"
var vmNameForInsert = "TestAccVcdVAppVmForInsert"
var TestAccVcdCatalogMediaForInsert = "TestAccVcdCatalogMediaBasicForInsert"
var TestAccVcdCatalogMediaDescriptionForInsert = "TestAccVcdCatalogMediaBasicDescriptionForInsert"
var TestAccVcdVAppVmNetForInsert = "TestAccVcdVAppVmNetForInsert"

func TestAccVcdMediaInsertBasic(t *testing.T) {
	var params = StringMap{
		"Org":              testConfig.VCD.Org,
		"Vdc":              testConfig.VCD.Vdc,
		"EdgeGateway":      testConfig.Networking.EdgeGateway,
		"Catalog":          testSuiteCatalogName,
		"CatalogItem":      testSuiteCatalogOVAItem,
		"VappName":         vappNameForInsert,
		"VmName":           vmNameForInsert,
		"CatalogMediaName": TestAccVcdCatalogMediaForInsert,
		"Description":      TestAccVcdCatalogMediaDescriptionForInsert,
		"MediaPath":        testConfig.Media.MediaPath,
		"UploadPieceSize":  testConfig.Media.UploadPieceSize,
		"UploadProgress":   testConfig.Media.UploadProgress,
		"InsertMediaName":  TestAccVcdMediaInsert,
		"NetworkName":      TestAccVcdVAppVmNetForInsert,
		"EjectForce":       true,
		"Tags":             "catalog",
	}

	configText := templateFill(testAccCheckVcdInsertEjectBasic, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccResourcesDestroyed,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMediaInserted("vcd_inserted_media."+TestAccVcdMediaInsert),
					testAccCheckMediaEjected("vcd_inserted_media."+TestAccVcdMediaInsert),
				),
			},
		},
	})
}

func testAccCheckMediaInserted(itemName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		injectItemRs, ok := s.RootModule().Resources[itemName]
		if !ok {
			return fmt.Errorf("not found: %s", itemName)
		}

		if injectItemRs.Primary.ID == "" {
			return fmt.Errorf("no media insert ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.VCD.Vdc, testConfig.VCD.Org, err)
		}

		vapp, err := vdc.FindVAppByName(vappNameForInsert)
		if err != nil {
			return err
		}

		vm, err := vdc.FindVMByName(vapp, vmNameForInsert)

		if err != nil {
			return err
		}

		for _, hardwareItem := range vm.VM.VirtualHardwareSection.Item {
			if hardwareItem.ResourceSubType == types.VMsCDResourceSubType {
				return nil
			}
		}
		return fmt.Errorf("no media inserted found")
	}
}

func testAccCheckMediaEjected(itemName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		injectItemRs, ok := s.RootModule().Resources[itemName]
		if !ok {
			return fmt.Errorf("not found: %s", itemName)
		}

		if injectItemRs.Primary.ID == "" {
			return fmt.Errorf("no media insert ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.VCD.Vdc, testConfig.VCD.Org, err)
		}

		vapp, err := vdc.FindVAppByName(vappNameForInsert)
		if err != nil {
			return err
		}

		vm, err := vdc.FindVMByName(vapp, vmNameForInsert)

		if err != nil {
			return err
		}

		for _, hardwareItem := range vm.VM.VirtualHardwareSection.Item {
			if hardwareItem.ResourceSubType == types.VMsCDResourceSubType {
				return nil
			}
		}
		return fmt.Errorf("no media inserted found")
	}
}

func testAccResourcesDestroyed(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)
	for _, rs := range s.RootModule().Resources {
		itemName := rs.Primary.Attributes["name"]
		if rs.Type != "vcd_inserted_media" && itemName != TestAccVcdMediaInsert {
			continue
		}

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.VCD.Vdc, testConfig.VCD.Org, err)
		}

		_, err = vdc.FindVAppByName(vappNameForInsert)
		if err != nil && !strings.Contains(err.Error(), "can't find vApp:") {
			return fmt.Errorf("vapp %s still exist and error: %#v", itemName, err)
		}

		_, err = vdc.GetOrgVdcNetworkByName(TestAccVcdVAppVmNetForInsert, false)
		if err == nil {
			return fmt.Errorf("network %s still exist and error: %#v", itemName, err)
		}
	}
	return nil
}

const testAccCheckVcdInsertEjectBasic = `
resource "vcd_network_routed" "{{.NetworkName}}" {
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
  name       = "{{.VappName}}"
  org        = "{{.Org}}"
  vdc        = "{{.Vdc}}"
  depends_on = ["vcd_network_routed.{{.NetworkName}}"]
}

resource "vcd_vapp_vm" "{{.VmName}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  vapp_name     = "${vcd_vapp.{{.VappName}}.name}"
  name          = "{{.VmName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  network_name  = "${vcd_network_routed.{{.NetworkName}}.name}"
  memory        = 1024
  cpus          = 1
  ip            = "10.10.102.161"
  power_on      = "false"
  depends_on    = ["vcd_vapp.{{.VappName}}"]
}

resource "vcd_catalog_media" "{{.CatalogMediaName}}" {
  org     = "{{.Org}}"
  catalog = "{{.Catalog}}"

  name                 = "{{.CatalogMediaName}}"
  description          = "{{.Description}}"
  media_path           = "{{.MediaPath}}"
  upload_piece_size    = {{.UploadPieceSize}}
  show_upload_progress = "{{.UploadProgress}}"
}

resource "vcd_inserted_media" "{{.InsertMediaName}}" {
  org     = "{{.Org}}"
  vdc     = "{{.Vdc}}"
  catalog = "{{.Catalog}}"
  name    = "{{.CatalogMediaName}}"

  vapp_name  = "${vcd_vapp.{{.VappName}}.name}"
  vm_name    = "${vcd_vapp_vm.{{.VmName}}.name}"
  depends_on = ["vcd_vapp_vm.{{.VmName}}", "vcd_catalog_media.{{.CatalogMediaName}}"]

  eject_force = "{{.EjectForce}}"
}
`
