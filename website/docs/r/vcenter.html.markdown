---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_vcenter"
sidebar_current: "docs-vcd-resource-vcenter"
description: |-
  Provides a resource to manage vCenters.
---

# vcd\_nsxt\_vcenter

Provides a resource to manage vCenters.

~> Only `System Administrator` can create this resource.

## Example Usage

```hcl
resource "vcd_vcenter" "test" {
  name                   = "TestAccVcdTmVcenter-rename"
  url                    = "https://host:443"
  auto_trust_certificate = true
  username               = "admim@vsphere.local"
  password               = "CHANGE-ME"
  is_enabled             = true
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for vCenter server
* `description` - (Optional) An optional description for vCenter server
* `username` - (Required) A username for authenticating to vCenter server
* `password` - (Required) A password for authenticating to vCenter server
* `url` - (Required) An URL of vCenter server
* `auto_trust_certificate` - (Required) Defines if the certificate of a given vCenter server should
  automatically be added to trusted certificate store. **Note:** not having the certificate trusted
  will cause malfunction.
* `is_enabled` - (Optional) Defines if the vCenter is enabled. Default `true`. The vCenter must
  always be disabled before removal (this resource will disable it automatically on destroy).

## Attribute Reference

The following attributes are exported on this resource:

* `max_virtual_services` - Maximum number of virtual services this NSX-T ALB Service Engine Group can run

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing vCenter configuration can be [imported][docs-import] into this resource via supplying
path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_vcenter.imported my-vcenter
```

The above would import the `my-vcenter` vCenter settings that are defined at provider level.