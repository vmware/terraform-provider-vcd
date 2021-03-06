// +build standaloneVm vm ALL functional
// +build !skipStandaloneVm

package vcd

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TestAccVcdStandaloneVmUpdateCustomization tests that setting attribute customizaton.force to `true`
// during update triggers VM customization and waits until it is completed.
// It is important to wait until the operation is completed to test what VM was properly handled before triggering
// power on and force customization. (VM must be un-deployed for customization to work, otherwise it would stay in
// "GC_PENDING" state for long time)
func TestAccVcdStandaloneVmUpdateCustomization(t *testing.T) {
	preTestChecks(t)
	var standaloneVmName = fmt.Sprintf("%s-%d", t.Name(), os.Getpid())

	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"Catalog":     testSuiteCatalogName,
		"CatalogItem": testSuiteCatalogOVAItem,
		"VMName":      standaloneVmName,
		"NetworkName": testConfig.TestEnvBuild.IsolatedNetwork,
		"Tags":        "standaloneVm vm",
	}

	configTextVM := templateFill(testAccCheckVcdVmUpdateCustomization, params)

	params["FuncName"] = t.Name() + "-step1"
	params["Customization"] = "true"
	params["SkipTest"] = "# skip-binary-test: customization.force=true must always request for update"
	configTextVMUpdateStep1 := templateFill(testAccCheckVcdVmCreateCustomization, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVM)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdStandaloneVmDestroy(standaloneVmName, "", ""),
		Steps: []resource.TestStep{
			// Step 0 - Create without customization flag
			resource.TestStep{
				Config: configTextVM,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdStandaloneVMCustomization("vcd_vm.test-vm", false),
					testAccCheckVcdStandaloneVmExists(standaloneVmName, "vcd_vm.test-vm", "", ""),
					resource.TestCheckResourceAttr("vcd_vm.test-vm", "name", standaloneVmName),
					resource.TestCheckResourceAttr("vcd_vm.test-vm", "network.#", "0"),

					resource.TestCheckResourceAttr("vcd_vm.test-vm", "customization.#", "1"),
				),
			},
			// Step 1 - Update - change network configuration and force customization
			resource.TestStep{
				Config: configTextVMUpdateStep1,
				// The plan should never be empty because force works as a flag and every update triggers "update"
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdStandaloneVMCustomization("vcd_vm.test-vm", true),
					testAccCheckVcdStandaloneVmExists(standaloneVmName, "vcd_vm.test-vm", "", ""),
					resource.TestCheckResourceAttr("vcd_vm.test-vm", "name", standaloneVmName),
					resource.TestCheckResourceAttr("vcd_vm.test-vm", "network.#", "1"),

					resource.TestCheckResourceAttr("vcd_vm.test-vm", "customization.#", "1"),
					resource.TestCheckResourceAttr("vcd_vm.test-vm", "customization.0.force", "false"),
				),
			},
		},
	})
	postTestChecks(t)
}

// TestAccVcdStandaloneVmCreateCustomization tests that setting attribute customizaton.force to `true`
// during create triggers VM customization and waits until it is completed.
// It is important to wait until the operation is completed to test what VM was properly handled before triggering
// power on and force customization. (VM must be un-deployed for customization to work, otherwise it would stay in
// "GC_PENDING" state for long time)
func TestAccVcdStandaloneVmCreateCustomization(t *testing.T) {
	preTestChecks(t)
	var standaloneVmName = fmt.Sprintf("%s-%d", t.Name(), os.Getpid())

	var params = StringMap{
		"Org":           testConfig.VCD.Org,
		"Vdc":           testConfig.VCD.Vdc,
		"EdgeGateway":   testConfig.Networking.EdgeGateway,
		"Catalog":       testSuiteCatalogName,
		"CatalogItem":   testSuiteCatalogOVAItem,
		"NetworkName":   testConfig.TestEnvBuild.IsolatedNetwork,
		"VMName":        standaloneVmName,
		"Tags":          "standaloneVm vm",
		"Customization": "true",
	}
	params["SkipTest"] = "# skip-binary-test: customization.force=true must always request for update"
	configTextVMUpdateStep2 := templateFill(testAccCheckVcdVmCreateCustomization, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdStandaloneVmDestroy(standaloneVmName, "", ""),
		Steps: []resource.TestStep{
			// Step 0 - Create new VM and force customization initially
			resource.TestStep{
				Config: configTextVMUpdateStep2,
				// The plan should never be empty because force works as a flag and every update triggers "update"
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdStandaloneVmExists(standaloneVmName, "vcd_vm.test-vm", "", ""),
					resource.TestCheckResourceAttr("vcd_vm.test-vm", "name", standaloneVmName),
					resource.TestCheckResourceAttr("vcd_vm.test-vm", "network.#", "1"),

					resource.TestCheckResourceAttr("vcd_vm.test-vm", "customization.#", "1"),
					// Always store 'customization.0.force=false' in statefile so that a diff is always triggered
					resource.TestCheckResourceAttr("vcd_vm.test-vm", "customization.0.force", "false"),
				),
			},
		},
	})
	postTestChecks(t)
}

