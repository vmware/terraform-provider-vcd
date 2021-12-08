---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_data_center_group"
sidebar_current: "docs-vcd-data-source-data-center-group"
description: |-
  Provides a data source to read data center groups.
---

# vcd\_data\_center\_group
Supported in provider *v3.5+* and VCD 10.2+.

Provides a data source to read data center group and reference in other resources.

~> Only `System Administrator` and `Org Users` with right `View VDC Group` can access data center groups using this data source.

## Example Usage

```hcl
data "vcd_org_vdc" "startVdc" {
  org  = "myOrg"
  name = "myVDC"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional)  - name of data center group
* `id` - (Optional)  - ID of data center group

`name` or `id` is required field.

## Attribute Reference

All the arguments and attributes defined in
[`vcd_data_center_group`](/providers/vmware/vcd/latest/docs/resources/vcd_data_center_group) resource are available.