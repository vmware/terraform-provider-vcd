---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_tm_provider_gateway"
sidebar_current: "docs-vcd-data-source-tm-provider-gateway"
description: |-
  Provides a VMware Cloud Foundation Tenant Manager Provider Gateway data source.
---

# vcd\_tm\_provider\_gateway

Provides a VMware Cloud Foundation Tenant Manager Provider Gateway data source.

## Example Usage

```hcl

```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which the edge gateway belongs. Optional if defined at provider level.
* `vdc` - (Optional) The name of VDC that owns the edge gateway. Optional if defined at provider level.
* `edge_gateway_id` - (Required) An ID of NSX-T Edge Gateway. Can be lookup up using
  [vcd_nsxt_edgegateway](/providers/vmware/vcd/latest/docs/data-sources/nsxt_edgegateway) data source

## Attribute Reference

All the arguments and attributes defined in
[`vcd_tm_provider_gateway`](/providers/vmware/vcd/latest/docs/resources/tm_provider_gateway)
resource are available.