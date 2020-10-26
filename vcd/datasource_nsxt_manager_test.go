// +build ALL nsxt functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdDatasourceNsxtManager(t *testing.T) {

	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	skipNoNsxtConfiguration(t)

	var params = StringMap{
		"FuncName":    t.Name(),
		"NsxtManager": testConfig.Nsxt.Manager,
		"Tags":        "nsxt",
	}

	configText := templateFill(testAccCheckVcdNsxtManager, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					// ID must match URN 'urn:vcloud:nsxtmanager:09722307-aee0-4623-af95-7f8e577c9ebc'
					resource.TestMatchResourceAttr("data.vcd_nsxt_manager.nsxt", "id",
						regexp.MustCompile(`urn:vcloud:nsxtmanager:[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					resource.TestCheckResourceAttr("data.vcd_nsxt_manager.nsxt", "name", params["NsxtManager"].(string)),
				),
			},
		},
	})
}

const testAccCheckVcdNsxtManager = `
data "vcd_nsxt_manager" "nsxt" {
  name = "{{.NsxtManager}}"
}
`
