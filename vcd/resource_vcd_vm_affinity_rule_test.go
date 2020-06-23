// +build vm functional affinity ALL

package vcd

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// affinityRuleData is the definition of a VM affinity rule
type affinityRuleData struct {
	name        string               // Name of the rule
	polarity    string               // Affinity or Anti-Affinity
	creationVms map[string]*types.VM // List of the VMs to add on creation
	updateVms   map[string]*types.VM // List of the VMs to add on update
}

// TestAccVcdVmAffinityRule creates the pre-requisites for the VM affinity rule test
// Creates several definitions, and calls runVmAffinityRuleTest for each one
func TestAccVcdVmAffinityRule(t *testing.T) {
	// This test requires access to the vCD before filling templates
	// Thus it won't run in the short test
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	if testConfig.VCD.Org == "" {
		t.Skip("[TestAccVcdVmAffinityRule] no Org found in configuration")
	}
	if testConfig.VCD.Vdc == "" {
		t.Skip("[TestAccVcdVmAffinityRule] no VDC found in configuration")
	}

	client := createTemporaryVCDConnection()

	_, vdc, err := client.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
	if err != nil {
		t.Errorf("error retrieving org and VDC: %s", err)
	}

	vappDefinition := map[string][]string{
		"Test_EmptyVmVapp1": []string{"Test_EmptyVm1a", "Test_EmptyVm1b"},
		"Test_EmptyVmVapp2": []string{"Test_EmptyVm2a", "Test_EmptyVm2b"},
		"Test_EmptyVmVapp3": []string{"Test_EmptyVm3a", "Test_EmptyVm3b"},
	}
	vappList, err := makeVappGroup("TestAccVcdVmAffinityRule", vdc, vappDefinition)
	if err != nil {
		t.Errorf("error creating vApp collection: %s", err)
		return
	}

	defer func() {
		if os.Getenv("GOVCD_KEEP_TEST_OBJECTS") != "" {
			if vcdTestVerbose {
				fmt.Printf("Skipping vApp removal: GOVCD_KEEP_TEST_OBJECTS was set\n")
			}
			return
		}
		for _, vapp := range vappList {
			if vcdTestVerbose {
				fmt.Printf("Removing vApp %s\n", vapp.VApp.Name)
			}
			task, err := vapp.Delete()
			if err == nil {
				_ = task.WaitTaskCompletion()
			}
		}
	}()

	affinityRuleList, err := vdc.GetAllVmAffinityRuleList()
	if err != nil {
		t.Skip("could not retrieve the list of existing VM affinity rules")
	}

	// Checks for usage of VM affinity rules or VMs that may conflict with this test.
	var foundNames []string
	type usedVm struct {
		vappName string
		vmName   string
		ruleName string
	}
	var seenVms = make(map[string]usedVm)

	for i := 1; i <= 4; i++ {
		name := fmt.Sprintf("Test_VmAffinityRule%d", i)
		for _, rule := range affinityRuleList {
			// Checks for names already used
			// While affinity rules admit duplicate names, this test needs unique ones, to test specific features
			if name == rule.Name {
				foundNames = append(foundNames, name)
			}
			// Checks that no VM in the pool created (or retrieved) for this test
			// is used in another affinity rule
			for _, vmInRule := range rule.VmReferences[0].VMReference {
				for _, vapp := range vappList {
					for _, vm := range vapp.VApp.Children.VM {
						if vmInRule.HREF == vm.HREF {
							seenVms[vapp.VApp.Name+"-"+vm.Name+"-"+rule.Name] = usedVm{vapp.VApp.Name, vm.Name, rule.Name}
						}
					}
				}
			}
		}
	}
	if len(foundNames) > 0 {
		t.Skipf("found VM affinity rules with conflicting names: %v", foundNames)
	}
	if len(seenVms) > 0 {
		fmt.Println(strings.Repeat("-", 50))
		for _, usedVm := range seenVms {
			fmt.Printf("%s\n", fmt.Sprintf("VM %s in vApp %s already used in rule %s", usedVm.vmName, usedVm.vappName, usedVm.ruleName))
		}
		t.Skip("VMs needed for this test are already used in one or more affinity rules")
	}

	// End of preparation
	// Now we run one test for each affinity rule definition

	t.Run("Affinity1", func(t *testing.T) {
		runVmAffinityRuleTest(affinityRuleData{
			name:     "Test_VmAffinityRule1",
			polarity: types.PolarityAffinity,
			creationVms: map[string]*types.VM{
				"Test_EmptyVm1a": vappList[0].VApp.Children.VM[0],
				"Test_EmptyVm1b": vappList[0].VApp.Children.VM[1],
			},
			updateVms: map[string]*types.VM{
				"Test_EmptyVm2a": vappList[1].VApp.Children.VM[0],
				"Test_EmptyVm2b": vappList[1].VApp.Children.VM[1],
			},
		}, t)
	})

	t.Run("Affinity2", func(t *testing.T) {
		runVmAffinityRuleTest(affinityRuleData{
			name:     "Test_VmAffinityRule2",
			polarity: types.PolarityAffinity,
			creationVms: map[string]*types.VM{
				"Test_EmptyVm1a": vappList[0].VApp.Children.VM[0],
				"Test_EmptyVm1b": vappList[0].VApp.Children.VM[1],
				"Test_EmptyVm2a": vappList[1].VApp.Children.VM[0],
			},
			updateVms: map[string]*types.VM{
				"Test_EmptyVm3a": vappList[2].VApp.Children.VM[0],
				"Test_EmptyVm3b": vappList[2].VApp.Children.VM[1],
			},
		}, t)
	})

	t.Run("Affinity3", func(t *testing.T) {
		runVmAffinityRuleTest(affinityRuleData{
			name:     "Test_VmAffinityRule3",
			polarity: types.PolarityAffinity,
			creationVms: map[string]*types.VM{
				"Test_EmptyVm1a": vappList[0].VApp.Children.VM[0],
				"Test_EmptyVm1b": vappList[0].VApp.Children.VM[1],
				"Test_EmptyVm2a": vappList[1].VApp.Children.VM[0],
				"Test_EmptyVm2b": vappList[1].VApp.Children.VM[1],
				"Test_EmptyVm3a": vappList[2].VApp.Children.VM[0],
				"Test_EmptyVm3b": vappList[2].VApp.Children.VM[1],
			},
			updateVms: nil,
		}, t)
	})

	t.Run("Anti-Affinity1", func(t *testing.T) {
		runVmAffinityRuleTest(affinityRuleData{
			name:     "Test_VmAffinityRule4",
			polarity: types.PolarityAntiAffinity,
			creationVms: map[string]*types.VM{
				"Test_EmptyVm1a": vappList[0].VApp.Children.VM[0],
				"Test_EmptyVm2a": vappList[1].VApp.Children.VM[0],
			},
			updateVms: map[string]*types.VM{
				"Test_EmptyVm3a": vappList[2].VApp.Children.VM[0],
			},
		}, t)
	})

}

