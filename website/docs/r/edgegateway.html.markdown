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

  name              = "my-egw"
  description       = "new edge gateway"
  backing_config    = "compact"
  default_gateway   = "my-ext-net1"
  external_networks = [ "my-ext-net1", "my-ext-net2" ]
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Required) The name of organization to which the VDC belongs.
* `vdc` - (Required) The name of VDC that owns the edge gateway.
* `name` - (Required) A unique name for the edge gateway.
* `external_networks` - (Required) An array of external network names.
* `backing_config` - (Required) Configuration of the vShield edge VM for this gateway. One of: `compact`, `full`.
* `default_gateway` - (Optional) Name of the external network to be used as default gateway. It must be included in the 
  list of `external_networks`. Providing an empty string or omitting the argument will create the edge gateway without a default gateway.
* `advanced_networking` - (Optional) True if the gateway uses advanced networking. Default is `false`.
* `ha_enabled` - (Optional) Enable high availability on this edge gateway. Default is `false`.
* `distributed_routing` - (Optional) If advanced networking enabled, also enable distributed routing. Default is `false`.


