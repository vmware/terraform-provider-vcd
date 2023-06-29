//go:build network || nsxt || ALL || functional

package vcd

import (
	"fmt"
	"testing"
	"time"

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

	// params["FuncName"] = t.Name() + "step2"
	// configText2 := templateFill(testAccVcdNsxtEdgegatewayDhcpForwardingStep2, params)
	// debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s\n", configText2)

	// params["FuncName"] = t.Name() + "step3"
	// configText3 := templateFill(testAccVcdNsxtEdgegatewayDhcpForwardingStep3, params)
	// debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s\n", configText3)

	// params["FuncName"] = t.Name() + "step4"
	// configText4 := templateFill(testAccVcdNsxtEdgegatewayDhcpForwardingStep4, params)
	// debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s\n", configText4)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNsxtEdgeDhcpForwardDestroy(testConfig.Nsxt.Vdc, testConfig.Nsxt.EdgeGateway),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_static_route.sr1", "id"),
					// resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_dhcp_forwarding.DhcpForwarding", "enabled", "true"),
					// resource.TestCheckTypeSetElemAttr("vcd_nsxt_edgegateway_dhcp_forwarding.DhcpForwarding", "dhcp_servers.*", "1.2.3.4"),
					// resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_dhcp_forwarding.DhcpForwarding", "dhcp_servers.#", "1"),
					// sleepTester(2*time.Minute),
				),
			},
			// {
			// 	Config: configText2,
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_dhcp_forwarding.DhcpForwarding", "enabled", "true"),
			// 		resource.TestCheckTypeSetElemAttr("vcd_nsxt_edgegateway_dhcp_forwarding.DhcpForwarding", "dhcp_servers.*", "1.2.3.4"),
			// 		resource.TestCheckTypeSetElemAttr("vcd_nsxt_edgegateway_dhcp_forwarding.DhcpForwarding", "dhcp_servers.*", "fe80::aaaa"),
			// 		resource.TestCheckTypeSetElemAttr("vcd_nsxt_edgegateway_dhcp_forwarding.DhcpForwarding", "dhcp_servers.*", "192.168.1.254"),
			// 		resource.TestCheckTypeSetElemAttr("vcd_nsxt_edgegateway_dhcp_forwarding.DhcpForwarding", "dhcp_servers.*", "0.0.0.0"),
			// 		resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_dhcp_forwarding.DhcpForwarding", "dhcp_servers.#", "4"),
			// 	),
			// },
			// {
			// 	Config: configText3,
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_dhcp_forwarding.DhcpForwarding", "enabled", "false"),
			// 		resource.TestCheckTypeSetElemAttr("vcd_nsxt_edgegateway_dhcp_forwarding.DhcpForwarding", "dhcp_servers.*", "1.2.3.4"),
			// 		resource.TestCheckTypeSetElemAttr("vcd_nsxt_edgegateway_dhcp_forwarding.DhcpForwarding", "dhcp_servers.*", "fe80::aaaa"),
			// 		resource.TestCheckTypeSetElemAttr("vcd_nsxt_edgegateway_dhcp_forwarding.DhcpForwarding", "dhcp_servers.*", "192.168.1.254"),
			// 		resource.TestCheckTypeSetElemAttr("vcd_nsxt_edgegateway_dhcp_forwarding.DhcpForwarding", "dhcp_servers.*", "0.0.0.0"),
			// 		resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_dhcp_forwarding.DhcpForwarding", "dhcp_servers.#", "4"),
			// 	),
			// },
			// {
			// 	Config: configText4,
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_dhcp_forwarding.DhcpForwarding", "enabled", "true"),
			// 		resource.TestCheckTypeSetElemAttr("vcd_nsxt_edgegateway_dhcp_forwarding.DhcpForwarding", "dhcp_servers.*", "1.2.3.4"),
			// 		resource.TestCheckTypeSetElemAttr("vcd_nsxt_edgegateway_dhcp_forwarding.DhcpForwarding", "dhcp_servers.*", "fe80::aaaa"),
			// 		// This is left on purpose, as right now if the forwarding service is disabled,
			// 		// IP addresses can't be deleted, if this fails, it means that the bug got fixed
			// 		resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_dhcp_forwarding.DhcpForwarding", "dhcp_servers.#", "2"),
			// 	),
			// },
			// {
			// 	ResourceName:            "vcd_nsxt_edgegateway_dhcp_forwarding.DhcpForwarding",
			// 	ImportState:             true,
			// 	ImportStateVerify:       true,
			// 	ImportStateIdFunc:       importStateIdOrgNsxtVdcObject(params["EdgeGw"].(string)),
			// 	ImportStateVerifyIgnore: []string{"org"},
			// },
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
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  name         = "{{.TestName}}-2"
  description  = "description-field"
  network_cidr = "192.168.1.0/24"

  next_hop {
	ip_address     = data.vcd_network_routed_v2.net.gateway
	admin_distance = 4

	scope {
		id   = data.vcd_network_routed_v2.net.id
		type = "NETWORK"
	}
  }
}
`

// const testAccVcdNsxtEdgegatewayDhcpForwardingStep2 = testAccVcdNsxtEdgegatewayDhcpForwardingData + `
// resource "vcd_nsxt_edgegateway_dhcp_forwarding" "DhcpForwarding" {
//   edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
//   enabled      = true
//   dhcp_servers = [
//     "1.2.3.4",
//     "fe80::aaaa",
//     "192.168.1.254",
//     "0.0.0.0",
//   ]
// }
// `