// runVmAffinityRuleTest runs the test for a VM affinity rule definition
func runVmAffinityRuleTest(data affinityRuleData, t *testing.T) {

	var creationVmNames []string
	var updateVmNames []string
	count := 0

	// Builds the list of VM IDs used for creation
	for vmName := range data.creationVms {
		creationVmNames = append(creationVmNames, fmt.Sprintf("data.vcd_vapp_vm.%s.id", vmName))
		if !(count == 0 && len(data.creationVms) > 2 && len(data.updateVms) == 0) {
			// if the update list is empty and the creation has more than 2 items,
			// the updated list will be the result of a deletion (the first VM will be skipped)
			updateVmNames = append(updateVmNames, fmt.Sprintf("data.vcd_vapp_vm.%s.id", vmName))
		}
		count++
	}

	// Finished the build of VM list for update, by appending the VMs indicated in the update list to
	// the ones in creation.
	for vmName := range data.updateVms {
		updateVmNames = append(updateVmNames, fmt.Sprintf("data.vcd_vapp_vm.%s.id", vmName))
	}
	expectedUpdateVMs := len(data.updateVms) + len(data.creationVms)
	if len(data.updateVms) == 0 {
		if len(data.creationVms) > 2 {
			expectedUpdateVMs = len(data.creationVms) - 1
		}
	}
	var params = StringMap{
		"Org":                    testConfig.VCD.Org,
		"Vdc":                    testConfig.VCD.Vdc,
		"AffinityRuleIdentifier": data.name,
		"AffinityRuleName":       data.name,
		"Polarity":               data.polarity,
		"Enabled":                "true",
		"Required":               "true",
		"VirtualMachineIds":      strings.Join(creationVmNames, ",\n    "),
		"FuncName":               data.name,
		"SkipNotice":             "# skip-binary-test: needs external resources",
	}

	configText := templateFill(testAccVmAffinityRuleBase+
		testAccVmAffinityRuleOperation+
		testAccVmAffinityRuleDataSource, params)

	params["FuncName"] = data.name + "-update"
	params["AffinityRuleName"] = data.name + "-update"
	params["VirtualMachineIds"] = strings.Join(updateVmNames, ",\n    ")
	params["Required"] = "false"
	params["Enabled"] = "false"
	params["SkipNotice"] = "# skip-binary-test: only for updates"
	updateText := templateFill(testAccVmAffinityRuleBase+testAccVmAffinityRuleOperation, params)

	debugPrintf("#[DEBUG] CREATION CONFIGURATION: %s", configText)
	debugPrintf("#[DEBUG] UPDATE CONFIGURATION: %s", updateText)

	var rule govcd.VmAffinityRule
	resourceName := "vcd_vm_affinity_rule." + data.name
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVmAffinityRuleDestroy(&rule, testConfig.VCD.Org, testConfig.VCD.Vdc),
		Steps: []resource.TestStep{
			// Test creation
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVmAffinityRuleExists(resourceName, testConfig.VCD.Org, testConfig.VCD.Vdc, &rule),
					resource.TestCheckResourceAttr(
						resourceName, "name", data.name),
					resource.TestCheckResourceAttr(
						resourceName, "enabled", fmt.Sprintf("%v", "true")),
					resource.TestCheckResourceAttr(
						resourceName, "required", fmt.Sprintf("%v", "true")),
					resource.TestCheckResourceAttr(
						resourceName, "virtual_machine_ids.#", fmt.Sprintf("%d", len(data.creationVms))),
					resource.TestCheckResourceAttr(
						resourceName, "polarity", data.polarity),
					resource.TestCheckOutput("name_of_rule_by_id", data.name),
					resource.TestCheckOutput("polarity_of_rule_by_name", data.polarity),
				),
			},
			// Tests update
			resource.TestStep{
				Config: updateText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVmAffinityRuleExists(resourceName, testConfig.VCD.Org, testConfig.VCD.Vdc, &rule),
					resource.TestCheckResourceAttr(
						resourceName, "name", data.name+"-update"),
					resource.TestCheckResourceAttr(
						resourceName, "enabled", fmt.Sprintf("%v", "false")),
					resource.TestCheckResourceAttr(
						resourceName, "required", fmt.Sprintf("%v", "false")),
					resource.TestCheckResourceAttr(
						resourceName, "virtual_machine_ids.#", fmt.Sprintf("%d", expectedUpdateVMs)),
					resource.TestCheckResourceAttr(
						resourceName, "polarity", data.polarity),
				),
			},
			// Tests import by name
			resource.TestStep{
				Config:            updateText,
				ResourceName:      "vcd_vm_affinity_rule." + data.name + "-import-name",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgVdcObject(testConfig, data.name+"-update"),
			},
			// Tests import by ID
			resource.TestStep{
				Config:            updateText,
				ResourceName:      "vcd_vm_affinity_rule." + data.name + "-import-id",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdByAffinityRule("vcd_vm_affinity_rule." + data.name),
			},
		},
	})
}

