//go:build tm || ALL || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNsxtManager(t *testing.T) {
	preTestChecks(t)

	skipIfNotSysAdmin(t)
	skipIfNotTm(t)

	if !testConfig.Tm.CreateNsxtManager {
		t.Skipf("Skipping vCenter creation")
	}

	var params = StringMap{
		"Testname": t.Name(),
		"Username": testConfig.Tm.NsxtManagerUsername,
		"Password": testConfig.Tm.NsxtManagerPassword,
		"Url":      testConfig.Tm.NsxtManagerUrl,

		"Tags": "tm",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdNsxtManagerStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdNsxtManagerStep2, params)

	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText2)
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
					resource.TestCheckResourceAttr("vcd_nsxt_manager.test", "name", params["Testname"].(string)),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcd_nsxt_manager.test", "data.vcd_nsxt_manager.test", []string{"%"}),
				),
			},
			{
				ResourceName:      "vcd_nsxt_manager.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     params["Testname"].(string),
			},
		},
	})

	postTestChecks(t)
}

const testAccVcdNsxtManagerStep1 = `
resource "vcd_nsxt_manager" "test" {
  name                   = "{{.Testname}}"
  description            = "terraform test"
  username               = "{{.Username}}"
  password               = "{{.Password}}"
  url                    = "{{.Url}}"
  network_provider_scope = ""
}
`

const testAccVcdNsxtManagerStep2 = testAccVcdNsxtManagerStep1 + `
data "vcd_nsxt_manager" "test" {
  name = vcd_nsxt_manager.test.name
}
`
