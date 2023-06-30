//go:build network || nsxt || ALL || functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdNsxtEdgeStaticRoute(t *testing.T) {
	preTestChecks(t)

	// Binary tests cannot be run for this test because it requires dedicated Tier-0 gateway which
	// is enabled using custom SDK functions
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// Requires VCD 10.4.0+
	vcdClient := createTemporaryVCDConnection(true)
	if vcdClient == nil {
		t.Skipf(t.Name() + " requires a connection to set the tests")
	}

	if vcdClient.Client.APIVCDMaxVersionIs("< 37.0") {
		t.Skipf("NSX-T Edge Gateway Static Routing requires VCD 10.4.0+ (API v37.0+)")
	}

	// Ensure Edge Gateway has a dedicated Tier 0 gateway (External network) as Static Route
	// configuration requires it. Restore it right after the test so that other tests are not
	// impacted.
	updateEdgeGatewayTier0Dedication(t, true)
	defer updateEdgeGatewayTier0Dedication(t, false)

	// String map to fill the template
	var params = StringMap{
		"TestName":      t.Name(),
		"Org":           testConfig.VCD.Org,
		"NsxtVdc":       testConfig.Nsxt.Vdc,
		"EdgeGw":        testConfig.Nsxt.EdgeGateway,
		"RoutedNetName": testConfig.Nsxt.RoutedNetwork,

		"Tags": "network nsxt",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccVcdNsxtEdgegatewayStaticRouteStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s\n", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdNsxtEdgegatewayStaticRouteStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s\n", configText2)

	params["FuncName"] = t.Name() + "step3"
	configText3DS := templateFill(testAccVcdNsxtEdgegatewayStaticRouteStep3DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s\n", configText3DS)

	params["FuncName"] = t.Name() + "step4"
	configText4 := templateFill(testAccVcdNsxtEdgegatewayStaticRouteStep4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s\n", configText4)

	params["FuncName"] = t.Name() + "step5"
	configText5DS := templateFill(testAccVcdNsxtEdgegatewayStaticRouteStep5DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s\n", configText5DS)

	routedNetGw := &testCachedFieldValue{}
	routedNetId := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNsxtEdgeStaticRouteDestroy(testConfig.Nsxt.Vdc, testConfig.Nsxt.EdgeGateway),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_static_route.sr1", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_static_route.sr1", "name", t.Name()+"-1"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_static_route.sr1", "description", "description-field"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_static_route.sr1", "network_cidr", "10.10.11.0/24"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_static_route.sr1", "next_hop.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway_static_route.sr1", "next_hop.*", map[string]string{
						"ip_address":     "4.3.2.1",
						"admin_distance": "1",
					}),

					routedNetGw.cacheTestResourceFieldValue("data.vcd_network_routed_v2.net", "gateway"),
					routedNetId.cacheTestResourceFieldValue("data.vcd_network_routed_v2.net", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_static_route.sr2", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_static_route.sr2", "name", t.Name()+"-2"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_static_route.sr2", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_static_route.sr2", "network_cidr", "192.168.1.0/24"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_static_route.sr2", "next_hop.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway_static_route.sr2", "next_hop.*", map[string]string{
						"ip_address":     routedNetGw.fieldValue,
						"admin_distance": "4",
						"scope.#":        "1",
						"scope.0.type":   "NETWORK",
						"scope.0.id":     routedNetId.fieldValue,
						"scope.0.name":   params["RoutedNetName"].(string),
					}),

					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway_static_route.sr2", "next_hop.*", map[string]string{
						"ip_address":     routedNetGw.fieldValue,
						"admin_distance": "3",
						"scope.#":        "1",
						"scope.0.type":   "NETWORK",
						"scope.0.id":     routedNetId.fieldValue,
						"scope.0.name":   params["RoutedNetName"].(string),
					}),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_static_route.sr1", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_static_route.sr1", "name", t.Name()+"-1"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_static_route.sr1", "description", "description-field-updated"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_static_route.sr1", "network_cidr", "10.10.11.0/24"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_static_route.sr1", "next_hop.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway_static_route.sr1", "next_hop.*", map[string]string{
						"ip_address":     "1.2.3.4",
						"admin_distance": "5",
					}),

					routedNetGw.cacheTestResourceFieldValue("data.vcd_network_routed_v2.net", "gateway"),
					routedNetId.cacheTestResourceFieldValue("data.vcd_network_routed_v2.net", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_static_route.sr2", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_static_route.sr2", "name", t.Name()+"-2"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_static_route.sr2", "description", "description-field"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_static_route.sr2", "network_cidr", "192.168.1.0/24"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_static_route.sr2", "next_hop.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway_static_route.sr2", "next_hop.*", map[string]string{
						"ip_address":     routedNetGw.fieldValue,
						"admin_distance": "2",
						"scope.#":        "1",
						"scope.0.type":   "NETWORK",
						"scope.0.id":     routedNetId.fieldValue,
						"scope.0.name":   params["RoutedNetName"].(string),
					}),
				),
			},
			{ // Import by Name
				ResourceName:      "vcd_nsxt_edgegateway_static_route.sr1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdNsxtEdgeGatewayObject(testConfig.Nsxt.EdgeGateway, t.Name()+"-1"),
			},
			{ // Import by Network CIDR
				ResourceName:      "vcd_nsxt_edgegateway_static_route.sr1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdNsxtEdgeGatewayObject(testConfig.Nsxt.EdgeGateway, "10.10.11.0/24"),
			},
			{
				Config: configText3DS,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_nsxt_edgegateway_static_route.by-name", "vcd_nsxt_edgegateway_static_route.sr1", nil),
					resourceFieldsEqual("data.vcd_nsxt_edgegateway_static_route.by-name-and-cidr", "vcd_nsxt_edgegateway_static_route.sr1", nil),
				),
			},
			{
				Config: configText4, // check that 2 static routes can have duplicate names
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_static_route.sr1", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_static_route.sr1", "name", t.Name()),

					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_static_route.sr2", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_static_route.sr2", "name", t.Name()),
				),
			},
			{
				Config: configText5DS, // validate that a single data source can be identified with name+network_cidr filtering
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_nsxt_edgegateway_static_route.by-name-and-cidr", "vcd_nsxt_edgegateway_static_route.sr1", nil),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtEdgegatewayStaticRoutePrereqs = `
data "vcd_org_vdc" "{{.NsxtVdc}}" {
  name = "{{.NsxtVdc}}"		
}
	
data "vcd_nsxt_edgegateway" "existing" {
  owner_id = data.vcd_org_vdc.{{.NsxtVdc}}.id
  name     = "{{.EdgeGw}}"
}

data "vcd_network_routed_v2" "net" {
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  name            = "{{.RoutedNetName}}"
}
`

const testAccVcdNsxtEdgegatewayStaticRouteStep1 = testAccVcdNsxtEdgegatewayStaticRoutePrereqs + `
resource "vcd_nsxt_edgegateway_static_route" "sr1" {
  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  name         = "{{.TestName}}-1"
  description  = "description-field"
  network_cidr = "10.10.11.0/24"

  next_hop {
	ip_address     = "4.3.2.1"
	admin_distance = 1
  }
}

resource "vcd_nsxt_edgegateway_static_route" "sr2" {
  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  name         = "{{.TestName}}-2"
  network_cidr = "192.168.1.0/24"

  next_hop {
    ip_address     = data.vcd_network_routed_v2.net.gateway
    admin_distance = 4

	scope {
	  id   = data.vcd_network_routed_v2.net.id
	  type = "NETWORK"
	}
  }

  next_hop {
    ip_address     = cidrhost(format("%s/%s",data.vcd_network_routed_v2.net.gateway,data.vcd_network_routed_v2.net.prefix_length),4)
    admin_distance = 3

	scope {
	  id   = data.vcd_network_routed_v2.net.id
	  type = "NETWORK"
	}
  }
}
`

const testAccVcdNsxtEdgegatewayStaticRouteStep2 = testAccVcdNsxtEdgegatewayStaticRoutePrereqs + `
resource "vcd_nsxt_edgegateway_static_route" "sr1" {
  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  name         = "{{.TestName}}-1"
  description  = "description-field-updated"
  network_cidr = "10.10.11.0/24"

  next_hop {
	ip_address     = "1.2.3.4"
	admin_distance = 5
  }
}

resource "vcd_nsxt_edgegateway_static_route" "sr2" {
  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  name         = "{{.TestName}}-2"
  description  = "description-field"
  network_cidr = "192.168.1.0/24"

  next_hop {
    ip_address     = data.vcd_network_routed_v2.net.gateway
    admin_distance = 2

	scope {
	  id   = data.vcd_network_routed_v2.net.id
	  type = "NETWORK"
	}
  }
}
`

const testAccVcdNsxtEdgegatewayStaticRouteStep3DS = testAccVcdNsxtEdgegatewayStaticRouteStep2 + `
# skip-binary-test: Data Source test
data "vcd_nsxt_edgegateway_static_route" "by-name" {
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  name            = "{{.TestName}}-1"
}

data "vcd_nsxt_edgegateway_static_route" "by-name-and-cidr" {
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  name            = "{{.TestName}}-1"
  network_cidr    = "10.10.11.0/24"
}
`

// create Static Routes with duplicate names to check that data source can filter on network_cidr
const testAccVcdNsxtEdgegatewayStaticRouteStep4 = testAccVcdNsxtEdgegatewayStaticRoutePrereqs + `
resource "vcd_nsxt_edgegateway_static_route" "sr1" {
  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  name         = "{{.TestName}}"
  description  = "description-field-updated"
  network_cidr = "10.10.11.0/24"

  next_hop {
	ip_address     = "1.2.3.4"
	admin_distance = 5
  }
}

resource "vcd_nsxt_edgegateway_static_route" "sr2" {
  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  name         = "{{.TestName}}"
  description  = "description-field"
  network_cidr = "192.168.1.0/24"

  next_hop {
    ip_address     = data.vcd_network_routed_v2.net.gateway
    admin_distance = 2

	scope {
	  id   = data.vcd_network_routed_v2.net.id
	  type = "NETWORK"
	}
  }
}
`

const testAccVcdNsxtEdgegatewayStaticRouteStep5DS = testAccVcdNsxtEdgegatewayStaticRouteStep4 + `
# skip-binary-test: Data Source test
# This data source would fail because it attemps to search only by name and there are 2 Static 
# Routes with the same name
#data "vcd_nsxt_edgegateway_static_route" "by-name" {
#  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
#  name            = "{{.TestName}}"
#}

data "vcd_nsxt_edgegateway_static_route" "by-name-and-cidr" {
  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  name            = "{{.TestName}}"
  network_cidr    = "10.10.11.0/24"
}
`

func testAccCheckNsxtEdgeStaticRouteDestroy(vdcOrVdcGroupName, edgeGatewayName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		vdcOrVdcGroup, err := lookupVdcOrVdcGroup(conn, testConfig.VCD.Org, vdcOrVdcGroupName)
		if err != nil {
			return fmt.Errorf("unable to find VDC or VDC group %s: %s", vdcOrVdcGroupName, err)
		}

		edge, err := vdcOrVdcGroup.GetNsxtEdgeGatewayByName(edgeGatewayName)
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, edgeGatewayName)
		}

		allStaticRoutes, err := edge.GetAllStaticRoutes(nil)
		if err != nil {
			return fmt.Errorf("unable to get Static Routes: %s", err)
		}

		if len(allStaticRoutes) > 0 {
			return fmt.Errorf("'%d' Static Routes still exist", len(allStaticRoutes))
		}

		return nil
	}
}
