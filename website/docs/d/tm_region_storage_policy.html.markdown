---
layout: "vcd"
page_title: "VMware Cloud Foundation Tenant Manager: vcd_tm_region_storage_policy"
sidebar_current: "docs-vcd-data-source-tm-region-storage-policy"
description: |-
  Provides a VMware Cloud Foundation Tenant Manager Region Storage Policy data source. This can be used to read Content Libraries.
---

# vcd\_tm\_region\_storage\_policy

Provides a VMware Cloud Foundation Tenant Manager Region Storage Policy data source. This can be used to read Region Storage Policies.

This data source is exclusive to **VMware Cloud Foundation Tenant Manager**. Supported in provider *v4.0+*

## Example Usage

```hcl
data "vcd_tm_region_storage_policy" "sp" {
  name = "vSAN Default Storage Policy"
}

resource "vcd_tm_content_library" "cl" {
  name = "My Library"
  description = "A simple library"
  storage_policy_ids = [
    data.vcd_tm_region_storage_policy.sp.id
  ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Region Storage Policy to read

## Attribute reference

// TODO: TM (Resource is not implemented yet)

All arguments and attributes defined in [the resource](/providers/vmware/vcd/latest/docs/resources/tm_region_storage_policy) are supported
as read-only (Computed) values.
