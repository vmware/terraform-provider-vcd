---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_tm_region"
sidebar_current: "docs-vcd-resource-tm-region"
description: |-
  Provides a resource to manage Regions in VMware Cloud Foundation Tenant Manager.
---

# vcd\_tm\_region

Provides a resource to manage Regions in VMware Cloud Foundation Tenant Manager.

~> Only `System Administrator` can create this resource.

## Example Usage

```hcl
data "vcd_tm_supervisor" "one" {
  name = "first-supervisor"

  depends_on = [vcd_vcenter.one]
}

resource "vcd_tm_region" "one" {
  name                 = "region-one"
  is_enabled           = true
  nsx_manager_id       = data.vcd_tm_nsxt_manager.test.id
  supervisor_ids       = [data.vcd_tm_supervisor.test.id]
  storage_policy_names = ["vSAN Default Storage Policy"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for Region
* `description` - (Optional) An optional description for Region
* `is_enabled` - (Optional) Defines if the Region is enabled. Default is `true`
* `nsx_manager_id` - (Required) NSX-T Manager assigned to this region. Can be looked up using
  [`vcd_tm_nsxt_manager`](/providers/vmware/vcd/latest/docs/data-sources/tm_nsxt_manager)
* `supervisor_ids` - (Required) A set of Supervisor IDs. At least one is required. Can be looked up
  using [`vcd_tm_supervisor`](/providers/vmware/vcd/latest/docs/data-sources/tm_supervisor)
* `storage_policy_names` - (Required) A set of Storage Policy names to be used for this region. At
  least one is required.

## Attribute Reference

The following attributes are exported on this resource:

* `cpu_capacity_mhz` - Total CPU resources in MHz available to this Region
* `cpu_reservation_capacity_mhz` - Total CPU reservation resources in MHz available to this Region
* `memory_capacity_mib` - Total memory resources (in mebibytes) available to this Region
* `memory_reservation_capacity_mib` - Total memory reservation resources (in mebibytes) available to this Region
* `status` - The creation status of the Region. Possible values are `READY`, `NOT_READY`, `ERROR`,
  `FAILED`. A Region needs to be ready and enabled to be usable

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing Region configuration can be [imported][docs-import] into this resource via supplying
path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_tm_region.imported my-region
```

The above would import the `my-region` Region settings