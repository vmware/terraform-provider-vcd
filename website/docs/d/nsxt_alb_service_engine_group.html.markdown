---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_alb_service_engine_group"
sidebar_current: "docs-vcd-datasource-nsxt-alb-service-engine-group"
description: |-
  Provides a data source to read NSX-T ALB Service Engine Groups. A Service Engine Group is an isolation domain that also
  defines shared service engine properties, such as size, network access, and failover. Resources in a service engine
  group can be used for different virtual services, depending on your tenant needs. These resources cannot be shared
  between different service engine groups.
---

# vcd\_nsxt\_alb\_service\_engine\_group

Supported in provider *v3.4+* and VCD 10.2+ with NSX-T and ALB.

Provides a data source to read NSX-T ALB Service Engine Groups. A Service Engine Group is an isolation domain that also
defines shared service engine properties, such as size, network access, and failover. Resources in a service engine
group can be used for different virtual services, depending on your tenant needs. These resources cannot be shared
between different service engine groups.

~> Only `System Administrator` can use this data source.

## Example Usage

```hcl
data "vcd_nsxt_alb_service_engine_group" "demo" {
  name            = "configured-service-engine-group"
  sync_on_refresh = false
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required)  - Name of Service Engine Group.
* `sync_on_refresh` (Optional) - A special argument that is not passed to VCD, but alters behaviour of this resource so
  that it performs a Sync operation on every Terraform refresh. *Note* this may impact refresh performance, but should
  ensure up-to-date information is read. Default is **false**.

## Attribute Reference

All the arguments and attributes defined in
[`vcd_nsxt_alb_service_engine_group`](/providers/vmware/vcd/latest/docs/resources/nsxt_alb_service_engine_group)
resource are available.
