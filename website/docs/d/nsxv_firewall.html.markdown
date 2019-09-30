---
layout: "vcd"
page_title: "vCloudDirector: vcd_nsxv_firewall"
sidebar_current: "docs-vcd-data-source-nsxv-firewall"
description: |-
  Provides a vCloud Director firewall data sourcefor advanced edge gateways (NSX-V). This can be
  used to read existing rule by ID and use its attributes in other resources.
---

# vcd\_nsxv\_firewall

Provides a vCloud Director firewall data sourcefor advanced edge gateways (NSX-V). This can be
used to read existing rule by ID and use its attributes in other resources.

~> **Note:** This data source requires advanced edge gateway. For non-advanced edge gateways please
use the [`vcd_firewall_rules`](/docs/providers/vcd/r/firewall_rules.html) resource.

## Example Usage

```hcl
data "vcd_nsxv_firewall" "my-rule" {
  org                 = "my-org"
  vdc                 = "my-org-vdc"
  edge_gateway        = "my-edge-gw"

  rule_id = "133048"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations.
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level.
* `edge_gateway` - (Required) The name of the edge gateway on which to apply the DNAT rule.
* `rule_id` - (Required) ID of firewall rule as shown in the UI.

## Attribute Reference

All the attributes defined in [`vcd_nsxv_firewall`](/docs/providers/vcd/r/nsxv_firewall.html)
resource are available.
