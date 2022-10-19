---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_catalog"
sidebar_current: "docs-vcd-resource-subscribed-catalog"
description: |-
  Provides a VMware Cloud Director subscribed catalog resource. This can be used to create, update, and delete a subscribed catalog.
---

# vcd\_subscribed\_catalog

Provides a VMware Cloud Director subscribed catalog resource. This can be used to create, update, and delete a subscribed catalog.

A `vcd_subscribed_catalog` is a catalog that was created by subscribing to another catalog. It can be used, to some extent,
like any other catalog, but users must keep in mind that this resource depends on the connection to another catalog, which
may not even be in the same VCD. For more information, see the full [Catalog subscription and sharing](/providers/vmware/vcd/latest/docs/guides/catalog_subscription_and_sharing) guide.

Supported in provider *v3.8+*

## Example creation usage

```hcl

data "vcd_catalog" "publisher" {
  org  = "some-other-org"  
  name = "publisher"
}

resource "vcd_subscribed_catalog" "subscriber" {
  org  = "my-org"
  name = "subscriber"

  delete_force     = "true"
  delete_recursive = "true"

  subscription_url      = data.vcd_catalog.publisher.publish_subscription_url
  make_local_copy       = true
  timeout               = 15
  subscription_password = var.subscription_password
}

```

## Example synchronisation usage

When a subscribed catalog needs to be updated (i.e. getting new items that were published in the original catalog), we can 
refresh the resource:

```hcl
data "vcd_catalog" "publisher" {
  org  = "some-other-org"  
  name = "publisher"
}

resource "vcd_subscribed_catalog" "subscriber" {
  org  = "my-org"
  name = "subscriber"

  delete_force     = "true"
  delete_recursive = "true"

  subscription_url      = data.vcd_catalog.publisher.publish_subscription_url
  make_local_copy       = false
  subscription_password = var.subscription_password

  sync_on_refresh         = true
  sync_all                = true
  
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `name` - (Required) Catalog name
* `storage_profile_id` - (Optional, *v3.1+*) Allows to set specific storage profile to be used for catalog. **Note.** Data
source [vcd_storage_profile](/providers/vmware/vcd/latest/docs/data-sources/storage_profile) can help to lookup storage profile ID.
* `delete_recursive` - (Required) When destroying use delete_recursive=True to remove the catalog and any objects it contains that are in a state that normally allows removal
* `delete_force` -(Required) When destroying use delete_force=True with delete_recursive=True to remove a catalog and any objects it contains, regardless of their state
* `subscription_password` - An optional password to access the catalog. Only ASCII characters are allowed in a valid password.

## Attribute Reference

* `description` -  Description of catalog. This is inherited from the publishing catalog
* `catalog_version` - Version number from this catalog.
* `owner_name` - Owner of the catalog.
* `number_of_vapp_templates` - Number of vApp templates available in this catalog.
* `number_of_media` - Number of media items available in this catalog.
* `is_shared` - Indicates if the catalog is shared.
* `is_published` - Indicates if this catalog is available for subscription.
* `publish_subscription_type` - Shows if the catalog is published, if it is a subscription from another one or none of those.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing subscribed catalog can be [imported][docs-import] into this resource via supplying the full dot separated path for a
catalog. For example, using this structure, representing an existing subscribed catalog that was **not** created using Terraform:

```hcl
resource "vcd_subscribed_catalog" "my-catalog" {
  org              = "my-org"
  name             = "my-catalog"
  delete_recursive = "true"
  delete_force     = "true"
  subscription_url = var.publish_subscription_url
}
```

You can import such catalog into terraform state using this command

```bash
terraform import vcd_subscribed_catalog.my-catalog my-org.my-catalog-name
# or
terraform import vcd_subscribed_catalog.my-catalog my-org.my-catalog-id
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the subscribed catalog as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the subscribed catalog's stored properties.

