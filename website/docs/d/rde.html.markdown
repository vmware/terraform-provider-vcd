---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_rde"
sidebar_current: "docs-vcd-data-source-rde"
description: |-
   Provides the capability of fetching an existing Runtime Defined Entity from VMware Cloud Director.
---

# vcd\_rde

Provides the capability of fetching an existing Runtime Defined Entity from VMware Cloud Director.
Requires system administrator privileges.

Supported in provider *v3.9+*

## Example Usage

```hcl
data "vcd_rde_type" "my-rde-type" {
  vendor    = "bigcorp"
  namespace = "tech"
  version   = "1.2.3"
}

data "vcd_rde" "my-rde" {
   rde_type_id = data.vcd_rde_type.my-rde-type.id
   name = "myRde1"
}

output "rde-json" {
  value = data.vcd_rde.my-rde.entity
}

output "rde-id" {
  value = data.vcd_rde.my-rde.id
}
```

## Argument Reference

The following arguments are supported:

* `rde_type_vendor` - (Required) The vendor of the [RDE Type](/providers/vmware/vcd/latest/docs/data-sources/rde_type) that the RDE is instance of.
* `rde_type_namespace` - (Required) The namespace of the [RDE Type](/providers/vmware/vcd/latest/docs/data-sources/rde_type) that the RDE is instance of.
* `rde_type_version` - (Required) The version of the [RDE Type](/providers/vmware/vcd/latest/docs/data-sources/rde_type) that the RDE is instance of.
* `name` - (Required) The name of the RDE to fetch.

## Attribute Reference

All the supported attributes are defined in the
[Defined Interface resource](/providers/vmware/vcd/latest/docs/resources/rde#argument-reference).
