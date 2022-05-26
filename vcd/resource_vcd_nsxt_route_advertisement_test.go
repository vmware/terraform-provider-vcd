package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func TestAccVcdNsxtRouteAdvertisement(t *testing.T) {
	preTestChecks(t)
	skipNoNsxtConfiguration(t)

	subnet1 := "192.168.1.0/24"
	subnet2 := "172.16.0.0/24"

	// String map to fill the template
	var params = StringMap{
		"Org":                            testConfig.VCD.Org,
		"NsxtVdc":                        testConfig.Nsxt.Vdc,
		"EdgeGw":                         testConfig.Nsxt.EdgeGateway,
		"RouteAdvertisementResourceName": t.Name(),
		"Enabled":                        "true",
		"Subnet1Cidr":                    subnet1,
		"Subnet2Cidr":                    subnet2,
	}

	configText1 := templateFill(testAccNsxtRouteAdvertisementCreation, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccNsxtRouteAdvertisementUpdate, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckNsxtRouteAdvertisement(testConfig.Nsxt.Vdc, testConfig.Nsxt.EdgeGateway),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
			{
				Config: configText2,
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
			// Import!!!!
		},
	})
}

const testAccNsxtRouteAdvertisementCreation = `
data "vcd_org_vdc" "{{.NsxtVdc}}" {
  org = "{{.Org}}"
  name = "{{.NsxtVdc}}"
}

data "vcd_nsxt_edgegateway" "{{.EdgeGw}}" {
  owner_id = data.vcd_org_vdc.{{.NsxtVdc}}.id
  name     = "{{.EdgeGw}}"
}

resource "vcd_nsxt_route_advertisement" "{{.RouteAdvertisementResourceName}}" {
  edge_gateway_id = data.vcd_nsxt_edgegateway.{{.EdgeGw}}.id
  enabled = {{.Enabled}}
  subnets = [{{.Subnet1Cidr}}]
}
`

const testAccNsxtRouteAdvertisementUpdate = `
data "vcd_org_vdc" "{{.NsxtVdc}}" {
  org = "{{.Org}}"
  name = "{{.NsxtVdc}}"
}

data "vcd_nsxt_edgegateway" "{{.EdgeGw}}" {
  owner_id = data.vcd_org_vdc.{{.NsxtVdc}}.id
  name     = "{{.EdgeGw}}"
}

resource "vcd_nsxt_route_advertisement" "{{.RouteAdvertisementResourceName}}" {
  edge_gateway_id = data.vcd_nsxt_edgegateway.{{.EdgeGw}}.id
  enabled = {{.Enabled}}
  subnets = [{{.Subnet1Cidr}}, {{.Subnet2Cidr}}]
}
`

func testAccCheckNsxtRouteAdvertisement(vdcName, edgeGatewayName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// TBD
		return nil
	}
}
