---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_alb_edgegateway_service_engine_group"
sidebar_current: "docs-vcd-datasource-nsxt-alb-edgegateway-service-engine-group"
description: |-
  Provides a datasource to read NSX-T ALB Service Engine Group assignment to NSX-T Edge Gateway.
---

# vcd\_nsxt\_alb\_edgegateway\_service\_engine\_group

Supported in provider *v3.5+* and VCD 10.2+ with NSX-T and ALB.

Provides a datasource to read NSX-T ALB Service Engine Group assignment to NSX-T Edge Gateway.

## Example Usage (Referencing Service Engine Group by ID)

```hcl
data "vcd_nsxt_edgegateway" "existing" {
  org = "my-org"
  vdc = "nsxt-vdc"

  name = "nsxt-gw"
}

data "vcd_nsxt_alb_service_engine_group" "first" {
  name = "first-se"
}

data "vcd_nsxt_alb_edgegateway_service_engine_group" "test" {
  edge_gateway_id         = data.vcd_nsxt_edgegateway.existing.id
  service_engine_group_id = data.vcd_nsxt_alb_service_engine_group.first.id
}
```

## Example Usage (Referencing Service Engine Group by Name)

```hcl
data "vcd_nsxt_edgegateway" "existing" {
  org = "my-org"
  vdc = "nsxt-vdc"

  name = "nsxt-gw"
}

data "vcd_nsxt_alb_edgegateway_service_engine_group" "test" {
  edge_gateway_id           = data.vcd_nsxt_edgegateway.existing.id
  service_engine_group_name = "known-service-engine-group-name"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which the edge gateway belongs. Optional if defined at provider level.
* `vdc` - (Optional) The name of VDC that owns the edge gateway. Optional if defined at provider level.
* `edge_gateway_id` - (Optional) An ID of NSX-T Edge Gateway. Can be lookup up using
  [vcd_nsxt_edgegateway](/providers/vmware/vcd/latest/docs/data-sources/nsxt_edgegateway) data source
* `service_engine_group_id` - (Required) An ID of NSX-T Service Engine Group. Can be looked up using
  [vcd_nsxt_alb_service_engine_group](/providers/vmware/vcd/latest/docs/data-sources/nsxt_alb_service_engine_group) data
  source. **Note** Either `service_engine_group_name` or `service_engine_group_id` require it.
* `service_engine_group_name` - (Optional) A Name of NSX-T Service Engine Group. **Note** Either
  `service_engine_group_name` or `service_engine_group_id` require it.

## Attribute Reference

All the arguments and attributes defined in
[`vcd_nsxt_alb_edgegateway_service_engine_group`](/providers/vmware/vcd/latest/docs/resources/nsxt_alb_edgegateway_service_engine_group)
resource are available.
