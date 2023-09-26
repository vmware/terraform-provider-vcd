//go:build network || nsxt || ALL || functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdNsxtEdgeL2VpnTunnel(t *testing.T) {
	preTestChecks(t)

	skipIfNotSysAdmin(t)

	// String map to fill the template
	var params = StringMap{
		"Org":                     testConfig.VCD.Org,
		"NsxtVdc":                 testConfig.Nsxt.Vdc,
		"EdgeGw":                  testConfig.Nsxt.EdgeGateway,
		"NetworkName":             testConfig.Nsxt.RoutedNetwork,
		"TunnelName":              t.Name(),
		"TunnelInterface":         "192.168.0.1/24",
		"RemoteEndpointIp":        "1.2.3.4",
		"TunnelInterfaceUpdated":  "192.169.0.1/22",
		"RemoteEndpointIpUpdated": "4.3.2.1",
		"PreSharedKey":            t.Name(),
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccVcdNsxtEdgegatewayL2VpnTunnelStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s\n", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdNsxtEdgegatewayL2VpnTunnelStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s\n", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceName := "vcd_nsxt_edgegateway_l2_vpn_tunnel." + params["TunnelName"].(string)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNsxtEdgeL2VpnTunnelDestroy(testConfig.Nsxt.Vdc, testConfig.Nsxt.EdgeGateway, t.Name()),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceName, "description", t.Name()),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "session_mode", "SERVER"),
					resource.TestCheckResourceAttr(resourceName, "remote_endpoint_ip", params["RemoteEndpointIp"].(string)),
					resource.TestCheckResourceAttr(resourceName, "tunnel_interface", params["TunnelInterface"].(string)),
					resource.TestCheckResourceAttr(resourceName, "stretched_network.#", "0"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceName, "description", t.Name()),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "session_mode", "SERVER"),
					resource.TestCheckResourceAttr(resourceName, "remote_endpoint_ip", params["RemoteEndpointIpUpdated"].(string)),
					resource.TestCheckResourceAttr(resourceName, "tunnel_interface", params["TunnelInterfaceUpdated"].(string)),
					resource.TestCheckResourceAttr(resourceName, "stretched_network.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdNsxtEdgeGatewayObject(params["EdgeGw"].(string), t.Name()),
				ImportStateVerifyIgnore: []string{"org"},
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtEdgegatewayL2VpnTunnelData = `
data "vcd_org_vdc" "{{.NsxtVdc}}" {
  name = "{{.NsxtVdc}}"		
}
	
data "vcd_nsxt_edgegateway" "{{.EdgeGw}}" {
  owner_id = data.vcd_org_vdc.{{.NsxtVdc}}.id
  name     = "{{.EdgeGw}}"
}

data "vcd_network_routed_v2" "{{.NetworkName}}" {
  edge_gateway_id = data.vcd_nsxt_edgegateway.{{.EdgeGw}}.id
  name            = "{{.NetworkName}}"
}
`

const testAccVcdNsxtEdgegatewayL2VpnTunnelStep1 = testAccVcdNsxtEdgegatewayL2VpnTunnelData + `
resource "vcd_nsxt_edgegateway_l2_vpn_tunnel" "{{.TunnelName}}" {
  name        = "{{.TunnelName}}"
  description = "{{.TunnelName}}"

  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.{{.EdgeGw}}.id

  session_mode              = "SERVER"
  enabled                   = true
  connector_initiation_mode = "INITIATOR"

  local_endpoint_ip  = data.vcd_nsxt_edgegateway.{{.EdgeGw}}.primary_ip
  tunnel_interface   = "{{.TunnelInterface}}"
  remote_endpoint_ip = "{{.RemoteEndpointIp}}"

  pre_shared_key = "{{.PreSharedKey}}"
}
`

const testAccVcdNsxtEdgegatewayL2VpnTunnelStep2 = testAccVcdNsxtEdgegatewayL2VpnTunnelData + `
resource "vcd_nsxt_edgegateway_l2_vpn_tunnel" "{{.TunnelName}}" {
  name        = "{{.TunnelName}}"
  description = "{{.TunnelName}}"

  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.{{.EdgeGw}}.id

  session_mode              = "SERVER"
  enabled                   = false
  connector_initiation_mode = "ON_DEMAND"

  local_endpoint_ip  = data.vcd_nsxt_edgegateway.{{.EdgeGw}}.primary_ip
  tunnel_interface   = "{{.TunnelInterfaceUpdated}}"
  remote_endpoint_ip = "{{.RemoteEndpointIpUpdated}}"

  pre_shared_key = "{{.PreSharedKey}}"

  stretched_network {
    network_id = data.vcd_network_routed_v2.{{.NetworkName}}.id
  }
}
`

func testAccCheckNsxtEdgeL2VpnTunnelDestroy(vdcOrVdcGroupName, edgeGatewayName, l2VpnTunnelName string) resource.TestCheckFunc {
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

		_, err = edge.GetL2VpnTunnelByName(l2VpnTunnelName)
		if govcd.ContainsNotFound(err) {
			return nil
		}

		return fmt.Errorf("tunnel still exists")
	}
}
