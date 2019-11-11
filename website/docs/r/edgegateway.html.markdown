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

~> **Note:** Load balancing capabilities will work only when edge gateway is `advanced`. Load
balancing settings will be **ignored** when it is not. Refer to [official vCloud Director documentation]
(https://docs.vmware.com/en/vCloud-Director/9.7/com.vmware.vcloud.tenantportal.doc/GUID-7E082E77-B459-4CE7-806D-2769F7CB5624.html) 
for more information.

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
* `fw_enabled` (Optional) Enable firewall. Default `true`. **Note:** Disabling Firewall will also
disable NAT and other NAT dependent features like Load Balancer.
* `fw_default_rule_logging_enabled` (Optional) Enable default firewall rule (last in the processing 
order) logging. Default `false`.
* `fw_default_rule_action` (Optional) Default firewall rule (last in the processing order) action.
One of `accept` or `deny`. Default `deny`.

## Attribute Reference

The following attributes are exported on this resource:

* `default_network_ip` (*v2.6+*) - IP address of edge gateway used for default network

## Importing

Supported in provider *v2.5+*

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing edge gateway can be [imported][docs-import] into this resource via supplying its path. 
The path for this resource is made of org-name.vdc-name.edge-name
For example, using this structure, representing an edge gateway that was **not** created using Terraform:

```hcl
resource "vcd_edgegateway" "tf-edgegateway" {
  name              = "my-edge-gw"
  org               = "my-org"
  vdc               = "my-vdc"
  configuration     = "COMPUTE"
  external_networks = ["COMPUTE"]
}
```

You can import such edge gateway into terraform state using this command

```
terraform import vcd_edgegateway.tf-edgegateway my-org.my-vdc.my-edge-gw
```
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After importing, if you run `terraform plan` you will see the rest of the values and modify the script accordingly for 
further operations.
