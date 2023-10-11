---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_segment_profile_template"
sidebar_current: "docs-vcd-resource-nsxt-segment-profile-template"
description: |-
  Provides a resource to manage NSX-T Segment Profile Templates.
---

# vcd\_nsxt\_segment\_profile\_template

Provides a resource to manage NSX-T Segment Profile Templates.

Supported in provider *v3.11+* and VCD 10.4.0+ with NSX-T. Requires System Administrator privileges.

## Example Usage (Example with all Segment Profiles)

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

resource "vcd_nsxt_segment_profile_template" "complete" {
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id

  name        = "my-first-segment-profile-template"
  description = "my description"

  ip_discovery_profile_id     = data.vcd_nsxt_segment_ip_discovery_profile.first.id
  mac_discovery_profile_id    = data.vcd_nsxt_segment_mac_discovery_profile.first.id
  spoof_guard_profile_id      = data.vcd_nsxt_segment_spoof_guard_profile.first.id
  qos_profile_id              = data.vcd_nsxt_segment_qos_profile.first.id
  segment_security_profile_id = data.vcd_nsxt_segment_security_profile.first.id
}
```

## Argument Reference

The following arguments are supported:

* `nsxt_manager_id` - (Required) NSX-T Manager ID (can be referenced using
  [`vcd_nsxt_manager`](/providers/vmware/vcd/latest/docs/data-sources/nsxt_manager) datasource)
* `name` - (Required) Name for Segment Profile Template
* `description` - (Optional) Description of Segment Profile Template
* `ip_discovery_profile_id` - (Optional) IP Discovery Profile ID. can be referenced using
  [`vcd_nsxt_segment_ip_discovery_profile`](/providers/vmware/vcd/latest/docs/data-sources/nsxt_segment_ip_discovery_profile)
* `mac_discovery_profile_id` - (Optional) IP Discovery Profile ID. can be referenced using
  [`vcd_nsxt_segment_mac_discovery_profile`](/providers/vmware/vcd/latest/docs/data-sources/nsxt_segment_mac_discovery_profile)
* `spoof_guard_profile_id` - (Optional) IP Discovery Profile ID. can be referenced using
  [`vcd_nsxt_segment_spoof_guard_profile`](/providers/vmware/vcd/latest/docs/data-sources/nsxt_segment_spoof_guard_profile)
* `qos_profile_id` - (Optional) IP Discovery Profile ID. can be referenced using
  [`vcd_nsxt_segment_qos_profile`](/providers/vmware/vcd/latest/docs/data-sources/nsxt_segment_qos_profile)
* `segment_security_profile_id` - (Optional) IP Discovery Profile ID. can be referenced using
  [`vcd_nsxt_segment_security_profile`](/providers/vmware/vcd/latest/docs/data-sources/nsxt_segment_security_profile)


## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing NSX-T Segment Profile Template configuration can be [imported][docs-import] into this
resource via supplying path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxt_segment_profile_template.imported segment-profile-name
```

The above would import the `segment-profile-name` NSX-T Segment Profile Template.
