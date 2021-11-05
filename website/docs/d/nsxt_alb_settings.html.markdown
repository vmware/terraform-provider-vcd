---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_alb_settings"
sidebar_current: "docs-vcd-datasource-nsxt-alb-settings"
description: |-
  Provides a data source to read NSX-T ALB General Settings for particular NSX-T Edge Gateway.
---

# vcd\_nsxt\_alb\_settings

Supported in provider *v3.5+* and VCD 10.2+ with NSX-T and ALB.

Provides a data source to read NSX-T ALB General Settings for particular NSX-T Edge Gateway.

## Example Usage

```hcl
data "vcd_nsxt_edgegateway" "existing" {
  org = "my-org"
  vdc = "nsxt-vdc"

  name = "nsxt-gw"
}

data "vcd_nsxt_alb_settings" "test" {
  org = "my-org"
  vdc = "nsxt-vdc"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which the edge gateway belongs. Optional if defined at provider level.
* `vdc` - (Optional) The name of VDC that owns the edge gateway. Optional if defined at provider level.
* `edge_gateway_id` - (Required) An ID of NSX-T Edge Gateway. Can be lookup up using
  [vcd_nsxt_edgegateway](/providers/vmware/vcd/latest/docs/data-sources/nsxt_edgegateway) data source

## Attribute Reference

All the arguments and attributes defined in
[`vcd_nsxt_alb_settings`](/providers/vmware/vcd/latest/docs/resources/nsxt_alb_settings) resource are available.