// const testAccVcdNsxtEdgegatewayDhcpForwardingStep3 = testAccVcdNsxtEdgegatewayDhcpForwardingData + `
// resource "vcd_nsxt_edgegateway_dhcp_forwarding" "DhcpForwarding" {
//   edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
//   enabled      = false
//   dhcp_servers = [
//     "1.2.3.4",
//     "fe80::aaaa",
//     "192.168.1.254",
//     "0.0.0.0",
//   ]
// }
// `

// const testAccVcdNsxtEdgegatewayDhcpForwardingStep4 = testAccVcdNsxtEdgegatewayDhcpForwardingData + `
// resource "vcd_nsxt_edgegateway_dhcp_forwarding" "DhcpForwarding" {
//   edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

//   enabled      = true
//   dhcp_servers = [
//     "1.2.3.4",
//     "fe80::aaaa",
//   ]
// }
// `

// func testAccCheckNsxtEdgeDhcpForwardDestroy(vdcOrVdcGroupName, edgeGatewayName string) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		conn := testAccProvider.Meta().(*VCDClient)

// 		vdcOrVdcGroup, err := lookupVdcOrVdcGroup(conn, testConfig.VCD.Org, vdcOrVdcGroupName)
// 		if err != nil {
// 			return fmt.Errorf("unable to find VDC or VDC group %s: %s", vdcOrVdcGroupName, err)
// 		}

// 		edge, err := vdcOrVdcGroup.GetNsxtEdgeGatewayByName(edgeGatewayName)
// 		if err != nil {
// 			return fmt.Errorf(errorUnableToFindEdgeGateway, edgeGatewayName)
// 		}

// 		dhcpForwardingConfig, err := edge.GetDhcpForwarder()
// 		if err != nil {
// 			return fmt.Errorf("unable to get DHCP forwarding config: %s", err)
// 		}

// 		if dhcpForwardingConfig.Enabled && dhcpForwardingConfig.DhcpServers != nil {
// 			return fmt.Errorf("DHCP forwarding configuration still exists")
// 		}

// 		return nil
// 	}
// }

func sleepTester(d time.Duration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fmt.Printf("sleeping %s\n", d.String())
		time.Sleep(d)
		fmt.Println("finished sleeping")
		return nil
	}
}
