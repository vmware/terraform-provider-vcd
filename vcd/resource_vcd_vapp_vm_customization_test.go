// +build vapp vm ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

// TestAccVcdVAppVmUpdateCustomization tests that setting attribute customizaton.force to `true`
// during update triggers VM customization and waits until it is completed.
// It is important to wait until the operation is completed to test what VM was properly handled before triggering
// power on and force customization. (VM must be un-deployed for customization to work, otherwise it would stay in
// "GC_PENDING" state for long time)
func TestAccVcdVAppVmUpdateCustomization(t *testing.T) {
	var (
		vapp        govcd.VApp
		vm          govcd.VM
		netVappName string = t.Name()
		netVmName1  string = t.Name() + "VM"
	)

	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"Catalog":     testSuiteCatalogName,
		"CatalogItem": testSuiteCatalogOVAItem,
		"VAppName":    netVappName,
		"VMName":      netVmName1,
		"Tags":        "vapp vm",
	}

	configTextVM := templateFill(testAccCheckVcdVAppVmUpdateCustomization, params)

	params["FuncName"] = t.Name() + "-step1"
	configTextVMUpdateStep1 := templateFill(testAccCheckVcdVAppVmUpdateCustomizationStep1, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVM)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdVAppVmDestroy(netVappName),
		Steps: []resource.TestStep{
			// Step 0 - Create without customization flag
			resource.TestStep{
				Config: configTextVM,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVMCustomization("vcd_vapp_vm.test-vm", false),
					testAccCheckVcdVAppVmExists(netVappName, netVmName1, "vcd_vapp_vm.test-vm", &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm", "name", netVmName1),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm", "network.#", "1"),

					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm", "customization.#", "0"),
				),
			},
			// Step 1 - Update - change network configuration and force customization
			resource.TestStep{
				Config: configTextVMUpdateStep1,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVMCustomization("vcd_vapp_vm.test-vm", true),
					testAccCheckVcdVAppVmExists(netVappName, netVmName1, "vcd_vapp_vm.test-vm", &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm", "name", netVmName1),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm", "network.#", "2"),

					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm", "customization.#", "1"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm", "customization.0.force", "true"),
				),
			},
		},
	})
}

// TestAccVcdVAppVmCreateCustomization tests that setting attribute customizaton.force to `true`
// during create triggers VM customization and waits until it is completed.
// It is important to wait until the operation is completed to test what VM was properly handled before triggering
// power on and force customization. (VM must be un-deployed for customization to work, otherwise it would stay in
// "GC_PENDING" state for long time)
func TestAccVcdVAppVmCreateCustomization(t *testing.T) {
	var (
		vapp        govcd.VApp
		vm          govcd.VM
		netVappName string = t.Name()
		netVmName1  string = t.Name() + "VM"
	)

	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"Catalog":     testSuiteCatalogName,
		"CatalogItem": testSuiteCatalogOVAItem,
		"VAppName":    netVappName,
		"VMName":      netVmName1,
		"Tags":        "vapp vm",
	}

	configTextVMUpdateStep2 := templateFill(testAccCheckVcdVAppVmCreateCustomization, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdVAppVmDestroy(netVappName),
		Steps: []resource.TestStep{
			// Step 0 - Create new VM and force customization initially
			resource.TestStep{
				Config: configTextVMUpdateStep2,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVAppVmExists(netVappName, netVmName1, "vcd_vapp_vm.test-vm2", &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm2", "name", netVmName1),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm2", "network.#", "1"),

					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm2", "customization.#", "1"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm2", "customization.0.force", "true"),
				),
			},
		},
	})
}

