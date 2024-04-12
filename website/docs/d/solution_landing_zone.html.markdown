---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_solution_landing_zone"
sidebar_current: "docs-vcd-data-source-solution-landing-zone"
description: |-
  Provides a data source to read VCD Solution Add-on Landing Zone.
---

# vcd\_solution\_landing\_zone

Supported in provider *v3.13+* and VCD 10.4+.

Provides a data source to read VCD Solution Add-on Landing Zone.

~> Only `System Administrator` can create this resource.

## Example Usage

```hcl
data "vcd_solution_landing_zone" "slz" {}
```

## Argument Reference

No arguments are required because this is a global configuration for VCD

## Attribute Reference

All the attributes defined in
[`vcd_solution_landing_zone`](/providers/vmware/vcd/latest/docs/resources/solution_landing_zone)
resource are available.
