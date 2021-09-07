---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_alb_controller"
sidebar_current: "docs-vcd-resource-nsxt-alb-controller"
description: |-
  Provides a resource to manage NSX-T ALB Controller for Providers. It helps to integrate VMware Cloud Director with NSX-T
  Advanced Load Balancer deployment. Controller instances are registered with VMware Cloud Director instance. Controller
  instances serve as a central control plane for the load-balancing services provided by NSX-T Advanced Load Balancer.
---

# vcd\_nsxt\_alb\_controller

Supported in provider *v3.4+* and VCD 10.2+ with NSX-T and ALB.

Provides a resource to manage NSX-T ALB Controller for Providers. It helps to integrate VMware Cloud Director with NSX-T
Advanced Load Balancer deployment. Controller instances are registered with VMware Cloud Director instance. Controller
instances serve as a central control plane for the load-balancing services provided by NSX-T Advanced Load Balancer.

~> Only `System Administrator` can create this resource.

## Example Usage (Adding NSX-T ALB Controller to provider)

```hcl
resource "vcd_nsxt_alb_controller" "first" {
  name         = "aviController1"
  description  = "first alb controller"
  url          = "https://my.controller"
  username     = "admin"
  password     = "CHANGE-ME"
  license_type = "ENTERPRISE"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for NSX-T ALB Controller
* `description` - (Optional) An optional description NSX-T ALB Controller
* `url` - (Required) The URL of ALB Controller
* `username` - (Required) The username for ALB Controller
* `password` - (Required) The password for ALB Controller. Password will not be refreshed.
* `license_type` - (Required) License type of ALB Controller (`ENTERPRISE` or `BASIC`)


## Attribute Reference

The following attributes are exported on this resource:

* `version` - ALB Controller version (e.g. 20.1.3)


## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing NSX-T ALB Controller configuration can be [imported][docs-import] into this resource
via supplying path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxt_alb_controller.imported my-controller-name
```

The above would import the `my-controller-name` NSX-T ALB controller settings that are defined at provider level.