// testAccCheckVcdVMCustomization functions acts as a check and a function which waits until
// the VM exits its original "GC_PENDING" state after provisioning. This is needed in order to
// be able to check that setting customization.force flag to `true` actually has impact on VM
// settings.
func testAccCheckVcdVMCustomization(node string, customizationPending bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[node]
		if !ok {
			return fmt.Errorf("not found: %s", node)
		}

		if rs.Primary.Attributes["vapp_name"] == "" {
			return fmt.Errorf("no vApp name specified: %+#v", rs)
		}

		if rs.Primary.Attributes["name"] == "" {
			return fmt.Errorf("no VM name specified: %+#v", rs)
		}

		conn := testAccProvider.Meta().(*VCDClient)

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.VCD.Vdc, testConfig.VCD.Org, err)
		}

		vapp, err := vdc.FindVAppByName(rs.Primary.Attributes["vapp_name"])
		if err != nil {
			return err
		}

		vm, err := vdc.FindVMByName(vapp, rs.Primary.Attributes["name"])

		if err != nil {
			return err
		}

		// When force customization was not explicitly triggered - wait until the VM exits from its original GC_PENDING
		// state after provisioning. This takes some time until the VM boots starts guest tools and reports success.
		if !customizationPending {
			// Not using maxRetryTimeout for timeout here because it would force for maxRetryTimeout to be quite long
			// time by default as it takes some time (around 150s during testing) for Photon OS to boot
			// first time and get rid of "GC_PENDING" state
			err = vm.BlockWhileGuestCustomizationStatus("GC_PENDING", minIfLess(300, conn.Client.MaxRetryTimeout))
			if err != nil {
				return err
			}
		}
		customizationStatus, err := vm.GetGuestCustomizationStatus()
		if err != nil {
			return fmt.Errorf("unable to get VM customization status: %s", err)
		}
		// At the stage where "GC_PENDING" should not be set. The state should be something else or this
		// is an error
		if !customizationPending && customizationStatus == "GC_PENDING" {
			return fmt.Errorf("customizationStatus should not be in pending state for vm %s", vm.VM.Name)
		}

		// Customization status of "GC_PENDING" is expected now and it is an error if something else is set
		if customizationPending && customizationStatus != "GC_PENDING" {
			return fmt.Errorf("customizationStatus should be 'GC_PENDING'instead of '%s' for vm %s",
				vm.VM.Name, customizationStatus)
		}

		if customizationPending && customizationStatus == "GC_PENDING" {
			err = vm.BlockWhileGuestCustomizationStatus("GC_PENDING", minIfLess(300, conn.Client.MaxRetryTimeout))
			if err != nil {
				return fmt.Errorf("timed out waiting for VM %s to leave 'GC_PENDING' state: %s", vm.VM.Name, err)
			}
		}

		return nil
	}
}

const testAccCheckVcdVAppVmCustomizationShared = `
resource "vcd_vapp" "test-vapp" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name       = "{{.VAppName}}"
}

resource "vcd_vapp_network" "vappNet" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name       = "vapp-net"
  vapp_name  = "${vcd_vapp.test-vapp.name}"
  gateway    = "192.168.2.1"
  netmask    = "255.255.255.0"
  dns1       = "192.168.2.1"
  dns2       = "192.168.2.2"
  dns_suffix = "mybiz.biz"

  static_ip_pool {
    start_address = "192.168.2.51"
    end_address   = "192.168.2.100"
  }
}
`

const testAccCheckVcdVAppVmUpdateCustomization = testAccCheckVcdVAppVmCustomizationShared + `
resource "vcd_vapp_vm" "test-vm" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name     = "${vcd_vapp.test-vapp.name}"
  name          = "{{.VMName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1

  network {
    type               = "vapp"
    name               = "${vcd_vapp_network.vappNet.name}"
    ip_allocation_mode = "POOL"
  }
}
`

const testAccCheckVcdVAppVmUpdateCustomizationStep1 = testAccCheckVcdVAppVmCustomizationShared + `
resource "vcd_vapp_vm" "test-vm" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name     = "${vcd_vapp.test-vapp.name}"
  name          = "{{.VMName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1

  network {
    type               = "vapp"
    name               = "${vcd_vapp_network.vappNet.name}"
    ip_allocation_mode = "POOL"
  }

  network {
    type               = "none"
    ip_allocation_mode = "NONE"
  }

  customization {
    force = true
  }
}
`

const testAccCheckVcdVAppVmCreateCustomization = testAccCheckVcdVAppVmCustomizationShared + `
resource "vcd_vapp_vm" "test-vm2" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name     = "${vcd_vapp.test-vapp.name}"
  name          = "{{.VMName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1

  network {
    type               = "vapp"
    name               = "${vcd_vapp_network.vappNet.name}"
    ip_allocation_mode = "POOL"
  }

  customization {
    force = true
  }
}
`
