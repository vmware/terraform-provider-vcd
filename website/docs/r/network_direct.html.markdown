---
layout: "vcd"
page_title: "vCloudDirector: vcd_network_direct"
sidebar_current: "docs-vcd-resource-network-direct"
description: |-
  Provides a vCloud Director VDC Network attached to an external one. This can be used to create, modify, and delete internal networks for vApps to connect.
---

# vcd\_network\_direct (*v2.0+*)

Provides a vCloud Director VDC Network directly connected to an external network. This can be used to create,
modify, and delete internal networks for vApps to connect.

## Example Usage

```hcl
resource "vcd_network_direct" "net" {
  org              = "my-org"
  vdc              = "my-vdc"
  name             = "my-net"
  external_network = "my-ext-net"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the network
* `external_network` - (Required) The name of the external network.
* `shared` - (Optional) Defines if this network is shared between multiple vDCs
  in the vOrg.  Defaults to `false`.

