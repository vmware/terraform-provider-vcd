//go:build ALL || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdDatasourceVmGroup(t *testing.T) {
	// Pre-checks
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	// Test configuration
	var params = StringMap{
		"ProviderVdcName": testConfig.VCD.NsxtProviderVdc.Name,
		"VmGroup":         testConfig.VCD.NsxtProviderVdc.PlacementPolicyVmGroup,
	}
	testParamsNotEmpty(t, params)
	configText := templateFill(testAccVcdDatasourceVmGroup, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// Test cases
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.vcd_vm_group.vm-group", "provider_vdc_id", "data.vcd_provider_vdc.pvdc1", "id"),
					resource.TestCheckResourceAttrSet("data.vcd_vm_group.vm-group", "name"),
					resource.TestCheckResourceAttrSet("data.vcd_vm_group.vm-group", "cluster_moref"),
					resource.TestCheckResourceAttrSet("data.vcd_vm_group.vm-group", "cluster_name"),
					resource.TestMatchResourceAttr("data.vcd_vm_group.vm-group", "vcenter_id", getUuidRegex("^", "$")),
					resource.TestMatchResourceAttr("data.vcd_vm_group.vm-group", "named_vm_group_id", getUuidRegex("^", "$")),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdDatasourceVmGroup = `
data "vcd_provider_vdc" "pvdc1" {
    name = "{{.ProviderVdcName}}"
}

data "vcd_vm_group" "vm-group" {
    name            = "{{.VmGroup}}"
    provider_vdc_id = data.vcd_provider_vdc.pvdc1.id
}
`
