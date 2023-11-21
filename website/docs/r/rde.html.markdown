---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_rde"
sidebar_current: "docs-vcd-resource-rde"
description: |-
   Provides the capability of creating, updating, and deleting Runtime Defined Entities in VMware Cloud Director.
---

# vcd\_rde

Provides the capability of creating, updating, and deleting Runtime Defined Entities in VMware Cloud Director.

-> VCD allows to have multiple RDEs of the same [RDE Type](/providers/vmware/vcd/latest/docs/resources/rde_type) with
the same name, meaning that they would be only distinguishable by their ID. This could lead to potential issues when fetching
a unique RDE with the data source, so take this trait into account when creating them.

Supported in provider *v3.9+*

## Example Usage with a JSON file

```hcl
data "vcd_rde_type" "my_type" {
  vendor    = "bigcorp"
  namespace = "tech1"
  version   = "1.2.3"
}

resource "vcd_rde" "my_rde" {
  org          = "my-org"
  rde_type_id  = data.vcd_rde_type.my-type.id
  name         = "My custom RDE"
  resolve      = true
  input_entity = file("${path.module}/entities/custom-rde.json")
}

output "computed_rde" {
  value = vcd_rde.my_rde.computed_entity
}
```

## Example Usage with a JSON template

Using the [`templatefile`](https://developer.hashicorp.com/terraform/language/functions/templatefile) Terraform function will
allow you to parameterize RDE creation with custom inputs, as follows:

```hcl
data "vcd_rde_type" "my_type" {
  vendor    = "bigcorp"
  namespace = "tech1"
  version   = "1.2.3"
}

resource "vcd_rde" "my_rde" {
  org         = "my-org"
  rde_type_id = data.vcd_rde_type.my-type.id
  name        = "My custom RDE"
  resolve     = true
  # Functions are evaluated before the dependency tree is calculated, so the file must exist and not be a reference to
  # a created Terraform resource.
  input_entity = templatefile("${path.module}/entities/custom-rde.json", {
    name          = var.name
    custom_field  = "This one is hardcoded"
    another_field = var.anoter_field
    replicas      = 2
  })
}

output "computed_rde" {
  value = vcd_rde.my_rde.computed_entity
}
```

## Example Usage with a URL that contains a schema file

```hcl
data "vcd_rde_type" "my_type" {
  vendor    = "bigcorp"
  namespace = "tech1"
  version   = "1.2.3"
}

resource "vcd_rde" "my-rde" {
  org         = "my-org"
  rde_type_id = data.vcd_rde_type.my-type.id
  name        = "My custom RDE"
  resolve     = true
  entity_url  = "https://just.an-example.com/entities/custom-rde.json"
}
```

## Example of Upgrade of the RDE Type Version

```hcl
data "vcd_rde_type" "my_type" {
  vendor    = "bigcorp"
  namespace = "tech1"
  version   = "1.0.0"
}

data "vcd_rde_type" "my_updated_type" {
  vendor    = "bigcorp"
  namespace = "tech1"
  version   = "1.1.0"
}

resource "vcd_rde" "my-rde" {
  org = "my-org"
  # Update from 'data.vcd_rde_type.my_type.id' to 'data.vcd_rde_type.my_updated_type.id' to upgrade the RDE Type version
  rde_type_id = data.vcd_rde_type.my_updated_type.id
  name        = "My custom RDE"
  resolve     = true # This will attempt to resolve after the version is updated
  entity_url  = "https://just.an-example.com/entities/custom-rde.json"
}

```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) Name of the [Organization](/providers/vmware/vcd/latest/docs/resources/org) that will own the RDE, optional if defined at provider level.
* `rde_type_id` - (Required) The ID of the [RDE Type](/providers/vmware/vcd/latest/docs/data-sources/rde_type) to instantiate. It only supports
  updating to a **newer/lower** `version` of the **same** RDE Type.
