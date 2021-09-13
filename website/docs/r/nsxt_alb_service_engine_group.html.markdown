---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_alb_service_engine_group"
sidebar_current: "docs-vcd-resource-nsxt-alb-service-engine-group"
description: |-
  Provides a resource to manage NSX-T ALB Service Engine Groups. A Service Engine Group is an isolation domain that also
  defines shared service engine properties, such as size, network access, and failover. Resources in a service engine
  group can be used for different virtual services, depending on your tenant needs. These resources cannot be shared
  between different service engine groups.
---

# vcd\_nsxt\_alb\_service\_engine\_group

Supported in provider *v3.4+* and VCD 10.2+ with NSX-T and ALB.

Provides a resource to manage NSX-T ALB Service Engine Groups. A Service Engine Group is an isolation domain that also
defines shared service engine properties, such as size, network access, and failover. Resources in a service engine
group can be used for different virtual services, depending on your tenant needs. These resources cannot be shared
between different service engine groups.

~> Only `System Administrator` can create this resource.

## Example Usage (Adding NSX-T ALB Service Engine Group)

```hcl
# Local variable is used to avoid direct reference and cover Terraform core bug https://github.com/hashicorp/terraform/issues/29484
# Even changing NSX-T ALB Controller name in UI, plan will cause to recreate all resources depending 
# on `vcd_nsxt_alb_importable_cloud` data source if this indirect reference (via local) variable is not used.
locals {
  controller_id = vcd_nsxt_alb_controller.first.id
}

resource "vcd_nsxt_alb_controller" "first" {
  name         = "aviController1"
  description  = "first alb controller"
  url          = "https://controller.myXZ"
  username     = "admin"
  password     = "Welcome@1234"
  license_type = "ENTERPRISE"
}

data "vcd_nsxt_alb_importable_cloud" "cld" {
  name          = "importable-cloud-name"
  controller_id = local.controller_id
}

resource "vcd_nsxt_alb_cloud" "first" {
  name        = "nsxt-cloud"
  description = "first alb cloud"

  controller_id       = vcd_nsxt_alb_controller.first.id
  importable_cloud_id = data.vcd_nsxt_alb_importable_cloud.cld.id
  network_pool_id     = data.vcd_nsxt_alb_importable_cloud.cld.network_pool_id
}

resource "vcd_nsxt_alb_service_engine_group" "first" {
  name                      = "demo-service-engine"
  description               = "Service Engine for Terraform documentation"
  alb_cloud_id              = vcd_nsxt_alb_cloud.first.id
  service_engine_group_name = "Default-Group"
  reservation_model         = "SHARED"
  sync_on_refresh           = false
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for NSX-T ALB Service Engine Group
* `description` - (Optional) An optional description NSX-T ALB Service Engine Group
* `alb_cloud_id` - (Required) A reference NSX-T ALB Cloud. Can be looked up using `vcd_nsxt_alb_cloud` resource or data
  source
* `reservation_model` - (Required) Definition if the Service Engine Group is `DEDICATED` or `SHARED`
* `service_engine_group_name` - (Required) Name of available Service Engine Group in ALB
* `sync_on_refresh` (Optional) - A special argument that is not passed to VCD, but alters behaviour of this resource so
  that it performs a Sync operation on every Terraform refresh. *Note* this may impact refresh performance, but should
  ensure up-to-date information is read. Default is **false**.

## Attribute Reference

The following attributes are exported on this resource:

* `max_virtual_services` - Maximum number of virtual services this NSX-T ALB Service Engine Group can run
* `reserved_virtual_services` - Number of reserved virtual services
* `deployed_virtual_services` - Number of deployed virtual services
* `ha_mode` defines High Availability Mode for Service Engine Group. One off:
  * ELASTIC_N_PLUS_M_BUFFER - Service Engines will scale out to N active nodes with M nodes as buffer.
  * ELASTIC_ACTIVE_ACTIVE - Active-Active with scale out.
  * LEGACY_ACTIVE_STANDBY - Traditional single Active-Standby configuration
* `overallocated` - Boolean value stating if there are more deployed virtual services than allocated ones

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing NSX-T ALB Service Engine Group configuration can be [imported][docs-import] into this resource
via supplying path for it. An example is
below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxt_alb_service_engine_group.imported my-service-engine-group-name
```

The above would import the `my-service-engine-group-name` NSX-T ALB controller settings that are defined at provider
level.
