---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_ip_space_uplink"
sidebar_current: "docs-vcd-resource-ip-space-uplink"
description: |-
  Provides a resource to manage IP Space Uplinks in External Networks (Provider Gateways).
---

# vcd\_ip\_space\_uplink

Provides a resource to manage IP Space Uplinks in External Networks (Provider Gateways).

~> Only `System Administrator` can create this resource.

## Example Usage (Adding IP Space Uplink to Provider Gateway)

```hcl
data "vcd_nsxt_manager" "main" {
  name = "nsxManager1"
}

data "vcd_nsxt_tier0_router" "router" {
  name            = "tier0Router"
  nsxt_manager_id = data.vcd_nsxt_manager.main.id
}

resource "vcd_ip_space" "space1" {
  name = "ip-space-1"
  type = "PUBLIC"

  internal_scope = ["192.168.1.0/24"]

  route_advertisement_enabled = false
}

resource "vcd_external_network_v2" "provider-gateway" {
  name = "ProviderGateway1"

  nsxt_network {
    nsxt_manager_id      = data.vcd_nsxt_manager.main.id
    nsxt_tier0_router_id = data.vcd_nsxt_tier0_router.router.id
  }

  use_ip_spaces = true
}

resource "vcd_ip_space_uplink" "u1" {
  name                = "uplink"
  description         = "uplink number one"
  external_network_id = vcd_external_network_v2.provider-gateway.id
  ip_space_id         = vcd_ip_space.space1.id
}
```

## Example Usage (Adding IP Space Uplink with Tier-0 Router Associated Interfaces)

```hcl
data "vcd_nsxt_tier0_interface" "one" {
  external_network_id = vcd_external_network_v2.provider-gateway.id
  name                = "interface-one"
}

resource "vcd_ip_space_uplink" "u1" {
  name                     = "uplink"
  external_network_id      = vcd_external_network_v2.provider-gateway.id
  ip_space_id              = vcd_ip_space.space1.id
  associated_interface_ids = [data.vcd_nsxt_tier0_interface.one.id]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A tenant facing name for IP Space Uplink
* `description` - (Optional) An optional description for IP Space Uplink
* `external_network_id` - (Required) External Network ID For IP Space Uplink configuration
* `ip_space_id` - (Required) IP Space ID configuration
* `associated_interface_ids_id` - (Optional; *v3.14+*, *VCD 10.5+*) A set of Tier-0 Router Interface
  IDs that can be associated with the Uplink. Data Source
  [vcd_nsxt_tier0_router_interface](/providers/vmware/vcd/latest/docs/data-sources/nsxt_tier0_router_interface)
  can help to look it up.

## Attribute Reference

The following attributes are exported on this resource:

* `ip_space_type` - Backing IP Space type
* `status` - Status of IP Space Uplink


## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing IP Space Uplink configuration can be [imported][docs-import] into this resource via
supplying path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_ip_space_uplink.imported external-network-name.ip-space-uplink-name
```

The above would import the `ip-space-uplink-name` IP Space Uplink that is set for
`external-network-name`
