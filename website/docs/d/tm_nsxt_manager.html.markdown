---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_tm_nsxt_manager"
sidebar_current: "docs-vcd-data-source-tm-nsxt-manager"
description: |-
  Provides a data source for available Tenant Manager NSX-T manager.
---

# vcd\_tm\_nsxt\_manager

Provides a data source for available Tenant Manager NSX-T manager.

Supported in provider *v3.0+*

~> **Note:** This resource uses new VMware Cloud Director
[OpenAPI](https://code.vmware.com/docs/11982/getting-started-with-vmware-cloud-director-openapi) and
requires at least VCD *10.1.1+* and NSX-T *3.0+*.

## Example Usage 

```hcl
data "vcd_nsxt_manager" "main" {
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
[`vcd_nsxt_manager`](/providers/vmware/vcd/latest/docs/resources/nsxt_manager#attribute-reference)
are supported.
