---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_rde_interface"
sidebar_current: "docs-vcd-data-source-rde-interface"
description: |-
   Provides the capability of fetching an existing Defined Interfaces from VMware Cloud Director.
---

# vcd\_rde\_interface

Provides the capability of fetching an existing Defined Interface from VMware Cloud Director.

Supported in provider *v3.9+*

## Example Usage

```hcl
data "vcd_rde_interface" "my-interface" {
  vendor    = "bigcorp"
  namespace = "tech1"
  version   = "1.2.3"
}

output "interface-name" {
   value = data.vcd_rde_interface.my-interface.name
}

output "interface-id" {
   value = data.vcd_rde_interface.my-interface.id
}
```

## Argument Reference

The following arguments are supported:

* `vendor` - (Required) The vendor of the Defined Interface.
* `namespace` - (Required) A unique namespace associated with the Defined Interface.
* `version` - (Required) The version of the Defined Interface. Must follow [semantic versioning](https://semver.org/) syntax.

## Attribute Reference

All the supported attributes are defined in the
[Defined Interface resource](/providers/vmware/vcd/latest/docs/resources/rde_interface#argument-reference).
