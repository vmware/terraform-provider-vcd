---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_route_advertisement"
sidebar_current: "docs-vcd-datasource-nsxt_route_advertisement"
description: |-
Provides a VMware Cloud Director data source for reading route advertisement in an NSX-T Edge Gateway.
---

# vcd\_nsxt\_route\_advertisement

Provides a VMware Cloud Director data source for reading route advertisement in an NSX-T Edge Gateway.

## Example Usage (Reading route advertisement from NSX-T Edge Gateway)

```hcl
data "vcd_vdc_group" "group1" {
  name = "my-vdc-group"
}

data "vcd_nsxt_edgegateway" "t1" {
  owner_id = data.vcd_vdc_group.group1.id
  name     = "my-nsxt-edge-gateway"
}

data "vcd_nsxt_route_advertisement" "route_advertisement" {
  edge_gateway_id = data.vcd_nsxt_edgegateway.t1.id
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
  when connected as sysadmin working across different organizations.
* `edge_gateway_id` - (Required) NSX-T Edge Gateway ID in which route advertisement is located.

## Attribute Reference

All the arguments and attributes defined in
[`vcd_nsxt_route_advertisement`](/providers/vmware/vcd/latest/docs/resources/nsxt_route_advertisement) resource are available.
