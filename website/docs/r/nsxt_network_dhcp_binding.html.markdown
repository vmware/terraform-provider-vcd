---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_network_dhcp_binding"
sidebar_current: "docs-vcd-resource-nsxt-network-dhcp-binding"
description: |-
  Provides a resource to manage NSX-T Org VDC network DHCP bindings.
---

# vcd\_nsxt\_network\_dhcp\_binding

Provides a resource to manage NSX-T Org VDC network DHCP bindings.

-> This resource requires VCD 10.3.1+

## Example Usage

```hcl
resource "vcd_nsxt_network_dhcp" "pools" {
  org = "cloud"

  org_network_id      = vcd_network_isolated_v2.net1.id
  mode                = "NETWORK"
  listener_ip_address = "7.1.1.254"

  pool {
    start_address = "7.1.1.100"
    end_address   = "7.1.1.110"
  }
}

resource "vcd_nsxt_network_dhcp_binding" "binding2" {
  org = "cloud"

  # Referencing vcd_nsxt_network_dhcp.pools.id instead of vcd_network_isolated_v2.net1.id because
  # DHCP service must be enabled on the network before DHCP bindings can be created and it makes
  # implicit dependencies work. One can reference `vcd_network_isolated_v2.net1.id` here and use
  # depends_on = [vcd_nsxt_network_dhcp.pools]
  org_network_id = vcd_nsxt_network_dhcp.pools.id

  name         = "DHCP Binding"
  description  = "DHCP binding description"
  binding_type = "IPV4"
  ip_address   = "7.1.1.190"
  lease_time   = 3600
  mac_address  = "00:11:22:33:44:66"
  dns_servers  = ["7.1.1.242", "7.1.1.243"]

  dhcp_v4_config {
    gateway_ip_address = "7.1.1.233"
    hostname           = "non-existent"
  }
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization. Optional if defined at provider level
* `org_network_id` - (Required) The ID of an Org VDC network. **Note**  (`.id` field) of
  `vcd_network_isolated_v2`, `vcd_network_routed_v2` and `vcd_nsxt_network_dhcp` suite here. It is
  more convenient to use reference to `vcd_nsxt_network_dhcp` ID because it makes sure that DHCP is
  enabled before configuring pools
* `binding_type` - (Required) One of `IPV4` or `IPV6`
* `ip_address` - (Required) IP address used for binding
* `mac_address` - (Required) MAC address used for binding
* `lease_time` - (Required) Lease time in seconds. Minimum `3600` seconds
* `dns_servers` - (Optional) A list of DNS servers. Maximum 2 can be specified
* `dhcp_v4_config` - (Optional) Additional configuration for IPv4 specific options. See [IPv4 block](#ipv4-block)

<a id="ipv4-block"></a>

## dhcp_v4_config

* `gateway_ip_address` - (Optional) Gateway IP address to use for the client
* `hostname` - (Optional) Hostname to be set for client

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing NSX-T DHCP Binding configuration can be [imported][docs-import] into this resource via
supplying path for it. An example is
below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxt_network_dhcp_binding.imported my-org.my-org-vdc-or-vdc-group.my-nsxt-vdc-network-name.my-binding-name
```

The above would import the `my-binding-name` NSX-T DHCP Binding configuration
