---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_manager"
sidebar_current: "docs-vcd-resource-nsxt-manager"
description: |-
  Provides a resource to manage NSX-T Managers.
---

# vcd\_nsxt\_manager

Provides a resource to manage NSX-T Managers.

~> Only `System Administrator` can create this resource.

## Example Usage

```hcl
resource "vcd_nsxt_manager" "test" {
  name                   = "TestAccVcdTmNsxtManager"
  description            = "terraform test"
  username               = "admin"
  password               = "CHANGE-ME"
  url                    = "https://HOST"
  auto_trust_certificate = true
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for NSX-T Manager
* `description` - (Optional) An optional description for NSX-T Manager
* `username` - (Required) A username for authenticating to NSX-T Manager
* `password` - (Required) A password for authenticating to NSX-T Manager
* `url` - (Required) An URL of NSX-T Manager
* `auto_trust_certificate` - (Required) Defines if the certificate of a given NSX-T Manager should
  automatically be added to trusted certificate store. **Note:** not having the certificate trusted
  will cause malfunction.
* `network_provider_scope` - (Optional) The network provider scope is the tenant facing name for the
  NSX Manager.

## Attribute Reference

The following attributes are exported on this resource:

* `status` - Status of NSX-T Manager. One of:
 * `PENDING` - Desired entity configuration has been received by system and is pending realization.
 * `CONFIGURING` - The system is in process of realizing the entity.
 * `REALIZED` - The entity is successfully realized in the system.
 * `REALIZATION_FAILED` - There are some issues and the system is not able to realize the entity.
 * `UNKNOWN` - Current state of entity is unknown.

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing NSX-T Manager configuration can be [imported][docs-import] into this resource via
supplying path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxt_manager.imported my-nsxt-manager
```

The above would import the `my-nsxt-manager` NSX-T Manager settings that are defined at provider
level.
