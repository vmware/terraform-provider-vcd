---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_alb_edgegateway_service_engine_group"
sidebar_current: "docs-vcd-resource-nsxt-alb-edge-service-engine-group"
description: |-
Provides a resource to manage NSX-T ALB Service Engine Group assignment to NSX-T Edge Gateway.
---

# vcd\_nsxt\_alb\_edgegateway\_service\_engine\_group

Supported in provider *v3.5+* and VCD 10.2+ with NSX-T and ALB.

Provides a resource to manage NSX-T ALB Service Engine Group assignment to NSX-T Edge Gateway.

~> Only `System Administrator` can create this resource.

## Example Usage (Enabling NSX-T ALB on NSX-T Edge Gateway)

```hcl
data "vcd_nsxt_edgegateway" "existing" {
  org  = "my-org"
  vdc  = "nsxt-vdc"

  name = "nsxt-gw"
}

resource "vcd_nsxt_alb_edgegateway_service_engine_group" "test" {
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  service_engine_group_id = vcd_nsxt_alb_service_engine_group.first.id
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which the edge gateway belongs. Optional if defined at provider level.
* `vdc` - (Optional) The name of VDC that owns the edge gateway. Optional if defined at provider level.
* `edge_gateway_id` - (Required) An ID of NSX-T Edge Gateway. Can be lookup up using
  [vcd_nsxt_edgegateway](/providers/vmware/vcd/latest/docs/data-sources/nsxt_edgegateway) data source
* `service_engine_group_id` - (Required) An ID of NSX-T Service Engine Group. Can be looked up using
  [vcd_nsxt_alb_service_engine_group](/providers/vmware/vcd/latest/docs/data-sources/nsxt_alb_service_engine_group) data source
* `max_virtual_services` - (Optional) Maximum amount of Virtual Services to run on this Service Engine Group. **Only for
  Shared Service Engine Groups**.
* `reserved_virtual_services` - (Optional) Number of reserved Virtual Services for this Edge Gateway. **Only for Shared
  Service Engine Groups.**

## Attribute reference

* `deployed_virtual_services` -  Number of deployed Virtual Services on this Service Engine Group.

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing NSX-T Edge Gateway ALB Service Engine Group assignment configuration can be [imported][docs-import] into
this resource via supplying
path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxt_alb_settings.imported my-org.my-vdc.my-nsxt-edge-gateway-name.service-engine-group-name
```

The above would import the NSX-T Edge Gateway ALB Service Engine Group assignment configuration for Service Engine Group
Name `service-engine-group-name` on  Edge Gateway named `my-nsxt-edge-gateway-name` in Org `my-org`
and VDC `my-vdc`.