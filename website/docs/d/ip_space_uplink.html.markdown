---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_ip_space_uplink"
sidebar_current: "docs-vcd-datasource-ip-space-uplink"
description: |-
  Provides a data source to read IP Space Uplinks in External Networks (Provider Gateways).
---

# vcd\_ip\_space\_uplink

Provides a data source to read IP Space Uplinks in External Networks (Provider Gateways).

## Example Usage

```hcl
data "vcd_ip_space_uplink" "u1" {
  name                = "ip-space-uplink-1"
  external_network_id = vcd_external_network_v2.provider-gateway.id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) - Name of IP Space Uplink
* `external_network_id` - (Required) Parent External Network ID

## Attribute Reference

All the arguments and attributes defined in
[`vcd_ip_space_uplink`](/providers/vmware/vcd/latest/docs/resources/ip_space_uplink) resource are available.
