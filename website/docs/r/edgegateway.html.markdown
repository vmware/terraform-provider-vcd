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

~> **Note:** To enable load balancing capabilities the edge gateway must be `advanced`. Refer to
[official vCloud Director documentation](https://docs.vmware.com/en/vCloud-Director/9.7/com.vmware.vcloud.tenantportal.doc/GUID-7E082E77-B459-4CE7-806D-2769F7CB5624.html) for more information.

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
* `lb_enabled` - (Optional) Enable load balancing. Default is `false`.
* `lb_acceleration_enabled` - (Optional) Enable to configure the load balancer to use the faster L4
engine rather than L7 engine. The L4 TCP VIP is processed before the edge gateway firewall so no 
`allow` firewall rule is required. Default is `false`. **Note:** L7 VIPs for HTTP and HTTPS are
processed after the firewall, so when Acceleration Enabled is not selected, an edge gateway firewall
rule must exist to allow access to the L7 VIP for those protocols. When Acceleration Enabled is
selected and the server pool is in non-transparent mode, an SNAT rule is added, so you must ensure
that the firewall is enabled on the edge gateway.
* `lb_logging_enabled` - (Optional) Enables the edge gateway load balancer to collect traffic logs.
Default is `false`.
* `lb_loglevel` - (Optional) Choose the severity of events to be logged. One of `emergency`,
`alert`, `critical`, `error`, `warning`, `notice`, `info`, `debug`


