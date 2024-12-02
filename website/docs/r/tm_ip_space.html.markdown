---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_tm_ip_space"
sidebar_current: "docs-vcd-resource-tm-ip-space"
description: |-
  Provides a VMware Cloud Foundation Tenant Manager IP Space resource.
---

# vcd\_tm\_ip\_space

Supported in provider *v3.4+* and VCD 10.2+ with NSX-T and ALB.

Provides a VMware Cloud Foundation Tenant Manager IP Space resource.

~> Only `System Administrator` can create this resource.

## Example Usage (Adding NSX-T ALB Service Engine Group)

```hcl

```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for NSX-T ALB Service Engine Group
* `description` - (Optional) An optional description NSX-T ALB Service Engine Group
* `alb_cloud_id` - (Required) A reference NSX-T ALB Cloud. Can be looked up using `vcd_nsxt_alb_cloud` resource or data
  source


## Attribute Reference

The following attributes are exported on this resource:

* `max_virtual_services` - Maximum number of virtual services this NSX-T ALB Service Engine Group can run
* `reserved_virtual_services` - Number of reserved virtual services
* `deployed_virtual_services` - Number of deployed virtual services
* `ha_mode` defines High Availability Mode for Service Engine Group. One off:
  * ELASTIC_N_PLUS_M_BUFFER - Service Engines will scale out to N active nodes with M nodes as buffer.
  * ELASTIC_ACTIVE_ACTIVE - Active-Active with scale out.
  * LEGACY_ACTIVE_STANDBY - Traditional single Active-Standby configuration

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing IP Space configuration can be [imported][docs-import] into this resource via supplying
path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_tm_ip_space.imported my-ip-space-name
```

The above would import the `my-ip-space-name` IP Space.