---
layout: "vcd"
page_title: "VMware Cloud Foundation Tenant Manager: vcd_tm_content_library"
sidebar_current: "docs-vcd-data-source-tm-content-library"
description: |-
  Provides a VMware Cloud Foundation Tenant Manager Content Library data source. This can be used to read Content Libraries.
---

# vcd\_content\_library

Provides a VMware Cloud Foundation Tenant Manager Content Library data source. This can be used to read Content Libraries.

This data source is exclusive to **VMware Cloud Foundation Tenant Manager**. Supported in provider *v4.0+*

## Example Usage

```hcl
data "vcd_tm_content_library" "cl" {
  name = "My Library"
}

output "is_shared" {
  value = data.vcd_tm_content_library.cl.is_shared
}
output "owner_org" {
  value = data.vcd_tm_content_library.cl.owner_org_id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Content Library to read

## Attribute reference

All arguments and attributes defined in [the resource](/providers/vmware/vcd/latest/docs/resources/tm_content_library) are supported
as read-only (Computed) values.
