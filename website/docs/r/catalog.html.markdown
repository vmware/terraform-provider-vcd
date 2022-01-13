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
  delete_recursive = "true"
  delete_force     = "true"
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

  delete_recursive = "true"
  delete_force     = "true"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `name` - (Required) Catalog name
* `description` - (Optional) Description of catalog
* `storage_profile_id` - (Optional, *v3.1+*) Allows to set specific storage profile to be used for catalog. **Note.** Data
source [vcd_storage_profile](/providers/vmware/vcd/latest/docs/data-sources/storage_profile) can help to lookup storage profile ID.
* `delete_recursive` - (Required) When destroying use delete_recursive=True to remove the catalog and any objects it contains that are in a state that normally allows removal
* `delete_force` -(Required) When destroying use delete_force=True with delete_recursive=True to remove a catalog and any objects it contains, regardless of their state
* `publish_enabled` - (Optional, *v3.6+*) Enable external publication. Default is `false`. 
* `cache_enabled` - (Optional, *v3.6+*) Enable early catalog export to optimize synchronization. Default is `false`.
* `preserve_identity_information` - (Optional, *v3.6+*) Enable include BIOS UUIDs and MAC addresses in the downloaded OVF package. Preserving the identity information limits the portability of the package and you should use it only when necessary. Default is `false`.
* `password` - (Optional, *v3.6+*) An optional password to access the catalog. Only ASCII characters are allowed in a valid password.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing catalog can be [imported][docs-import] into this resource via supplying the full dot separated path for a
catalog. For example, using this structure, representing an existing catalog that was **not** created using Terraform:

```hcl
resource "vcd_catalog" "my-catalog" {
  org              = "my-org"
  name             = "my-catalog"
  delete_recursive = "true"
  delete_force     = "true"
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

