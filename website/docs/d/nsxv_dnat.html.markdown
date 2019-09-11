---
layout: "vcd"
page_title: "vCloudDirector: vcd_nsxv_dnat"
sidebar_current: "docs-vcd-data-source-nsxv-dnat"
description: |-
  Provides a vCloud Director DNAT data source. This can be used to create, modify,
  and delete destination NATs to map external IPs to a VM.
---

# vcd\_nsxv\_dnat

Provides a vCloud Director DNAT data source. This can be used to read existing rule by ID and
use its attributes in other resources.

## Example Usage

```hcl
data "vcd_nsxv_dnat" "my-rule" {
  org                 = "my-org"
  vdc                 = "my-org-vdc"
  edge_gateway        = "my-edge-gw"

  rule_id = "197864"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations.
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level.
* `edge_gateway` - (Required) The name of the edge gateway on which to apply the DNAT rule.
* `rule_id` - (Required) ID of DNAT rule as shown in the UI

## Attribute Reference

All the attributes defined in `vcd_nsxv_dnat` resource are available.
