---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_rde_type"
sidebar_current: "docs-vcd-data-source-rde-type"
description: |-
   Provides the capability of fetching an existing Runtime Defined Entity type from VMware Cloud Director.
---

# vcd\_rde\_type

Provides the capability of fetching an existing Runtime Defined Entity type from VMware Cloud Director.
Requires system administrator privileges.

Supported in provider *v3.9+*

## Example Usage

```hcl
data "vcd_rde_type" "my-rde-type" {
  vendor    = "bigcorp"
  namespace = "tech"
  version   = "1.2.3"
}

output "type-name" {
  value = data.vcd_rde_type.my-rde-type.name
}

output "type-id" {
  value = data.vcd_rde_type.my-rde-type.id
}
```

## Argument Reference

The following arguments are supported:

* `vendor` - (Required) The vendor of the Runtime Defined Entity type.
* `namespace` - (Required) A unique namespace associated with the Runtime Defined Entity type.
* `version` - (Required) The version of the Runtime Defined Entity type. Must follow [semantic versioning](https://semver.org/) syntax.

## Attribute Reference

All the supported attributes are defined in the
[Defined Interface resource](/providers/vmware/vcd/latest/docs/resources/rde_type#argument-reference).
