//go:build gateway || network || nsxt || ALL || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdDataSourceNsxtEdgeRateLimiting is a test for datasource
// vcd_nsxt_edgegateway_rate_limiting It only check if ingress and egress profile IDs are empty
// ("unlimited" rate). Other values are tested in resource test.
func TestAccVcdDataSourceNsxtEdgeRateLimiting(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	// String map to fill the template
	var params = StringMap{
		"Org":                  testConfig.VCD.Org,
		"NsxtVdc":              testConfig.Nsxt.Vdc,
		"NsxtVdcGroup":         testConfig.Nsxt.VdcGroup,
		"NsxtEdgeGwInVdcGroup": testConfig.Nsxt.VdcGroupEdgeGateway,
		"NsxtEdgeGw":           testConfig.Nsxt.EdgeGateway,
		"TestName":             t.Name(),
		"NsxtManager":          testConfig.Nsxt.Manager,

		"Tags": "network nsxt",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdDataSourceNsxtEdgeRateLimitingStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_edgegateway_rate_limiting.testing-in-vdc", "id"),
					resource.TestCheckResourceAttr("data.vcd_nsxt_edgegateway_rate_limiting.testing-in-vdc", "ingress_profile_id", ""),
					resource.TestCheckResourceAttr("data.vcd_nsxt_edgegateway_rate_limiting.testing-in-vdc", "egress_profile_id", ""),

					resource.TestCheckResourceAttrSet("data.vcd_nsxt_edgegateway_rate_limiting.testing-in-vdc-group", "id"),
					resource.TestCheckResourceAttr("data.vcd_nsxt_edgegateway_rate_limiting.testing-in-vdc-group", "ingress_profile_id", ""),
					resource.TestCheckResourceAttr("data.vcd_nsxt_edgegateway_rate_limiting.testing-in-vdc-group", "egress_profile_id", ""),
				),
			},
		},
	})
}

const testAccVcdDataSourceNsxtEdgeRateLimitingStep1 = `
data "vcd_vdc_group" "g1" {
  org  = "{{.Org}}"
  name = "{{.NsxtVdcGroup}}"
}

data "vcd_nsxt_edgegateway" "testing-in-vdc-group" {
  org      = "{{.Org}}"
  owner_id = data.vcd_vdc_group.g1.id

  name = "{{.NsxtEdgeGwInVdcGroup}}"
}

data "vcd_org_vdc" "v1" {
  org  = "{{.Org}}"
  name = "{{.NsxtVdc}}"
}

data "vcd_nsxt_edgegateway" "testing-in-vdc" {
  org      = "{{.Org}}"
  owner_id = data.vcd_org_vdc.v1.id

  name = "{{.NsxtEdgeGw}}"
}

data "vcd_nsxt_edgegateway_rate_limiting" "testing-in-vdc" {
  org               = "{{.Org}}"
  edge_gateway_id   = data.vcd_nsxt_edgegateway.testing-in-vdc.id
}

data "vcd_nsxt_edgegateway_rate_limiting" "testing-in-vdc-group" {
  org               = "{{.Org}}"
  edge_gateway_id   = data.vcd_nsxt_edgegateway.testing-in-vdc-group.id
}
`
