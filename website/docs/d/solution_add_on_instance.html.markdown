---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_solution_add_on_instance"
sidebar_current: "docs-vcd-data-source-solution-add-on-instance"
description: |-
  Provides a data source to read Solution Add-On Instances in Cloud Director. A Solution Add-On Instance
  is created from an existing Solution Add-On by supplying configuration values of that particular instance.
---

# vcd\_solution\_add\_on\_instance

Supported in provider *v3.13+* and VCD 10.4.1+.

Provides a data source to read Solution Add-On Instances in Cloud Director. A Solution Add-On
Instance is created from an existing Solution Add-On by supplying configuration values of that
particular instance.

## Example Usage

```hcl
data "vcd_solution_add_on_instance" "dse14" {
  name = "MyDseInstance"
}
```

## Attribute Reference

All the arguments and attributes defined in
[`vcd_solution_add_on_instance`](/providers/vmware/vcd/latest/docs/resources/solution_add_on_instance)
resource are available.
