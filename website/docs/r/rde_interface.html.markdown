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

* `vendor` - (Required) The vendor of the RDE Interface.
* `nss` - (Required) A unique namespace associated with the RDE Interface.
* `version` - (Required) The version of the RDE Interface. Must follow [semantic versioning](https://semver.org/) syntax.
* `name` - (Required) The name of the RDE Interface.

## Attribute Reference

* `readonly` - Specifies if the RDE Interface can be only read.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing RDE Interface can be [imported][docs-import] into this resource via supplying its vendor, nss and version, which
unequivocally identifies it.
For example, using this structure, representing an existing RDE Interface that was **not** created using Terraform:

```hcl
resource "vcd_rde_interface" "outer_interface" {
  vendor    = "bigcorp"
  nss = "tech"
  version   = "4.5.6"
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