* `name` - (Required) The name of the Runtime Defined Entity. It can be non-unique.
* `resolve` - (Required) If `true`, the Runtime Defined Entity will be resolved by this provider. If `false`, it won't be
  resolved and must be done either by an external component action or by an update. The Runtime Defined Entity can't be
  deleted until the input_entity is resolved by either party, unless `resolve_on_removal=true`. See [RDE resolution](#rde-resolution) for more details.
* `resolve_on_removal` - (Optional) If `true`, the Runtime Defined Entity will be resolved before it gets deleted, to ensure forced deletion. Destroy will fail if it is not resolved. It is `false` by default.
* `input_entity` - (Optional) A string that specifies a valid JSON for the RDE. It can be retrieved with functions such as `file`, `templatefile`... Either `input_entity` or `input_entity_url` is required.
* `input_entity_url` - (Optional) The URL that points to a valid JSON for the RDE. Either `input_entity` or `input_entity_url` is required.
  The referenced JSON will be downloaded on every read operation, and it will break Terraform operations if these contents are no longer present on the remote site.
  If you can't guarantee this, it is safer to use `input_entity`.
* `external_id` - (Optional) An external input_entity's ID that this Runtime Defined Entity may have a relation to.
* `metadata_entry` - (Optional; *v3.11+*) A set of metadata entries to assign. See [Metadata](#metadata) section for details.

## Attribute Reference

The following attributes are supported:

* `computed_entity` - The real state of this RDE in VCD. See [Input entity vs Computed entity](#input-entity-vs-computed-entity) below for details.
* `entity_in_sync` - It's `true` when `computed_entity` is equal to either `input_entity` or the contents of `input_entity_url`,
  meaning that the computed RDE retrieved from VCD is synchronized with the input RDE.
* `owner_user_id` - The ID of the [Organization user](/providers/vmware/vcd/latest/docs/resources/org_user) that owns this Runtime Defined Entity.
* `org_id` - The ID of the [Organization](/providers/vmware/vcd/latest/docs/resources/org) to which the Runtime Defined Entity belongs.
* `state` - Specifies whether the entity is correctly resolved or not. When created it will be in `PRE_CREATED` state.
  If the entity is correctly validated against its [RDE Type](/providers/vmware/vcd/latest/docs/resources/rde_type) schema, the state will be `RESOLVED`,
  otherwise it will be `RESOLUTION_ERROR`.

<a id="input-entity-vs-computed-entity"></a>
## Input entity vs Computed entity

There is a common use case for RDEs where they are used by 3rd party components that perform continuous updates on them,
which are expected and desired. This conflicts with Terraform way of working, as doing a `terraform apply` would then
perform actions to remove every single change done by those 3rd party tools, which we don't want in this case.

To add compatibility with this scenario, there are two important arguments, `input_entity`/`input_entity_url`,
and two important computed attribute, `computed_entity` and `entity_in_sync`.

If your RDE is intended to be managed **only and exclusively** by Terraform, the contents of the input JSON should
always match with those retrieved into `computed_entity`, and this will be reflected in the `entity_in_sync` attribute,
which should be always `true`.

Otherwise, only `computed_entity` will reflect the current state of the RDE in VCD and `entity_in_sync` will be `false`, whereas
`input_entity` and `input_entity_url` will only specify the RDE contents that were used either on creation or in a deliberate
update that will cause the RDE contents to be **completely overridden**.

As per this last point, one needs to be careful when updating `input_entity` or `input_entity_url`, as Terraform will apply
whatever changes were done, ignoring the real state from `computed_entity`. To perform a real update, one
needs to check the contents of the `computed_entity` and do some diff with the original input.

In other words:

~> When you want to update an RDE and `entity_in_sync` is `false`, you should always merge the contents
of `computed_entity` and `input_entity` to avoid overriding the whole entity by mistake with an old value.

<a id="rde-resolution"></a>
## RDE resolution

When a RDE is created, its `state` will be `PRE_CREATED`, which means that the entity JSON was not validated against the
[RDE Type](/providers/vmware/vcd/latest/docs/resources/rde_type) schema. After resolution, `state` should be either `RESOLVED`
or `RESOLUTION_ERROR` if the input JSON doesn't match the schema.

The RDE must be eventually resolved to be used or deleted, and this operation can be done either by Terraform with
`resolve=true`, or by a 3rd party actor that will do it behind the scenes at some point (in this case, the Terraform resource
should have `resolve=false` to avoid being resolved).
In this last scenario, it is advisable to mark `resolve_on_removal=true` so Terraform can delete the RDE even if it was not
resolved by anyone.

<a id="metadata"></a>
## Metadata

The `metadata_entry` is a set of metadata entries that have the following structure:

* `key` - (Required) Key of this metadata entry.
* `value` - (Required) Value of this metadata entry.
* `type` - (Optional) Type of this metadata entry. One of: `StringEntry`, `NumberEntry`, `BoolEntry`. Defaults to `StringEntry`.
* `domain` - (Optional) Only meaningful for providers. Allows them to share entries with their tenants. Currently, accepted values are: `TENANT`, `PROVIDER`. Defaults to `TENANT`.
* `readonly` - (Optional) `true` if the metadata entry is read only. Defaults to `false`.
* `persistent` - (Optional) `true` if the metadata is persistent. Persistent entries can be copied over on some entity operation
  (e.g. Creating a copy of a VDC, capturing a vApp to a template, instantiating a catalog item as a VM...). Defaults to `false`.
* `id` - (Computed) Read-only identifier for this metadata entry.

Example:

```hcl
resource "vcd_rde" "my-rde" {
  org         = "my-org"
  rde_type_id = data.vcd_rde_type.my-type.id
  name        = "My custom RDE"
  resolve     = true
  entity_url  = "https://just.an-example.com/entities/custom-rde.json"
  metadata_entry {
    key      = "foo"
    type     = "StringEntry"
    value    = "bar"
    domain   = "TENANT"
    readonly = true
  }
  metadata_entry {
    key      = "bar"
    type     = "NumberEntry"
    value    = "42"
    domain   = "TENANT"
    readonly = true
  }
}
```

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

-> Note: VCD allows to have many Runtime Defined Entities from a given type with the same name. The only way to differentiate
them is with their unique ID.

An existing Runtime Defined Entity can be [imported][docs-import] into this resource via supplying its `vendor`, `nss`,
`version` and `name`. As this can identify not only one RDE but **many**, a `position` is also needed in the import process.
For example, using this structure, representing an existing Runtime Defined Entity that was **not** created using Terraform:

```hcl
resource "vcd_rde" "outer_rde" {
  rde_type_id = data.my_rde_type.id
  name        = "foo"
  # ...
}
```

You can import such Runtime Defined Entity into Terraform state using this command

```
terraform import vcd_rde.outer_rde bigcorp.tech.4.5.6.foo.1
```

Where `vendor=bigcorp`, `nss=tech`, `version=4.5.6`, `name=foo` and we want the first retrieved RDE (`position=1`) in case
there's more than one with that combination of type parameters and name.

To know how many RDEs are available in VCD with the given combination of type parameters and name, one can do:

```
terraform import vcd_rde.outer_rde list@bigcorp.tech.4.5.6.foo
```
It will return a list of IDs. Then one can import again specifying the position, or directly with the ID:

```
terraform import vcd_rde.outer_rde urn:vcloud:entity:bigcorp:tech:a074f9e9-5d76-4f1e-8c37-f4e8b28e51ff
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the Runtime Defined Entity as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the Runtime Defined Entity's stored properties.
