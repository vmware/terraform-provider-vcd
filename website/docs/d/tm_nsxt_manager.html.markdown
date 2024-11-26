---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_tm_nsxt_manager"
sidebar_current: "docs-vcd-data-source-tm-nsxt-manager"
description: |-
  Provides a data source for available Tenant Manager NSX-T manager.
---

# vcd\_tm\_nsxt\_manager

Provides a data source for available Tenant Manager NSX-T manager.

## Example Usage 

```hcl
data "vcd_tm_nsxt_manager" "main" {
  name = "nsxt-manager-one"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) NSX-T manager name

## Attribute reference

* `id` - ID of the manager
* `href` - Full URL of the manager

All attributes defined in
[`vcd_tm_nsxt_manager`](/providers/vmware/vcd/latest/docs/resources/tm_nsxt_manager#attribute-reference)
are supported.
