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
  value = data.vcd_catalog_media.size
}

output "type_is_iso" {
  value = data.vcd_catalog_media.is_iso
}


```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level
* `catalog` - (Required) The name of the catalog where media file is
* `name` - (Required) Media name in catalog

## Attribute reference

All attributes defined in [catalog_media](/docs/providers/vcd/r/catalog_media.html#attribute-reference) are supported.
