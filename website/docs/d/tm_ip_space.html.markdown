---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_tm_ip_space"
sidebar_current: "docs-vcd-data-source-tm-ip-space"
description: |-
  Provides a VMware Cloud Foundation Tenant Manager IP Space data source.
---

# vcd\_tm\_ip\_space

Provides a VMware Cloud Foundation Tenant Manager IP Space data source.

## Example Usage

```hcl
data "vcd_tm_region" "demo" {
  name = "demo-region"
}

data "vcd_tm_ip_space" "demo" {
  name      = "demo-ip-space"
  region_id = data.vcd_tm_region.region.id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of IP Space
* `region_id` - (Optional) The Region ID that has this IP Space definition. Can be looked up using
  [`vcd_tm_region`](/providers/vmware/vcd/latest/docs/data-sources/tm_region)

## Attribute Reference

All the arguments and attributes defined in
[`vcd_tm_ip_space`](/providers/vmware/vcd/latest/docs/resources/tm_ip_space) resource are available.
