//go:build vdc || ALL || functional
// +build vdc ALL functional

package vcd

import (
	"fmt"
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

	configText := templateFill(testAccCheckVmPlacementPolicy_basic, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION - creation: %s", configText)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		//	CheckDestroy:      testAccCheckVmSizingPolicyDestroyed,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVmPlacementPolicy_basic = `
data "vcd_provider_vdc" "pvdc" {
  name = "{{.PvdcName}}"
}

resource "vcd_vm_placement_policy" "{{.PolicyName}}" {
  name        = "{{.PolicyName}}"
  description = "{{.Description}}"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
  vm_group_ids = [ "{{.VmGroupId}}" ]
}
`
