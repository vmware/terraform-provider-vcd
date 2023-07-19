---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_distributed_firewall_rule"
sidebar_current: "docs-vcd-data-source-nsxt-distributed-firewall-rule"
description: |-
  The Distributed Firewall data source reads a single rule for a particular VDC Group.
---

# vcd\_nsxt\_distributed\_firewall\_rule

The Distributed Firewall data source reads a single rule for a particular VDC Group.

-> There is a different data source
[`vcd_nsxt_distributed_firewall`](/providers/vmware/vcd/latest/docs/data-sources/nsxt_distributed_firewall)
resource available that can fetch all firewall rules.

## Example Usage

```hcl
data "vcd_vdc_group" "g1" {
  org  = "my-org" # Optional, can be inherited from Provider configuration
  name = "my-vdc-group"
}

data "vcd_nsxt_distributed_firewall_rule" "r1" {
  org          = "my-org" # Optional, can be inherited from Provider configuration
  vdc_group_id = data.vcd_vdc_group.g1.id

  name = "rule1"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization in which Distributed Firewall is located. Optional if
  defined at provider level.
* `vdc_group_id` - (Required) The ID of a VDC Group
* `name` - (Required) The name of firewall rule

## Attribute Reference

All the arguments and attributes defined in
[`vcd_nsxt_distributed_firewall_rule`](/providers/vmware/vcd/latest/docs/resources/nsxt_distributed_firewall_rule)
resource are available.
