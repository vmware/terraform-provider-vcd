---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_catalog_media"
sidebar_current: "docs-vcd-resource-catalog-media"
description: |-
  Provides a VMware Cloud Director media resource. This can be used to upload and delete media (ISO) file inside a catalog.
---

# vcd\_catalog\_media

Provides a VMware Cloud Director media resource. This can be used to upload media to catalog and delete it.

Supported in provider *v2.0+*

## Example Usage

```hcl
data "vcd_catalog" "my-catalog" {
  org  = "my-org"
  name = "my-catalog"
}

resource "vcd_catalog_media" "myNewMedia" {
  org        = "my-org"
  catalog_id = data.vcd_catalog.my-catalog.id

  name                 = "my iso"
  description          = "new os versions"
  media_path           = "/home/user/file.iso"
  upload_piece_size    = 10

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
* `catalog` - (Optional; Deprecated) The name of the catalog where to upload media file. It's mandatory if `catalog_id` is not used.
* `catalog_id` - (Optional; *v3.8.2+*) The ID of the catalog where to upload media file. It's mandatory if `catalog` field is not used.
* `name` - (Required) Media file name in catalog
* `description` - (Optional) - Description of media file
* `media_path` - (Required) - Absolute or relative path to file to upload
* `upload_piece_size` - (Optional) - size in MB for splitting upload size. It can possibly impact upload performance. Default 1MB.
* `show_upload_progress` - (Optional) - Default false. Allows to see upload progress. (See note below)
* `metadata` - (Deprecated; *v2.5+*) Use `metadata_entry` instead. Key value map of metadata to assign
* `metadata_entry` - (Optional; *v3.8+*) A set of metadata entries to assign. See [Metadata](#metadata) section for details.

## Attribute reference

Supported in provider *v2.5+*

* `is_iso` - (Computed) returns True if this media file is ISO
* `owner_name` - (Computed) returns owner name
* `is_published` - (Computed) returns True if this media file is in a published catalog
* `creation_date` - (Computed) returns creation date
* `size` - (Computed) returns media storage in Bytes
* `status` - (Computed) returns media status
* `storage_profile_name` - (Computed) returns storage profile name

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
resource "vcd_catalog_media" "example" {
  # ...
  metadata_entry {
    key         = "foo"
    type        = "MetadataStringValue"
    value       = "bar"
    user_access = "PRIVATE"
    is_system   = "true" # Requires System admin privileges
  }

  metadata_entry {
    key         = "myBool"
    type        = "MetadataBooleanValue"
    value       = "true"
    user_access = "READWRITE"
    is_system   = "false"
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

Supported in provider *v2.5+*

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing media item can be [imported][docs-import] into this resource via supplying its path.
The path for this resource is made of org-name.catalog-name.media-name
For example, using this structure, representing a media item that was **not** created using Terraform:

```hcl
resource "vcd_catalog_media" "tf-mymedia" {
  org     = "my-org"
  catalog = "my-catalog"
  name    = "my-media"
}
```

You can import such catalog media into terraform state using this command

```
terraform import vcd_catalog_media.tf-mymedia my-org.my-catalog.my-media
```

[docs-import]:https://www.terraform.io/docs/import/

After importing, if you run `terraform plan` you will see the rest of the values and modify the script accordingly for
further operations.
