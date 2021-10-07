//go:build nsxt || alb || ALL || functional
// +build nsxt alb ALL functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdNsxtAlbGeneralSettingsDS assumes that NSX-T ALB is not configured and General Settings shows "Inactive"
func TestAccVcdNsxtAlbGeneralSettingsDS(t *testing.T) {
	vcdClient := createTemporaryVCDConnection()
	if vcdClient.Client.APIVCDMaxVersionIs("< 35.0") {
		t.Skip(t.Name() + " requires at least API v35.0 (vCD 10.2+)")
	}
	skipNoNsxtAlbConfiguration(t)

	// String map to fill the template
	var params = StringMap{
		"Org":     testConfig.VCD.Org,
		"NsxtVdc": testConfig.Nsxt.Vdc,
		"EdgeGw":  testConfig.Nsxt.EdgeGateway,
		"Tags":    "nsxt alb",
	}

	configText1 := templateFill(testAccVcdNsxtAlbGeneralSettingsDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },

		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.vcd_nsxt_alb_general_settings.test", "is_active", "false"),
					resource.TestCheckResourceAttr("data.vcd_nsxt_alb_general_settings.test", "service_network_specification", ""),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtAlbGeneralSettingsDS = `
data "vcd_nsxt_edgegateway" "existing" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  name = "{{.EdgeGw}}"
}

data "vcd_nsxt_alb_general_settings" "test" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
}
`
