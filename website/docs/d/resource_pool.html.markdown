---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_resource_pool"
sidebar_current: "docs-vcd-data-source-resource-pool"
description: |-
  Provides a data source for a resource pool attached to a vCenter.
---

# vcd\_resource\_pool

Provides a data source for a resource pool attached to a vCenter.

Supported in provider *v3.10+*


## Example Usage

```hcl
data "vcd_vcenter" "vcenter1" {
  name = "vc1"
}

data "vcd_resource_pool" "rp1" {
  name       = "resource-pool-for-vcd-01"
  vcenter_id = data.vcd_vcenter.vcenter1.id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) resource pool name.
* `vcenter_id` - (Required) ID of the vCenter to which this resource pool belongs.

## Attribute reference

* `hardware_version` -  default hardware version available to this resource pool.
