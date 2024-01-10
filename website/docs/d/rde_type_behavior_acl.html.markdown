---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_rde_type_behavior_acl"
sidebar_current: "docs-vcd-data-source-rde-type-behavior-acl"
description: |-
  Provides the capability of fetching the RDE Type Behavior Access Levels from VMware Cloud Director.
---

# vcd\_rde\_type\_behavior\_acl

Provides the capability of fetching the RDE Type Behavior Access Levels from VMware Cloud Director.

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

data "vcd_rde_type_behavior_acl" "my_behavior_acl" {
  rde_type_id = data.vcd_rde_type.my_type.id
  behavior_id = data.vcd_rde_interface_behavior.my_interface_behavior.id
}

output "access_levels" {
  value = data.vcd_rde_type_behavior_acl.my_behavior_acl.access_level_ids
}

```

## Argument Reference

The following arguments are supported:

* `rde_type_id` - (Required) The ID of the RDE Type
* `behavior_id` - (Required) The ID of either a [RDE Type Behavior](/providers/vmware/vcd/latest/docs/resources/rde_type_behavior)
  or a [RDE Interface Behavior](/providers/vmware/vcd/latest/docs/resources/rde_interface_behavior)

## Attribute Reference

* `access_level_ids` - Set of Access Level IDs associated to the Behavior defined in `behavior_id` argument
