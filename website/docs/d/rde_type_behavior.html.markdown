---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_rde_type_behavior"
sidebar_current: "docs-vcd-data-source-rde-type-behavior"
description: |-
  Provides the capability of fetching the Behavior overrides from an existing RDE Type in VMware Cloud Director.
---

# vcd\_rde\_type\_behavior

Provides the capability of fetching the Behavior overrides from an existing RDE Type in VMware Cloud Director.

Supported in provider *v3.10+*. Requires System Administrator privileges.

## Example Usage

```hcl
data "vcd_rde_interface" "my_interface" {
  vendor  = "vmware"
  nss     = "k8s"
  version = "1.0.0"
}

data "vcd_rde_interface_behavior" "my_interface_behavior" {
  interface_id = data.vcd_rde_interface.my_interface.id
  name         = "createKubeConfig"
}

data "vcd_rde_type" "my_type" {
  vendor  = "vmware"
  nss     = "k8s"
  version = "1.2.0"
}

data "vcd_rde_type_behavior" "my_behavior" {
  rde_type_id               = data.vcd_rde_type.my_type.id
  rde_interface_behavior_id = data.vcd_rde_interface_behavior.my_interface_behavior.id
}

output "execution_id" {
  value = data.vcd_rde_type_behavior.my_behavior.execution.id
}
```

## Argument Reference

The following arguments are supported:

* `rde_type_id` - (Required) The ID of the [RDE Type](/providers/vmware/vcd/latest/docs/data-sources/rde_type) that owns the Behavior override to fetch
* `rde_interface_behavior_id` - (Required) The ID of the original [RDE Interface Behavior](/providers/vmware/vcd/latest/docs/data-sources/rde_interface_behavior)

## Attribute Reference

All the supported attributes are defined in the
[RDE Type Behavior resource](/providers/vmware/vcd/latest/docs/resources/rde_type_behavior#argument-reference).
