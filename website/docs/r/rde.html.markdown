---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_rde"
sidebar_current: "docs-vcd-resource-rde"
description: |-
   Provides the capability of creating, updating, and deleting Runtime Defined Entities in VMware Cloud Director.
---

# vcd\_rde\_type

Provides the capability of creating, updating, and deleting Runtime Defined Entities in VMware Cloud Director.
Requires system administrator privileges.

Supported in provider *v3.9+*

## Example Usage with a schema file

```hcl
data "vcd_rde_type" "my-type" {
  vendor    = "bigcorp"
  namespace = "tech1"
  version   = "1.2.3"
}

resource "vcd_rde" "my-rde" {
  rde_type_id   = data.vcd_rde_type.my-type.id
  name          = "My custom RDE"
  resolve       = true
  entity        = file("${path.module}/entities/custom-rde.json")
}
```

## Example Usage with a URL that contains a schema file

```hcl
data "vcd_rde_type" "my-type" {
  vendor    = "bigcorp"
  namespace = "tech1"
  version   = "1.2.3"
}

resource "vcd_rde" "my-rde" {
  rde_type_id   = data.vcd_rde_type.my-type.id
  name          = "My custom RDE"
  resolve       = true
  entity_url    = "https://just.an-example.com/entities/custom-rde.json"
}
```

## Argument Reference

The following arguments are supported:

* `rde_type_id` - (Required) The ID of the type of the Runtime Defined Entity. You can use the [`vcd_rde_type`](/providers/vmware/vcd/latest/docs/data-sources/rde_type) data source to retrieve it.
* `name` - (Required) The name of the Runtime Defined Entity.
* `resolve` - (Required) If `true`, the Runtime Defined Entity will be resolved by this provider. If `false`, it won't be
  resolved and must be either done by an external component or with an update. The Runtime Defined Entity can't be
  deleted until the entity is resolved by either party.
* `entity` - (Optional) A string that specifies a valid JSON for the entity. It can be retrieved with functions such as `file`, `templatefile`... Either `entity` or `entity_url` is required.
* `entity_url` - (Optional) The URL that points to a valid JSON for the entity. Either `entity` or `entity_url` is required.
  If `entity_url` is used, the downloaded schema will be computed in the `entity` attribute.
* `external_id` - (Optional) An external entity's ID that this Runtime Defined Entity may have a relation to.
* `metadata_entry` - (Optional) A set of metadata entries to assign. See [Metadata](#metadata) section for details.

## Attribute Reference

The following attributes are supported:

* `owner_id` - The ID of the owner of this Runtime Defined Entity, corresponds to a [Organization user](/providers/vmware/vcd/latest/docs/resources/org_user).
* `org_id` - The ID of the [Organization](/providers/vmware/vcd/latest/docs/resources/org) to which the Runtime Defined Entity belongs.
* `state` - If the specified JSON in either `entity` or `entity_url` is correct, the state will be `RESOLVED`, otherwise it will be `RESOLUTION_ERROR`. If an entity in an `RESOLUTION_ERROR` state, it will require to be updated to a correct JSON to be usable.

<a id="metadata"></a>
## Metadata

The `metadata_entry` is a set of metadata entries that have the following structure:

* `key` - (Required) Key of this metadata entry.
* `value` - (Required) Value of this metadata entry.
* `type` - (Optional) Type of this metadata entry. One of: `StringEntry`, `NumberEntry`, `BoolEntry`. Defaults to `StringEntry`.
* `domain` - (Optional) Only meaningful for providers. Allows them to share entries with their tenants. Currently, accepted values are: `TENANT`, `PROVIDER`. Defaults to `TENANT`.
* `readonly` - (Optional) `true` if the metadata entry is read only. Defaults to `false`.

Example:

```hcl
resource "vcd_rde" "my-rde" {
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

!!!!!!!!!!!!!!!!! TODO !!!!!!!!!!!!!

~> Note: VCD allows to have many Runtime Defined Entities from a given type with the same name. Due to limitations in the
way that Terraform works, during an import, the chosen RDE will be the first one that VCD returns.

!!!! TODO: Maybe we can put a selector in the import chain????

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the Runtime Defined Entity as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the Runtime Defined Entity's stored properties.
