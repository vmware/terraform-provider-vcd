// +build ALL nsxt functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdDatasourceNsxtTier0Router checks if datasource can find existing regular Tier-0 router
// provided it is specified in configuration
func TestAccVcdDatasourceNsxtTier0Router(t *testing.T) {
	testAccVcdDatasourceNsxtTier0Router(t, testConfig.Nsxt.Tier0router)
}

// TestAccVcdDatasourceNsxtTier0Router checks if datasource can find existing VRF Tier-0 router
// provided it is specified in configuration
func TestAccVcdDatasourceNsxtTier0RouterVrf(t *testing.T) {
	testAccVcdDatasourceNsxtTier0Router(t, testConfig.Nsxt.Tier0routerVrf)
}

func testAccVcdDatasourceNsxtTier0Router(t *testing.T, tier0RouterName string) {

	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	skipNoNsxtConfiguration(t)

	var params = StringMap{
		"FuncName":        t.Name(),
		"NsxtManager":     testConfig.Nsxt.Manager,
		"NsxtTier0Router": tier0RouterName,
		"Tags":            "nsxt",
	}

	configText := templateFill(testAccCheckVcdNsxtTier0Router, params)

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
					resource.TestCheckResourceAttr("data.vcd_nsxt_tier0_router.router", "name", params["NsxtTier0Router"].(string)),
				),
			},
		},
	})
}

const testAccCheckVcdNsxtTier0Router = `
data "vcd_nsxt_manager" "nsxt" {
  name = "{{.NsxtManager}}"
}
data "vcd_nsxt_tier0_router" "router" {
  name            = "{{.NsxtTier0Router}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}
`
