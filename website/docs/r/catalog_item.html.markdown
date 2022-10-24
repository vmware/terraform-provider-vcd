---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_catalog_item"
sidebar_current: "docs-vcd-resource-catalog-item"
description: |-
  Provides a VMware Cloud Director catalog item resource. This can be used to upload and delete OVA file inside a catalog.
---

# vcd\_catalog\_item

Provides a VMware Cloud Director catalog item resource. This can be used to upload OVA to catalog and delete it.

Supported in provider *v2.0+*

## Example Usage

```hcl
resource "vcd_catalog_item" "myNewCatalogItem" {
  org     = "my-org"
  catalog = "my-catalog"

  name                 = "my ova"
  description          = "new vapp template"
  ova_path             = "/home/user/file.ova"
  upload_piece_size    = 10
  show_upload_progress = true

  metadata_entry {
    key   = "license"
    value = "public"
  }
  metadata_entry {
    key   = "version"
    value = "v1"
  }
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `catalog` - (Required) The name of the catalog where to upload OVA file
* `name` - (Required) Item name in catalog
* `description` - (Optional) Description of item
* `ova_path` - (Optional) Absolute or relative path to file to upload
* `ovf_url` - (Optional; *v3.6+*) URL to OVF file. Only OVF (not OVA) files are supported by VCD uploading by URL
* `upload_piece_size` - (Optional) - Size in MB for splitting upload size. It can possibly impact upload performance. Default 1MB.
* `show_upload_progress` - (Optional) - Default false. Allows seeing upload progress. (See note below)
* `metadata` - (Optional; *v2.5+*) Key value map of metadata to assign to the associated vApp Template
* `metadata_entry` - (Optional; *v3.8+*) A set of metadata entries to assign to the Catalog Item. See [Metadata](#metadata) section for details.
* `catalog_item_metadata` - (Deprecated; *v3.7+*) Use `metadata_entry` instead.  Key value map of metadata to assign to the Catalog Item

<a id="metadata"></a>
## Metadata

The `metadata_entry` (*v3.8+*) is a set of metadata entries that have the following structure:

* `key` - (Required) Key of this metadata entry.
* `value` - (Required) Value of this metadata entry.
* `type` - (Required) Type of this metadata entry. One of: `MetadataStringValue`, `MetadataNumberValue`, `MetadataDateTimeValue`, `MetadataBooleanValue`.
* `user_access` - (Required) User access level for this metadata entry. One of: `PRIVATE` (hidden), `READONLY` (read only), `READWRITE` (read/write).
* `is_system` - (Required) Domain for this metadata entry. true if it belongs to `SYSTEM`, false if it belongs to `GENERAL`.

~> Note that `is_system` requires System Administrator privileges, and not all `user_access` options support it.
   You may use `is_system = true` with `user_access = "PRIVATE"` or `user_access = "READONLY"`.

Example:

```hcl
resource "vcd_catalog_item" "example" {
  # ...
  metadata_entry {
    key         = "foo"
    type        = "MetadataStringValue"
    value       = "bar"
    user_access = "PRIVATE"
    is_system   = "true" # Requires System admin privileges
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

### A note about upload progress

Until version 3.5.0, the progress was optionally shown on the screen. Due to changes in the terraform tool, such operation
is no longer possible. The progress messages are thus written to the log file (`go-vcloud-director.log`) using a special
tag `[SCREEN]`. To see the progress at run time, users can run the command below in a separate terminal window while 
`terraform apply` is working:

```
$ tail -f go-vcloud-director.log | grep '\[SCREEN\]'
```

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing catalog item can be [imported][docs-import] into this resource via supplying the full dot separated path for a
catalog item. For example, using this structure, representing an existing catalog item that was **not** created using Terraform:

```hcl
resource "vcd_catalog_item" "my-item" {
  org      = "my-org"
  catalog  = "my-catalog"
  name     = "my-item"
  ova_path = "guess"
}
```

You can import such catalog item into terraform state using this command

```
terraform import vcd_catalog_item.my-item my-org.my-catalog.my-item
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the catalog item as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the item's stored properties.
