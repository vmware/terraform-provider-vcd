---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_tm_supervisor"
sidebar_current: "docs-vcd-datasource-tm-supervisor"
description: |-
  Provides a data source to read Supervisors.
---

# vcd\_tm\_supervisor

Provides a data source to read Supervisors.

## Example Usage

```hcl
data "vcd_tm_supervisor" "one" {
  name = "my-supervisor-name"

  depends_on = [vcd_vcenter.one]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of Supervisor.

## Attribute Reference

* `vcenter_id` - vCenter server ID that contains this Supervisor
* `region_id` - Region ID that consumes this Supervisor
