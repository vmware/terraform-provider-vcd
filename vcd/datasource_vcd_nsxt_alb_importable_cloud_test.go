//go:build nsxt || alb || ALL || functional
// +build nsxt alb ALL functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNsxtAlbImportableCloudDS(t *testing.T) {
	preTestChecks(t)

	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	skipNoNsxtAlbConfiguration(t)

	// String map to fill the template
	var params = StringMap{
		"ControllerName":     t.Name(),
		"ControllerUrl":      testConfig.Nsxt.NsxtAlbControllerUrl,
		"ControllerUsername": testConfig.Nsxt.NsxtAlbControllerUser,
		"ControllerPassword": testConfig.Nsxt.NsxtAlbControllerPassword,
		"ImportableCloud":    testConfig.Nsxt.NsxtAlbImportableCloud,
		"Tags":               "alb nsxt",
	}
	vcdClient := createTemporaryVCDConnection(true)
	// From API v37.0 onwards, license_type is no longer used
	params["LicenseType"] = ""
	if vcdClient != nil && vcdClient.Client.APIVCDMaxVersionIs("< 37.0") {
		params["LicenseType"] = "license_type = \"ENTERPRISE\""
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdNsxtAlbImportableCloud, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdAlbControllerDestroy("vcd_nsxt_alb_controller.first"),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("data.vcd_nsxt_alb_importable_cloud.cld", "id", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("data.vcd_nsxt_alb_importable_cloud.cld", "controller_id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("data.vcd_nsxt_alb_importable_cloud.cld", "name", testConfig.Nsxt.NsxtAlbImportableCloud),

					resource.TestCheckResourceAttr("data.vcd_nsxt_alb_importable_cloud.cld", "already_imported", "false"),
					resource.TestMatchResourceAttr("data.vcd_nsxt_alb_importable_cloud.cld", "network_pool_name", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("data.vcd_nsxt_alb_importable_cloud.cld", "network_pool_id", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("data.vcd_nsxt_alb_importable_cloud.cld", "transport_zone_name", regexp.MustCompile(`\d*`)),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtAlbImportableCloudPrereqs = `
resource "vcd_nsxt_alb_controller" "first" {
  name         = "{{.ControllerName}}"
  description  = "first alb controller"
  url          = "{{.ControllerUrl}}"
  username     = "{{.ControllerUsername}}"
  password     = "{{.ControllerPassword}}"
  {{.LicenseType}}
}
`
const testAccVcdNsxtAlbImportableCloud = testAccVcdNsxtAlbImportableCloudPrereqs + `
# skip-binary-test: Data Source test
data "vcd_nsxt_alb_importable_cloud" "cld" {
  name          = "{{.ImportableCloud}}"
  controller_id = vcd_nsxt_alb_controller.first.id
}
`
