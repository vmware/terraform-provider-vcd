---
layout: "vcd"
page_title: "VMware Cloud Foundation Tenant Manager: vcd_tm_content_library_item"
sidebar_current: "docs-vcd-data-source-tm-content-library-item"
description: |-
  Provides a VMware Cloud Foundation Tenant Manager Content Library Item data source. This can be used to read Content Library Items.
---

# vcd\_content\_library\_item

Provides a VMware Cloud Foundation Tenant Manager Content Library Item data source. This can be used to read Content Library Items.

This data source is exclusive to **VMware Cloud Foundation Tenant Manager**. Supported in provider *v4.0+*

## Example Usage

```hcl
data "vcd_tm_content_library" "cl" {
  name = "My Library"
}

data "vcd_tm_content_library_item" "cli" {
  name               = "My Library Item"
  content_library_id = data.vcd_tm_content_library.cl.id
}

output "is_published" {
  value = data.vcd_tm_content_library_item.cli.is_published
}
output "image_identifier" {
  value = data.vcd_tm_content_library_item.cli.image_identifier
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Content Library to read
* `content_library_id` - (Required) ID of the Content Library that this item belongs to. Can be obtained with [a data source](/providers/vmware/vcd/latest/docs/data-sources/tm_content_library)

## Attribute reference

All arguments and attributes defined in [the resource](/providers/vmware/vcd/latest/docs/resources/tm_content_library_item) are supported
as read-only (Computed) values.
