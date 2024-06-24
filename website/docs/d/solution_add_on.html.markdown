---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_solution_add_on"
sidebar_current: "docs-vcd-data-source-solution-add-on"
description: |-
  Provides a data source to read Solution Add-Ons in Cloud Director. A Solution Add-On is the
  representation of a solution that is custom built for Cloud Director in the Cloud
  Director extensibility ecosystem. A Solution Add-On can encapsulate UI and API Cloud
  Director extensions together with their backend services and lifecycle management. Solution
  Add-Ons are distributed as .iso files. A Solution Add-on can contain numerous
  elements: UI plugins, vApps, users, roles, runtime defined entities, and more.
---

# vcd\_solution\_add\_on

Supported in provider *v3.13+* and VCD 10.4.1+.

Provides a data source to read Solution Add-Ons in Cloud Director. A Solution Add-On is the
representation of a solution that is custom built for Cloud Director in the Cloud
Director extensibility ecosystem. A Solution Add-On can encapsulate UI and API Cloud Director
extensions together with their backend services and lifecycle management. Solution аdd-оns are
distributed as .iso files. A Solution Add-On can contain numerous elements: UI plugins, vApps,
users, roles, runtime defined entities, and more.

## Example Usage

```hcl
data "vcd_solution_add_on" "dse14" {
  name = "vmware.ds-1.4.0-23376809"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Solution Add-On name, e.g. `vmware.ds-1.4.0-23376809`. Solution Add-On
  resource [`vcd_solution_add_on`](/providers/vmware/vcd/latest/docs/resources/solution_add_on)
  `import` with `list@` capability can help listing available names.


## Attribute Reference

All the arguments and attributes defined in
[`vcd_solution_add_on`](/providers/vmware/vcd/latest/docs/resources/solution_add_on) resource are
available.