// testAccCheckVcdVmAffinityRuleExists checks that the VM affinity rule was created
// The data collected is saved into the 'rule' object, to be using during destruction checks
func testAccCheckVcdVmAffinityRuleExists(resourceName string, orgName, vdcName string, rule *govcd.VmAffinityRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no VM affinity rule ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		org, err := conn.VCDClient.GetAdminOrgByName(orgName)
		if err != nil {
			return fmt.Errorf("error: could not find Org: %s", err)
		}
		vdc, err := org.GetVDCByName(vdcName, false)
		if err != nil {
			return fmt.Errorf("error: could not find VDC: %s", err)
		}

		vmAffinityRule, err := vdc.GetVmAffinityRuleById(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error: could not find VM affinity rule: %s", err)
		}

		*rule = *vmAffinityRule

		return nil
	}
}

// testAccCheckVmAffinityRuleDestroy checks that the affinity rule was destroyed
// The 'rule' object was created when checking for creation
func testAccCheckVmAffinityRuleDestroy(rule *govcd.VmAffinityRule, orgName, vdcName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if rule == nil || rule.VmAffinityRule == nil {
			return fmt.Errorf("affinity rule passed to destroy check is null")
		}
		conn := testAccProvider.Meta().(*VCDClient)
		org, err := conn.VCDClient.GetAdminOrgByName(orgName)
		if err != nil {
			return fmt.Errorf("error: could not find Org: %s", err)
		}
		vdc, err := org.GetVDCByName(vdcName, false)
		if err != nil {
			return fmt.Errorf("error: could not find VDC: %s", err)
		}

		var vmAffinityRule *govcd.VmAffinityRule
		for N := 0; N < 15; N++ {

			vmAffinityRule, err = vdc.GetVmAffinityRuleById(rule.VmAffinityRule.ID)
			if err != nil && vmAffinityRule == nil {
				break
			}
			time.Sleep(time.Second)
		}

		if err != govcd.ErrorEntityNotFound {
			return fmt.Errorf("VM affinity rule %s was not destroyed", rule.VmAffinityRule.Name)
		}
		if vmAffinityRule != nil {
			return fmt.Errorf("rule %s was found", rule.VmAffinityRule.Name)
		}
		return nil
	}
}

