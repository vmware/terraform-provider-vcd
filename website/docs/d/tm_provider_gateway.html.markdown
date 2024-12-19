---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_tm_provider_gateway"
sidebar_current: "docs-vcd-data-source-tm-provider-gateway"
description: |-
  Provides a VMware Cloud Foundation Tenant Manager Provider Gateway data source.
---

# vcd\_tm\_provider\_gateway

Provides a VMware Cloud Foundation Tenant Manager Provider Gateway data source.

## Example Usage

```hcl
data "vcd_tm_region" "demo" {
  name = "region-one"
}

data "vcd_tm_provider_gateway" "demo" {
  name      = "Demo Provider Gateway"
  region_id = data.vcd_tm_region.demo.id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of Provider Gateway
* `region_id` - (Required) An ID of Region. Can be looked up using
  [vcd_tm_region](/providers/vmware/vcd/latest/docs/data-sources/tm_region) data source


## Attribute Reference

All the arguments and attributes defined in
[`vcd_tm_provider_gateway`](/providers/vmware/vcd/latest/docs/resources/tm_provider_gateway)
resource are available.