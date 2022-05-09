//go:build network || nsxt || ALL || functional
// +build network nsxt ALL functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"strings"
	"testing"
)

func TestAccVcdSecurityTag(t *testing.T) {
	tag1 := strings.ToLower(t.Name() + "-tag1") // security tags are always lowercase in serverside
	tag2 := strings.ToLower(t.Name() + "-tag2")
	vAppName := t.Name() + "-vapp"
	firstVMName := t.Name() + "-vm"
	secondVMName := t.Name() + "-vm2"

	var params = StringMap{
		"Org":          testConfig.VCD.Org,
		"Vdc":          testConfig.Nsxt.Vdc,
		"VappName":     vAppName,
		"VmName":       firstVMName,
		"VmName2":      secondVMName,
		"ComputerName": t.Name() + "-vm",
		"Catalog":      testConfig.VCD.Catalog.Name,
		"Media":        testConfig.Media.MediaName,
		"SecurityTag1": tag1,
		"SecurityTag2": tag2,
		"FuncName":     t.Name(),
	}

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	vcdClient := createTemporaryVCDConnection(false)
	if vcdClient.Client.APIVCDMaxVersionIs("< 36.0") {
		t.Skip(t.Name() + " requires at least API v36.0 (VCD 10.3+)")
	}

	configText := templateFill(testAccSecurityTag, params)

	params["FuncName"] = t.Name() + "-update"
	configTextUpdate := templateFill(testAccSecurityTagUpdate, params)

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSecurityTagDestroy(tag1),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityTagCreated(tag1),
					testAccCheckSecurityTagOnVMCreated(tag1, vAppName, firstVMName),
					testAccCheckSecurityTagOnVMCreated(tag1, vAppName, secondVMName),
				),
			},
			{
				Config: configTextUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityTagCreated(tag1, tag2),
					testAccCheckSecurityTagOnVMCreated(tag1, vAppName, firstVMName),
					testAccCheckSecurityTagOnVMCreated(tag1, vAppName, secondVMName),
				),
			},
			{
				ResourceName:      fmt.Sprintf("vcd_security_tag.%s", tag1),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     testConfig.VCD.Org + "." + tag1,
			},
		},
	})
	postTestChecks(t)
}

// Templates to use on different stages
const testAccSecurityTag = `
resource "vcd_vapp" "{{.VappName}}" {
  name = "{{.VappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
}

resource "vcd_vapp_vm" "{{.VmName}}" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  vapp_name     = vcd_vapp.{{.VappName}}.name
  name          = "{{.VmName}}"
  computer_name = "emptyVM"
  memory        = 2048
  cpus          = 2
  cpu_cores     = 1

  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"
  catalog_name     = "{{.Catalog}}"
  boot_image       = "{{.Media}}"
}

resource "vcd_vapp_vm" "{{.VmName2}}" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  vapp_name     = vcd_vapp.{{.VappName}}.name
  name          = "{{.VmName2}}"
  computer_name = "emptyVM2"
  memory        = 2048
  cpus          = 2
  cpu_cores     = 1

  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"
  catalog_name     = "{{.Catalog}}"
  boot_image       = "{{.Media}}"
}

resource "vcd_security_tag" "{{.SecurityTag1}}" {
  org    = "{{.Org}}"
  name   = "{{.SecurityTag1}}"
  vm_ids = [vcd_vapp_vm.{{.VmName}}.id, vcd_vapp_vm.{{.VmName2}}.id]
}
`

const testAccSecurityTagUpdate = `
resource "vcd_vapp" "{{.VappName}}" {
  name = "{{.VappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
}

resource "vcd_vapp_vm" "{{.VmName}}" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  vapp_name     = vcd_vapp.{{.VappName}}.name
  name          = "{{.VmName}}"
  computer_name = "emptyVM"
  memory        = 2048
  cpus          = 2
  cpu_cores     = 1

  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"
  catalog_name     = "{{.Catalog}}"
  boot_image       = "{{.Media}}"
}

resource "vcd_vapp_vm" "{{.VmName2}}" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  vapp_name     = vcd_vapp.{{.VappName}}.name
  name          = "{{.VmName2}}"
  computer_name = "emptyVM2"
  memory        = 2048
  cpus          = 2
  cpu_cores     = 1

  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"
  catalog_name     = "{{.Catalog}}"
  boot_image       = "{{.Media}}"
}

resource "vcd_security_tag" "{{.SecurityTag1}}" {
  org    = "{{.Org}}"
  name   = "{{.SecurityTag1}}"
  vm_ids = [vcd_vapp_vm.{{.VmName}}.id, vcd_vapp_vm.{{.VmName2}}.id]
}

resource "vcd_security_tag" "{{.SecurityTag2}}" {
  name = "{{.SecurityTag2}}"
  vm_ids = [vcd_vapp_vm.{{.VmName}}.id]
}
`

func testAccCheckSecurityTagDestroy(securityTags ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		org, err := conn.GetOrgByName(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf("error retrieving the Org - %s", testConfig.VCD.Org)
		}

		securityTagValues, err := org.GetAllSecurityTagValues(nil)
		if err != nil {
			return fmt.Errorf("error retrieving Org %s security tags", testConfig.VCD.Org)
		}

		for _, tag := range securityTags {
			for _, securityTagValue := range securityTagValues {
				if securityTagValue.Tag == tag {
					return fmt.Errorf("found tag %s after destroying", tag)
				}
			}
		}
		return nil
	}
}

func testAccCheckSecurityTagCreated(securityTags ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		org, err := conn.GetOrgByName(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf("error retrieving the Org - %s", testConfig.VCD.Org)
		}

		securityTagValues, err := org.GetAllSecurityTagValues(nil)
		if err != nil {
			return fmt.Errorf("error retrieving Org %s security tags", testConfig.VCD.Org)
		}

		for _, tag := range securityTags {
			for i, securityTagValue := range securityTagValues {
				if securityTagValue.Tag == tag {
					break
				}
				if i == len(securityTagValues)-1 {
					return fmt.Errorf("tag %s wasn't found", tag)
				}
			}
		}
		return nil
	}
}

func testAccCheckSecurityTagOnVMCreated(securityTag, vAppName, VMName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		org, err := conn.GetOrgByName(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf("error retrieving the Org %s - %s", testConfig.VCD.Org, err)
		}

		vdc, err := org.GetVDCByName(testConfig.Nsxt.Vdc, false)
		if err != nil {
			return fmt.Errorf("error retrieving the VDC %s - %s", testConfig.Nsxt.Vdc, err)
		}

		vApp, err := vdc.GetVAppByName(vAppName, false)
		if err != nil {
			return fmt.Errorf("error retrieving the vApp %s - %s", vAppName, err)
		}

		vm, err := vApp.GetVMByName(VMName, false)
		if err != nil {
			return fmt.Errorf("error retrieving the VM %s - %s", VMName, err)
		}

		securityTaggedEntities, err := org.GetAllSecurityTaggedEntitiesByName(securityTag)
		if err != nil {
			return fmt.Errorf("error retrieving security tagged entities with tag name %s - %s", VMName, err)
		}

		for _, taggedVM := range securityTaggedEntities {
			if taggedVM.ID == vm.VM.ID {
				return nil
			}
		}
		return fmt.Errorf("the VM %s is not tagged with security tag %s", VMName, securityTag)
	}
}
