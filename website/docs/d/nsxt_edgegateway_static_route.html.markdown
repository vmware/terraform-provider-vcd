---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_edgegateway_static_route"
sidebar_current: "docs-vcd-data-source-nsxt-edge-static-route"
description: |-
  Provides a data source to read NSX-T Edge Gateway Static Routes.
---

# vcd\_nsxt\_edgegateway\_static\_route

Supported in provider *v3.10+* and VCD 10.4.0+ with NSX-T.

Provides a data source to read NSX-T Edge Gateway Static Routes.

## Example Usage (by Name only)

```hcl
data "vcd_nsxt_edgegateway_static_route" "by-name" {
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  name            = "existing-static-route"
}
```

## Example Usage (by Name and Network Cidr )

```hcl
data "vcd_nsxt_edgegateway_static_route" "by-name-and-cidr" {
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  name            = "duplicate-name-sr"
  network_cidr    = "10.10.11.0/24"
}
```

## Argument Reference

The following arguments are supported:

* `edge_gateway_id` - (Required) NSX-T Edge Gateway ID
* `name` - (Required) Name of Static Route. **Note** names *can be duplicate* and one can use
  `network_cidr` to make filtering more precise
* `network_cidr` - (Optional) Network CIDR for Static Route

## Attribute Reference

All the arguments and attributes defined in
[`vcd_nsxt_edgegateway_static_route`](/providers/vmware/vcd/latest/docs/resources/nsxt_edgegateway_static_route)
resource are available.
