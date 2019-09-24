---
layout: "vcd"
page_title: "vCloudDirector: vcd_network_direct"
sidebar_current: "docs-vcd-data-source-network-direct"
description: |-
  Provides a vCloud Director Org VDC Network attached to an external one. This can be used to reference internal networks for vApps to connect.
---

# vcd\_network\_direct

Provides a vCloud Director Org VDC Network data source directly connected to an external network. This can be used to reference
internal networks for vApps to connect.

Supported in provider *v2.5+*


## Example Usage

```hcl
data "vcd_network_direct" "net" {
  org  = "my-org"
  vdc  = "my-vdc"
  name = "my-net"
}

# Get the name of the external network from the data source
# and use it to establish a second data source
output "external_network" {
  value = data.vcd_network_direct.net.external_network
}

data "vcd_external_network" "external_network1" {
  name = "${data.vcd_network_direct.net.external_network}"
}

# From the second data source we extract the basic networking info
output "gateway" {
  value = data.vcd_external_network.external_network1.ip_scope.0.gateway
}
output "netmask" {
  value = data.vcd_external_network.external_network1.ip_scope.0.netmask
}
output "DNS" {
  value = data.vcd_external_network.external_network1.ip_scope.0.dns1
}
output "external_ip" {
  value = data.vcd_external_network.external_network1.ip_scope.0.static_ip_pool.0.start_address
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level.
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level.
* `name` - (Required) A unique name for the network

## Attribute Reference

* `external_network` -  The name of the external network.
* `shared` -  Defines if this network is shared between multiple vDCs in the vOrg.
