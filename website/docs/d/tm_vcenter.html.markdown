---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_tm_vcenter"
sidebar_current: "docs-vcd-data-source-tm-vcenter"
description: |-
  Provides a data source for vCenter server attached to VCD.
---

# vcd\_tm\_vcenter

Provides a data source for vCenter server attached to VCD.

Supported in provider *v3.0+*


## Example Usage

```hcl
data "vcd_tm_vcenter" "vc" {
  name = "vcenter-one"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) vCenter name

## Attribute reference

All attributes defined in
[`vcd_tm_vcenter`](/providers/vmware/vcd/latest/docs/resources/tm_vcenter#attribute-reference) are
supported.
