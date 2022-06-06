---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_edgegateway_bgp_configuration"
sidebar_current: "docs-vcd-data-source-nsxt-edgegateway-bgp-configuration"
description: |-
  Provides a data source to read BGP configuration on NSX-T Edge Gateway that has a dedicated Tier-0 
  Gateway or VRF.
---

# vcd\_nsxt\_edgegateway\_bgp\_configuration

Provides a data source to read BGP configuration on NSX-T Edge Gateway that has a dedicated Tier-0
  Gateway or VRF. BGP makes core routing decisions by using a table of IP networks, or prefixes,
which designate multiple routes between autonomous systems (AS).

## Example Usage

```hcl
data "vcd_org_vdc" "nsxt-vdc" {
  org  = "my-org"
  name = "my-vdc"
}

data "vcd_nsxt_edgegateway" "existing" {
  org      = "my-org"
  owner_id = data.vcd_org_vdc.nsxt-vdc.id

  name = "main"
}

data "vcd_nsxt_edgegateway_bgp_configuration" "testing" {
  org = "my-org"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which the edge gateway belongs. Optional if defined at provider level.
* `edge_gateway_id` - (Required) An ID of NSX-T Edge Gateway. Can be lookup up using
  [vcd_nsxt_edgegateway](/providers/vmware/vcd/latest/docs/data-sources/nsxt_edgegateway) data source

## Attribute Reference

All the arguments and attributes defined in
[`vcd_nsxt_edgegateway_bgp_configuration`](/providers/vmware/vcd/latest/docs/resources/nsxt_edgegateway_bgp_configuration)
resource are available.
