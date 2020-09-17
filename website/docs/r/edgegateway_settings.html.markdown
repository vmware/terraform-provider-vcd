---
layout: "vcd"
page_title: "vCloudDirector: vcd_edgegateway_settings"
sidebar_current: "docs-vcd-resource-edgegateway-settings"
description: |-
  Provides a vCloud Director edge gateway global settings. This can be used to update global edge gateways settings related to firewall and load balancing.
---

# vcd\_edgegateway\_settings

Provides a resource that can update vCloud Director edge gateway global settings either as system administrator or as
organization user.

The main use case of this resource is to allow both providers and tenants to change edge gateways global settings (such as
enabling load balancing or firewall) when the edge gateway was created outside of terraform.

Supported in provider *v3.0+*

## Example Usage

```hcl
data "vcd_edgegateway" "egw" {
  name = "my-egw"
}

resource "vcd_edgegateway_settings" "egw-settings" {
  edge_gateway_id         = data.vcd_edgegateway.egw.id
  lb_enabled              = true
  lb_acceleration_enabled = true
  lb_logging_enabled      = true
  lb_loglevel             = "debug"

  fw_enabled                      = true
  fw_default_rule_logging_enabled = true
  fw_default_rule_action          = "accept"
}
```

-> **Tip #1:** Although this resource changes values in the edge gateway referenced as a data source, Terraform state
of the edge gateway doesn't get updated. To reconcile the state of the data source with the values modified in vcd_edgegateway_settings,
you need to run `terraform refresh` after `apply`.

-> **Tip #2** Although tenants can enable load balancing using this resource, they can't set the LB log related properties.

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which the VDC belongs. Optional if defined at provider level.
* `vdc` - (Optional) The name of VDC that owns the edge gateway. Optional if defined at provider level. 
* `edge_gateway_name` - (Optional) A unique name for the edge gateway. (Required if `edge_gateway_id` is not set)
* `edge_gateway_id` - (Optional) The edge gateway ID. (Required if `edge_gateway_name` is not set)
* `lb_enabled` - (Optional) Enable load balancing. Default is `false`.
* `lb_acceleration_enabled` - (Optional) Enable to configure the load balancer.
* `lb_logging_enabled` - (Optional) Enables the edge gateway load balancer to collect traffic logs.
Default is `false`. Note: **only system administrators can change this property**. It is ignored for Org users.
* `lb_loglevel` - (Optional) Choose the severity of events to be logged. One of `emergency`,
`alert`, `critical`, `error`, `warning`, `notice`, `info`, `debug`. Note: **only system administrators can change this property**. It is ignored for Org users.
* `fw_enabled` (Optional) Enable firewall. Default `true`.
* `fw_default_rule_logging_enabled` (Optional) Enable default firewall rule (last in the processing 
order) logging. Default `false`.
* `fw_default_rule_action` (Optional) Default firewall rule (last in the processing order) action.
One of `accept` or `deny`. Default `deny`.
