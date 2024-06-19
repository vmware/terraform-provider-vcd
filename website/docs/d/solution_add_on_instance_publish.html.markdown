---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_solution_add_on_instance_publish"
sidebar_current: "docs-vcd-data-source-solution-add-on-instance-publish"
description: |-
  Provides a data source to read publishing configuration of Solution Add-On Instances in Cloud Director.
---

# vcd\_solution\_add\_on\_instance\_publish

Supported in provider *v3.13+* and VCD 10.4.1+.

Provides a data source to read publishing configuration of Solution Add-On Instances in Cloud Director.

## Example Usage

```hcl
data "vcd_solution_add_on_instance_publish" "public" {
  add_on_instance_name = "MyDseInstanceName"
}
```

## Argument Reference

The following arguments are supported:

* `add_on_instance_name` - (Required) The name of Solution Add-On Instance


## Attribute Reference

All the arguments and attributes defined in
[`vcd_solution_add_on_instance_publish`](/providers/vmware/vcd/latest/docs/resources/solution_add_on_instance_publish) resource are
available.