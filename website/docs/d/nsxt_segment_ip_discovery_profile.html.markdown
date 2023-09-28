---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_segment_ip_discovery_profile"
sidebar_current: "docs-vcd-data-source-nsxt-segment-ip-discovery-profile"
description: |-
  Provides a VMware Cloud Director NSX-T IP Discovery Profile data source. This can be used to read NSX-T Segment Profile definitions.
---

# vcd\_nsxt\_segment\_ip\_discovery\_profile

Provides a VMware Cloud Director NSX-T IP Discovery Profile data source. This can be used to read NSX-T Segment Profile definitions.

Supported in provider *v3.11+*.

## Example Usage (IP Discovery Profile)

```hcl
data "vcd_nsxt_manager" "nsxt" {
  name = "nsxManager1"
}

data "vcd_nsxt_segment_ip_discovery_profile" "first" {
  name            = "ip-discovery-profile-0"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}
```


## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of Segment Profile
* `nsxt_manager_id` - (Optional) Segment Profile search context. One of `nsxt_manager_id`, `vdc_id`, `vdc_group_id` is required
* `vdc_id` - (Optional) Segment Profile search context. One of `nsxt_manager_id`, `vdc_id`, `vdc_group_id` is required
* `vdc_group_id` - (Optional) Segment Profile search context. One of `nsxt_manager_id`, `vdc_id`, `vdc_group_id` is required


## Attribute reference

* `` 