// testAccCheckVcdStandaloneVMCustomization functions acts as a check and a function which waits until
// the VM exits its original "GC_PENDING" state after provisioning. This is needed in order to
// be able to check that setting customization.force flag to `true` actually has impact on VM
// settings.
func testAccCheckVcdStandaloneVMCustomization(node string, customizationPending bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[node]
		if !ok {
			return fmt.Errorf("not found: %s", node)
		}

		if rs.Primary.Attributes["name"] == "" {
			return fmt.Errorf("no VM name specified: %+#v", rs)
		}

		conn := testAccProvider.Meta().(*VCDClient)

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.VCD.Vdc, testConfig.VCD.Org, err)
		}

		vm, err := vdc.QueryVmByName(rs.Primary.Attributes["name"])

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
				customizationStatus, vm.VM.Name)
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

const testAccCheckVcdVmUpdateCustomization = `
resource "vcd_vm" "test-vm" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name          = "{{.VMName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1

}
`

const testAccCheckVcdVmCreateCustomization = `
{{.SkipTest}}
resource "vcd_vm" "test-vm" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name          = "{{.VMName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1

  customization {
    force = {{.Customization}}
  }

  network {
    type               = "org"
    name               = "{{.NetworkName}}"
    ip_allocation_mode = "POOL"
  }
}
`

