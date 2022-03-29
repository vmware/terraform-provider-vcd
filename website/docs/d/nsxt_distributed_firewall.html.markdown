---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_distributed_firewall"
sidebar_current: "docs-vcd-data-source-nsxt-distributed-firewall"
description: |-
  The distributed firewall data source reads all defined rules for a particular VDC Group.
---

# vcd\_nsxt\_distributed\_firewall

The distributed firewall data source reads all defined rules for a particular VDC Group.

## Example Usage

```hcl
data "vcd_vdc_group" "g1" {
  org  = "my-org"
  name = "my-vdc-group"
}
data "vcd_nsxt_distributed_firewall" "t1" {
  org          = "my-org"
  vdc_group_id = data.vcd_vdc_group.g1.id
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which the edge gateway belongs. Optional if defined
  at provider level.
* `vdc_group_id` - (Required) The ID of a VDC Group

## Attribute Reference

All the arguments and attributes defined in
[`vcd_nsxt_distributed_firewall`](/providers/vmware/vcd/latest/docs/resources/nsxt_distributed_firewall)
resource are available.
