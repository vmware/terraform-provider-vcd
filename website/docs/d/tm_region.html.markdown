---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_tm_region"
sidebar_current: "docs-vcd-datasource-tm-region"
description: |-
  Provides a data source to read Regions in VMware Cloud Foundation Tenant Manager.
---

# vcd\_tm\_region

Provides a data source to read Regions in VMware Cloud Foundation Tenant Manager.

## Example Usage

```hcl
data "vcd_tm_region" "one" {
  name = "region-one"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name of existing Region

## Attribute Reference

All the arguments and attributes defined in
[`vcd_tm_region`](/providers/vmware/vcd/latest/docs/resources/tm_region) resource are available.
