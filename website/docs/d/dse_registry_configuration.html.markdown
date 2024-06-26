---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_dse_registry_configuration"
sidebar_current: "docs-vcd-data-source-dse-registry-configuration"
description: |-
  Provides a data source to read Data Solution Extension (DSE) registry configuration.
---

# vcd\_dse\_registry\_configuration

Supported in provider *v3.13+* and VCD 10.5.0+ with Data Solution Extension.

Provides a data source to read Data Solution Extension (DSE) registry configuration.

~> Only `System Administrator` can use this data source.

## Example Usage

```hcl
data "vcd_dse_registry_configuration" "mongodb" {
  name = "MongoDB"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of Data Solution as it appears in repository configuration


## Attribute Reference

All the arguments and attributes defined in
[`vcd_dse_registry_configuration`](/providers/vmware/vcd/latest/docs/resources/dse_registry_configuration) resource are available.
