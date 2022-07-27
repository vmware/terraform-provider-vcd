---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_edgegateway_bgp_neighbor"
sidebar_current: "docs-vcd-resource-nsxt-edgegateway-bgp-neighbor"
description: |-
  Provides a data source to read NSX-T Edge Gateway BGP Neighbors and their configuration.
---

# vcd\_nsxt\_edgegateway\_bgp\_neighbor

Supported in provider *v3.7+* and VCD 10.2+ with NSX-T

Provides a data source to read NSX-T Edge Gateway BGP Neighbors and their configuration.

## Example Usage

```hcl
data "vcd_nsxt_edgegateway" "existing" {
  org = "my-org"
  vdc = "nsxt-vdc"

  name = "nsxt-gw"
}

data "vcd_nsxt_edgegateway_bgp_neighbor" "first" {
  org = "my-org"
  vdc = "nsxt-vdc"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  ip_address = "192.168.102.45"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which the edge gateway belongs. Optional if defined at provider level.
* `edge_gateway_id` - (Required) An ID of NSX-T Edge Gateway. Can be looked up using
  [vcd_nsxt_edgegateway](/providers/vmware/vcd/latest/docs/data-sources/nsxt_edgegateway) data source
* `ip_address` - (Required) An IP Address (IPv4 or IPv6) of existing BGP Neighbor in specified Edge Gateway

## Attribute Reference

All the arguments and attributes defined in
[`vcd_nsxt_edgegateway_bgp_neighbor`](/providers/vmware/vcd/latest/docs/resources/nsxt_edgegateway_bgp_neighbor)
resource are available.
