---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_org_vdc_nsxt_network_profile"
sidebar_current: "docs-vcd-datasource-nsxt-network-segment-profile"
description: |-
  Provides a data source to read Segment Profile configuration for NSX-T Org VDC networks.
---

# vcd\_nsxt\_network\_segment\_profile

Provides a data source to read Segment Profile configuration for NSX-T Org VDC networks.

Supported in provider *v3.11+* and VCD 10.4.0+ with NSX-T.

## Example Usage

```hcl
data "vcd_nsxt_network_segment_profile" "custom-prof" {
  org            = "my-org"
  org_network_id = vcd_network_routed_v2.net1.id
}
```


## Argument Reference

The following arguments are supported:

* `org` - (Optional, but required if not set at provider level) Org name 
* `org_network_id` - (Required) Org VDC Network ID

## Attribute Reference
 
All the arguments and attributes defined in
[`vcd_nsxt_network_segment_profile`](/providers/vmware/vcd/latest/docs/resources/nsxt_network_segment_profile)
resource are available.
