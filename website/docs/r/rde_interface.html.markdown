---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_rde_interface"
sidebar_current: "docs-vcd-resource-rde-interface"
description: |-
   Provides the capability of creating, updating, and deleting Runtime Defined Entity Interfaces in VMware Cloud Director.
---

# vcd\_rde\_interface

Provides the capability of creating, updating, and deleting Runtime Defined Entity Interfaces in VMware Cloud Director.

A Runtime Defined Entity Interface is specified unequivocally by 3 elements: `vendor`, `nss` and `version`. This
3-tuple must be unique. See the examples section for more details.

-> Creating, updating and deleting RDE Interfaces requires System administrator privileges.

Supported in provider *v3.9+*

## Example Usage

```hcl
resource "vcd_rde_interface" "my_interface1" {
  vendor  = "bigcorp"
  nss     = "tech"
  version = "1.2.3"
  name    = "BigCorp Interface"
}

resource "vcd_rde_interface" "my_interface2" {
  vendor  = "bigcorp"
  nss     = "tech"
  version = "1.2.4"
  name    = "Another BigCorp Interface"
}
```

The second interface is valid because the version is different. If we set `version = "1.2.3"` in `my_interface2`,
creation would fail despite having a different name, as the 3-tuple `vendor`, `nss` and `version` would be the same.

## Argument Reference

The following arguments are supported:

-> The 3-tuple of `vendor`, `nss` and `version` specifies a unique RDE Interface.

* `vendor` - (Required) The vendor of the RDE Interface. Only alphanumeric characters, underscores and hyphens allowed.
* `nss` - (Required) A unique namespace associated with the RDE Interface. Only alphanumeric characters, underscores and hyphens allowed.
* `version` - (Required) The version of the RDE Interface. Must follow [semantic versioning](https://semver.org/) syntax.
* `name` - (Required) The name of the RDE Interface.
* `behavior` - (Optional; *v3.10+*) A block that defines a RDE Interface Behavior. Only System Administrators can read and manage Behaviors.
  See [Behaviors](#behaviors) for more information.

## Attribute Reference

* `readonly` - Specifies if the RDE Interface can be only read.

## Behaviors

-> Available since v3.10+. Only System Administrators can manage Behaviors from RDE Interfaces.

You can define Behaviors on a RDE Interface with one or more of the `behavior` blocks, which contain the following
attributes:

~> If the RDE Interface is being used by one or more [RDE Types](/providers/vmware/vcd/latest/docs/resources/rde_types),
then Behaviors **can't be added or removed**, and only the `execution` attribute can be updated for existing ones.
Keep this in mind when creating a RDE Interface with Behaviors.

* `name` - (Required) Name of the Behavior
* `execution` - (Required) A map that defines the execution elements of the Behavior
* `description` - (Required) A description specifying the contract of the Behavior
* `ìd` - ID of the Behavior. This is auto-generated and read-only
* `ref` - The Behavior invocation reference to be used for polymorphic behavior invocations. This is auto-generated and read-only

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing RDE Interface can be [imported][docs-import] into this resource via supplying its `vendor`, `nss` and `version`, which
unequivocally identifies it.
For example, using this structure, representing an existing RDE Interface that was **not** created using Terraform:

```hcl
resource "vcd_rde_interface" "outer_interface" {
  vendor  = "bigcorp"
  nss     = "tech"
  version = "4.5.6"
}
```

You can import such RDE Interface into Terraform state using this command

```
terraform import vcd_rde_interface.outer_interface bigcorp.tech.4.5.6
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the RDE Interface as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the RDE Interface's stored properties.
