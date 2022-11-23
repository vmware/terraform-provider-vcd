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
resource "vcd_subscribed_catalog" "subscriber" {
  org  = "my-org"
  name = "subscriber"

  delete_force     = "true"
  delete_recursive = "true"

  subscription_url      = var.subscription_url
  make_local_copy       = true
  subscription_password = var.subscription_password
}

```

## Example synchronisation usage

When a subscribed catalog needs to be updated (i.e. getting new items that were published in the original catalog), we can 
refresh the resource:

```hcl
resource "vcd_subscribed_catalog" "subscriber" {
  org  = "my-org"
  name = "subscriber"

  delete_force     = "true"
  delete_recursive = "true"

  subscription_url      = var.subscription_url
  make_local_copy       = false
  subscription_password = var.subscription_password

  sync_on_refresh = true
  sync_all        = true
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level.
* `name` - (Required) Catalog name
* `storage_profile_id` - (Optional) Allows to set specific storage profile to be used for catalog.
* `delete_recursive` - (Optional) When destroying use `delete_recursive=true` to remove the catalog and any objects it contains that are in a state that normally allows removal.
* `delete_force` - (Optional) When destroying use `delete_force=true` with `delete_recursive=true` to remove a catalog and any objects it contains, regardless of their state.
* `subscription_password` - (Optional) An optional password to access the catalog. Only ASCII characters are allowed in a valid password. 
  The password is only required when set by the publishing catalog. Passing in six asterisks '******' indicates to keep current password. 
  Passing in an empty string indicates to remove password.
* `subscription_url` - (Required) The URL to subscribe to the external catalog.
* `make_local_copy` - (Optional) If `true`, subscription to a catalog creates a local copy of all items. Defaults to `false`, which does not create a local copy of catalog items unless a sync operation is performed.
  It can only be `false` if the user configured in the provider is the System administrator.
* `sync_on_refresh` - (Optional) Boolean value that shows if sync should be performed on every refresh.
* `cancel_failed_tasks` - (Optional) When `true`, the subscribed catalog will attempt canceling failed tasks.
* `sync_all` - (Optional) If `true`, synchronise this catalog and all items. 
* `sync_catalog` - (Optional) If `true`, synchronise this catalog. Not to be used when `sync_all` is set. This operation fetches the list of items. If `make_local_copy` is set, it also synchronises all the items.
* `sync_all_vapp_templates` - (Optional) If `true`, synchronise all vApp templates. Not to be used when `sync_all` is set.
* `sync_all_media_items` - (Optional) If `true`, synchronise all media items. Not to be used when `sync_all` is set.
* `sync_vapp_templates` - (Optional) Synchronise a list of vApp templates. Not to be used when `sync_all` or `sync_all_vapp_templates` are set.
* `sync_media_items` - (Optional) Synchronise a list of media items. Not to be used when `sync_all` or `sync_all_media_items` are set.
* `store_tasks` - (Optional) if `true`, saves the list of tasks to a file for later update.
 
## Attribute Reference

* `description` -  Description of catalog. This is inherited from the publishing catalog and updated on sync.
* `metadata` -  Optional metadata of the catalog. This is inherited from the publishing catalog and updated on sync.
* `catalog_version` - Version number from this catalog. This is inherited from the publishing catalog and updated on sync.
* `owner_name` - Owner of the catalog.
* `number_of_vapp_templates` - Number of vApp templates available in this catalog.
* `number_of_media` - Number of media items available in this catalog.
* `vapp_template_list` List of vApp template names in this catalog, in alphabetical order.
* `media_item_list` List of media item names in this catalog, in alphabetical order.
* `is_shared` - Indicates if the catalog is shared.
* `is_published` - Indicates if this catalog is available for subscription. (Always false)
* `publish_subscription_type` - Shows if the catalog is published, if it is a subscription from another one or none of those. (Always `SUBSCRIBED`)
* `href` - the catalog's Hyper reference.
* `created` - Date and time of catalog creation. This is the creation date of the subscription, not the original published catalog.
* `running_tasks` - List of running synchronization tasks that are still running. They can refer to the catalog or any of its catalog items.
* `failed_tasks` - List of synchronization tasks that are have failed. They can refer to the catalog or any of its catalog items.
* `tasks_file_name` Where the running tasks IDs have been stored. Only if `store_tasks` is set.

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

