---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_catalog"
sidebar_current: "docs-vcd-datasource-subscribed-catalog"
description: |-
  Provides a VMware Cloud Director subscribed catalog data source. This can be used to read a subscribed catalog.
---

# vcd\_subscribed\_catalog

Provides a VMware Cloud Director subscribed catalog resource. This can be used to read a subscribed catalog.

A `vcd_subscribed_catalog` is a catalog that was created by subscribing to another catalog. It can be used, to some extent,
like any other catalog, but users must keep in mind that this resource depends on the connection to another catalog, which
may not even be in the same VCD. For more information, see the full [Catalog subscription and sharing](/providers/vmware/vcd/latest/docs/guides/catalog_subscription_and_sharing) guide.

Supported in provider *v3.8+*

## Example

```hcl

data  "vcd_subscribed_catalog" "subscriber" {
  org  = "my-org"
  name = "subscriber"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `name` - (Optional) Catalog name. Required if `filter` is not set.
* `filter` - (Optional) Retrieves the data source using one or more filter parameters

## Attribute Reference
* `storage_profile_id` - Allows to set specific storage profile to be used for catalog.
* `subscription_password` - An optional password to access the catalog. It will always return an empty string.
* `subscription_url` - The URL to subscribe to the external catalog.
* `description` - Description of catalog. This is inherited from the publishing catalog
* `metadata` - Optional metadata of the catalog. This is inherited from the publishing catalog
* `catalog_version` - Version number from this catalog.
* `owner_name` - Owner of the catalog.
* `number_of_vapp_templates` - Number of vApp templates available in this catalog.
* `number_of_media` - Number of media items available in this catalog.
* `vapp_template_list` List of vApp templates in this catalog
* `media_item_list` List of media items in this catalog
* `is_shared` - Indicates if the catalog is shared.
* `is_published` - Indicates if this catalog is available for subscription.
* `publish_subscription_type` - Shows if the catalog is published, if it is a subscription from another one or none of those.
* `href` - the Hyper reference of the catalog.
* `created` - Date and time of catalog creation.
* `running_tasks` - List of running synchronization tasks that are still running. They can refer to the catalog or any of its catalog items.
* `failed_tasks` - List of synchronization tasks that are have failed. They can refer to the catalog or any of its catalog items.

## Filter arguments

* `name_regex` (Optional) matches the name using a regular expression.
* `date` (Optional) is an expression starting with an operator (`>`, `<`, `>=`, `<=`, `==`), followed by a date, with
  optional spaces in between. For example: `> 2020-02-01 12:35:00.523Z`
  The filter recognizes several formats, but one of `yyyy-mm-dd [hh:mm[:ss[.nnnZ]]]` or `dd-MMM-yyyy [hh:mm[:ss[.nnnZ]]]`
  is recommended.
  Comparison with equality operator (`==`) need to define the date to the microseconds.
* `latest` (Optional) If `true`, retrieve the latest item among the ones matching other parameters. If no other parameters
  are set, it retrieves the newest item.
* `earliest` (Optional) If `true`, retrieve the earliest item among the ones matching other parameters. If no other parameters
  are set, it retrieves the oldest item.
* `metadata` (Optional) One or more parameters that will match metadata contents.

See [Filters reference](/providers/vmware/vcd/latest/docs/guides/data_source_filters) for details and examples.

