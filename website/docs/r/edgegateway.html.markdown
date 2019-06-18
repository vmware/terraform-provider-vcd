---
layout: "vcd"
page_title: "vCloudDirector: vcd_edgegateway"
sidebar_current: "docs-vcd-resource-edgegateway"
description: |-
  Provides a vCloud Director edge gateway. This can be used to create and delete edge gateways connected to one or more external networks.
---

# vcd\_edgegateway

Provides a vCloud Director edge gateway directly connected to one or more external networks. This can be used to create
and delete edge gateways for Org VDC networks to connect.

Supported in provider *v2.4+*

~> **Note:** Only `System Administrator` can create an edge gateway.
You must use `System Adminstrator` account in `provider` configuration
and then provide `org` and `vdc` arguments for edge gateway to work.

## Example Usage

```hcl
resource "vcd_edgegateway" "egw" {
  org = "my-org"
  vdc = "my-vdc"

  name                    = "my-egw"
  description             = "new edge gateway"
  configuration           = "compact"
  default_gateway_network = "my-ext-net1"
  external_networks       = [ "my-ext-net1", "my-ext-net2" ]
  advanced                = true
}

resource "vcd_network_routed" "rnet1" {
  name         = "rnet1"
  org          = "my-org"
  vdc          = "my-vdc"
  edge_gateway = "${vcd_edgegateway.egw.name}"
  gateway      = "192.168.2.1"

  static_ip_pool {
    start_address = "192.168.2.2"
    end_address   = "192.168.2.100"
  }
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which the VDC belongs. Optional if defined at provider level.
* `vdc` - (Optional) The name of VDC that owns the edge gateway. Optional if defined at provider level. 
* `name` - (Required) A unique name for the edge gateway.
* `external_networks` - (Required) An array of external network names.
* `configuration` - (Required) Configuration of the vShield edge VM for this gateway. One of: `compact`, `full` ("Large"), `x-large`, `full4` ("Quad Large").
* `default_gateway_network` - (Optional) Name of the external network to be used as default gateway. It must be included in the
  list of `external_networks`. Providing an empty string or omitting the argument will create the edge gateway without a default gateway.
* `advanced` - (Optional) True if the gateway uses advanced networking. Default is `true`.
* `ha_enabled` - (Optional) Enable high availability on this edge gateway. Default is `false`.
* `distributed_routing` - (Optional) If advanced networking enabled, also enable distributed routing. Default is `false`.


