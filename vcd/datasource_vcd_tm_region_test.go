//go:build tm || ALL || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdTmRegionDs(t *testing.T) {
	preTestChecks(t)

	skipIfNotSysAdmin(t)
	skipIfNotTm(t)

	var params = StringMap{
		"Region": testConfig.Tm.Region,

		"Tags": "tm",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(TestAccVcdTmRegionDsStep1, params)

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
					resource.TestCheckResourceAttr("data.vcd_tm_region.test", "name", params["Region"].(string)),
					resource.TestCheckResourceAttr("data.vcd_tm_region.test", "description", ""),
					resource.TestCheckResourceAttrSet("data.vcd_tm_region.test", "is_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_tm_region.test", "nsx_manager_id"),
					resource.TestCheckResourceAttrSet("data.vcd_tm_region.test", "cpu_capacity_mhz"),
					resource.TestCheckResourceAttrSet("data.vcd_tm_region.test", "cpu_reservation_capacity_mhz"),
					resource.TestCheckResourceAttrSet("data.vcd_tm_region.test", "memory_capacity_mib"),
					resource.TestCheckResourceAttrSet("data.vcd_tm_region.test", "memory_reservation_capacity_mib"),
					resource.TestCheckResourceAttrSet("data.vcd_tm_region.test", "status"),
					resource.TestCheckResourceAttrSet("data.vcd_tm_region.test", "supervisors.#"),
					resource.TestCheckResourceAttrSet("data.vcd_tm_region.test", "storage_policies.#"),
				),
			},
		},
	})

	postTestChecks(t)
}

const TestAccVcdTmRegionDsStep1 = `
data "vcd_tm_region" "test" {
  name = "{{.Region}}"
}
`
