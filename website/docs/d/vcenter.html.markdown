---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_vcenter"
sidebar_current: "docs-vcd-data-source-vcenter"
description: |-
  Provides a data source for vCenter server attached to VCD.
---

# vcd\_vcenter

Provides a data source for vCenter server attached to VCD.

Supported in provider *v3.0+*


## Example Usage

```hcl
data "vcd_vcenter" "vc" {
  name = "vcenter-one"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) vCenter name

## Attribute reference

Only ID is set to be able and reference in other resources or data sources.
