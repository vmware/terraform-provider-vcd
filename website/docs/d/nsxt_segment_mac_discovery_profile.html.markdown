---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_segment_mac_discovery_profile"
sidebar_current: "docs-vcd-data-source-nsxt-segment-mac-discovery-profile"
description: |-
  Provides a VMware Cloud Director NSX-T MAC Discovery Profile data source. This can be used to read NSX-T Segment Profile definitions.
---

# vcd\_nsxt\_segment\_mac\_discovery\_profile

Provides a VMware Cloud Director NSX-T MAC Discovery Profile data source. This can be used to read NSX-T Segment Profile definitions.

Supported in provider *v3.11+*.

## Example Usage (MAC Discovery Profile)

```hcl
data "vcd_nsxt_manager" "nsxt" {
  name = "nsxManager1"
}

data "vcd_nsxt_segment_mac_discovery_profile" "first" {
  name            = "mac-discovery-profile-0"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}
```


## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of Segment Profile
* `nsxt_manager_id` - (Optional) Segment Profile search context. Use when searching by NSX-T manager.
* `vdc_id` - (Optional) Segment Profile search context. Use when searching by VDC
* `vdc_group_id` - (Optional) Segment Profile search context. Use when searching by VDC group

-> Note: only one of `nsxt_manager_id`, `vdc_id`, `vdc_group_id` can be used

## Attribute reference

* `description` - Description of MAC Discovery Profile
* `is_mac_change_enabled` - Defines whether source MAC address change is enabled
* `is_mac_learning_enabled` - Defines whether source MAC address learning is enabled
* `is_unknown_unicast_flooding_enabled` - Defines whether unknown unicast flooding rule is enabled.
  This allows flooding for unlearned MAC for ingress traffic
* `mac_learning_aging_time` - Aging time in seconds for learned MAC address. Indicates how long
  learned MAC address remain
* `mac_limit` - The maximum number of MAC addresses that can be learned on this port
* `mac_policy` - The policy after MAC Limit is exceeded. It can be either `ALLOW` or `DROP`