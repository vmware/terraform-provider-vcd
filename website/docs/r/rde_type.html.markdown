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

## Example Usage with a schema file

```hcl
data "vcd_rde_interface" "my-interface" {
  vendor    = "bigcorp"
  namespace = "tech1"
  version   = "1.2.3"  
}

resource "vcd_rde_type" "my-rde-type" {
  vendor        = "vmware"
  namespace     = "vcd"
  version       = "4.5.6"
  name          = "My VMware RDE Type"
  interface_ids = [ data.vcd_rde_interface.my-interface.id ]
  schema        = file("${path.module}/schemas/my-type-schema.json")
}
```

## Example Usage with a URL that contains a schema file

```hcl
data "vcd_rde_interface" "my-interface" {
  vendor    = "bigcorp"
  namespace = "tech1"
  version   = "1.2.3"  
}

resource "vcd_rde_type" "my-rde-type" {
  vendor        = "vmware"
  namespace     = "vcd"
  version       = "4.5.6"
  name          = "My VMware RDE Type"
  interface_ids = [ data.vcd_rde_interface.my-interface.id ]
  schema_url    = "https://just.an-example.com/schemas/my-type-schema.json"
}
```

## Argument Reference

The following arguments are supported:

* `vendor` - (Required) The vendor of the Runtime Defined Entity type.
* `namespace` - (Required) A unique namespace associated with the Runtime Defined Entity type.
* `version` - (Required) The version of the Runtime Defined Entity type. Must follow [semantic versioning](https://semver.org/) syntax.
* `interface_ids` - (Required) The set of [Defined Interfaces](/providers/vmware/vcd/latest/docs/resources/rde_interface) that this Runtime Defined Entity type will use.
* `name` - (Required) The name of the Runtime Defined Entity type.
* `description` - (Optional) The description of the Runtime Defined Entity type.
* `schema` - (Optional) A string that specifies a valid JSON schema. It can be retrieved with functions such as `file`, `templatefile`... Either `schema` or `schema_url` is required.
* `schema_url` - (Optional) The URL that points to a valid JSON schema. Either `schema` or `schema_url` is required.
* `external_id` - (Optional) An external entity's id that this Runtime Defined Entity type may apply to.
* `inherited_version` - (Optional) To be used when creating a new version of a Runtime Defined Entity type.
  Specifies the version of the type that will be the template for the authorization configuration of the new version.
  The Type ACLs and the access requirements of the Type Behaviors of the new version will be copied from those of the inherited version.
  If not set, then the new type version will not inherit another version and will have the default authorization settings, just like the first version of a new type.

## Attribute Reference

The following attributes are supported:

* `readonly` - True if the Runtime Defined Entity type cannot be modified.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing Runtime Defined Entity type can be [imported][docs-import] into this resource via supplying its vendor, namespace and version, which
unequivocally identifies it.
For example, using this structure, representing an existing Runtime Defined Entity type that was **not** created using Terraform:

```hcl
resource "vcd_rde_type" "outer-rde-type" {
  vendor    = "bigcorp"
  namespace = "tech"
  version   = "4.5.6"
}
```

You can import such Runtime Defined Entity type into Terraform state using this command

```
terraform import vcd_rde_type.outer-rde-type bigcorp.tech.4.5.6
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the Runtime Defined Entity type as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the Runtime Defined Entity type's stored properties.
