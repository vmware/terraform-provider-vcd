---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_rde_interface"
sidebar_current: "docs-vcd-data-source-rde-interface"
description: |-
   Provides the capability of fetching an existing Runtime Defined Entity Interface from VMware Cloud Director.
---

# vcd\_rde\_interface

Provides the capability of fetching an existing Runtime Defined Entity Interface from VMware Cloud Director.

Supported in provider *v3.9+*

## Example Usage

```hcl
data "vcd_rde_interface" "my_interface" {
  vendor  = "bigcorp"
  nss     = "tech"
  version = "1.2.3"
}

output "interface_name" {
  value = data.vcd_rde_interface.my_interface.name
}

output "interface_id" {
  value = data.vcd_rde_interface.my_interface.id
}
```

## Argument Reference

The following arguments are supported:

* `vendor` - (Required) The vendor of the RDE Interface.
* `nss` - (Required) A unique namespace associated with the RDE Interface.
* `version` - (Required) The version of the RDE Interface. Must follow [semantic versioning](https://semver.org/) syntax.

## Attribute Reference

All the supported attributes are defined in the
[Defined Interface resource](/providers/vmware/vcd/latest/docs/resources/rde_interface#argument-reference).

~> Note that `behavior` blocks are only available since *v3.10+*
