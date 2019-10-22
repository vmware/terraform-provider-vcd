---
layout: "vcd"
page_title: "vCloudDirector: vcd_edgegateway"
sidebar_current: "docs-vcd-data-source-edgegateway"
description: |-
  Provides an edge gateway data source.
---

# vcd\_edgegateway

Provides a vCloud Director edge gateway data source, directly connected to one or more external networks. This can be used to reference
edge gateways for Org VDC networks to connect.

Supported in provider *v2.5+*

## Example Usages

```hcl
data "vcd_edgegateway" "mygw" {
    name = "mygw"
    org  = "myorg"
    vdc  = "myvdc"
}
locals {
    network_index = index(data.vcd_edgegateway.mygw.external_networks, "My External Network Name")
}
output "ip_address" {
  value = element(data.vcd_edgegateway.mygw.external_networks_ip, local.network_index)
  depends_on = [data.vcd_edgegateway.mygw]
}
output "netmask" {
  value = element(data.vcd_edgegateway.mygw.external_networks_netmask, local.network_index)
  depends_on = [data.vcd_edgegateway.mygw]
}
output "gateway" {
  value = element(data.vcd_edgegateway.mygw.external_networks_gateway, local.network_index)
  depends_on = [data.vcd_edgegateway.mygw]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the edge gateway.
* `org` - (Optional) The name of organization to which the VDC belongs. Optional if defined at provider level.
* `vdc` - (Optional) The name of VDC that owns the edge gateway. Optional if defined at provider level. 

## Attribute Reference

All attributes defined in [edge gateway resource](/docs/providers/vcd/r/edgegateway.html#attribute-reference) are supported.