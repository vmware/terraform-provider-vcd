---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_rde_type_behavior_acl"
sidebar_current: "docs-vcd-resource-rde-type-behavior-acl"
description: |-
   Provides the capability of managing RDE Type Behavior and RDE Interface Behavior Access Levels in VMware Cloud Director.
---

# vcd\_rde\_type\_behavior\_acl

Provides the capability of managing [RDE Type Behavior](/providers/vmware/vcd/latest/docs/resources/rde_type_behavior)
and [RDE Interface Behavior](/providers/vmware/vcd/latest/docs/resources/rde_interface_behavior) Access Levels in VMware Cloud Director.

Supported in provider *v3.10+*. Requires System administrator privileges.

## Example Usage

```hcl
resource "vcd_rde_interface" "interface" {
  nss     = "nss"
  version = "1.0.0"
  vendor  = "vendor"
  name    = "MyInterface"
}

resource "vcd_rde_interface_behavior" "behavior1" {
  rde_interface_id = vcd_rde_interface.interface.id
  name             = "MyInterfaceBehavior1"
  description      = "My Behavior Contract goes here"
  execution = {
    "id" : "MyActivity1"
    "type" : "noop"
  }
}

resource "vcd_rde_interface_behavior" "behavior2" {
  rde_interface_id = vcd_rde_interface.interface.id
  name             = "MyInterfaceBehavior2"
  description      = "My Behavior Contract goes here"
  execution = {
    "id" : "MyActivity2"
    "type" : "noop"
  }
}

resource "vcd_rde_type" "type" {
  nss           = "nss"
  version       = "1.0.0"
  vendor        = "vendor"
  name          = "MyType"
  description   = "Type Description"
  interface_ids = [vcd_rde_interface.interface.id]
  schema        = file("/home/foo/my_rde_type.json")

  # Behaviors can't be created after the RDE Interface is used by a RDE Type
  # so we need to depend on the Behaviors to wait for them to be created first.
  depends_on = [vcd_rde_interface_behavior.behavior1, vcd_rde_interface_behavior.behavior2]
}

resource "vcd_rde_type_behavior" "behavior_override" {
  rde_type_id               = vcd_rde_type.type.id
  rde_interface_behavior_id = vcd_rde_interface_behavior.behavior1.id
  description               = "MyTypeBehaviorOverride"
  execution = {
    "id" : "MyActivityOverride"
    "type" : "noop"
  }
}

# Access levels for the Behavior override defined in a RDE Type
resource "vcd_rde_type_behavior_acl" "acl1" {
  rde_type_id      = vcd_rde_type.type.id
  behavior_id      = vcd_rde_type_behavior.behavior_override.id
  access_level_ids = ["urn:vcloud:accessLevel:FullControl", "urn:vcloud:accessLevel:ReadOnly"]
}

# Access levels for the Behavior defined in a RDE Interface
resource "vcd_rde_type_behavior_acl" "acl2" {
  rde_type_id      = vcd_rde_type.type.id
  behavior_id      = vcd_rde_interface_behavior.behavior2.id
  access_level_ids = ["urn:vcloud:accessLevel:FullControl"]
}
```

## Argument Reference

The following arguments are supported:

* `rde_type_id` - (Required) The ID of the RDE Type
* `behavior_id` - (Required) The ID of either a [RDE Type Behavior](/providers/vmware/vcd/latest/docs/resources/rde_type_behavior)
  or a [RDE Interface Behavior](/providers/vmware/vcd/latest/docs/resources/rde_interface_behavior)
* `access_level_ids` - (Required) Set of Access Level IDs to associate to the Behavior defined in `behavior_id` argument

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

Existing RDE Type Behavior Access Levels can be [imported][docs-import] into this resource via supplying the related RDE Type `vendor`, `nss` and `version`, and
the Behavior `name`.
For example, using this structure, representing some existing RDE Type Behavior Access Levels that were **not** created using Terraform:

```hcl
resource "vcd_rde_type_behavior_acl" "my_acl" {
  rde_type_id = data.vcd_rde_type.my_rde_type.id
  behavior_id = data.vcd_rde_interface_behavior.my_interface_behavior.id
}
```

You can import such RDE Type into Terraform state using this command

```
terraform import vcd_rde_type.outer_type vmware.k8s.1.0.0.createKubeConfig
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the RDE Type Behavior Access Levels as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the RDE Type Behavior Access Levels' stored properties.
