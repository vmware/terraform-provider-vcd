//go:build gateway || network || nsxt || ALL || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdDataSourceNsxtEdgeRateLimiting is a test for datasource
// vcd_nsxt_edgegateway_rate_limiting It only check if ingress and egress profile IDs are empty
// ("unlimited" rate). Other values are tested in resource test.
func TestAccVcdDataSourceNsxtEdgeDhcpForwarding(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	vcdClient := createTemporaryVCDConnection(true)
	if vcdClient == nil {
		t.Skip(acceptanceTestsSkipped)
	}
	if vcdClient.Client.APIVCDMaxVersionIs("< 36.1") {
		t.Skipf("This test tests VCD 10.3.1+ (API V36.1+) features. Skipping.")
	}

	// String map to fill the template
	var params = StringMap{
		"Org":                  testConfig.VCD.Org,
		"NsxtVdc":              testConfig.Nsxt.Vdc,
		"NsxtVdcGroup":         testConfig.Nsxt.VdcGroup,
		"NsxtEdgeGwInVdcGroup": testConfig.Nsxt.VdcGroupEdgeGateway,
		"NsxtEdgeGw":           testConfig.Nsxt.EdgeGateway,
		"TestName":             t.Name(),
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdDataSourceNsxtEdgeDhcpForwarding, params)
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
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_edgegateway_dhcp_forwarding.testing-in-vdc", "id"),
					resource.TestCheckResourceAttr("data.vcd_nsxt_edgegateway_dhcp_forwarding.testing-in-vdc", "enabled", "false"),

					resource.TestCheckResourceAttrSet("data.vcd_nsxt_edgegateway_dhcp_forwarding.testing-in-vdc-group", "id"),
					resource.TestCheckResourceAttr("data.vcd_nsxt_edgegateway_dhcp_forwarding.testing-in-vdc-group", "enabled", "false"),
				),
			},
		},
	})
}

const testAccVcdDataSourceNsxtEdgeDhcpForwarding = `
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

data "vcd_nsxt_edgegateway_dhcp_forwarding" "testing-in-vdc" {
  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.testing-in-vdc.id
}

data "vcd_nsxt_edgegateway_dhcp_forwarding" "testing-in-vdc-group" {
  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.testing-in-vdc-group.id
}
`