// importStateIdByAffinityRule runs the import of a VM affinity rule using the resource ID
func importStateIdByAffinityRule(resource string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resource)
		}

		if rs.Primary.ID == "" {
			return "", fmt.Errorf("no ID is set for %s resource", resource)
		}

		importId := testConfig.VCD.Org + "." + testConfig.VCD.Vdc + "." + rs.Primary.ID
		if testConfig.VCD.Org == "" || testConfig.VCD.Vdc == "" || rs.Primary.ID == "" {
			return "", fmt.Errorf("missing information to generate import path: %s", importId)
		}
		return importId, nil
	}
}

// makeEmptyVapp creates a given vApp without any VM
func makeEmptyVapp(vdc *govcd.Vdc, name string) (*govcd.VApp, error) {

	err := vdc.ComposeRawVApp(name)
	if err != nil {
		return nil, err
	}
	vapp, err := vdc.GetVAppByName(name, true)
	if err != nil {
		return nil, err
	}
	initialVappStatus, err := vapp.GetStatus()
	if err != nil {
		return nil, err
	}
	if initialVappStatus != "RESOLVED" {
		err = vapp.BlockWhileStatus(initialVappStatus, testConfig.Provider.MaxRetryTimeout)
		if err != nil {
			return nil, err
		}
	}
	return vapp, nil
}

// makeEmptyVm creates an empty VM inside a given vApp
func makeEmptyVm(vapp *govcd.VApp, name string) (*govcd.VM, error) {
	newDisk := types.DiskSettings{
		AdapterType:       "5",
		SizeMb:            int64(100),
		BusNumber:         0,
		UnitNumber:        0,
		ThinProvisioned:   takeBoolPointer(true),
		OverrideVmDefault: true}
	requestDetails := &types.RecomposeVAppParamsForEmptyVm{
		CreateItem: &types.CreateItem{
			Name:                      name,
			NetworkConnectionSection:  &types.NetworkConnectionSection{},
			Description:               "created by makeEmptyVm",
			GuestCustomizationSection: nil,
			VmSpecSection: &types.VmSpecSection{
				Modified:          takeBoolPointer(true),
				Info:              "Virtual Machine specification",
				OsType:            "debian10Guest",
				NumCpus:           takeIntPointer(1),
				NumCoresPerSocket: takeIntPointer(1),
				CpuResourceMhz:    &types.CpuResourceMhz{Configured: 1},
				MemoryResourceMb:  &types.MemoryResourceMb{Configured: 512},
				MediaSection:      nil,
				DiskSection:       &types.DiskSection{DiskSettings: []*types.DiskSettings{&newDisk}},
				HardwareVersion:   &types.HardwareVersion{Value: "vmx-13"},
				VmToolsVersion:    "",
				VirtualCpuType:    "VM32",
				TimeSyncWithHost:  nil,
			},
			BootImage: nil,
		},
		AllEULAsAccepted: true,
	}

	vm, err := vapp.AddEmptyVm(requestDetails)
	if err != nil {
		return nil, err
	}

	return vm, nil
}

