---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_catalog_vapp_template"
sidebar_current: "docs-vcd-resource-catalog-vapp_template"
description: |-
  Provides a VMware Cloud Director vApp Template resource. This can be used to upload and delete OVA files inside a catalog.
---

# vcd\_catalog\_vapp\_template

Provides a VMware Cloud Director vApp Template resource. This can be used to upload OVA to catalog and delete it.

Supported in provider *v3.8+*

## Example Usage

```hcl
data "vcd_catalog" "my-catalog" {
  org  = "my-org"
  name = "my-catalog"
}

resource "vcd_catalog_vapp_template" "myNewVappTemplate" {
  org        = "my-org"
  catalog_id = data.vcd_catalog.my-catalog.id

  name                 = "my ova"
  description          = "new vapp template"
  ova_path             = "/home/user/file.ova"
  upload_piece_size    = 10
  show_upload_progress = true

  metadata = {
    license = "public"
    version = "v1"
  }
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `catalog_id` - (Required) ID of the Catalog where to upload the OVA file
* `name` - (Required) vApp Template name in Catalog
* `description` - (Optional) Description of the vApp Template. Not to be used with `ovf_url` when target OVA has a description
* `ova_path` - (Optional) Absolute or relative path to file to upload
* `ovf_url` - (Optional) URL to OVF file. Only OVF (not OVA) files are supported by VCD uploading by URL
* `upload_piece_size` - (Optional) - Size in MB for splitting upload size. It can possibly impact upload performance. Default 1MB
* `metadata` -  (Deprecated) Use `metadata_entry` instead. Key/value map of metadata to assign to the associated vApp Template
* `metadata_entry` - (Optional; *v3.8+*) A set of metadata entries to assign. See [Metadata](#metadata) section for details.

## Attribute Reference

* `vdc_id` - The VDC ID to which this vApp Template belongs
* `vm_names` - Set of VM names within the vApp template
* `created` - Timestamp of when the vApp Template was created

<a id="metadata"></a>
## Metadata

The `metadata_entry` (*v3.8+*) is a set of metadata entries that have the following structure:

* `key` - (Required) Key of this metadata entry.
* `value` - (Required) Value of this metadata entry.
* `type` - (Required) Type of this metadata entry. One of: `MetadataStringValue`, `MetadataNumberValue`, `MetadataDateTimeValue`, `MetadataBooleanValue`.
* `user_access` - (Required) User access level for this metadata entry. One of: `PRIVATE` (hidden), `READONLY` (read only), `READWRITE` (read/write).
* `is_system` - (Required) Domain for this metadata entry. true if it belongs to `SYSTEM`, false if it belongs to `GENERAL`.

Example:

```hcl
resource "vcd_catalog_vapp_template" "example" {
  # ...
  metadata_entry {
    key         = "foo"
    type        = "MetadataStringValue"
    value       = "bar"
    user_access = "PRIVATE"
    is_system   = "true"
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

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing vApp Template can be [imported][docs-import] into this resource via supplying the full dot separated path for a
vApp Template. For example, using this structure, representing an existing vAppTemplate that was **not** created using Terraform:

```hcl
data "vcd_catalog" "my-catalog" {
  org  = "my-org"
  name = "my-catalog"
}

resource "vcd_catalog_vapp_template" "my-vapp-template" {
  org        = "my-org"
  catalog_id = data.vcd_catalog.my-catalog.id
  name       = "my-vapp-template"
  ova_path   = "guess"
}
```

You can import such vApp Template into terraform state using this command

```
terraform import vcd_catalog_vapp_template.my-vapp-template my-org.my-catalog.my-vapp-template
```

You can also import a vApp Template using a VDC name instead of a Catalog name:

```
terraform import vcd_catalog_vapp_template.my-vapp-template my-org.my-vdc.my-vapp-template
```


NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the vApp Template as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the vApp Template's stored properties.
