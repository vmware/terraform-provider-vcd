//go:build network || nsxt || ALL || functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v3/govcd"
)

// TestAccVcdNsxtEdgeL2VpnTunnel tests the functionality of the L2 VPN Tunnel for both SERVER and CLIENT sessions
// It works in the following order:
// Create both SERVER and CLIENT sessions
// Try updating the CLIENT session, we can't update the SERVER session at the
// same time as it alters the peer_code and causes an inconsistent plan
// Remove the CLIENT session and update the SERVER session in different ways.
func TestAccVcdNsxtEdgeL2VpnTunnel(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	// String map to fill the template
	var params = StringMap{
		"Org":                     testConfig.VCD.Org,
		"NsxtVdc":                 testConfig.Nsxt.Vdc,
		"ExtNet":                  testConfig.Nsxt.ExternalNetwork,
		"EdgeGw":                  testConfig.Nsxt.EdgeGateway,
		"NewEdgeGw":               t.Name() + "-gw",
		"NetworkName":             testConfig.Nsxt.RoutedNetwork,
		"NewNetworkName":          testConfig.Nsxt.RoutedNetwork + "-new",
		"ServerTunnelName":        t.Name() + "-server",
		"ClientTunnelName":        t.Name() + "-client",
		"TunnelInterface":         "192.168.0.1/24",
		"RemoteEndpointIp":        "1.2.3.4",
		"TunnelInterfaceUpdated":  "192.169.0.1/22",
		"RemoteEndpointIpUpdated": "4.3.2.1",
		"PreSharedKey":            t.Name(),
		"PreSharedKeyUpdated":     t.Name() + "-update",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccVcdNsxtEdgegatewayL2VpnTunnelStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s\n", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdNsxtEdgegatewayL2VpnTunnelStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s\n", configText2)

	params["FuncName"] = t.Name() + "step3"
	configText3 := templateFill(testAccVcdNsxtEdgegatewayL2VpnTunnelStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s\n", configText3)

	params["FuncName"] = t.Name() + "step4"
	configText4 := templateFill(testAccVcdNsxtEdgegatewayL2VpnTunnelStep4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s\n", configText4)

	params["FuncName"] = t.Name() + "step5"
	configText5 := templateFill(testAccVcdNsxtEdgegatewayL2VpnTunnelStep5, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s\n", configText5)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	serverTunnelName := "vcd_nsxt_edgegateway_l2_vpn_tunnel." + params["ServerTunnelName"].(string)
	clientTunnelName := "vcd_nsxt_edgegateway_l2_vpn_tunnel." + params["ClientTunnelName"].(string)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNsxtEdgeL2VpnTunnelDestroy(testConfig.Nsxt.Vdc, testConfig.Nsxt.EdgeGateway, t.Name()),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(serverTunnelName, "name", params["ServerTunnelName"].(string)),
					resource.TestCheckResourceAttr(serverTunnelName, "description", params["ServerTunnelName"].(string)),
					resource.TestCheckResourceAttr(serverTunnelName, "enabled", "true"),
					resource.TestCheckResourceAttr(serverTunnelName, "session_mode", "SERVER"),
					resource.TestCheckResourceAttr(serverTunnelName, "connector_initiation_mode", "INITIATOR"),
					resource.TestCheckResourceAttr(serverTunnelName, "remote_endpoint_ip", params["RemoteEndpointIp"].(string)),
					resource.TestCheckResourceAttr(serverTunnelName, "tunnel_interface", params["TunnelInterface"].(string)),
					resource.TestCheckResourceAttr(serverTunnelName, "pre_shared_key", params["PreSharedKey"].(string)),
					resource.TestCheckResourceAttr(serverTunnelName, "stretched_network.#", "0"),

					resource.TestCheckResourceAttr(clientTunnelName, "name", params["ClientTunnelName"].(string)),
					resource.TestCheckResourceAttr(clientTunnelName, "description", params["ClientTunnelName"].(string)),
					resource.TestCheckResourceAttr(clientTunnelName, "enabled", "true"),
					resource.TestCheckResourceAttr(clientTunnelName, "session_mode", "CLIENT"),
					resource.TestCheckResourceAttr(clientTunnelName, "remote_endpoint_ip", params["RemoteEndpointIp"].(string)),
					resource.TestCheckResourceAttr(clientTunnelName, "stretched_network.#", "0"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(serverTunnelName, "name", params["ServerTunnelName"].(string)),
					resource.TestCheckResourceAttr(serverTunnelName, "description", params["ServerTunnelName"].(string)),
					resource.TestCheckResourceAttr(serverTunnelName, "enabled", "true"),
					resource.TestCheckResourceAttr(serverTunnelName, "session_mode", "SERVER"),
					resource.TestCheckResourceAttr(serverTunnelName, "connector_initiation_mode", "INITIATOR"),
					resource.TestCheckResourceAttr(serverTunnelName, "remote_endpoint_ip", params["RemoteEndpointIp"].(string)),
					resource.TestCheckResourceAttr(serverTunnelName, "tunnel_interface", params["TunnelInterface"].(string)),
					resource.TestCheckResourceAttr(serverTunnelName, "pre_shared_key", params["PreSharedKey"].(string)),
					resource.TestCheckResourceAttr(serverTunnelName, "stretched_network.#", "0"),

					resource.TestCheckResourceAttr(clientTunnelName, "name", params["ClientTunnelName"].(string)),
					resource.TestCheckResourceAttr(clientTunnelName, "description", params["ClientTunnelName"].(string)),
					resource.TestCheckResourceAttr(clientTunnelName, "enabled", "true"),
					resource.TestCheckResourceAttr(clientTunnelName, "session_mode", "CLIENT"),
					resource.TestCheckResourceAttr(clientTunnelName, "remote_endpoint_ip", params["RemoteEndpointIpUpdated"].(string)),
					resource.TestCheckResourceAttr(clientTunnelName, "stretched_network.#", "1"),
					resource.TestCheckTypeSetElemAttr(clientTunnelName, "stretched_network.*", "1"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(serverTunnelName, "name", params["ServerTunnelName"].(string)),
					resource.TestCheckResourceAttr(serverTunnelName, "description", params["ServerTunnelName"].(string)),
					resource.TestCheckResourceAttr(serverTunnelName, "enabled", "true"),
					resource.TestCheckResourceAttr(serverTunnelName, "session_mode", "SERVER"),
					resource.TestCheckResourceAttr(serverTunnelName, "connector_initiation_mode", "INITIATOR"),
					resource.TestCheckResourceAttr(serverTunnelName, "remote_endpoint_ip", params["RemoteEndpointIp"].(string)),
					resource.TestCheckResourceAttr(serverTunnelName, "tunnel_interface", params["TunnelInterface"].(string)),
					resource.TestCheckResourceAttr(serverTunnelName, "pre_shared_key", params["PreSharedKey"].(string)),
					resource.TestCheckResourceAttr(serverTunnelName, "stretched_network.#", "0"),

					resource.TestCheckResourceAttr(clientTunnelName, "name", params["ClientTunnelName"].(string)),
					resource.TestCheckResourceAttr(clientTunnelName, "description", params["ClientTunnelName"].(string)),
					resource.TestCheckResourceAttr(clientTunnelName, "enabled", "true"),
					resource.TestCheckResourceAttr(clientTunnelName, "session_mode", "CLIENT"),
					resource.TestCheckResourceAttr(clientTunnelName, "remote_endpoint_ip", params["RemoteEndpointIp"].(string)),
					resource.TestCheckResourceAttr(clientTunnelName, "stretched_network.#", "0"),
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(serverTunnelName, "name", params["ServerTunnelName"].(string)),
					resource.TestCheckResourceAttr(serverTunnelName, "description", params["ServerTunnelName"].(string)),
					resource.TestCheckResourceAttr(serverTunnelName, "enabled", "false"),
					resource.TestCheckResourceAttr(serverTunnelName, "session_mode", "SERVER"),
					resource.TestCheckResourceAttr(serverTunnelName, "connector_initiation_mode", "ON_DEMAND"),
					resource.TestCheckResourceAttr(serverTunnelName, "remote_endpoint_ip", params["RemoteEndpointIpUpdated"].(string)),
					resource.TestCheckResourceAttr(serverTunnelName, "tunnel_interface", params["TunnelInterfaceUpdated"].(string)),
					resource.TestCheckResourceAttr(serverTunnelName, "pre_shared_key", params["PreSharedKeyUpdated"].(string)),
					resource.TestCheckResourceAttr(serverTunnelName, "stretched_network.#", "1"),
				),
			},
			{
				Config: configText5,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(serverTunnelName, "name", params["ServerTunnelName"].(string)),
					resource.TestCheckResourceAttr(serverTunnelName, "description", params["ServerTunnelName"].(string)),
					resource.TestCheckResourceAttr(serverTunnelName, "enabled", "true"),
					resource.TestCheckResourceAttr(serverTunnelName, "session_mode", "SERVER"),
					resource.TestCheckResourceAttr(serverTunnelName, "connector_initiation_mode", "INITIATOR"),
					resource.TestCheckResourceAttr(serverTunnelName, "remote_endpoint_ip", params["RemoteEndpointIp"].(string)),
					resource.TestCheckResourceAttr(serverTunnelName, "tunnel_interface", params["TunnelInterface"].(string)),
					resource.TestCheckResourceAttr(serverTunnelName, "pre_shared_key", params["PreSharedKey"].(string)),
					resource.TestCheckResourceAttr(serverTunnelName, "stretched_network.#", "0"),
				),
			},
			{
				ResourceName:            serverTunnelName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdNsxtEdgeGatewayObject(params["EdgeGw"].(string), params["ServerTunnelName"].(string)),
				ImportStateVerifyIgnore: []string{"org", "pre_shared_key"},
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtEdgegatewayL2VpnTunnelData = `
data "vcd_external_network_v2" "{{.ExtNet}}" {
  name = "{{.ExtNet}}"
}

data "vcd_org_vdc" "{{.NsxtVdc}}" {
  name = "{{.NsxtVdc}}"		
}
	
data "vcd_nsxt_edgegateway" "{{.EdgeGw}}" {
  owner_id = data.vcd_org_vdc.{{.NsxtVdc}}.id
  name     = "{{.EdgeGw}}"
}

resource "vcd_nsxt_edgegateway" "{{.NewEdgeGw}}" {
  org             = "{{.Org}}"
  owner_id        = data.vcd_org_vdc.{{.NsxtVdc}}.id
  name            = "{{.NewEdgeGw}}"

  external_network_id = data.vcd_external_network_v2.{{.ExtNet}}.id

  total_allocated_ip_count = 1

  subnet_with_total_ip_count {
     gateway       = tolist(data.vcd_external_network_v2.{{.ExtNet}}.ip_scope)[0].gateway
     prefix_length = tolist(data.vcd_external_network_v2.{{.ExtNet}}.ip_scope)[0].prefix_length
  }
}

resource "vcd_network_routed_v2" "{{.NewNetworkName}}" {
  edge_gateway_id = vcd_nsxt_edgegateway.{{.NewEdgeGw}}.id

  name          = "{{.NewNetworkName}}"
  gateway       = "192.168.1.1"
  prefix_length = 24
}

data "vcd_network_routed_v2" "{{.NetworkName}}" {
  edge_gateway_id = data.vcd_nsxt_edgegateway.{{.EdgeGw}}.id
  name            = "{{.NetworkName}}"
}
`

const testAccVcdNsxtEdgegatewayL2VpnTunnelStep1 = testAccVcdNsxtEdgegatewayL2VpnTunnelData + `
resource "vcd_nsxt_edgegateway_l2_vpn_tunnel" "{{.ServerTunnelName}}" {
  name        = "{{.ServerTunnelName}}"
  description = "{{.ServerTunnelName}}"

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

resource "vcd_nsxt_edgegateway_l2_vpn_tunnel" "{{.ClientTunnelName}}" {
  name        = "{{.ClientTunnelName}}"
  description = "{{.ClientTunnelName}}"

  org             = "{{.Org}}"
  edge_gateway_id = vcd_nsxt_edgegateway.{{.NewEdgeGw}}.id

  session_mode = "CLIENT"
  enabled      = true

  local_endpoint_ip  = vcd_nsxt_edgegateway.{{.NewEdgeGw}}.primary_ip
  remote_endpoint_ip = "{{.RemoteEndpointIp}}"

  peer_code  = vcd_nsxt_edgegateway_l2_vpn_tunnel.{{.ServerTunnelName}}.peer_code
  depends_on = [vcd_nsxt_edgegateway_l2_vpn_tunnel.{{.ServerTunnelName}}]
}
`

const testAccVcdNsxtEdgegatewayL2VpnTunnelStep2 = testAccVcdNsxtEdgegatewayL2VpnTunnelData + `
resource "vcd_nsxt_edgegateway_l2_vpn_tunnel" "{{.ServerTunnelName}}" {
  name        = "{{.ServerTunnelName}}"
  description = "{{.ServerTunnelName}}"

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

resource "vcd_nsxt_edgegateway_l2_vpn_tunnel" "{{.ClientTunnelName}}" {
  name        = "{{.ClientTunnelName}}"
  description = "{{.ClientTunnelName}}"

  org             = "{{.Org}}"
  edge_gateway_id = vcd_nsxt_edgegateway.{{.NewEdgeGw}}.id

  session_mode = "CLIENT"
  enabled      = true

  local_endpoint_ip  = vcd_nsxt_edgegateway.{{.NewEdgeGw}}.primary_ip
  remote_endpoint_ip = "{{.RemoteEndpointIpUpdated}}"

  stretched_network {
    network_id = vcd_network_routed_v2.{{.NewNetworkName}}.id
    tunnel_id  = 1 
  }

  peer_code  = vcd_nsxt_edgegateway_l2_vpn_tunnel.{{.ServerTunnelName}}.peer_code
  depends_on = [vcd_nsxt_edgegateway_l2_vpn_tunnel.{{.ServerTunnelName}}]
}
`

const testAccVcdNsxtEdgegatewayL2VpnTunnelStep3 = testAccVcdNsxtEdgegatewayL2VpnTunnelData + `
resource "vcd_nsxt_edgegateway_l2_vpn_tunnel" "{{.ServerTunnelName}}" {
  name        = "{{.ServerTunnelName}}"
  description = "{{.ServerTunnelName}}"

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

resource "vcd_nsxt_edgegateway_l2_vpn_tunnel" "{{.ClientTunnelName}}" {
  name        = "{{.ClientTunnelName}}"
  description = "{{.ClientTunnelName}}"

  org             = "{{.Org}}"
  edge_gateway_id = vcd_nsxt_edgegateway.{{.NewEdgeGw}}.id

  session_mode = "CLIENT"
  enabled      = true

  local_endpoint_ip  = vcd_nsxt_edgegateway.{{.NewEdgeGw}}.primary_ip
  remote_endpoint_ip = "{{.RemoteEndpointIp}}"

  peer_code  = vcd_nsxt_edgegateway_l2_vpn_tunnel.{{.ServerTunnelName}}.peer_code
  depends_on = [vcd_nsxt_edgegateway_l2_vpn_tunnel.{{.ServerTunnelName}}]
}
`

const testAccVcdNsxtEdgegatewayL2VpnTunnelStep4 = testAccVcdNsxtEdgegatewayL2VpnTunnelData + `
resource "vcd_nsxt_edgegateway_l2_vpn_tunnel" "{{.ServerTunnelName}}" {
  name        = "{{.ServerTunnelName}}"
  description = "{{.ServerTunnelName}}"

  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.{{.EdgeGw}}.id

  session_mode              = "SERVER"
  enabled                   = false
  connector_initiation_mode = "ON_DEMAND"

  local_endpoint_ip  = data.vcd_nsxt_edgegateway.{{.EdgeGw}}.primary_ip
  tunnel_interface   = "{{.TunnelInterfaceUpdated}}"
  remote_endpoint_ip = "{{.RemoteEndpointIpUpdated}}"

  stretched_network {
    network_id = data.vcd_network_routed_v2.{{.NetworkName}}.id
  }

  pre_shared_key = "{{.PreSharedKeyUpdated}}"
}
`

const testAccVcdNsxtEdgegatewayL2VpnTunnelStep5 = testAccVcdNsxtEdgegatewayL2VpnTunnelData + `
resource "vcd_nsxt_edgegateway_l2_vpn_tunnel" "{{.ServerTunnelName}}" {
  name        = "{{.ServerTunnelName}}"
  description = "{{.ServerTunnelName}}"

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