// makeVappGroup creates multiple vApps, each with several VMs,
// as defined in `groupDefinition`.
// Returns a list of vApps
func makeVappGroup(label string, vdc *govcd.Vdc, groupDefinition map[string][]string) ([]*govcd.VApp, error) {
	var vappList []*govcd.VApp
	for vappName, vmNames := range groupDefinition {

		existingVapp, err := vdc.GetVAppByName(vappName, false)
		if err == nil {

			if existingVapp.VApp.Children == nil || len(existingVapp.VApp.Children.VM) == 0 {
				return nil, fmt.Errorf("found vApp %s but without VMs", vappName)
			}
			foundVms := 0
			for _, vmName := range vmNames {
				for _, existingVM := range existingVapp.VApp.Children.VM {
					if existingVM.Name == vmName {
						foundVms++
					}
				}
			}
			if foundVms < 2 {
				return nil, fmt.Errorf("found vApp %s but with %d VMs instead of 2 ", vappName, foundVms)
			}

			vappList = append(vappList, existingVapp)
			if vcdTestVerbose {
				fmt.Printf("Using existing vApp %s\n", vappName)
			}
			continue
		}
		if vcdTestVerbose {
			fmt.Printf("Creating vApp %s\n", vappName)
		}
		vapp, err := makeEmptyVapp(vdc, vappName)
		if err != nil {
			return nil, err
		}
		for _, vmName := range vmNames {
			if vcdTestVerbose {
				fmt.Printf("\tCreating VM %s/%s\n", vappName, vmName)
			}
			_, err := makeEmptyVm(vapp, vmName)
			if err != nil {
				return nil, err
			}
		}
		vappList = append(vappList, vapp)
	}
	return vappList, nil
}

// testAccVmAffinityRuleBase is the fixed part of the test.
// It uses the vApps and VMs created by the test preparation procedure
// The data sources listed here are used in the dynamic test to reference
// the VM needed for the affinity rules
const testAccVmAffinityRuleBase = `
data "vcd_vapp" "Test_EmptyVmVapp1" {
  name = "Test_EmptyVmVapp1"
}

data "vcd_vapp" "Test_EmptyVmVapp2" {
  name = "Test_EmptyVmVapp2"
}

data "vcd_vapp" "Test_EmptyVmVapp3" {
  name = "Test_EmptyVmVapp3"
}

data "vcd_vapp_vm" "Test_EmptyVm1a" {
  vapp_name = data.vcd_vapp.Test_EmptyVmVapp1.name
  name      = "Test_EmptyVm1a"
}

data "vcd_vapp_vm" "Test_EmptyVm1b" {
  vapp_name = data.vcd_vapp.Test_EmptyVmVapp1.name
  name      = "Test_EmptyVm1b"
}

data "vcd_vapp_vm" "Test_EmptyVm2a" {
  vapp_name = data.vcd_vapp.Test_EmptyVmVapp2.name
  name      = "Test_EmptyVm2a"
}

data "vcd_vapp_vm" "Test_EmptyVm2b" {
  vapp_name = data.vcd_vapp.Test_EmptyVmVapp2.name
  name      = "Test_EmptyVm2b"
}

data "vcd_vapp_vm" "Test_EmptyVm3a" {
  vapp_name = data.vcd_vapp.Test_EmptyVmVapp3.name
  name      = "Test_EmptyVm3a"
}

data "vcd_vapp_vm" "Test_EmptyVm3b" {
  vapp_name = data.vcd_vapp.Test_EmptyVmVapp3.name
  name      = "Test_EmptyVm3b"
}
`

// testAccVmAffinityRuleOperation is the dynamic part of the test
// This template is filled for every affinity rule definition
const testAccVmAffinityRuleOperation = `
{{.SkipNotice}}
resource "vcd_vm_affinity_rule" "{{.AffinityRuleIdentifier}}" {

  org      = "{{.Org}}"
  vdc      = "{{.Vdc}}"
  name     = "{{.AffinityRuleName}}"
  required = {{.Required}}
  enabled  = {{.Enabled}}
  polarity = "{{.Polarity}}"

  virtual_machine_ids = [
    {{.VirtualMachineIds}}
  ]
}
`

const testAccVmAffinityRuleDataSource = `
{{.SkipNotice}}
data "vcd_vm_affinity_rule" "ds_affinity_rule_by_name" {
	name = vcd_vm_affinity_rule.{{.AffinityRuleIdentifier}}.name
}

data "vcd_vm_affinity_rule" "ds_affinity_rule_by_id" {
	rule_id = vcd_vm_affinity_rule.{{.AffinityRuleIdentifier}}.id
}

output "polarity_of_rule_by_name" {
	value = data.vcd_vm_affinity_rule.ds_affinity_rule_by_name.polarity
}

output "name_of_rule_by_id" {
	value = data.vcd_vm_affinity_rule.ds_affinity_rule_by_id.name
}
`
