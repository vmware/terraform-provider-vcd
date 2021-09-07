---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_alb_cloud"
sidebar_current: "docs-vcd-resource-nsxt-alb-cloud"
description: |-
  Provides a resource to manage NSX-T ALB Clouds for Providers. An NSX-T Cloud is a service provider-level construct that
  consists of an NSX-T Manager and an NSX-T Data Center transport zone.
---

# vcd\_nsxt\_alb\_cloud

Supported in provider *v3.4+* and VCD 10.2+ with NSX-T and ALB.

Provides a resource to manage NSX-T ALB Clouds for Providers. An NSX-T Cloud is a service provider-level construct that
consists of an NSX-T Manager and an NSX-T Data Center transport zone.

~> Only `System Administrator` can create this resource.

## Example Usage (Adding NSX-T ALB Cloud)

```hcl
data "vcd_nsxt_alb_controller" "main" {
  name = "aviController1"
}

data "vcd_nsxt_alb_importable_cloud" "cld" {
  name          = "NSXT Importable Cloud"
  controller_id = vcd_nsxt_alb_controller.first.id
}

resource "vcd_nsxt_alb_cloud" "first" {
  name        = "nsxt-cloud"
  description = "NSX-T ALB Cloud"

  controller_id       = data.vcd_nsxt_alb_controller.main.id
  importable_cloud_id = data.vcd_nsxt_alb_importable_cloud.cld.id
  network_pool_id     = data.vcd_nsxt_alb_importable_cloud.cld.network_pool_id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for NSX-T ALB Cloud
* `description` - (Optional) An optional description NSX-T ALB Cloud
* `controller_id` - (Required) ALB Controller ID
* `importable_cloud_id` - (Required) Importable Cloud ID. Can be looked up using `vcd_nsxt_alb_importable_cloud` data
  source
* `network_pool_id` - (Required) Network pool ID for ALB Cloud. Can be looked up using `vcd_nsxt_alb_importable_cloud` data
  source


## Attribute Reference

The following attributes are exported on this resource:

* `health_status` - HealthStatus contains status of the Load Balancer Cloud. Possible values are:
  * UP - The cloud is healthy and ready to enable Load Balancer for an Edge Gateway
  * DOWN - The cloud is in a failure state. Enabling Load balancer on an Edge Gateway may not be possible
  * RUNNING - The cloud is currently processing. An example is if it's enabling a Load Balancer for an Edge Gateway
  * UNAVAILABLE - The cloud is unavailable
  * UNKNOWN - The cloud state is unknown
* `health_message` - DetailedHealthMessage contains detailed message on the health of the Cloud
* `network_pool_name` - Network Pool Name used by the Cloud


## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing NSX-T ALB Cloud configuration can be [imported][docs-import] into this resource
via supplying path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxt_alb_cloud.imported my-alb-cloud-name
```

The above would import the `my-alb-cloud-name` NSX-T ALB cloud settings that are defined at provider level.