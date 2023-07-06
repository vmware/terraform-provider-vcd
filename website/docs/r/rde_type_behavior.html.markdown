---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_rde_type_behavior"
sidebar_current: "docs-vcd-resource-rde-type-behavior"
description: |-
   Provides the capability of managing RDE Type Behaviors in VMware Cloud Director.
---

# vcd\_rde\_type\_behavior

Provides the capability of managing RDE Type Behaviors in VMware Cloud Director, which override an existing [RDE Interface
Behavior](/providers/vmware/vcd/latest/docs/resources/rde_interface_behavior).

Supported in provider *v3.10+*. Requires System administrator privileges.

## Example Usage With Execution Override

```hcl
data "vcd_rde_interface" "my_interface" {
  vendor  = "bigcorp"
  nss     = "tech1"
  version = "1.2.3"
}

resource "vcd_rde_interface_behavior" "my_interface_behavior" {
  rde_interface_id = vcd_rde_interface.my_interface.id
  name             = "MyBehavior"
  description      = "Adds a node to the cluster.\nParameters:\n  clusterId: the ID of the cluster\n  node: The node address\n"
  execution = {
    "id" : "addNode"
    "type" : "Activity"
  }
}

resource "vcd_rde_type" "my_rde_type" {
  vendor        = "vmware"
  nss           = "vcd"
  version       = "4.5.6"
  name          = "My VMware RDE Type"
  interface_ids = [data.vcd_rde_interface.my_interface.id]
  schema        = file("${path.module}/schemas/my-type-schema.json")

  # Behaviors can't be created after the RDE Interface is used by a RDE Type
  # so we need to depend on the Behavior to wait for it to be created first.
  depends_on = [vcd_rde_interface_behavior.my_interface_behavior]
}

resource "vcd_rde_type_behavior" "my_rde_type_behavior" {
  rde_type_id               = vcd_rde_type.my_rde_type.id
  rde_interface_behavior_id = vcd_rde_interface_behavior.my_interface_behavior.id
  execution = {
    "id" : "addNodeOverrided"
    "type" : "Activity"
  }
}
```

## Argument Reference

The following arguments are supported:

* `rde_type_id` - (Required) The ID of the RDE Type that owns the Behavior
* `rde_interface_behavior_id` - (Required) The ID of the [RDE Interface Behavior](/providers/vmware/vcd/latest/docs/resources/rde_interface_behavior) to override
* `execution` - (Required) A map that specifies the Behavior execution mechanism.
  You can find more information about the different execution types, like `WebHook`, `noop`, `Activity`, `MQTT`, `VRO`, `AWSLambdaFaaS`
  and others [in the Extensibility SDK documentation](https://vmware.github.io/vcd-ext-sdk/docs/defined_entities_api/behaviors)
* `description` - (Optional) The description of the RDE Type Behavior.

## Attribute Reference

* `name` - Name of the overridden Behavior
* `ref` - The Behavior invocation reference to be used for polymorphic behavior invocations

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing RDE Type Behavior can be [imported][docs-import] into this resource via supplying the related RDE Type `vendor`, `nss` and `version`, and
the Behavior `name`.
For example, using this structure, representing an existing RDE Type Behavior that was **not** created using Terraform:

```hcl
resource "vcd_rde_type_behavior" "my_rde_type_behavior" {
  rde_type_id               = data.vcd_rde_type.my_rde_type.id
  rde_interface_behavior_id = data.vcd_rde_interface_behavior.my_interface_behavior.id
}
```

You can import such RDE Type into Terraform state using this command

```
terraform import vcd_rde_type.outer_type vmware.k8s.1.0.0.createKubeConfig
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the RDE Type Behavior as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the RDE Type Behavior's stored properties.
