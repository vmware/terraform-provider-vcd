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
data "vcd_tm_region" "region" {
  name = "my-region"
}

data "vcd_tm_region_storage_policy" "sp" {
  region_id = data.vcd_tm_region.region.id
  name      = "vSAN Default Storage Policy"
}

resource "vcd_tm_content_library" "cl" {
  name        = "My Library"
  description = "A simple library"
  storage_policy_ids = [
    data.vcd_tm_region_storage_policy.sp.id
  ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Region Storage Policy to read
* `region_id` - (Required) The ID of the Region where the Storage Policy belongs

## Attribute reference

* `description` - Description of the Region Storage Policy
* `status` - The creation status of the Region Storage Policy. Can be `NOT_READY` or `READY`
* `storage_capacity_mb` - Storage capacity in megabytes for this Region Storage Policy
* `storage_consumed_mb` - Consumed storage in megabytes for this Region Storage Policy