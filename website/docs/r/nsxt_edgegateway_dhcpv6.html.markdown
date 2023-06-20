---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_edgegateway_dhcpv6"
sidebar_current: "docs-vcd-resource-nsxt-edge-dhcpv6"
description: |-
  Provides a resource to manage DHCPv6 configuration for NSX-T Edge Gateways.
---

# vcd\_nsxt\_edgegateway\_dhcpv6

Provides a resource to manage DHCPv6 configuration for NSX-T Edge Gateways.

## Example Usage (DHCPv6 mode)

```hcl
esource "vcd_nsxt_edgegateway_dhcpv6" "dhcpv6-mode" {
  org             = "cloud"
  edge_gateway_id = data.vcd_nsxt_edgegateway.testing-in-vdc.id

  enabled = true
  mode    = "DHCPv6"
}
```

## Example Usage (SLAAC mode)

```hcl
resource "vcd_nsxt_edgegateway_dhcpv6" "slaac-mode" {
  org             = "datacloud"
  edge_gateway_id = data.vcd_nsxt_edgegateway.testing-in-vdc.id

  enabled      = true
  mode         = "SLAAC"
  domain_names = ["non-existing.org.tld","fake.org.tld"]
  dns_servers  = ["2001:4860:4860::8888","2001:4860:4860::8844"]
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Required) Org in which the NSX-T Edge Gateway is located
* `edge_gateway_id` - (Required) NSX-T Edge Gateway ID
* `enabled` - (Required) Boolean flag if DHCPv6 is enabled or disabled. **Note**
* `mode` - (Required) One of `SLAAC` (Stateless Address Autoconfiguration) or `DHCPv6` (Dynamic Host
  Configuration Protocol)
* `domain_names` - (Optional) Set of domain names (only applicable for `DHCPv6` mode)
* `dns_servers` - (Optional) Set of IPv6 DNS servers (only applicable for `DHCPv6` mode)

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing NSX-T Edge Gateway DHCPv6 configuration can be [imported][docs-import] into this
resource via supplying path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxt_edgegateway_dhcpv6.imported my-org.nsxt-vdc.nsxt-edge
```

The above would import the `nsxt-edge` Edge Gateway DHCPv6 configuration for this particular
Edge Gateway.
