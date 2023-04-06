---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_edgegateway_rate_limiting"
sidebar_current: "docs-vcd-data-source-nsxt-edge-rate-limiting"
description: |-
  Provides a data source to read NSX-T Edge Gateway Rate Limiting (QoS) configuration.
---

# vcd\_nsxt\_edgegateway\_rate\_limiting

Supported in provider *v3.9+* and VCD 10.3.2+ with NSX-T.

Provides a data source to read NSX-T Edge Gateway Rate Limiting (QoS) configuration.

## Example Usage

```hcl
data "vcd_org_vdc" "v1" {
  org  = "datacloud"
  name = "nsxt-vdc-datacloud"
}

data "vcd_nsxt_edgegateway" "in-vdc" {
  org      = "datacloud"
  owner_id = data.vcd_org_vdc.v1.id

  name = "nsxt-gw-datacloud"
}

data "vcd_nsxt_edgegateway_rate_limiting" "in-vdc" {
  org             = "datacloud"
  edge_gateway_id = data.vcd_nsxt_edgegateway.in-vdc.id
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Required) Org in which the NSX-T Edge Gateway is located
* `edge_gateway_id` - (Required) NSX-T Edge Gateway ID

## Attribute Reference

All the arguments and attributes defined in
[`vcd_nsxt_edgegateway_rate_limiting`](/providers/vmware/vcd/latest/docs/resources/nsxt_edgegateway_rate_limiting)
resource are available.
