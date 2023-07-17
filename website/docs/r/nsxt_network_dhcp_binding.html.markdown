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

## Example Usage (IPv4 binding)

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

## Example Usage (IPv6 binding)

```hcl
resource "vcd_network_routed_v2" "ipv6-dualstack" {
  org  = "cloud"
  name = "Dual stack routed network"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  gateway       = "192.168.1.1"
  prefix_length = 24
  static_ip_pool {
    start_address = "192.168.1.10"
    end_address   = "192.168.1.20"
  }

  dual_stack_enabled      = true
  secondary_gateway       = "2002:0:0:1234:abcd:ffff:c0a6:121"
  secondary_prefix_length = 124

  secondary_static_ip_pool {
    start_address = "2002:0:0:1234:abcd:ffff:c0a6:122"
    end_address   = "2002:0:0:1234:abcd:ffff:c0a6:124"
  }
}

resource "vcd_nsxt_edgegateway_dhcpv6" "test" {
  org             = "cloud"
  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  enabled = true
  # Bindings can be configured only in `DHCPv6` mode
  mode = "DHCPv6"
}

resource "vcd_nsxt_network_dhcp" "routed-ipv6-dual-stack" {
  org = "cloud"

  org_network_id      = vcd_network_routed_v2.ipv6-dualstack.id
  mode                = "NETWORK"
  listener_ip_address = "2002:0:0:1234:abcd:ffff:c0a6:129"

  pool {
    start_address = "2002:0:0:1234:abcd:ffff:c0a6:125"
    end_address   = "2002:0:0:1234:abcd:ffff:c0a6:126"
  }

  depends_on = [vcd_nsxt_edgegateway_dhcpv6.test]
}

resource "vcd_nsxt_network_dhcp_binding" "ipv6-binding1" {
  org = "cloud"

  org_network_id = vcd_nsxt_network_dhcp.routed-ipv6-dual-stack.id

  name         = "IPv6 DHCP Binding-1"
  binding_type = "IPV6"
  ip_address   = "2002:0:0:1234:abcd:ffff:c0a6:127"
  lease_time   = 3600
  mac_address  = "00:11:22:33:44:66"

  dhcp_v6_config {
    sntp_servers = ["4b0d:74eb:ee01:0ff4:ab1b:f7cc:4d74:d2a3", "cc80:5498:18da:0883:d78a:4e4b:754d:df47"]
    domain_names = ["non-existing.org.tld", "fake.org.tld"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization. Optional if defined at provider level
* `org_network_id` - (Required) The ID of an Org VDC network. **Note**  (`.id` field) of
  `vcd_network_isolated_v2`, `vcd_network_routed_v2` or `vcd_nsxt_network_dhcp` can be referenced
  here. It is more convenient to use reference to `vcd_nsxt_network_dhcp` ID because it makes sure
  that DHCP is enabled before configuring pools
* `binding_type` - (Required) One of `IPV4` or `IPV6`
* `ip_address` - (Required) IP address used for binding
* `mac_address` - (Required) MAC address used for binding
* `lease_time` - (Required) Lease time in seconds. Minimum `3600` seconds
* `dns_servers` - (Optional) A list of DNS servers. Maximum 2 can be specified
* `dhcp_v4_config` - (Optional) Additional configuration for IPv4 specific options. See [IPv4 block](#ipv4-block)
* `dhcp_v6_config` - (Optional, *v3.10+*) Additional configuration for IPv6 specific options. See [IPv6 block](#ipv6-block)

<a id="ipv4-block"></a>

## IPv4 block (dhcp_v4_config)

* `gateway_ip_address` - (Optional) Gateway IP address to use for the client
* `hostname` - (Optional) Hostname to be set for client

<a id="ipv6-block"></a>
## IPv6 block (dhcp_v6_config)

* `sntp_servers` - (Optional) A set of SNTP servers
* `domain_names` - (Optional) A set of domain names

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
