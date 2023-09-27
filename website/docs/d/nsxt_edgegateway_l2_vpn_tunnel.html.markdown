---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_edgegateway_l2_vpn_tunnel"
sidebar_current: "docs-vcd-datasource-nsxt-edgegateway-l2-vpn-tunnel"
description: |-
  Provides a datasource to read NSX-T Edge Gateway L2 VPN Tunnel sessions and their configurations.
---

# vcd\_nsxt\_edgegateway\_l2\_vpn\_tunnel

Supported in provider *v3.11+* and VCD *10.3.1+* with NSX-T

Provides a datasource to read NSX-T Edge Gateway L2 VPN Tunnel sessions and their configurations.

## Example Usage (Reading a tunnel's server session to get the peer code for the client session)

```hcl
data "vcd_org_vdc" "vdc1" {
  name = "existing-vdc"
}

data "vcd_nsxt_edgegateway" "testing" {
  org      = "myorg"
  owner_id = data.vcd_org_vdc.vdc1.id
  name     = "nsxt-edge-gateway"
}

data "vcd_nsxt_edgegateway_l2_vpn_tunnel" "server-session" {
  org = "datacloud"
  edge_gateway_id = data.vcd_nsxt_edgegateway.server-testing.id

  name = "server-session"
}

resource "vcd_nsxt_edgegateway_l2_vpn_tunnel" "client-session" {
  org = "datacloud"

  # Note that this is different, as one edge gateway can only function
  # in SERVER or CLIENT mode.
  edge_gateway_id = data.vcd_nsxt_edgegateway.client-testing.id

  session_mode = "CLIENT"
  enabled      = true

  # must be sub-allocated on the Edge Gateway
  local_endpoint_ip  = "101.22.30.3"
  remote_endpoint_ip = "1.2.2.3"

  peer_code = data.vcd_nsxt_edgegateway_l2_vpn_tunnel.server-session.peer_code
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at 
  provider level. Useful when connected as sysadmin working across different organisations
* `edge_gateway_id` - (Required) The ID of the edge gateway (NSX-T only). 
  Can be looked up using `vcd_nsxt_edgegateway` datasource
* `name` - (Required) The name of the tunnel.

## Attribute Reference

All properties defined in [vcd_nsxt_edgegateway_l2_vpn_tunnel](/providers/vmware/vcd/latest/docs/resources/nsxt_edgegateway_l2_vpn_tunnel)
resource are available.

