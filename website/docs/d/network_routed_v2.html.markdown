---
layout: "vcd"
page_title: "vCloudDirector: vcd_network_routed_v2"
sidebar_current: "docs-vcd-data-source-network-routed-v2"
description: |-
  Provides a vCloud Director Org VDC routed Network. This can be used to reference internal networks for vApps to connect.
---

# vcd\_network\_routed\_v2

Provides a VMware Cloud Director Org VDC routed Network data source. This can be used to reference internal networks for vApps to connect.

Supported in provider *v3.2+* for both NSX-T and NSX-V VDCs.

## Example Usage

```hcl
data "vcd_network_routed_v2" "net" {
  org  = "my-org" # Optional
  vdc  = "my-vdc" # Optional
  name = "my-net"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level
* `name` - (Required) A unique name for the network (optional when `filter` is used)
* `filter` - (Optional; *2.9+*) Retrieves the data source using one or more filter parameters

## Attribute reference

All attributes defined in [routed network resource](/docs/providers/vcd/r/network_routed_v2.html#attribute-reference) are supported.

## Filter arguments

(Supported in provider *v2.9+*)

* `name_regex` (Optional) matches the name using a regular expression.
* `ip` (Optional) matches the IP of the resource using a regular expression.
* `metadata` (Optional) One or more parameters that will match metadata contents.

See [Filters reference](/docs/providers/vcd/guides/data_source_filters.html) for details and examples.
