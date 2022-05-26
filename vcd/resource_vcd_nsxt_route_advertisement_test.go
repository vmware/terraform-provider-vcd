package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"regexp"
	"strconv"
	"testing"
)

func TestAccVcdNsxtRouteAdvertisement(t *testing.T) {
	preTestChecks(t)
	skipNoNsxtConfiguration(t)

	isRouteAdvertisementEnable := true
	subnet1 := "192.168.1.0/24"
	subnet2 := "192.168.2.0/24"

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"NsxtVdc":     testConfig.Nsxt.Vdc,
		"EdgeGw":      testConfig.Nsxt.EdgeGateway,
		"Enabled":     strconv.FormatBool(isRouteAdvertisementEnable),
		"Subnet1Cidr": subnet1,
		"Subnet2Cidr": subnet2,
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
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_route_advertisement.testing", "id", regexp.MustCompile(`^urn:vcloud:gateway:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_route_advertisement.testing", "enabled", strconv.FormatBool(isRouteAdvertisementEnable)),
					resource.TestCheckResourceAttr("vcd_nsxt_route_advertisement.testing", "subnets.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_route_advertisement.testing", "subnets.0", subnet1),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_route_advertisement.testing", "id", regexp.MustCompile(`^urn:vcloud:gateway:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_route_advertisement.testing", "enabled", strconv.FormatBool(isRouteAdvertisementEnable)),
					resource.TestCheckResourceAttr("vcd_nsxt_route_advertisement.testing", "subnets.#", "2"),
					resource.TestMatchResourceAttr("vcd_nsxt_route_advertisement.testing", "subnets.0", regexp.MustCompile(`^192.168.[1-2].0/24$`)),
					resource.TestMatchResourceAttr("vcd_nsxt_route_advertisement.testing", "subnets.1", regexp.MustCompile(`^192.168.[1-2].0/24$`)),
				),
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

resource "vcd_nsxt_route_advertisement" "testing" {
  edge_gateway_id = data.vcd_nsxt_edgegateway.{{.EdgeGw}}.id
  enabled = {{.Enabled}}
  subnets = ["{{.Subnet1Cidr}}"]
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

resource "vcd_nsxt_route_advertisement" "testing" {
  edge_gateway_id = data.vcd_nsxt_edgegateway.{{.EdgeGw}}.id
  enabled = {{.Enabled}}
  subnets = ["{{.Subnet1Cidr}}", "{{.Subnet2Cidr}}"]
}
`

func testAccCheckNsxtRouteAdvertisement(vdcName, edgeGatewayName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, vdcName)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, vdcName, testConfig.VCD.Org, err)
		}

		edge, err := vdc.GetNsxtEdgeGatewayByName(edgeGatewayName)
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, edgeGatewayName)
		}

		routeAdvertisement, err := edge.GetNsxtRouteAdvertisement(true)
		if err != nil {
			return fmt.Errorf("error trying to retrieve route advertisement - %s", err)
		}

		if routeAdvertisement.Enable {
			return fmt.Errorf("error destroying route advertisement. Wanted routeAdvertisement.Enable false, Got %t", routeAdvertisement.Enable)
		}

		if routeAdvertisement.Subnets != nil && len(routeAdvertisement.Subnets) > 0 {
			return fmt.Errorf("error destroying route advertisement. Wanted 0 routeAdvertisement.Subnets, got %d", len(routeAdvertisement.Subnets))
		}

		return nil
	}
}
