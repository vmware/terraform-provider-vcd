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
* `nsxt_manager_id` - (Optional) Segment Profile search context. Use when searching by NSX-T manager.
* `vdc_id` - (Optional) Segment Profile search context. Use when searching by VDC
* `vdc_group_id` - (Optional) Segment Profile search context. Use when searching by VDC group

-> Note: only one of `nsxt_manager_id`, `vdc_id`, `vdc_group_id` can be used


## Attribute reference

* `description` -  Description of IP Discovery Profile
* `arp_binding_limit` - Indicates the number of ARP snooped IP addresses to be remembered per
  logical port
* `arp_binding_timeout` - ARP and ND (Neighbor Discovery) cache timeout (in minutes)
* `is_arp_snooping_enabled` - Defines whether ARP snooping is enabled
* `is_dhcp_snooping_v4_enabled` - Defines whether DHCP snooping for IPv4 is enabled
* `is_dhcp_snooping_v6_enabled` - Defines whether DHCP snooping for IPv6 is enabled
* `is_duplicate_ip_detection_enabled` - Defines whether duplicate IP detection is enabled. Duplicate
  IP detection is used to determine if there is any IP conflict with any other port on the same
  logical switch. If a conflict is detected, then the IP is marked as a duplicate on the port where
  the IP was discovered last
* `is_nd_snooping_enabled` - Defines whether ND (Neighbor Discovery) snooping is enabled. If true,
  this method will snoop the NS (Neighbor Solicitation) and NA (Neighbor Advertisement) messages in
  the ND (Neighbor Discovery Protocol) family of messages which are transmitted by a VM. From the NS
  messages, we will learn about the source which sent this NS message. From the NA message, we will
  learn the resolved address in the message which the VM is a recipient of. Addresses snooped by
  this method are subject to TOFU
* `is_tofu_enabled` - Defines whether `Trust on First Use(TOFU)` paradigm is enabled
* `is_vmtools_v4_enabled` - Defines whether fetching IPv4 address using vm-tools is enabled. This
  option is only supported on ESX where vm-tools is installed
* `is_vmtools_v6_enabled` - Defines whether fetching IPv6 address using vm-tools is enabled. This
  will learn the IPv6 addresses which are configured on interfaces of a VM with the help of the
  VMTools software
* `nd_snooping_limit` - Maximum number of ND (Neighbor Discovery Protocol) snooped IPv6 addresses
