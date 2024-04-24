---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_solution_add_on"
sidebar_current: "docs-vcd-data-source-solution-add-on"
description: |-
  Provides a data source to read Solution Add-Ons in Cloud Director. A solution add-on is the
  representation of a solution that is custom built for VMware Cloud Director in the VMware Cloud
  Director extensibility ecosystem. A solution add-on can encapsulate UI and API VMware Cloud
  Director extensions together with their backend services and lifecycle management. Solution
  аdd-оns are distributed as .iso files. A solution add-on can contain numerous
  elements: UI plugins, vApps, users, roles, runtime defined entities, and more.
---

# vcd\_solution\_add\_on

Supported in provider *v3.13+* and VCD 10.4.1+.

Provides a data source to read Solution Add-Ons in Cloud Director. A solution add-on is the
representation of a solution that is custom built for VMware Cloud Director in the VMware Cloud
Director extensibility ecosystem. A solution add-on can encapsulate UI and API VMware Cloud Director
extensions together with their backend services and lifecycle management. Solution аdd-оns are
distributed as .iso files. A solution add-on can contain numerous elements: UI plugins, vApps,
users, roles, runtime defined entities, and more.

## Example Usage

```hcl

```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which the edge gateway belongs. Optional if defined at provider level.


## Attribute Reference

All the arguments and attributes defined in
[`vcd_solution_add_on`](/providers/vmware/vcd/latest/docs/resources/solution_add_on) resource are
available.