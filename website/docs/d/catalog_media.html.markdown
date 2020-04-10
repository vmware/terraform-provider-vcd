---
layout: "vcd"
page_title: "vCloudDirector: vcd_catalog_media"
sidebar_current: "docs-vcd-data-source-catalog-media"
description: |-
  Provides a catalog media data source.
---

# vcd\_catalog\_media

Provides a vCloud Director Catalog media data source. A Catalog media can be used to reference a catalog media and use its 
data within other resources or data sources.

Supported in provider *v2.5+*

## Example Usage

```hcl
data "vcd_catalog_media" "existing-media" {
  org     = "my-org"
  catalog = "my-cat"
  name    = "my-media"
}

output "media_size" {
  value = data.vcd_catalog_media.existing-media.size
}

output "type_is_iso" {
  value = data.vcd_catalog_media.existing-media.is_iso
}


```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level
* `catalog` - (Required) The name of the catalog where media file is
* `name` - (Required) Media name in catalog (Optional when `filter` is used)
* `filter` - (Optional; *2.9+*) Retrieves the data source using one or more filter parameters

## Attribute reference

All attributes defined in [catalog_media](/docs/providers/vcd/r/catalog_media.html#attribute-reference) are supported.

## Filter arguments

(supported in provider *v2.9+*)

* `name_regex` (Optional) matches the name using a regular expression.
* `date` (Optional) is an expression containing an operator (`>`, `<`, `>=`, `<=`, `==`) and a date. Several formats 
  are recognized, but one of `yyyy-mm-dd [hh:mm[:ss]]` or `dd-MMM-yyyy [hh:mm[:ss]]` is recommended.
* `latest` (Optional) If `true`, retrieve the latest item among the ones matching other parameters. If no other parameters
  are set, it retrieves the newest item.
* `metadata` (Optional) One or more parameters that will match metadata contents.

See [Filters reference](/docs/providers/vcd/guides/filters.html) for details and examples.

