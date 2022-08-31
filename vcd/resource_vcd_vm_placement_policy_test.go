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
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
	}
	if testConfig.VCD.ProviderVdc.Name == "" {
		t.Skip("Variable providerVdc.Name must be set to run VDC tests")
	}

	// Get VM Group ID by PlacementPolicyVmGroup
	vcdClient := createTemporaryVCDConnection(true)
	vmGroupId := ""
	if vcdClient != nil {
		vmGroup, err := vcdClient.VCDClient.GetVmGroupByName(testConfig.TestEnvBuild.PlacementPolicyVmGroup)
		if err != nil {
			t.Skip(t.Name() + " could not find VM Group in testEnvBuild.placementPolicyVmGroup required to test VM Placement Policies")
		}
		vmGroupId = vmGroup.VmGroup.NamedVmGroupId
	}
	if vmGroupId == "" {
		t.Skip(t.Name() + " could not find VM Group in testEnvBuild.placementPolicyVmGroup required to test VM Placement Policies")
	}

	var params = StringMap{
		"PvdcName":    testConfig.VCD.NsxtProviderVdc.Name,
		"VmGroupId":   fmt.Sprintf("urn:vcloud:namedVmGroup:%s", vmGroupId),
		"PolicyName":  t.Name(),
		"Description": t.Name() + "_description",
	}
	testParamsNotEmpty(t, params)
	policyName := "vcd_vm_placement_policy." + params["PolicyName"].(string)
	datasourcePolicyName := "data.vcd_vm_placement_policy.data-" + params["PolicyName"].(string)
	configText := templateFill(testAccCheckVmPlacementPolicy_create, params)
	params["FuncName"] = t.Name() + "-Update"
	configTextUpdate := templateFill(testAccCheckVmPlacementPolicy_update, params)

	debugPrintf("#[DEBUG] CONFIGURATION - creation: %s", configText)

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
					resource.TestMatchResourceAttr(policyName, "vm_group_ids.0", regexp.MustCompile(`urn:vcloud:namedVmGroup:[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
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
					resource.TestMatchResourceAttr(policyName, "vm_group_ids.0", regexp.MustCompile(`urn:vcloud:namedVmGroup:[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					resourceFieldsEqual(policyName, datasourcePolicyName, nil),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVmPlacementPolicy_create = `
data "vcd_provider_vdc" "pvdc" {
  name = "{{.PvdcName}}"
}

resource "vcd_vm_placement_policy" "{{.PolicyName}}" {
  name        = "{{.PolicyName}}"
  description = "{{.Description}}"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
  vm_group_ids = [ "{{.VmGroupId}}" ]
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

resource "vcd_vm_placement_policy" "{{.PolicyName}}" {
  name        = "{{.PolicyName}}-update"
  description = "{{.Description}}-update"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
  vm_group_ids = [ "{{.VmGroupId}}" ]
}

data "vcd_vm_placement_policy" "data-{{.PolicyName}}" {
	name = vcd_vm_placement_policy.{{.PolicyName}}.name
    provider_vdc_id = vcd_vm_placement_policy.{{.PolicyName}}.provider_vdc_id
}
`

func testAccCheckVmPlacementPolicyDestroyed(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)
	var err error
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_vm_placement_policy" && rs.Primary.Attributes["name"] != "TestAccVcdVmPlacementPolicy" {
			continue
		}

		_, err = conn.Client.GetVdcComputePolicyById(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("VM sizing policy %s still exists", rs.Primary.ID)
		}
	}

	return nil
}
