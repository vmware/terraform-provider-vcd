---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_segment_security_profile"
sidebar_current: "docs-vcd-data-source-nsxt-segment-security-profile"
description: |-
  Provides a VMware Cloud Director NSX-T Segment Security Profile data source. This can be used to read NSX-T Segment Profile definitions.
---

# vcd\_nsxt\_segment\_security\_profile

Provides a VMware Cloud Director NSX-T Segment Security Profile data source. This can be used to read NSX-T Segment Profile definitions.

Supported in provider *v3.11+*.

## Example Usage (Segment Security Profile)

```hcl
data "vcd_nsxt_manager" "nsxt" {
  name = "nsxManager1"
}

data "vcd_nsxt_segment_security_profile" "first" {
  name            = "segment-security-profile-0"
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

* `description` - Description of Segment Security Profile
* `bpdu_filter_allow_list` - Pre-defined list of allowed MAC addresses to be excluded from BPDU filtering.
* `is_bpdu_filter_enabled` - Defines whether BPDU filter is enabled.
* `is_dhcp_v4_client_block_enabled` - Defines whether DHCP Client block IPv4 is enabled. This filters DHCP Client IPv4 traffic.
* `is_dhcp_v6_client_block_enabled` - Defines whether DHCP Client block IPv6 is enabled. This filters DHCP Client IPv6 traffic.
* `is_dhcp_v4_server_block_enabled` - Defines whether DHCP Server block IPv4 is enabled. This filters DHCP Server IPv4 traffic.
* `is_dhcp_v6_server_block_enabled` - Defines whether DHCP Server block IPv6 is enabled. This filters DHCP Server IPv6 traffic.
* `is_non_ip_traffic_block_enabled` - Defines whether non IP traffic block is enabled. If true, it blocks all traffic except IP/(G)ARP/BPDU.
* `is_ra_guard_enabled` - Defines whether Router Advertisement Guard is enabled. This filters DHCP Server IPv6 traffic.
* `is_rate_limitting_enabled` - Defines whether Rate Limiting is enabled.
* `rx_broadcast_limit` - Incoming broadcast traffic limit in packets per second.
* `rx_multicast_limit` - Incoming multicast traffic limit in packets per second.
* `tx_broadcast_limit` - Outgoing broadcast traffic limit in packets per second.
* `tx_multicast_limit` - Outgoing multicast traffic limit in packets per second.
