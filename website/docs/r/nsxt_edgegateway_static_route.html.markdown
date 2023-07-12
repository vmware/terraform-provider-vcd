---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_edgegateway_static_route"
sidebar_current: "docs-vcd-resource-nsxt-edgegateway-static-route"
description: |-
  Provides a resource to manage NSX-T Edge Gateway Static Routes.
---

# vcd\_nsxt\_edgegateway\_static\_route

Supported in provider *v3.10+* and VCD 10.4.0+ with NSX-T.

Provides a resource to manage NSX-T Edge Gateway Static Routes.

## Example Usage

```hcl
resource "vcd_nsxt_edgegateway_static_route" "sr" {
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  name         = "some-static-route"
  description  = "description for the route"
  network_cidr = "192.168.1.0/24"

  next_hop {
    ip_address     = "1.1.1.1"
    admin_distance = 4

    scope {
      id   = data.vcd_network_routed_v2.net.id
      type = "NETWORK"
    }
  }

  next_hop {
    ip_address     = "1.1.1.254"
    admin_distance = 3

    scope {
      id   = data.vcd_network_routed_v2.net.id
      type = "NETWORK"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `edge_gateway_id` - (Required) NSX-T Edge Gateway ID
* `name` - (Required) Name for NSX-T Edge Gateway Static Route
* `description` - (Optional) Description for NSX-T Edge Gateway Static Route
* `network_cidr` - (Required) Specifies network prefix in CIDR format. Both IPv4 and IPv6 formats
  are supported.
* `next_hop` - (Required) A set of next hops to use within the static route. At least one is
  required. See [Next Hop](#next-hop) for definition structure.

<a id="next-hop"></a>
## Next Hop

Each member `next_hop` contains the following attributes:

* `ip_address` - (Required) IP address for next hop gateway IP Address for the Static Route
* `admin_distance` - (Required) Admin distance is used to choose which route to use when there are
  multiple routes for a specific network. The lower the admin distance, the higher the preference
  for the route. Starts with 1.
* `scope` - (Optional) Scope holds a reference to an entity where the next hop of a Static Route is
reachable. In general, the reference should be an Org VDC network or segment backed external
network, but scope could also reference a SYSTEM_OWNED entity if the next hop is configured outside
of VCD. See [Next Hop Scope](#next-hop-scope) for definition structure.

<a id="next-hop-scope"></a>
## Next Hop Scope

* `id` - (Required) ID of Org VDC network or segment backed external network
* `type` - (Required) Type of backing entity. In general this will be `NETWORK` but can become
  `SYSTEM_OWNED` if the Static Route is modified outside of VCD

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing NSX-T Edge Static Route configuration can be [imported][docs-import] into this
resource via supplying path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxt_edgegateway_static_route.imported my-org.nsxt-vdc.nsxt-edge.static-route-name
```

or 

```
terraform import vcd_nsxt_edgegateway_static_route.imported my-org.nsxt-vdc.nsxt-edge.static-route-cidr
```

The above would import the `static-route-name` or `static-route-cidr` Edge Gateway Static Route
configuration for this particular Edge Gateway.
