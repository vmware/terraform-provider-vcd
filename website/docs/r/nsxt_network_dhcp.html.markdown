---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_network_dhcp"
sidebar_current: "docs-vcd-resource-nsxt-network-dhcp"
description: |-
  Provides a resource to manage DHCP pools for NSX-T Org VDC networks.
---

# vcd\_nsxt\_network\_dhcp

Provides a resource to manage DHCP pools for NSX-T Org VDC networks.

## Example Usage 1 (Routed Org VDC Network with EDGE mode)

```hcl
resource "vcd_network_routed_v2" "parent-network" {
  name = "nsxt-routed-dhcp"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  gateway       = "7.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "7.1.1.10"
    end_address   = "7.1.1.20"
  }
}

resource "vcd_nsxt_network_dhcp" "pools" {
  org_network_id = vcd_network_routed_v2.parent-network.id

  pool {
    start_address = "7.1.1.100"
    end_address   = "7.1.1.110"
  }

  pool {
    start_address = "7.1.1.111"
    end_address   = "7.1.1.112"
  }
}
```

## Example Usage 2 (Isolated Org VDC Network with NETWORK mode)
```hcl
resource "vcd_network_isolated_v2" "net1" {
  org      = "cloud"
  owner_id = vcd_org_vdc.with-edge-cluster.id # VDC ID with Edge Cluster configured
  name     = "private-network"

  gateway       = "7.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "7.1.1.10"
    end_address   = "7.1.1.20"
  }
}

resource "vcd_nsxt_network_dhcp" "pools" {
  org = "cloud"
  vdc = vcd_org_vdc.with-edge-cluster.name

  org_network_id      = vcd_network_isolated_v2.net1.id
  mode                = "NETWORK"
  listener_ip_address = "7.1.1.254"

  pool {
    start_address = "7.1.1.100"
    end_address   = "7.1.1.110"
  }
}
```

## Example Usage 3 (Routed Org VDC Network with RELAY mode)
```hcl
resource "vcd_nsxt_edgegateway_dhcp_forwarding" "dhcp-forwarding" {
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  
  enabled      = true
  dhcp_servers = [
    "65.43.21.0",
    "fe80::abcd",
  ]
}

resource "vcd_network_routed_v2" "net1" {
  org  = "cloud"
  vdc  = "nsxt-vdc-cloud"
  name = "nsxt-routed-dhcp"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  gateway       = "7.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "7.1.1.10"
    end_address   = "7.1.1.20"
  }
}

resource "vcd_nsxt_network_dhcp" "pools" {
  org = "cloud"
  vdc = "nsxt-vdc-cloud"

  org_network_id = vcd_network_routed_v2.net1.id

  # DHCP forwarding must be configured on NSX-T Edge Gateway
  # for RELAY mode
  mode = "RELAY"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
  when connected as sysadmin working across different organisations.
* `org_network_id` - (Required) ID of parent Org VDC Routed network.
* `pool` - (Optional) One or more blocks to define DHCP pool ranges. Must not be set when
  `mode=RELAY`. See [Pools](#pools) and example for usage details.
* `mode` - (Optional; *v3.8+*) One of `EDGE`, `NETWORK` or `RELAY`. Default is `EDGE`
  * `EDGE` can be used with Routed Org VDC networks.
  * `NETWORK` can be used for Isolated and Routed Org VDC networks. It requires
    `listener_ip_address` to be set and Edge Cluster must be assigned to VDC. 
  * `RELAY` can be used with Routed Org VDC networks, but requires DHCP forwarding configuration in
    NSX-T Edge Gateway.
* `listener_ip_address` - (Optional; *v3.8+*) IP address of DHCP server in network. Must match
  subnet. **Only** used when `mode=NETWORK`.
* `lease_time` - (Optional; *v3.8+*; VCD `10.3.1+`) - Lease time in seconds. Minimum value is 60s
  and maximum is 4294967295s (~ 49 days).
* `dns_servers` - (Optional; *v3.7+*; VCD `10.3.1+`) - The DNS server IPs to be assigned by this
  DHCP service. Maximum two values. 

## Pools

* `start_address` - (Required) Start address of DHCP pool range
* `end_address` - (Required) End address of DHCP pool range

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing DHCP configuration can be [imported][docs-import] into this resource
via supplying the full dot separated path for your Org VDC network. An example is
below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxt_network_dhcp.imported my-org.my-org-vdc-or-vdc-group.my-nsxt-vdc-network-name
```

The above would import the DHCP config settings that are defined on VDC network
`my-nsxt-vdc-network-name` which is configured in organization named `my-org` and VDC or VDC Group
named `my-org-vdc-or-vdc-group`.
