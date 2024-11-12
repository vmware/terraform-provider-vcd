//go:build tm || ALL || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdTmVdcDs(t *testing.T) {
	preTestChecks(t)

	skipIfNotSysAdmin(t)
	skipIfNotTm(t)

	var params = StringMap{
		"Vdc": testConfig.Tm.Vdc,

		"Tags": "tm",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(TestAccVcdTmVdcDsStep1, params)

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vcd_tm_org_vdc.test", "name", params["Vdc"].(string)),
					resource.TestCheckResourceAttr("data.vcd_tm_org_vdc.test", "description", ""),
					resource.TestCheckResourceAttrSet("data.vcd_tm_org_vdc.test", "is_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_tm_org_vdc.test", "org_id"),
					resource.TestCheckResourceAttrSet("data.vcd_tm_org_vdc.test", "region_id"),
					resource.TestCheckResourceAttrSet("data.vcd_tm_org_vdc.test", "status"),
					resource.TestCheckResourceAttrSet("data.vcd_tm_org_vdc.test", "supervisor_ids.#"),
					resource.TestCheckResourceAttrSet("data.vcd_tm_org_vdc.test", "zone_resource_allocations.#"),
				),
			},
		},
	})

	postTestChecks(t)
}

const TestAccVcdTmVdcDsStep1 = `
data "vcd_tm_org_vdc" "test" {
  name = "{{.Vdc}}"
}
`
