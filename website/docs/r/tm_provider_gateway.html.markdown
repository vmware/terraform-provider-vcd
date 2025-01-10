---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_tm_provider_gateway"
sidebar_current: "docs-vcd-resource-tm-provider-gateway"
description: |-
  Provides a VMware Cloud Foundation Tenant Manager Provider Gateway resource.
---

# vcd\_tm\_provider\_gateway

Provides a VMware Cloud Foundation Tenant Manager Provider Gateway resource.

## Example Usage

```hcl
data "vcd_tm_region" "demo" {
  name = "region-one"
}

data "vcd_tm_tier0_gateway" "demo" {
  name      = "my-tier0-gateway"
  region_id = data.vcd_tm_region.demo.id
}

data "vcd_tm_ip_space" "demo" {
  name      = "demo-ip-space"
  region_id = data.vcd_tm_region.demo.id
}

data "vcd_tm_ip_space" "demo2" {
  name      = "demo-ip-space-2"
  region_id = data.vcd_tm_region.demo.id
}

resource "vcd_tm_provider_gateway" "demo" {
  name                  = "Demo Provider Gateway"
  description           = "Terraform Provider Gateway"
  region_id             = data.vcd_tm_region.demo.id
  nsxt_tier0_gateway_id = data.vcd_tm_tier0_gateway.demo.id
  ip_space_ids          = [data.vcd_tm_ip_space.demo.id, data.vcd_tm_ip_space.demo2.id]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for Provider Gateway
* `description` - (Optional) A description for Provider Gateway
* `region_id` - (Required) A Region ID for Provider Gateway
* `nsxt_tier0_gateway_id` - (Required) An existing NSX-T Tier 0 Gateway
* `ip_space_ids` - (Required) A set of IP Space IDs that should be assigned to this Provider Gateway

## Attribute Reference

* `status` - Current status of the entity. Possible values are:
 * `PENDING` - Desired entity configuration has been received by system and is pending realization
 * `CONFIGURING` - The system is in process of realizing the entity
 * `REALIZED` - The entity is successfully realized in the system
 * `REALIZATION_FAILED` - There are some issues and the system is not able to realize the entity
 * `UNKNOWN` - Current state of entity is unknown

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing Provider Gateway configuration can be [imported][docs-import] into this resource via
supplying path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_tm_provider_gateway.imported my-region-name.my-provider-gateway
```

The above would import the `my-provider-gateway` Provider Gateway in Region `my-region-name`
