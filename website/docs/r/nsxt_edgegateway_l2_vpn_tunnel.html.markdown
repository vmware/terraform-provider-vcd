---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_edgegateway_l2_vpn_tunnel"
sidebar_current: "docs-vcd-resource-nsxt-edgegateway-l2-vpn-tunnel"
description: |-
  Provides a resource to manage NSX-T Edge Gateway L2 VPN Tunnel sessions and their configurations.
---

# vcd\_nsxt\_edgegateway\_l2\_vpn\_tunnel

Supported in provider *v3.11+* and VCD 10.3.1+ with NSX-T

Provides a resource to manage NSX-T Edge Gateway L2 VPN Tunnel sessions and their configurations.

## Example Usage (Both server and client tunnel sessions connecting both Edge Gateways)

```hcl
resource "vcd_nsxt_edgegateway_l2_vpn_tunnel" "server-session" {
  org = "datacloud"

  edge_gateway_id = data.vcd_nsxt_edgegateway.server-testing.id

  name = "server-session"
  description = "example description"

  session_mode             = "SERVER"
  enabled                  = true
  connector_initiator_mode = "ON_DEMAND"

  # must be sub-allocated on the Edge Gateway
  local_endpoint_ip  = "10.10.50.2"
  tunnel_interface   = "192.168.0.1/24"
  remote_endpoint_ip = "1.2.2.3"

  stretched_network {
    network_id = data.vcd_routed_network_v2.test_network_server.id
  }

  pre_shared_key = "secret_passphrase"
}

resource "vcd_nsxt_edgegateway_l2_vpn_tunnel" "client-session" {
  org = "datacloud"

  # Note that this is different, as one edge gateway can only function
  # in SERVER or CLIENT mode.
  edge_gateway_id = data.vcd_nsxt_edgegateway.client-testing.id

  name = "client-session"
  description = "example description"

  session_mode = "CLIENT"
  enabled      = true

  # must be sub-allocated on the Edge Gateway
  local_endpoint_ip  = "101.22.30.3"
  remote_endpoint_ip = "1.2.2.3"

  stretched_network {
    network_id = data.vcd_routed_network_v2.test_network_client.id
    # CLIENT sessions need to define a tunnel ID for every stretched network
    tunnel_id  = 1
  }

  stretched_network {
    network_id = data.vcd_routed_network_v2.test_network_client_other.id
    tunnel_id  = 2
  }

  # Be aware, that if there are changes in the `server-session`, the peer_code
  # will be updated aswell, so `terraform apply` needs to be run twice
  peer_code = vcd_nsxt_edgegateway_l2_vpn_tunnel.server-session.peer_code
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at 
  provider level. Useful when connected as sysadmin working across different organisations
* `edge_gateway_id` - (Required) The ID of the edge gateway (NSX-T only). 
  Can be looked up using `vcd_nsxt_edgegateway` datasource
* `name` - (Required) The name of the tunnel.
* `description` - (Optional) The description of the tunnel.
* `session_mode` - (Required) Mode of the tunnel session (SERVER or CLIENT)
* `enabled` - (Optional) State of the session (Set to `true` by default)
* `connector_initiator_mode` - (Required for `SERVER` sessions) Mode in which 
  the connection is formed. Only relevant to `SERVER` sessions. One of:
	* INITIATOR - Local endpoint initiates tunnel setup and will also respond to 
  incoming tunnel setup requests from the peer gateway.
	* RESPOND_ONLY - Local endpoint shall only respond to incoming tunnel setup 
  requests, it shall not initiate the tunnel setup.
	* ON_DEMAND - In this mode local endpoint will initiate tunnel creation once 
  first packet matching the policy rule is received, and will also respond to 
  incoming initiation requests.
* `local_endpoint_ip` - (Required) The IP address corresponding to the Edge 
  Gateway the tunnel is being configured on. The IP must be sub-allocated 
  on the Edge Gateway.
* `remote_endpoint_ip` - (Required) The IP address of the remote endpoint, which 
corresponds to the device on the remote site terminating the VPN tunnel.
* `tunnel_interface` - (Optional) The network CIDR block over which the session 
  interfaces. Relevant only for SERVER session modes. If not provided, Cloud 
  Director will attempt to automatically allocate a tunnel interface.
* pre_shared_key - (Required for `SERVER` sessions) The key that is used for 
  authenticating the connection, only needed for `SERVER` sessions.
* peer_code - (Optional) Encoded string that contains the whole configuration 
  of a `SERVER` session, including the pre-shared key, so it is user's 
  responsibility to secure it.

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing L2 VPN Tunnel configuration can be [imported][docs-import] into this resource
via supplying path for it. An example is
below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxt_edgegateway_l2_vpn_tunnel.imported `my-org.my-vdc-or-vdc-group.my-edge-gateway.l2_vpn_tunnel`
```

The above would import the `l2_vpn_tunnel` L2 VPN Tunnel that is defined in
`my-edge-gateway` NSX-T Edge Gateway. Edge Gateway should be located in `my-vdc-or-vdc-group` VDC or
VDC Group in Org `my-org`
