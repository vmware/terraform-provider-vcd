---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_rde_interface"
sidebar_current: "docs-vcd-resource-rde-interface"
description: |-
   Provides the capability of creating, updating, and deleting Defined Interfaces in VMware Cloud Director.
---

# vcd\_rde\_interface

Provides the capability of creating, updating, and deleting Defined Interfaces in VMware Cloud Director.
Creating, updating and deleting Defined Interfaces require System administrator privileges.

Supported in provider *v3.9+*

## Example Usage

```hcl
resource "vcd_rde_interface" "my_interface" {
  vendor    = "bigcorp"
  namespace = "tech"
  version   = "1.2.3"
  name      = "BigCorp Interface"
}
```

## Argument Reference

The following arguments are supported:

* `vendor` - (Required) The vendor of the Defined Interface.
* `namespace` - (Required) A unique namespace associated with the Defined Interface.
* `version` - (Required) The version of the Defined Interface. Must follow [semantic versioning](https://semver.org/) syntax.
* `name` - (Required) The name of the Defined Interface.

## Attribute Reference

* `readonly` - Specifies if the Defined Interface can be only read.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing Defined Interface can be [imported][docs-import] into this resource via supplying its vendor, namespace and version, which
unequivocally identifies it.
For example, using this structure, representing an existing Defined Interface that was **not** created using Terraform:

```hcl
resource "vcd_rde_interface" "outer_interface" {
  vendor    = "bigcorp"
  namespace = "tech"
  version   = "4.5.6"
}
```

You can import such Defined Interface into Terraform state using this command

```
terraform import vcd_rde_interface.outer_interface bigcorp.tech.4.5.6
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the Defined Interface as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the Defined Interface's stored properties.
