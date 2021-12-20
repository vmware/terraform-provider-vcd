//go:build nsxt || alb || ALL || functional
// +build nsxt alb ALL functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdNsxtAlbSettingsDS assumes that NSX-T ALB is not configured and General Settings shows "Inactive"
func TestAccVcdNsxtAlbSettingsDS(t *testing.T) {
	preTestChecks(t)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	vcdClient := createTemporaryVCDConnection(false)
	if vcdClient.Client.APIVCDMaxVersionIs("< 35.0") {
		t.Skip(t.Name() + " requires at least API v35.0 (VCD 10.2+)")
	}
	skipNoNsxtAlbConfiguration(t)

	// String map to fill the template
	var params = StringMap{
		"Org":     testConfig.VCD.Org,
		"NsxtVdc": testConfig.Nsxt.Vdc,
		"EdgeGw":  testConfig.Nsxt.EdgeGateway,
		"Tags":    "nsxt alb",
	}

	configText1 := templateFill(testAccVcdNsxtAlbSettingsDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },

		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.vcd_nsxt_alb_settings.test", "is_active", "false"),
					resource.TestCheckResourceAttr("data.vcd_nsxt_alb_settings.test", "service_network_specification", ""),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtAlbSettingsDS = `
data "vcd_nsxt_edgegateway" "existing" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  name = "{{.EdgeGw}}"
}

data "vcd_nsxt_alb_settings" "test" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
}
`
