---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_manager"
sidebar_current: "docs-vcd-data-source-nsxt-manager"
description: |-
  Provides a data source for available NSX-T manager.
---

# vcd\_nsxt\_manager

Provides a data source for NSX-T manager.

Supported in provider *v3.0+*

## Example Usage 

```hcl
data "vcd_nsxt_manager" "main" {
  name = "nsxt-manager-one"
}
```


## Argument Reference

The following arguments are supported:

* `name` - (Required) Organization VDC name

## Attribute reference

Only ID is set to be able and reference in other resources or data sources.
