---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_org_vdc_nsxt_network_profile"
sidebar_current: "docs-vcd-resource-nsxt-network-segment-profile"
description: |-
  Provides a resource to configure Segment Profiles for NSX-T Org VDC networks.
---

# vcd\_nsxt\_network\_segment\_profile

Provides a resource to configure Segment Profiles for NSX-T Org VDC networks.

Supported in provider *v3.11+* and VCD 10.4.0+ with NSX-T.

## Example Usage (Segment Profile Template assignment to Org VDC Network)

```hcl
data "vcd_nsxt_segment_profile_template" "complete" {
  name = "complete-profile"
}

data "vcd_nsxt_edgegateway" "existing" {
  org  = "my-org"
  name = "my-gw"
}

resource "vcd_network_routed_v2" "net1" {
  org  = "my-org"
  name = "routed-net"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  gateway       = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "1.1.1.10"
    end_address   = "1.1.1.20"
  }
}

resource "vcd_nsxt_network_segment_profile" "custom-prof" {
  org            = "my-org"
  org_network_id = vcd_network_routed_v2.net1.id

  segment_profile_template_id = data.vcd_nsxt_segment_profile_template.complete.id
}
```

## Example Usage (Custom Segment Profile assignment to Org VDC Network)

```hcl
data "vcd_nsxt_manager" "nsxt" {
  name = "nsxManager1"
}

data "vcd_nsxt_segment_ip_discovery_profile" "first" {
  name            = "ip-discovery-profile-0"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_mac_discovery_profile" "first" {
  name            = "mac-discovery-profile-0"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_spoof_guard_profile" "first" {
  name            = "spoof-guard-profile-0"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_qos_profile" "first" {
  name            = "qos-profile-0"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_security_profile" "first" {
  name            = "segment-security-profile-0"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_edgegateway" "existing" {
  org  = "my-org"
  name = "nsxt-gw-v40"
}

resource "vcd_network_routed_v2" "net1" {
  org  = "my-org"
  name = "routed-net"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  gateway       = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "1.1.1.10"
    end_address   = "1.1.1.20"
  }
}

resource "vcd_nsxt_network_segment_profile" "custom-prof" {
  org            = "my-org"
  org_network_id = vcd_network_routed_v2.net1.id

  ip_discovery_profile_id     = data.vcd_nsxt_segment_ip_discovery_profile.first.id
  mac_discovery_profile_id    = data.vcd_nsxt_segment_mac_discovery_profile.first.id
  spoof_guard_profile_id      = data.vcd_nsxt_segment_spoof_guard_profile.first.id
  qos_profile_id              = data.vcd_nsxt_segment_qos_profile.first.id
  segment_security_profile_id = data.vcd_nsxt_segment_security_profile.first.id
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level
* `org_network_id` - (Required) Org VDC Network ID
* `segment_profile_template_id` - (Optional) Segment Profile Template ID to be applied for this Org
  VDC Network
* `ip_discovery_profile_id` - (Optional) IP Discovery Profile ID. Can be referenced using
  [`vcd_nsxt_segment_ip_discovery_profile`](/providers/vmware/vcd/latest/docs/data-sources/nsxt_segment_ip_discovery_profile)
  data source.
* `mac_discovery_profile_id` - (Optional) MAC Discovery Profile ID. Can be referenced using
  [`vcd_nsxt_segment_mac_discovery_profile`](/providers/vmware/vcd/latest/docs/data-sources/nsxt_segment_mac_discovery_profile)
  data source.
* `spoof_guard_profile_id` - (Optional) Spoof Guard Profile ID. Can be referenced using
  [`vcd_nsxt_segment_spoof_guard_profile`](/providers/vmware/vcd/latest/docs/data-sources/nsxt_segment_spoof_guard_profile)
  data source.
* `qos_profile_id` - (Optional) QoS Profile ID. Can be referenced using
  [`vcd_nsxt_segment_qos_profile`](/providers/vmware/vcd/latest/docs/data-sources/nsxt_segment_qos_profile)
  data source.
* `segment_security_profile_id` - (Optional) Segment Security Profile ID. Can be referenced using
  [`vcd_nsxt_segment_security_profile`](/providers/vmware/vcd/latest/docs/data-sources/nsxt_segment_security_profile)
  data source.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing NSX-T Org VDC Network Segment Profile configuration can be [imported][docs-import] into
this resource via supplying the full dot separated path for Org VDC Network. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxt_network_segment_profile.my-profile org-name.vdc-org-vdc-group-name.org_network_name
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
