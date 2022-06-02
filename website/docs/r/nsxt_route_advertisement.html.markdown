---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_route_advertisement"
sidebar_current: "docs-vcd-resource-nsxt_route_advertisement"
description: |-
Provides a VMware Cloud Director resource for setting route advertisement in an NSX-T Edge Gateway.
---

# vcd\_nsxt\_route\_advertisement

Provides a VMware Cloud Director resource for setting route advertisement in an NSX-T Edge Gateway.

~> **Note:** This resource requires an NSX-T Edge Gateway. Also, for this resource to work appropriately, the option "Dedicate Tier-0 Gateway" must be enabled. Otherwise, route advertisement creation will fail.

## Example Usage 1 (Enable route advertisement and publish 192.168.1.0/24)

```hcl
data "vcd_org_vdc" "my_vdc" {
  org  = "my-org" #optional
  name = "my-vdc"
}

data "vcd_nsxt_edgegateway" "my_edge_gateway" {
  owner_id = data.vcd_org_vdc.my_vdc.id
  name     = "my-nsxt-edge-gateway"
}

resource "vcd_nsxt_route_advertisement" "my_route_advertisement" {
  edge_gateway_id = data.vcd_nsxt_edgegateway.my_edge_gateway.id
  enabled         = true
  subnets         = ["192.168.1.0/24"]
}
```

## Example Usage 2 (Enable route advertisement and publish 192.168.1.0/24 and 192.168.2.0/24)

```hcl
data "vcd_org_vdc" "my_vdc" {
  org  = "my-org" #optional
  name = "my-vdc"
}

data "vcd_nsxt_edgegateway" "my_edge_gateway" {
  owner_id = data.vcd_org_vdc.my_vdc.id
  name     = "my-nsxt-edge-gateway"
}

resource "vcd_nsxt_route_advertisement" "my_route_advertisement" {
  edge_gateway_id = data.vcd_nsxt_edgegateway.my_edge_gateway.id
  enabled         = true
  subnets         = ["192.168.1.0/24", "192.168.2.0/24"]
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
  when connected as sysadmin working across different organizations.
* `edge_gateway_id` - (Required) NSX-T Edge Gateway ID in which route advertisement is located.
* `enabled` - (Optional) Define if route advertisement is active. Default `true`.
* `subnets` - (Optional) Set of subnets that will be advertised to Tier-0 gateway. Leaving it empty means none.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing NSX-T route advertisement can be [imported][docs-import] into this resource
via supplying the full dot separated path for NSX-T Edge Gateway. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxt_route_advertisement.my-route-advertisement my-org.my-org-vdc-org-vdc-group-name.my-edge-gw
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
