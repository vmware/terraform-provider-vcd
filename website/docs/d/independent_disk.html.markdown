---
layout: "vcd"
page_title: "vCloudDirector: vcd_independent_disk"
sidebar_current: "docs-vcd-data-source-independent-disk"
description: |-
  Provides a independent disk data source.
---

# vcd\_independent\_disk

Provides a vCloud Director Independent disk data source. A independent disk data source can be used to reference an independent disk and use its 
data within other resources or data sources.

Supported in provider *v2.5+*

## Example Usage

```hcl
data "vcd_independent_disk" "existing-disk" {
  org     = "my-org"
  vdc     = "my-vdc"
  name    = "my-disk"
}
output "disk-size" {
  value = data.vcd_independent_disk.size_in_bytes
}
output "type_is_attached" {
  value = data.vcd_independent_disk.is_attached
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level
* `name` - (Required) Disk name

## Attribute reference

All attributes defined in [independent disk](/docs/providers/vcd/r/independent_disk.html#attribute-reference) are supported.