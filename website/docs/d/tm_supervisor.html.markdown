---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_tm_supervisor"
sidebar_current: "docs-vcd-datasource-tm-supervisor"
description: |-
  Provides a data source to read Supervisors in VMware Cloud Foundation Tenant Manager.
---

# vcd\_tm\_supervisor

Provides a data source to read Supervisors in VMware Cloud Foundation Tenant Manager.

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
