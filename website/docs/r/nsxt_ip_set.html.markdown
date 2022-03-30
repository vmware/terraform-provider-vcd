---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_ip_set"
sidebar_current: "docs-vcd-resource-nsxt-ip-set"
description: |-
  Provides a resource to manage NSX-T IP Set. IP Sets are groups of objects to which the firewall rules apply. Combining
  multiple objects into IP Sets helps reduce the total number of firewall rules to be created.
---

# vcd\_nsxt\_ip\_set

Supported in provider *v3.3+* and VCD 10.1+ with NSX-T backed VDCs.

Provides a resource to manage NSX-T IP Set. IP Sets are groups of objects to which the firewall rules apply. Combining 
multiple objects into IP Sets helps reduce the total number of firewall rules to be created.

-> Starting with **v3.6.0** `vcd_nsxt_ip_set` added support for VDC Groups.
The `vdc` field (in resource or inherited from provider configuration) is deprecated, as `vcd_nsxt_ip_set` will
inherit the VDC Group or VDC membership from a parent Edge Gateway specified in the `edge_gateway_id` field.
More about VDC Group support in a [VDC Groups guide](/providers/vmware/vcd/latest/docs/guides/vdc_groups).

## Example Usage (IP Set with multiple IP address ranges defined)

```hcl
data "vcd_nsxt_edgegateway" "main" {
  org  = "my-org" # Optional
  name = "main-edge"
}

resource "vcd_nsxt_ip_set" "set1" {
  org = "my-org" # Optional

  edge_gateway_id = data.vcd_nsxt_edgegateway.main.id

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
* `vdc` - (Deprecated;Optional) The name of VDC to use, optional if defined at provider level. **Deprecated**
  in favor of `edge_gateway_id` field.
* `name` - (Required) A unique name for IP Set
* `description` - (Optional) An optional description of the IP Set
* `edge_gateway_id` - (Required) The ID of the edge gateway (NSX-T only). Can be looked up using
  `vcd_nsxt_edgegateway` data source.
* `ip_addresses` (Optional) A set of IP addresses, subnets or ranges (IPv4 or IPv6)

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing IP Set configuration can be [imported][docs-import] into this resource
via supplying the full dot separated path for your IP Set name. An example is
below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxt_ip_set.imported my-org.my-org-vdc-name.my-nsxt-edge-gateway-name.my-ip-set-name
or
terraform import vcd_nsxt_ip_set.imported my-org.my-vdc-group-name.my-nsxt-edge-gateway-name.my-ip-set-name
```

The above would import the `my-ip-set-name` IP Set config settings that are defined
on NSX-T Edge Gateway `my-nsxt-edge-gateway` which is configured in organization named `my-org` and
VDC named `my-org-vdc-name` or VDC Group `my-vdc-group-name`.
