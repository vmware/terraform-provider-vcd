---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_segment_qos_profile"
sidebar_current: "docs-vcd-data-source-nsxt-segment-qos-profile"
description: |-
  Provides a VMware Cloud Director NSX-T QoS Profile data source. This can be used to read NSX-T Segment Profile definitions.
---

# vcd\_nsxt\_segment\_qos\_profile

Provides a VMware Cloud Director NSX-T QoS Profile data source. This can be used to read NSX-T Segment Profile definitions.

Supported in provider *v3.11+*.

## Example Usage (QoS Profile)

```hcl
data "vcd_nsxt_manager" "nsxt" {
  name = "nsxManager1"
}

data "vcd_nsxt_segment_qos_profile" "first" {
  name            = "qos-profile-0"
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

* `description` - Description of QoS Profile
* `class_of_service` - Class of service groups similar types of traffic in the network and each type
  of traffic is treated as a class with its own level of service priority. The lower priority
  traffic is slowed down or in some cases dropped to provide better throughput for higher priority
  traffic.
* `dscp_priority` - A Differentiated Services Code Point (DSCP) priority
  Profile. 
* `dscp_trust_mode` - A Differentiated Services Code Point (DSCP) trust mode. Values are below:
  * `TRUSTED` - With Trusted mode the inner header DSCP value is applied to the outer IP header for
    IP/IPv6 traffic. For non IP/IPv6 traffic, the outer IP header takes the default value.
  * `UNTRUSTED` - Untrusted mode is supported on overlay-based and VLAN-based logical port. 
* `egress_rate_limiter_avg_bandwidth` - Average egress bandwidth in Mb/s.
* `egress_rate_limiter_burst_size` - Egress burst size in bytes.
* `egress_rate_limiter_peak_bandwidth` - Peak egress bandwidth in Mb/s.
* `ingress_broadcast_rate_limiter_avg_bandwidth` - Average ingress broadcast bandwidth in Mb/s.
* `ingress_broadcast_rate_limiter_burst_size` - Ingress broadcast burst size in bytes.
* `ingress_broadcast_rate_limiter_peak_bandwidth` - Peak ingress broadcast bandwidth in Mb/s.
* `ingress_rate_limiter_avg_bandwidth` - Average ingress bandwidth in Mb/s.
* `ingress_rate_limiter_burst_size` - Iingress burst size in bytes.
* `ingress_rate_limiter_peak_bandwidth` - Peak ingress broadcast bandwidth in Mb/s.
