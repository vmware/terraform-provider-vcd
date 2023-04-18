//go:build network || nsxt || vm || ALL || functional

package vcd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
		"Catalog":      testConfig.VCD.Catalog.NsxtBackedCatalogName,
		"Media":        testConfig.Media.NsxtBackedMediaName,
		"SecurityTag1": tag1,
		"SecurityTag2": tag2,
		"FuncName":     t.Name(),
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccSecurityTag, params)

	params["FuncName"] = t.Name() + "-update"
	configTextUpdate := templateFill(testAccSecurityTagUpdate, params)

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
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

// TestAccVcdVappVmWithSecurityTags tests security_tags field of a vcd_vapp_vm
// There is still a vcd_security_tag resource which is responsible for managing
// the tag in VCD as security_tags field only creates and doesn't remove the
// tag itself if it is removed from a VM to not break other VMs.
func TestAccVcdVappVmWithSecurityTags(t *testing.T) {
	tag1 := strings.ToLower(t.Name() + "-tag1") // security tags are always lowercase in serverside
	tag2 := strings.ToLower(t.Name() + "-tag2")
	vAppName := t.Name() + "-vapp"
	vmName := t.Name() + "-vm"

	var params = StringMap{
		"Org":          testConfig.VCD.Org,
		"Vdc":          testConfig.Nsxt.Vdc,
		"vappName":     vAppName,
		"vmName":       vmName,
		"computerName": t.Name() + "-vm",
		"securityTag1": tag1,
		"securityTag2": tag2,
		"FuncName":     t.Name(),
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVappVmWithSecurityTagsStep1, params)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVappVmUpdateSecurityTagsStep2, params)

	params["FuncName"] = t.Name() + "step3"
	configText3 := templateFill(testAccVappVmUpdateSecurityTagsStep3, params)

	params["FuncName"] = t.Name() + "step4DS"
	configText4DS := templateFill(testAccVappVmUpdateSecurityTagsStep4DS, params)

	params["FuncName"] = t.Name() + "step5"
	configText5 := templateFill(testAccVappVmUpdateSecurityTagsStep5, params)

	params["FuncName"] = t.Name() + "step6"
	configText6 := templateFill(testAccVappVmUpdateSecurityTagsStep6, params)

	debugPrintf("#[DEBUG] CONFIGURATION step 1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step 2: %s\n", configText2)
	debugPrintf("#[DEBUG] CONFIGURATION step 3: %s\n", configText3)
	debugPrintf("#[DEBUG] CONFIGURATION step 4: %s\n", configText4DS)
	debugPrintf("#[DEBUG] CONFIGURATION step 5: %s\n", configText5)
	debugPrintf("#[DEBUG] CONFIGURATION step 6: %s\n", configText6)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceName := "vcd_vapp_vm." + vmName
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		// We don't use CheckDestroy to assert that tags are destroyed because VCD takes some time
		// to clean up orphaned tags
		Steps: []resource.TestStep{
			{ // VM with tag
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityTagCreated(tag1),
					testAccCheckSecurityTagOnVMCreated(tag1, vAppName, vmName),
					resource.TestCheckResourceAttr(resourceName, "security_tags.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "security_tags.*", tag1),
				),
			},
			{ // VM tag changed
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "security_tags.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "security_tags.*", tag2),
				),
			},
			{ // Standalone VM with 1 tag and vApp VM with 2 security tags
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "security_tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "security_tags.*", tag1),
					resource.TestCheckTypeSetElemAttr(resourceName, "security_tags.*", tag2),
					// Standalone VM
					resource.TestCheckResourceAttr("vcd_vm.standalone", "security_tags.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_vm.standalone", "security_tags.*", tag2),
				),
			},
			{ // Test that standalone and vApp VM data sources report tags correctly
				Config: configText4DS,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "security_tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "security_tags.*", tag1),
					resource.TestCheckTypeSetElemAttr(resourceName, "security_tags.*", tag2),
					resource.TestCheckResourceAttr("data.vcd_vapp_vm.with-tags", "security_tags.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.vcd_vapp_vm.with-tags", "security_tags.*", tag1),
					resource.TestCheckTypeSetElemAttr("data.vcd_vapp_vm.with-tags", "security_tags.*", tag2),
					resourceFieldsEqual("data.vcd_vapp_vm.with-tags", resourceName, []string{"%"}),

					// Standalone VM
					resource.TestCheckResourceAttr("data.vcd_vm.with-tags", "security_tags.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.vcd_vm.with-tags", "security_tags.*", tag2),
					resourceFieldsEqual("data.vcd_vm.with-tags", "vcd_vm.standalone", []string{"%"}),
				),
			},
			{ // stop managing `security_tags` - tags from previous setting should still be available
				Config: configText5,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "security_tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "security_tags.*", tag1),
					resource.TestCheckTypeSetElemAttr(resourceName, "security_tags.*", tag2),
					testAccCheckSecurityTagOnVMCreated(tag1, vAppName, vmName),
					testAccCheckSecurityTagOnVMCreated(tag2, vAppName, vmName),
				),
			},
			{ // Removing security tags using `security_tags = []` notation
				Config: configText6,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "security_tags.#", "0"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVappVmWithSecurityTagsStep1 = `
# skip-binary-test: Too similar to step 3
resource "vcd_vapp" "{{.vappName}}" {
  name = "{{.vappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
}

resource "vcd_vapp_vm" "{{.vmName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"
  vapp_name        = vcd_vapp.{{.vappName}}.name
  name             = "{{.vmName}}"
  computer_name    = "{{.computerName}}"
  memory           = 1024
  cpus             = 2
  cpu_cores        = 1
  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"
  security_tags    = ["{{.securityTag1}}"]
}
`

const testAccVappVmUpdateSecurityTagsStep2 = `
# skip-binary-test: Too similar to step 3
resource "vcd_vapp" "{{.vappName}}" {
  name = "{{.vappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
}

resource "vcd_vapp_vm" "{{.vmName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"
  
  vapp_name        = vcd_vapp.{{.vappName}}.name
  name             = "{{.vmName}}"
  computer_name    = "{{.computerName}}"
  memory           = 1024
  cpus             = 2
  cpu_cores        = 1
  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"
  security_tags    = ["{{.securityTag2}}"]
}
`

const testAccVappVmUpdateSecurityTagsStep3 = `
resource "vcd_vapp" "{{.vappName}}" {
  name = "{{.vappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
}

resource "vcd_vapp_vm" "{{.vmName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"
  
  vapp_name        = vcd_vapp.{{.vappName}}.name
  name             = "{{.vmName}}"
  computer_name    = "{{.computerName}}"
  memory           = 1024
  cpus             = 2
  cpu_cores        = 1
  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"
  security_tags    = ["{{.securityTag1}}","{{.securityTag2}}"]
}


resource "vcd_vm" "standalone" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name             = "{{.vmName}}-standalone"
  computer_name    = "{{.computerName}}"
  memory           = 1024
  cpus             = 2
  cpu_cores        = 1
  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"
  security_tags    = ["{{.securityTag2}}"]
}
`
const testAccVappVmUpdateSecurityTagsStep4DS = testAccVappVmUpdateSecurityTagsStep3 + `
# skip-binary-test: Data Source test
data "vcd_vapp_vm" "with-tags" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"
  
  vapp_name = vcd_vapp.{{.vappName}}.name
  name      = "{{.vmName}}"
}

data "vcd_vm" "with-tags" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"
  
  name = "{{.vmName}}-standalone"
}
`

const testAccVappVmUpdateSecurityTagsStep5 = `
# skip-binary-test: only useful for update
resource "vcd_vapp" "{{.vappName}}" {
  name = "{{.vappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
}

resource "vcd_vapp_vm" "{{.vmName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"
  
  vapp_name        = vcd_vapp.{{.vappName}}.name
  name             = "{{.vmName}}"
  computer_name    = "{{.computerName}}"
  memory           = 1024
  cpus             = 2
  cpu_cores        = 1
  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"
}
`

const testAccVappVmUpdateSecurityTagsStep6 = `
resource "vcd_vapp" "{{.vappName}}" {
  name = "{{.vappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
}

resource "vcd_vapp_vm" "{{.vmName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"
  
  vapp_name        = vcd_vapp.{{.vappName}}.name
  name             = "{{.vmName}}"
  computer_name    = "{{.computerName}}"
  memory           = 1024
  cpus             = 2
  cpu_cores        = 1
  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"
  security_tags    = []
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
