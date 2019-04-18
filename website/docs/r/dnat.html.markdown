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

  edge_gateway  = "Edge Gateway Name"
  network_name  = "Org vDC network name"
  external_ip   = "78.101.10.20"
  port          = -1                    # "-1" == "any"
  internal_ip   = "10.10.0.5"
  protocol      = "ICMP"
  icmp_sub_type = "router-solicitation"
}
```

## Argument Reference

The following arguments are supported:

* `edge_gateway` - (Required) The name of the edge gateway on which to apply the DNAT
* `network_name` - (Required *v2.2+*) The name of the organization network name on which to apply the DNAT
* `external_ip` - (Required) One of the external IPs available on your Edge Gateway
* `port` - (Required) The port number to map. -1 translates to "any"
* `translated_port` - (Optional) The port number to map
* `internal_ip` - (Required) The IP of the VM to map to
* `protocol` - (Optional; *v2.0+*) The protocol type. Possible values are TCP, UDP, TCPUDP, ICMP, ANY. TCP is default to be backward compatible with previous version
* `icmp_sub_type` - (Optional; *v2.0+*) The name of ICMP type. Possible values are   address-mask-request, destination-unreachable, echo-request, echo-reply, parameter-problem, redirect, router-advertisement, router-solicitation, source-quench, time-exceeded, timestamp-request, timestamp-reply, any
* `org` - (Optional; *v2.0+*) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `vdc` - (Optional; *v2.0+*) The name of VDC to use, optional if defined at provider level
