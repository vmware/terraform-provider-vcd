---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_edgegateway_qos_profile"
sidebar_current: "docs-vcd-data-source-nsxt-qos-profile"
description: |-
  Provides a data source to read NSX-T Edge Gateway QoS profiles. Which can be used to modify NSX-T 
  Edge Gateway Rate Limiting (QoS) configuration.
---

# vcd\_nsxt\_edgegateway\_qos\_profile

Supported in provider *v3.9+* and VCD 10.3.2+ with NSX-T.

Provides a data source to read NSX-T Edge Gateway QoS profiles. Which can be used to modify NSX-T
Edge Gateway Rate Limiting (QoS) configuration.

## Example Usage

```hcl
data "vcd_nsxt_manager" "first" {
  name = "nsxt-manager-name"
}

data "vcd_nsxt_edgegateway_qos_profile" "qos-1" {
  nsxt_manager_id = data.nsxt_manager_id.first.id
  name = "qos-policy-1"
}
```

## Argument Reference

The following arguments are supported:

* `nsxt_manager_id` - (Required) NSX-T Manager where the QoS profile is defined in (can be looked up
using `vcd_nsxt_manager` data source)
* `name` - (Required) QoS Profile Name


## Attribute Reference

* `description` - Description of QoS Profile
* `committed_bandwidth` - Committed bandwith specificd in Mb/s
* `burst_size` - Burst size defined in bytes
* `excess_action` - Excess action defines action on traffic exceeding bandwidth
