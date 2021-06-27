---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_ipsec_vpn_tunnel"
sidebar_current: "docs-vcd-data-source-nsxt-ipsec-vpn-tunnel"
description: |-
  Provides a data source to read NSX-T IPsec VPN Tunnel. You can configure site-to-site connectivity between an NSX-T Data
  Center Edge Gateway and remote sites. The remote sites must use NSX-T Data Center, have third-party hardware routers,
  or VPN gateways that support IPSec.
---

# vcd\_nsxt\_ipsec\_vpn\_tunnel

Supported in provider *v3.3+* and VCD 10.1+ with NSX-T backed VDCs.

Provides a data source to read NSX-T IPsec VPN Tunnel. You can configure site-to-site connectivity between an NSX-T Data
Center Edge Gateway and remote sites. The remote sites must use NSX-T Data Center, have third-party hardware routers,
or VPN gateways that support IPSec.

## Example Usage

```hcl
data "vcd_nsxt_ipsec_vpn_tunnel" "tunnel1" {
  org = "my-org"
  vdc = "my-org-vdc"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  name = "tunnel-1"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
  when connected as sysadmin working across different organisations.
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level.
* `edge_gateway_id` - (Required) The ID of the edge gateway (NSX-T only). Can be looked up using `vcd_nsxt_edgegateway`
  data source
* `name` - (Required)  - Name of existing IPsec VPN Tunnel

## Attribute Reference

All the arguments and attributes defined in
[`vcd_nsxt_ipsec_vpn_tunnel`](/docs/providers/vcd/r/nsxt_ipsec_vpn_tunnel.html) resource are available.
