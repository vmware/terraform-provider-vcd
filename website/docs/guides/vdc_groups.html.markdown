---
layout: "vcd"
page_title: "VMware Cloud Director: VDC Groups"
sidebar_current: "docs-vcd-guides-vdc-groups"
description: |-
  Provides guidance to VDC Group support
---

# VDC Groups in VMware Cloud Director 10.2+

## About 

Starting with version 10.2, VMware Cloud Director supports data center group networking backed by
NSX-T Data Center.

To create a network across multiple organization VDCs, you first group the VDCs and then create a
group network that is shared with them.

Data center group networks backed by NSX-T Data Center provide level-2 network sharing, single
active egress point configuration, and distributed firewall (DFW) rules that are applied across a
data center group.

### Data center group 

A data center group acts as a cross-VDC router that provides centralized networking administration,
egress point configuration, and east-west traffic between all networks within the group. A data
center group can contain between one and 16 VDCs that you configure to share an active egress point. 

### Availability zone 

An availability zone represents the compute clusters or compute fault domains that are available to
the network. By default, the availability zone is the provider VDC. 

### Egress point 

An existing NSX-T Data Center Edge Gateway that you configure to connect a data center group to an
external network. 

## Requirements

VDC Group support requires:

* Terraform Provider VCD 3.6+
* VMware Cloud Director 10.2+

## Terraform Provider VCD support 

This document describes features that were introduced in Terraform Provider VCD 3.6.0+ for VDC Group
support. Earlier versions of Terraform Provider VCD do not support VDC Groups.

### List of resources that support VDC Groups (NSX-T only)

The following list of resources (and their corresponding data sources) support NSX-T VDC Groups (no
NSX-V support is provided):

* [vcd_nsxt_edgegateway](/providers/vmware/vcd/latest/docs/resources/nsxt_edgegateway)
* [vcd_network_routed_v2](/providers/vmware/vcd/latest/docs/resources/network_routed_v2)

The next sub-sections will cover some specific overview for each resource

#### Resource vcd_nsxt_edgegateway

New fields for VDC Groups:

* `owner_id` (replaces deprecated `vdc` field in resource and inherited from provider
  configuration). This field now supports both - VDC and VDC Group IDs. 
* `starting_vdc_id` is an optional field and is only useful if `owner_id` is a VDC Group. NSX-T Edge
  Gateway cannot be created directly in VDC Group - at first it must originate in a VDC (which is a
  member of destination VDC Group). The initial VDC defines Egress point for traffic and picking
  right VDC might be important when VDC Group spans multiple availability zones. When this field is
  not specified, a random member of destination VDC Group will be picked for Edge Gateway creation
  and then immediately moved to VDC Group as specified in `owner_id`.

#### Resource vcd_network_routed_v2

Terraform Provider VCD 3.6.0 changes behavior of `vcd_network_routed_v2` resource. It __does not
require__ to specify `vdc` or `owner_id` fields. Instead, it inherits VDC or VDC Group membership
directly from parent Edge Gateway (specified in `edge_gateway_id`). The reason for this is that
routed Org VDC networks travel to and from VDC Groups with parent Edge Gateway and this does not
work well with Terraform concept.

## References

* [VMware Cloud Director Documentation about VDC
  Groups](https://docs.vmware.com/en/VMware-Cloud-Director/10.2/VMware-Cloud-Director-Tenant-Portal-Guide/GUID-E8A8CD70-31AD-4592-B520-34E3B7DC4E6E.html)
