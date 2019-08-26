---
layout: "vcd"
page_title: "vCloudDirector: vcd_catalog"
sidebar_current: "docs-vcd-data-source-catalog"
description: |-
  Provides a catalog data source.
---

# vcd\_catalog

Provides a vCloud Director Catalog data source. A Catalog can be used to manage catalog items and media items.

Supported in provider *v2.5+*

## Example Usage

```hcl
data "vcd_catalog" "my-cat" {
  org  = "my-org"
  name = "my-cat"
}

resource "vcd_catalog_item" "myItem" {
  org     = "${data.vcd_catalog.my-cat.org}"
  catalog = "${data.vcd_catalog.my-cat.name}"

  name                 = "myItem"
  description          = "Belongs to ${data.vcd_catalog.my-cat.id}"
  ova_path             = "/path/to/test_vapp_template.ova"
  upload_piece_size    = 5
  show_upload_progress = "true"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional, but required if not set at provider level) Org name 
* `name` - (Required) Catalog name

## Attribute Reference

* `description` - Catalog description.
