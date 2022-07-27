---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_edgegateway_bgp_neighbor"
sidebar_current: "docs-vcd-resource-nsxt-edgegateway-bgp-neighbor"
description: |-
  Provides a resource to manage NSX-T Edge Gateway BGP Neighbors and their configuration.
---

# vcd\_nsxt\_edgegateway\_bgp\_neighbor

Supported in provider *v3.7+* and VCD 10.2+ with NSX-T

Provides a resource to manage NSX-T Edge Gateway BGP Neighbors and their configuration.

## Example Usage (BGP Neighbor configuration with route filtering with referenced BGP IP Prefix List)

```hcl
resource "vcd_nsxt_edgegateway_bgp_neighbor" "neighbor-one" {
  org = "datacloud"

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  ip_address       = "1.1.1.1"
  remote_as_number = "62513"

  keep_alive_timer      = 78
  hold_down_timer       = 400
  graceful_restart_mode = "GRACEFUL_AND_HELPER"
  allow_as_in           = false
  bfd_enabled           = true
  bfd_interval          = 800
  bfd_dead_multiple     = 5

  route_filtering              = "IPV4"
  in_filter_ip_prefix_list_id  = data.vcd_nsxt_edgegateway_bgp_ip_prefix_list.in-1.id
  out_filter_ip_prefix_list_id = data.vcd_nsxt_edgegateway_bgp_ip_prefix_list.out-1.id
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
  when connected as sysadmin working across different organisations
* `edge_gateway_id` - (Required) The ID of the edge gateway (NSX-T only). Can be looked up using
  `vcd_nsxt_edgegateway` datasource
* `ip_address` - (Required) BGP Neighbor IP Address (IPv4 or IPv6)
* `remote_as_number` - (Required) BGP Neighbor Remote Autonomous System (AS) Number
* `password` - (Optional) BGP Neighbor Password
* `keep_alive_timer` - (Optional) Time interval (in seconds) between sending keep-alive messages to a BGP peer
* `hold_down_timer` - (Optional) Time interval (in seconds) before declaring a BGP peer dead
* `graceful_restart_mode` - (Optional) BGP Neighbor Graceful Restart Mode. One of:
  * `DISABLE` - Overrides the global edge gateway settings and disables graceful restart mode for this neighbor.
  * `HELPER_ONLY` - Overrides the global edge gateway settings and configures graceful restart mode as Helper only for this neighbor.
  * `GRACEFUL_AND_HELPER` - Overrides the global edge gateway settings and configures graceful restart mode as Graceful restart and Helper for this neighbor.
* `allow_as_in` - (Optional) BGP Allow-as-in feature is used to allow the BGP speaker to accept the BGP updates even if its own BGP AS number is in the AS-Path attribute.
* `bfd_enabled` - (Optional) Should Bidirectional Forwarding Detection (BFD) be enabled 
* `bfd_interval` - (Optional) Time interval (in milliseconds) between heartbeat packets
* `bfd_dead_multiple` - (Optional) Number of times a heartbeat packet is missed before BFD declares that the neighbor is down
* `route_filtering` - (Optional) Route filtering by IP Address family. One of `DISABLED`, `IPV4`, `IPV6`
* `in_filter_ip_prefix_list_id` - (Optional) The ID of the IP Prefix List to be used for filtering incoming BGP routes
* `out_filter_ip_prefix_list_id` - (Optional) The ID of the IP Prefix List to be used for filtering outgoing BGP routes

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing BGP IP Prefix List configuration can be [imported][docs-import] into this resource
via supplying path for it. An example is
below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxt_edgegateway_bgp_neighbor.imported `my-org.my-vdc-or-vdc-group.my-edge-gateway.bgp-neighbor-ip`
```

The above would import the `bgp-neighbor-ip` BGP Neighbor that is defined in
`my-edge-gateway` NSX-T Edge Gateway. Edge Gateway should be located in `my-vdc-or-vdc-group` VDC or
VDC Group in Org `my-org`