// TestAccVcdStandaloneVmCustomizationSettings tests out possible customization options
func TestAccVcdStandaloneVmCustomizationSettings(t *testing.T) {
	preTestChecks(t)
	var standaloneVmName = fmt.Sprintf("%s-%d", t.Name(), os.Getpid())

	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"Catalog":     testSuiteCatalogName,
		"CatalogItem": testSuiteCatalogOVAItem,
		"VMName":      standaloneVmName,
		"Tags":        "standaloneVm vm",
	}

	configTextVM := templateFill(testAccCheckVcdVmUpdateCustomizationSettings, params)

	params["FuncName"] = t.Name() + "-step1"
	configTextVMStep1 := templateFill(testAccCheckVcdVmUpdateCustomizationSettingsStep1, params)

	params["FuncName"] = t.Name() + "-step2"
	configTextVMStep2 := templateFill(testAccCheckVcdVmUpdateCustomizationSettingsStep2, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVM)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdStandaloneVmDestroy(standaloneVmName, "", ""),
		Steps: []resource.TestStep{
			// Step 1
			resource.TestStep{
				Config: configTextVM,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdStandaloneVmExists(standaloneVmName, "vcd_vm.test-vm", "", ""),
					resource.TestCheckResourceAttr("vcd_vm.test-vm", "name", standaloneVmName),
					resource.TestCheckResourceAttr("vcd_vm.test-vm", "network.#", "0"),

					resource.TestCheckResourceAttr("vcd_vm.test-vm", "customization.#", "1"),
					resource.TestCheckResourceAttr("vcd_vm.test-vm", "customization.0.enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vm.test-vm", "customization.0.change_sid", "true"),
					resource.TestCheckResourceAttr("vcd_vm.test-vm", "customization.0.allow_local_admin_password", "false"),
					resource.TestCheckResourceAttr("vcd_vm.test-vm", "customization.0.must_change_password_on_first_login", "true"),
					resource.TestCheckResourceAttr("vcd_vm.test-vm", "customization.0.number_of_auto_logons", "4"),
				),
			},
			// Step 2 - join org domain (does not fail because enabled=false even though OS is not windows)
			resource.TestStep{
				// Taint:  []string{"vcd_vm.test-vm"},
				// Taint does not work in SDK 2.1.0 therefore every test step has resource address changed to force
				// recreation of the VM
				Config: configTextVMStep1,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdStandaloneVmExists(standaloneVmName, "vcd_vm.test-vm-step2", "", ""),
					resource.TestCheckResourceAttr("vcd_vm.test-vm-step2", "name", standaloneVmName),
					resource.TestCheckResourceAttr("vcd_vm.test-vm-step2", "network.#", "0"),

					resource.TestCheckResourceAttr("vcd_vm.test-vm-step2", "customization.#", "1"),
					resource.TestCheckResourceAttr("vcd_vm.test-vm-step2", "customization.0.enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vm.test-vm-step2", "customization.0.admin_password", "some password"),
					resource.TestCheckResourceAttr("vcd_vm.test-vm-step2", "customization.0.join_domain", "true"),
					resource.TestCheckResourceAttr("vcd_vm.test-vm-step2", "customization.0.join_org_domain", "true"),
				),
			},
			// Step 3 - join org domain enabled
			resource.TestStep{
				// Taint:  []string{"vcd_vm.test-vm"},
				// Taint does not work in SDK 2.1.0 therefore every test step has resource address changed to force
				// recreation of the VM
				Config: configTextVMStep2,
				// Our testing suite does not have Windows OS to actually try domain join so the point of this test is
				// to prove that values are actually set and try to be applied on vCD.
				ExpectError: regexp.MustCompile(`Join Domain is not supported for OS type .*`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdStandaloneVmExists(standaloneVmName, "vcd_vm.test-vm-step3", "", ""),
					resource.TestCheckResourceAttr("vcd_vm.test-vm-step3", "name", standaloneVmName),
					resource.TestCheckResourceAttr("vcd_vm.test-vm-step3", "network.#", "0"),

					resource.TestCheckResourceAttr("vcd_vm.test-vm-step3", "customization.#", "1"),
					resource.TestCheckResourceAttr("vcd_vm.test-vm-step3", "customization.0.enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vm.test-vm-step3", "customization.0.join_domain", "true"),
					resource.TestCheckResourceAttr("vcd_vm.test-vm-step3", "customization.0.join_domain_name", "UnrealDomain"),
					resource.TestCheckResourceAttr("vcd_vm.test-vm-step3", "customization.0.join_domain_user", "NoUser"),
					resource.TestCheckResourceAttr("vcd_vm.test-vm-step3", "customization.0.join_domain_password", "NoPass"),
					resource.TestCheckResourceAttr("vcd_vm.test-vm-step3", "customization.0.join_domain_account_ou", "ou=IT,dc=some,dc=com"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdVmUpdateCustomizationSettings = `
resource "vcd_vm" "test-vm" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name          = "{{.VMName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1

  customization {
	enabled                             = true
	change_sid                          = true
	allow_local_admin_password          = false
	must_change_password_on_first_login = true
	auto_generate_password              = true
	number_of_auto_logons               = 4
  }
}
`

const testAccCheckVcdVmUpdateCustomizationSettingsStep1 = `
# skip-binary-test: it will fail on purpose
resource "vcd_vm" "test-vm-step2" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name          = "{{.VMName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1

  customization {
	enabled         = false
	admin_password  = "some password"
	auto_generate_password = false
	join_domain     = true
	join_org_domain = true
  }
}
`

const testAccCheckVcdVmUpdateCustomizationSettingsStep2 = `
# skip-binary-test: it will fail on purpose
resource "vcd_vm" "test-vm-step3" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name          = "{{.VMName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1

  customization {
	enabled                = true
	join_domain            = true
	join_domain_name       = "UnrealDomain"
	join_domain_user       = "NoUser"
	join_domain_password   = "NoPass"
	join_domain_account_ou = "ou=IT,dc=some,dc=com"
  }
}
`
