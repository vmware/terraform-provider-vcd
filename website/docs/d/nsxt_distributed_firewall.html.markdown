---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_distributed_firewall"
sidebar_current: "docs-vcd-data-source-nsxt-distributed-firewall"
description: |-
  The Distributed Firewall data source reads all defined rules for a particular VDC Group.
---

# vcd\_nsxt\_distributed\_firewall

The Distributed Firewall data source reads all defined rules for a particular VDC Group.

## Example Usage

```hcl
data "vcd_vdc_group" "g1" {
  org  = "my-org" # Optional, can be inherited from Provider configuration
  name = "my-vdc-group"
}

data "vcd_nsxt_distributed_firewall" "t1" {
  org          = "my-org"  # Optional, can be inherited from Provider configuration
  vdc_group_id = data.vcd_vdc_group.g1.id
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization in which Distributed Firewall is located. Optional if
  defined at provider level.
* `vdc_group_id` - (Required) The ID of a VDC Group

## Attribute Reference

All the arguments and attributes defined in
[`vcd_nsxt_distributed_firewall`](/providers/vmware/vcd/latest/docs/resources/nsxt_distributed_firewall)
resource are available.
