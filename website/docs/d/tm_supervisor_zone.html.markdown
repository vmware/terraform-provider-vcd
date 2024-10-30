---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_tm_supervisor_zone"
sidebar_current: "docs-vcd-datasource-tm-supervisor-zone"
description: |-
  Provides a data source to read Supervisor Zone.
---

# vcd\_tm\_supervisor\_zone

Provides a data source to read Supervisor Zone.

## Example Usage

```hcl
data "vcd_tm_supervisor" "one" {
  name = "my-supervisor-name"

  depends_on = [vcd_vcenter.one]
}

data "vcd_tm_supervisor_zone" "one" {
  supervisor_id = data.vcd_tm_supervisor.one.id
  name = "domain-c8"
}
```

## Argument Reference

The following arguments are supported:

* `supervisor_id` - (Required) ID of parent Supervisor
* `name` - (Required) The name of Supervisor Zone

## Attribute Reference

* `vcenter_id` - vCenter server ID that contains this Supervisor
* `region_id` - Region ID that consumes this Supervisor
* `cpu_capacity_mhz` - The CPU capacity (in MHz) in this zone. Total CPU consumption in this zone
  cannot cross this limit.
* `cpu_used_mhz` - Total CPU used (in MHz) in this zone.
* `memory_capacity_mib` - The memory capacity (in mebibytes) in this zone. Total memory consumption
  in this zone cannot cross this limit.
* `memory_used_mib` - Total memory used (in mebibytes) in this zone.
