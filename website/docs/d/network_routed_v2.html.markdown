---
layout: "vcd"
page_title: "vCloudDirector: vcd_network_routed_v2"
sidebar_current: "docs-vcd-data-source-network-routed-v2"
description: |-
  Provides a vCloud Director Org VDC routed Network. This can be used to reference internal networks for vApps to connect.
---

# vcd\_network\_routed\_v2

Provides a vCloud Director Org VDC routed Network data source. This can be used to reference internal networks for vApps to connect.

Supported in provider *v3.2+* for both NSX-T and NSX-V VDCs.

## Example Usage

```hcl
data "vcd_network_routed_v2" "net" {
  org  = "my-org" # Optional
  vdc  = "my-vdc" # Optional
  name = "my-net"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level
* `name` - (Required) A unique name for the network

## Attribute reference

All attributes defined in [routed network resource](/docs/providers/vcd/r/network_routed_v2.html#attribute-reference) are supported.
