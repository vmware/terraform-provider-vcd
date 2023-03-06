---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_rde_type"
sidebar_current: "docs-vcd-resource-rde-type"
description: |-
   Provides the capability of creating, updating, and deleting Runtime Defined Entity types in VMware Cloud Director.
---

# vcd\_rde\_type

Provides the capability of creating, updating, and deleting Runtime Defined Entity types in VMware Cloud Director.
Requires system administrator privileges.

Supported in provider *v3.9+*

## Example Usage with a local schema file

```hcl
data "vcd_rde_interface" "my_interface" {
  vendor  = "bigcorp"
  nss     = "tech1"
  version = "1.2.3"
}

resource "vcd_rde_type" "my_rde_type" {
  vendor        = "vmware"
  nss           = "vcd"
  version       = "4.5.6"
  name          = "My VMware RDE Type"
  interface_ids = [data.vcd_rde_interface.my_interface.id]
  schema        = file("${path.module}/schemas/my-type-schema.json")
}
```

## Example Usage with a URL hosting the schema file

```hcl
data "vcd_rde_interface" "my_interface" {
  vendor  = "bigcorp"
  ns      = "tech1"
  version = "1.2.3"
}

resource "vcd_rde_type" "my_rde_type" {
  vendor        = "vmware"
  nss           = "vcd"
  version       = "4.5.6"
  name          = "My VMware RDE Type"
  interface_ids = [data.vcd_rde_interface.my_interface.id]
  schema_url    = "https://just.an-example.com/schemas/my-type-schema.json"
}
```

## Argument Reference

The following arguments are supported:

-> The 3-tuple of `vendor`, `nss` and `version` specifies a unique RDE Type.

* `vendor` - (Required) The vendor of the Runtime Defined Entity Type. Only alphanumeric characters, underscores and hyphens allowed.
* `nss` - (Required) A unique namespace associated with the Runtime Defined Entity Type. Only alphanumeric characters, underscores and hyphens allowed.
* `version` - (Required) The version of the Runtime Defined Entity Type. Must follow [semantic versioning](https://semver.org/) syntax.
* `name` - (Required) The name of the Runtime Defined Entity Type.
* `description` - (Optional) The description of the Runtime Defined Entity Type.
* `interface_ids` - (Optional) The set of [Defined Interfaces](/providers/vmware/vcd/latest/docs/resources/rde_interface) that this Runtime Defined Entity Type will use.
* `schema` - (Optional) A string that specifies a valid JSON schema. It can be retrieved with Terraform functions such as `file`, `templatefile`, etc. Either `schema` or `schema_url` is required.
* `schema_url` - (Optional) The URL that points to a valid JSON schema. Either `schema` or `schema_url` is required.
  If `schema_url` is used, the downloaded schema will be computed in the `schema` attribute.
* `external_id` - (Optional) An external entity's ID that this Runtime Defined Entity Type may apply to.
* `inherited_version` - (Optional) To be used when creating a new version of a Runtime Defined Entity Type.
  Specifies the version of the type that will be the template for the authorization configuration of the new version.
  The Type ACLs and the access requirements of the Type Behaviors of the new version will be copied from those of the inherited version.
  If not set, then the new type version will not inherit another version and will have the default authorization settings, just like the first version of a new type.

## Attribute Reference

The following attributes are supported:

* `readonly` - True if the Runtime Defined Entity Type cannot be modified.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing Runtime Defined Entity Type can be [imported][docs-import] into this resource via supplying its vendor, nss and version, which
unequivocally identifies it.
For example, using this structure, representing an existing Runtime Defined Entity Type that was **not** created using Terraform:

```hcl
resource "vcd_rde_type" "outer_rde_type" {
  vendor  = "bigcorp"
  nss     = "tech"
  version = "4.5.6"
}
```

You can import such Runtime Defined Entity Type into Terraform state using this command

```
terraform import vcd_rde_type.outer_rde_type bigcorp.tech.4.5.6
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the Runtime Defined Entity Type as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the Runtime Defined Entity Type's stored properties.
