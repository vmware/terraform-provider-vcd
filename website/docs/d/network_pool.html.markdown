---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_network_pool"
sidebar_current: "docs-vcd-data-source-network-pool"
description: |-
  Provides a data source for a network pool attached to a VCD.
---

# vcd\_network\_pool

Provides a data source for a network pool attached to a VCD.

Supported in provider *v3.10+*

## Example Usage

```hcl
data "vcd_network_pool" "np1" {
  name = "NSX-T Overlay 1"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) network pool name.

## Attribute reference

* `description` -  description of this network pool
* `type` -  Type of this network pool
