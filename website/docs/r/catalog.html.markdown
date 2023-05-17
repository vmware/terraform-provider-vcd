---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_catalog"
sidebar_current: "docs-vcd-resource-catalog"
description: |-
  Provides a VMware Cloud Director catalog resource. This can be used to create and delete a catalog.
---

# vcd\_catalog

Provides a VMware Cloud Director catalog resource. This can be used to create and delete a catalog.

Supported in provider *v2.0+*

## Example Usage

```hcl
resource "vcd_catalog" "myNewCatalog" {
  org = "my-org"

  name             = "my-catalog"
  description      = "catalog for files"
  delete_recursive = true
  delete_force     = true
}
```

## Example Usage (Custom storage profile)

```hcl

data "vcd_storage_profile" "sp1" {
  org  = "my-org"
  vdc  = "my-vdc"
  name = "ssd-storage-profile"
}

resource "vcd_catalog" "myNewCatalog" {
  org = "my-org"

  name               = "my-catalog"
  description        = "catalog for files"
  storage_profile_id = data.vcd_storage_profile.sp1.id

  delete_recursive = true
  delete_force     = true
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organizations. 
   When using a catalog shared from another organization, this field must have the name of that one, not the current one.
   If you don't know the name of the sharing org, and put the current one, an error message will list the possible names.
* `name` - (Required) Catalog name
* `description` - (Optional) Description of catalog
* `storage_profile_id` - (Optional, *v3.1+*) Allows to set specific storage profile to be used for catalog. **Note.** Data
source [vcd_storage_profile](/providers/vmware/vcd/latest/docs/data-sources/storage_profile) can help to lookup storage profile ID.
* `delete_recursive` - (Required) When destroying use delete_recursive=True to remove the catalog and any objects it contains that are in a state that normally allows removal
* `delete_force` - (Required) When destroying use `delete_force=true` with `delete_recursive=true` to remove a catalog and any objects it contains, regardless of their state
* `publish_enabled` - (Optional, *v3.6+*) Enable allows to publish a catalog externally to make its vApp templates and media files available for subscription by organizations outside the Cloud Director installation. Default is `false`. 
* `cache_enabled` - (Optional, *v3.6+*) Enable early catalog export to optimize synchronization. Default is `false`. It is recommended to set it to `true` when publishing the catalog.
* `preserve_identity_information` - (Optional, *v3.6+*) Enable include BIOS UUIDs and MAC addresses in the downloaded OVF package. Preserving the identity information limits the portability of the package, and you should use it only when necessary. Default is `false`.
* `password` - (Optional, *v3.6+*) An optional password to access the catalog. Only ASCII characters are allowed in a valid password.
* `metadata` - (Deprecated; *v3.6+*) Use `metadata_entry` instead. Key value map of metadata to assign.
* `metadata_entry` - (Optional; *v3.8+*) A set of metadata entries to assign. See [Metadata](#metadata) section for details.
* `metadata_entry_ignore` - (Optional; *3.10+*) A set of metadata entries that must be ignored by Terraform. See [Metadata](#metadata) section for details.

## Attribute Reference

* `catalog_version` - (*v3.6+*) Version number from this catalog.
* `owner_name` - (*v3.6+*) Owner of the catalog.
* `number_of_vapp_templates` - (*v3.6+*) Number of vApp templates available in this catalog.
* `vapp_template_list` (*v3.8+*) List of vApp template names in this catalog, in alphabetical order.
* `media_item_list` (*v3.8+*) List of media item names in this catalog, in alphabetical order.
* `number_of_media` - (*v3.6+*) Number of media items available in this catalog.
* `is_shared` - (*v3.6+*) Indicates if the catalog is shared.
* `is_published` - (*v3.6+*) Indicates if this catalog is shared to all organizations.
* `is_local` - (*v3.8.1+*) Indicates if this catalog was created in the current organization.
* `created` - (*v3.6+*) Date and time of catalog creation
* `publish_subscription_type` - (*v3.6+*) Shows if the catalog is `PUBLISHED`, if it is a subscription from another one (`SUBSCRIBED`), or none of those (`UNPUBLISHED`).
* `publish_subscription_url` - (*v3.8+*) URL to which other catalogs can subscribe.

<a id="metadata"></a>
## Metadata

The `metadata_entry` (*v3.8+*) attribute is a set of metadata entries that have the following structure:

* `key` - (Required) Key of this metadata entry.
* `value` - (Required) Value of this metadata entry.
* `type` - (Required) Type of this metadata entry. One of: `MetadataStringValue`, `MetadataNumberValue`, `MetadataDateTimeValue`, `MetadataBooleanValue`.
* `user_access` - (Required) User access level for this metadata entry. One of: `PRIVATE` (hidden), `READONLY` (read only), `READWRITE` (read/write).
* `is_system` - (Required) Domain for this metadata entry. true if it belongs to `SYSTEM`, false if it belongs to `GENERAL`.

~> Note that `is_system` requires System Administrator privileges, and not all `user_access` options support it.
   You may use `is_system = true` with `user_access = "PRIVATE"` or `user_access = "READONLY"`.

Example:

```hcl
resource "vcd_catalog" "example" {
  # ...
  metadata_entry {
    key         = "foo"
    type        = "MetadataStringValue"
    value       = "bar"
    user_access = "PRIVATE"
    is_system   = true # Requires System admin privileges
  }

  metadata_entry {
    key         = "myBool"
    type        = "MetadataBooleanValue"
    value       = "true"
    user_access = "READWRITE"
    is_system   = false
  }
}
```

To remove all metadata one needs to specify an empty `metadata_entry`, like:

```
metadata_entry {}
```

The same applies also for deprecated `metadata` attribute:

```
metadata = {}
```

To ignore any metadata entry of your choice, you may use the `metadata_entry_ignore` (*v3.10+*) attribute.
The structure is the same as `metadata_entry`, but both `key` and `value` support regular expressions for filtering.
Each element of the structure will be combined with the others by using an `AND` logical operator. For example:

```hcl
resource "vcd_catalog" "example" {
  # ...
  metadata_entry_ignore {
    key         = "foo.*"
    value       = "bar"
    user_access = "PRIVATE"
  }
}
```

This will make Terraform ignore the metadata entries which key matches `foo.*` AND the value is `bar` AND the user access is `PRIVATE`.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing catalog can be [imported][docs-import] into this resource via supplying the full dot separated path for a
catalog. For example, using this structure, representing an existing catalog that was **not** created using Terraform:

```hcl
resource "vcd_catalog" "my-catalog" {
  org              = "my-org"
  name             = "my-catalog"
  delete_recursive = true
  delete_force     = true
}
```

You can import such catalog into terraform state using this command

```bash
terraform import vcd_catalog.my-catalog my-org.my-catalog
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the catalog as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the catalog's stored properties.

