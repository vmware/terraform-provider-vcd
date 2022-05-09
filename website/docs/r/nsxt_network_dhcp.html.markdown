---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_network_dhcp"
sidebar_current: "docs-vcd-resource-nsxt-network-dhcp"
description: |-
  Provides a resource to manage DHCP pools for NSX-T Org VDC Routed network.
---

# vcd\_nsxt\_network\_dhcp

Provides a resource to manage DHCP pools for NSX-T Org VDC Routed network.


## Example Usage 1

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

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
  when connected as sysadmin working across different organisations.
* `org_network_id` - (Required) ID of parent Org VDC Routed network
* `pool` - (Required) One or more blocks to define DHCP pool ranges. See [Pools](#pools) and example 
for usage details.
* `dns_servers` - (Optional; *v3.7+*) - The DNS server IPs to be assigned by this DHCP service. Maximum two values. 
This argument is supported from VCD 10.3.1+.

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
