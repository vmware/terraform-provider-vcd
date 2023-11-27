---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_edgegateway_bgp_ip_prefix_list"
sidebar_current: "docs-vcd-datasource-nsxt-edgegateway-bgp-ip-prefix-list"
description: |-
  Provides a data source to manage NSX-T Edge Gateway BGP IP Prefix Lists. IP prefix lists can
  contain single or multiple IP addresses and can be used to assign BGP neighbors with access
  permissions for route advertisement.
---

# vcd\_nsxt\_edgegateway\_bgp\_ip\_prefix\_list

Supported in provider *v3.7+* and VCD 10.2+ with NSX-T

Provides a resource to manage NSX-T Edge Gateway BGP IP Prefix Lists. IP prefix lists can contain 
single or multiple IP addresses and can be used to assign BGP neighbors with access permissions 
for route advertisement.

## Example Usage

```hcl
data "vcd_vdc_group" "g1" {
  org  = "my-org"
  name = "my-vdc-group"
}

data "vcd_nsxt_edgegateway" "testing" {
  org      = "my-org"
  owner_id = data.vcd_vdc_group.g1.id

  name = "my-edge-gateway"
}

data "vcd_nsxt_edgegateway_bgp_ip_prefix_list" "testing" {
  org             = "my-org"
  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  name = "my-bgp-prefix-list"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which the edge gateway belongs. Optional if defined at provider level.
* `edge_gateway_id` - (Required) An ID of NSX-T Edge Gateway. Can be lookup up using
  [vcd_nsxt_edgegateway](/providers/vmware/vcd/latest/docs/data-sources/nsxt_edgegateway) data source
* `name` - (Required) A name of existing BGP IP Prefix List in specified Edge Gateway

## Attribute Reference

All the arguments and attributes defined in
[`vcd_nsxt_edgegateway_bgp_ip_prefix_list`](/providers/vmware/vcd/latest/docs/resources/nsxt_edgegateway_bgp_ip_prefix_list)
resource are available.
