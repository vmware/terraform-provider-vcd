---
layout: "vcd"
page_title: "vCloudDirector: vcd_network_direct"
sidebar_current: "docs-vcd-resource-network-direct"
description: |-
  Provides a vCloud Director Org VDC Network attached to an external one. This can be used to create, modify, and delete internal networks for vApps to connect.
---

# vcd\_network\_direct

Provides a vCloud Director Org VDC Network directly connected to an external network. This can be used to create,
modify, and delete internal networks for vApps to connect.

Supported in provider *v2.0+*

## Example Usage

```hcl
resource "vcd_network_direct" "net" {
  org = "my-org" # Optional
  vdc = "my-vdc" # Optional

  name             = "my-net"
  external_network = "my-ext-net"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional; *v2.0+*) The name of organization to use, optional if defined at provider level. Useful when 
  connected as sysadmin working across different organisations
* `vdc` - (Optional; *v2.0+*) The name of VDC to use, optional if defined at provider level
* `name` - (Required) A unique name for the network
* `external_network` - (Required) The name of the external network.
* `shared` - (Optional) Defines if this network is shared between multiple vDCs
  in the vOrg.  Defaults to `false`.

