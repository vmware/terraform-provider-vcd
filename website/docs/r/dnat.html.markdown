---
layout: "vcd"
page_title: "vCloudDirector: vcd_dnat"
sidebar_current: "docs-vcd-resource-dnat"
description: |-
  Provides a vCloud Director DNAT resource. This can be used to create, modify, and delete destination NATs to map external IPs to a VM.
---

# vcd\_dnat

Provides a vCloud Director DNAT resource. This can be used to create, modify,
and delete destination NATs to map an external IP/port to an internal IP/port.

~> **Note:** DEPRECATED: This resource may corrupt UI edited NAT rules when used with advanced
edge gateways. Please use [`vcd_nsxv_dnat`](/docs/providers/vcd/r/nsxv_dnat.html) in that case.

~> **Note:** From v2.4+ `protocol` requires lower case values. This may result in invalid configuration if upper case was used previously.

!> **Warning:** When advanced edge gateway is used and the rule is updated using UI, then ID mapping will be lost and Terraform won't find the rule anymore and remove it from state.  

## Example Usage

```hcl
resource "vcd_dnat" "web" {
  org = "my-org" # Optional
  vdc = "my-vdc" # Optional

  edge_gateway    = "Edge Gateway Name"
  external_ip     = "78.101.10.20"
  port            = 80
  internal_ip     = "10.10.0.5"
  translated_port = 8080
}

resource "vcd_dnat" "forIcmp" {
  org = "my-org" # Optional
  vdc = "my-vdc" # Optional
  
  network_name = "my-external-network"
  network_type = "ext"

  edge_gateway  = "Edge Gateway Name"
  external_ip   = "78.101.10.20"
  port          = -1                    # "-1" == "any"
  internal_ip   = "10.10.0.5"
  protocol      = "icmp"
  icmp_sub_type = "router-solicitation"
}
```

## Argument Reference

The following arguments are supported:

* `edge_gateway` - (Required) The name of the edge gateway on which to apply the DNAT
* `external_ip` - (Required) One of the external IPs available on your Edge Gateway
* `port` - (Required) The port number to map. -1 translates to "any"
* `translated_port` - (Optional) The port number to map
* `internal_ip` - (Required) The IP of the VM to map to
* `protocol` - (Optional; *v2.0+*) The protocol type. Possible values are `tcp`, `udp`, `tcpupd`, `icmp`, `any`. `tcp` is default to be backward compatible with previous version
* `icmp_sub_type` - (Optional; *v2.0+*) The name of ICMP type. Possible values are   address-mask-request, destination-unreachable, echo-request, echo-reply, parameter-problem, redirect, router-advertisement, router-solicitation, source-quench, time-exceeded, timestamp-request, timestamp-reply, any
* `network_type` - (Optional; *v2.4+*) Type of the network on which to apply the NAT rule. Possible values `org` or `ext`. `ext` requires system administrator privileges. *`network_type` will be a required field in the next major version.*
* `network_name` - (Optional; *v2.4+*) The name of the network on which to apply the SNAT. *`network_name` will be a required field in the next major version.*
* `org` - (Optional; *v2.0+*) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `vdc` - (Optional; *v2.0+*) The name of VDC to use, optional if defined at provider level
* `description` - (Optional; *v2.4+*) - Description of item