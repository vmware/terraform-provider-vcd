---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_network_routed_v2"
sidebar_current: "docs-vcd-data-source-network-routed-v2"
description: |-
  Provides a VMware Cloud Director Org VDC routed Network data source to read data or reference  existing network
  (backed by NSX-T or NSX-V).
---

# vcd\_network\_routed\_v2

Provides a VMware Cloud Director Org VDC routed Network data source to read data or reference  existing network
(backed by NSX-T or NSX-V).

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
* `filter` - (Optional) Retrieves the data source using one or more filter parameters

## Attribute reference

All attributes defined in [routed network resource](/providers/vmware/vcd/latest/docs/resources/network_routed_v2#attribute-reference) are supported.

## Filter arguments

* `name_regex` (Optional) matches the name using a regular expression.
* `ip` (Optional) matches the IP of the resource using a regular expression.

See [Filters reference](/providers/vmware/vcd/latest/docs/guides/data_source_filters) for details and examples.
