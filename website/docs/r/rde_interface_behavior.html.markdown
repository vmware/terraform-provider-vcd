---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_rde_interface_behavior"
sidebar_current: "docs-vcd-resource-rde-interface-behavior"
description: |-
   Provides the capability of managing RDE Interface Behaviors in VMware Cloud Director.
---

# vcd\_rde\_interface\_behavior

Provides the capability of managing RDE Interface Behaviors in VMware Cloud Director.

Supported in provider *v3.10+*. Requires System administrator privileges.

~> Be aware that Behaviors can only be created and deleted when no [RDE Types](/providers/vmware/vcd/latest/docs/resources/rde_type) are using the Interface where they are defined.
If you want to use RDE Types with Behaviors, you should use `depends_on` as seen in the example [here](/providers/vmware/vcd/latest/docs/resources/rde_interface_behavior#example-usage)

## Example Usage

```hcl
resource "vcd_rde_interface" "my_interface" {
  vendor  = "bigcorp"
  nss     = "tech"
  version = "1.2.3"
  name    = "BigCorp Interface"
}

resource "vcd_rde_interface_behavior" "my_behavior" {
  interface_id = vcd_rde_interface.my_interface.id
  name         = "MyBehavior"
  description  = "Adds a node to the cluster.\nParameters:\n  clusterId: the ID of the cluster\n  node: The node address\n"
  execution = {
    "id" : "MyExecution"
    "type" : "Activity"
  }
}

resource "vcd_rde_interface_behavior" "my_behavior2" {
  interface_id = vcd_rde_interface.my_interface.id
  name         = "MyBehavior2"
  execution = {
    "id" : "MyExecution2"
    "type" : "noop"
  }
}

resource "vcd_rde_interface_behavior" "my_behavior3" {
  interface_id = vcd_rde_interface.my_interface.id
  name         = "MyBehavior3"
  execution = {
    "type" : "WebHook",
    "id" : "testWebHook",
    "href" : "https://hooks.slack.com:443/services/T07UZFN0N/B01EW5NC42D/rfjhHCGIwzuzQFrpPZiuLkIX",
    "_internal_key" : "secretKey"
  }
}
```

## Argument Reference

The following arguments are supported:

* `interface_id` - (Required) The ID of the RDE Interface that owns the Behavior
* `name` - (Required) Name of the Behavior
* `description` - (Optional) A description specifying the contract of the Behavior
* `execution` - (Required) A map that specifies the Behavior execution mechanism.
  You can find more information about the different execution types, like `WebHook`, `noop`, `Activity`, `MQTT`, `VRO`, `AWSLambdaFaaS`
  and others [in the Extensibility SDK documentation](https://vmware.github.io/vcd-ext-sdk/docs/defined_entities_api/behaviors)

## Attribute Reference

* `ref` - The Behavior invocation reference to be used for polymorphic behavior invocations

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing RDE Interface Behavior can be [imported][docs-import] into this resource via supplying the related RDE Interface `vendor`, `nss` and `version`, and
the Behavior `name`.
For example, using this structure, representing an existing RDE Interface Behavior that was **not** created using Terraform:

```hcl
resource "vcd_rde_interface_behavior" "outer_interface" {
  interface_id = "urn:vcloud:interface:vmware:k8s:1.0.0"
  name         = "createKubeConfig"
}
```

You can import such RDE Interface into Terraform state using this command

```
terraform import vcd_rde_interface.outer_interface vmware.k8s.1.0.0.createKubeConfig
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the RDE Interface Behavior as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the RDE Interface Behavior's stored properties.
