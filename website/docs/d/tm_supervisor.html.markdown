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
data "vcd_vcenter" "one" {
  name = "vcenter-one"
}

data "vcd_tm_supervisor" "one" {
  name       = "my-supervisor-name"
  vcenter_id = data.vcd_vcenter.one.id

  depends_on = [vcd_vcenter.one]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of Supervisor
* `vcenter_id` - (Required) vCenter server ID that contains this Supervisor

## Attribute Reference

* `region_id` - Region ID that consumes this Supervisor
