---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_ip_set"
sidebar_current: "docs-vcd-resource-nsxt-ip-set"
description: |-
  Provides a resource to manage NSX-T IP Set. IP sets are groups of objects to which the firewall rules apply. Combining
  multiple objects into IP sets helps reduce the total number of firewall rules to be created.
---

# vcd\_nsxt\_ip\_set

Supported in provider *v3.3+* and VCD 10.1+ with NSX-T backed VDCs.

Provides a resource to manage NSX-T IP Set. IP sets are groups of objects to which the firewall rules apply. Combining 
multiple objects into IP sets helps reduce the total number of firewall rules to be created.

## Example Usage (IP set with multiple IP address ranges defined)

```hcl
resource "vcd_nsxt_ip_set" "set1" {
  org = "my-org"
  vdc = "my-org-vdc"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  name        = "first-ip-set"
  description = "IP Set containing IPv4 and IPv6 ranges"

  ip_addresses = [
    "12.12.12.1",
    "10.10.10.0/24",
    "11.11.11.1-11.11.11.2",
    "2001:db8::/48",
    "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
  when connected as sysadmin working across different organisations.
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level.
* `name` - (Required) A unique name for IP Set
* `description` - (Optional) An optional description of the IP Set
* `edge_gateway_id` - (Required) The ID of the edge gateway (NSX-T only). Can be looked up using
  `vcd_nsxt_edgegateway` data source
* `ip_addresses` (Optional) A set of IP addresses, subnets or ranges (IPv4 or IPv6)

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing IP Set configuration can be [imported][docs-import] into this resource
via supplying the full dot separated path for your IP Set name. An example is
below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxt_ip_set.imported my-org.my-org-vdc.my-nsxt-edge-gateway.my-ip-set-name
```

The above would import the `my-ip-set-name` IP Set config settings that are defined
on NSX-T Edge Gateway `my-nsxt-edge-gateway` which is configured in organization named `my-org` and
VDC named `my-org-vdc`.
