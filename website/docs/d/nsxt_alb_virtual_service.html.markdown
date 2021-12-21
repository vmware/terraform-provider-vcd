---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_alb_virtual_service"
sidebar_current: "docs-vcd-datasource-nsxt-alb-virtual-service"
description: |-
  Provides a data source to read NSX-T ALB Virtual services for particular NSX-T Edge Gateway. A virtual service
  advertises an IP address and ports to the external world and listens for client traffic. When a virtual service receives
  traffic, it directs it to members in ALB Pool.
---

# vcd\_nsxt\_alb\_virtual\_service

Supported in provider *v3.5+* and VCD 10.2+ with NSX-T and ALB.

Provides a data source to read NSX-T ALB Virtual services for particular NSX-T Edge Gateway. A virtual service
advertises an IP address and ports to the external world and listens for client traffic. When a virtual service receives
traffic, it directs it to members in ALB Pool.

## Example Usage

```hcl
data "vcd_nsxt_edgegateway" "existing" {
  org = "my-org"
  vdc = "nsxt-vdc"

  name = "nsxt-gw"
}

data "vcd_nsxt_alb_virtual_service" "test" {
  org = "dainius"
  vdc = "nsxt-vdc-dainius"

  edge_gateway_id = vcd_nsxt_edgegateway.existing.id
  name            = "virutal-service-name"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which the edge gateway belongs. Optional if defined at provider level
* `vdc` - (Optional) The name of VDC that owns the edge gateway. Optional if defined at provider level
* `edge_gateway_id` - (Required) An ID of NSX-T Edge Gateway. Can be lookup up using
  [vcd_nsxt_edgegateway](/providers/vmware/vcd/latest/docs/data-sources/nsxt_edgegateway) data source
* `name` - (Required) The name of ALB Virtual Service

## Attribute Reference

All the arguments and attributes defined in
[`vcd_nsxt_alb_virtual_service`](/providers/vmware/vcd/latest/docs/resources/nsxt_alb_virtual_service) resource are
available.
