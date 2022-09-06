//go:build vdc || ALL || functional
// +build vdc ALL functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdVmPlacementPolicy(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
	}
	if testConfig.VCD.ProviderVdc.Name == "" {
		t.Skip("Variable providerVdc.Name must be set to run VDC tests")
	}

	var params = StringMap{
		"PvdcName":    testConfig.VCD.NsxtProviderVdc.Name,
		"PolicyName":  t.Name(),
		"VmGroup":     testConfig.TestEnvBuild.PlacementPolicyVmGroup,
		"Description": t.Name() + "_description",
	}
	testParamsNotEmpty(t, params)
	policyName := "vcd_vm_placement_policy." + params["PolicyName"].(string)
	datasourcePolicyName := "data.vcd_vm_placement_policy.data-" + params["PolicyName"].(string)
	configText := templateFill(testAccCheckVmPlacementPolicy_create, params)
	params["FuncName"] = t.Name() + "-Update"
	configTextUpdate := templateFill(testAccCheckVmPlacementPolicy_update, params)

	debugPrintf("#[DEBUG] CONFIGURATION - creation: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVmPlacementPolicyDestroyed,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(policyName, "id", regexp.MustCompile(`urn:vcloud:vdcComputePolicy:[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					resource.TestCheckResourceAttr(policyName, "name", params["PolicyName"].(string)),
					resource.TestCheckResourceAttr(policyName, "description", params["Description"].(string)),
					resource.TestMatchResourceAttr(policyName, "provider_vdc_id", regexp.MustCompile(`urn:vcloud:providervdc:[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					resource.TestCheckResourceAttr(policyName, "vm_group_ids.#", "1"),
					resource.TestMatchResourceAttr(policyName, "vm_group_ids.0", regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					resourceFieldsEqual(policyName, datasourcePolicyName, nil),
				),
			},
			{
				Config: configTextUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(policyName, "id", regexp.MustCompile(`urn:vcloud:vdcComputePolicy:[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					resource.TestCheckResourceAttr(policyName, "name", params["PolicyName"].(string)+"-update"),
					resource.TestCheckResourceAttr(policyName, "description", params["Description"].(string)+"-update"),
					resource.TestMatchResourceAttr(policyName, "provider_vdc_id", regexp.MustCompile(`urn:vcloud:providervdc:[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					resource.TestCheckResourceAttr(policyName, "vm_group_ids.#", "1"),
					resource.TestMatchResourceAttr(policyName, "vm_group_ids.0", regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					resourceFieldsEqual(policyName, datasourcePolicyName, nil),
				),
			},
			// Tests import by id
			{
				ResourceName:      policyName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateComputePolicyByIdOrName(policyName, true),
			},
			// Tests import by name
			{
				ResourceName:      policyName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateComputePolicyByIdOrName(policyName, false),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVmPlacementPolicy_create = `
data "vcd_provider_vdc" "pvdc" {
  name = "{{.PvdcName}}"
}

data "vcd_vm_group" "vm-group" {
  name            = "{{.VmGroup}}"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
}

resource "vcd_vm_placement_policy" "{{.PolicyName}}" {
  name        = "{{.PolicyName}}"
  description = "{{.Description}}"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
  vm_group_ids = [ data.vcd_vm_group.vm-group.id ]
}

data "vcd_vm_placement_policy" "data-{{.PolicyName}}" {
	name = vcd_vm_placement_policy.{{.PolicyName}}.name
    provider_vdc_id = vcd_vm_placement_policy.{{.PolicyName}}.provider_vdc_id
}
`

const testAccCheckVmPlacementPolicy_update = `
data "vcd_provider_vdc" "pvdc" {
  name = "{{.PvdcName}}"
}

data "vcd_vm_group" "vm-group" {
  name            = "{{.VmGroup}}"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
}

resource "vcd_vm_placement_policy" "{{.PolicyName}}" {
  name        = "{{.PolicyName}}-update"
  description = "{{.Description}}-update"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
  vm_group_ids = [ data.vcd_vm_group.vm-group.id ]
}

data "vcd_vm_placement_policy" "data-{{.PolicyName}}" {
	name = vcd_vm_placement_policy.{{.PolicyName}}.name
    provider_vdc_id = vcd_vm_placement_policy.{{.PolicyName}}.provider_vdc_id
}
`

// TestAccVcdVmPlacementPolicyWithoutDescription checks that a VM Placement Policy without description specified in the HCL
// corresponds to a VM Placement Policy with an empty description in VCD.
func TestAccVcdVmPlacementPolicyWithoutDescription(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
	}
	if testConfig.VCD.ProviderVdc.Name == "" {
		t.Skip("Variable providerVdc.Name must be set to run VDC tests")
	}

	var params = StringMap{
		"PvdcName":    testConfig.VCD.NsxtProviderVdc.Name,
		"PolicyName":  t.Name(),
		"VmGroup":     testConfig.TestEnvBuild.PlacementPolicyVmGroup,
	}
	testParamsNotEmpty(t, params)
	policyName := "vcd_vm_placement_policy." + params["PolicyName"].(string)
	configText := templateFill(testAccCheckVmPlacementPolicyWithoutDescription, params)

	debugPrintf("#[DEBUG] CONFIGURATION - creation: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVmPlacementPolicyDestroyed,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(policyName, "description", ""),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVmPlacementPolicyWithoutDescription = `
data "vcd_provider_vdc" "pvdc" {
  name = "{{.PvdcName}}"
}

data "vcd_vm_group" "vm-group" {
  name            = "{{.VmGroup}}"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
}

resource "vcd_vm_placement_policy" "{{.PolicyName}}" {
  name        = "{{.PolicyName}}"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
  vm_group_ids = [ data.vcd_vm_group.vm-group.id ]
}
`

func testAccCheckVmPlacementPolicyDestroyed(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)
	var err error
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_vm_placement_policy" && rs.Primary.Attributes["name"] != "TestAccVcdVmPlacementPolicy" {
			continue
		}

		_, err = conn.GetVdcComputePolicyV2ById(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("VM Placement Policy %s still exists", rs.Primary.ID)
		}
	}
	return nil
}
