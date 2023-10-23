---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_org_vdc_nsxt_network_profile"
sidebar_current: "docs-vcd-data-source-nsxt-segment-profile-template"
description: |-
  Provides a data source to read Network Profile for NSX-T VDCs.
---

# vcd\_org\_vdc\_nsxt\_network\_profile

Provides a data source to read Network Profile for NSX-T VDCs.

Supported in provider *v3.11+* and VCD 10.4.0+ with NSX-T.

## Example Usage

```hcl
data "vcd_org_vdc_nsxt_network_profile" "nsxt" {
  org = "my-org"
  vdc = "my-vdc"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level

## Attribute reference

* `edge_cluster_id` - An ID of NSX-T Edge Cluster which should provide vApp
  Networking Services or DHCP for Isolated Networks. This value **might be unavailable** in data
  source if user has insufficient rights.
* `vdc_networks_default_segment_profile_template_id` - Default Segment Profile ID for all Org VDC
  networks withing this VDC
* `vapp_networks_default_segment_profile_template_id`- Default Segment Profile ID for all vApp
  networks withing this VDC

All other attributes are defined in [organization VDC network profile
resource](/providers/vmware/vcd/latest/docs/resources/org_vdc_nsxt_network_profile.html#attribute-reference)
are supported.

