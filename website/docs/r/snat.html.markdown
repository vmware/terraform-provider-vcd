---
layout: "vcd"
page_title: "vCloudDirector: vcd_snat"
sidebar_current: "docs-vcd-resource-snat"
description: |-
  Provides a vCloud Director SNAT resource. This can be used to create, modify, and delete source NATs to allow vApps to send external traffic.
---

# vcd\_snat

Provides a vCloud Director SNAT resource. This can be used to create, modify,
and delete source NATs to allow vApps to send external traffic.

## Example Usage

```hcl
resource "vcd_snat" "outbound" {
  edge_gateway = "Edge Gateway Name"
  external_ip  = "78.101.10.20"
  internal_ip  = "10.10.0.0/24"
}
```

## Argument Reference

The following arguments are supported:

* `edge_gateway` - (Required) The name of the edge gateway on which to apply the SNAT
* `network_name` - (Optional; *v2.2+*) The name of the network on which to apply the SNAT. *`network_name` will be required field in the next major version.*
* `network_type` - (Optional; *v2.2+*) Type of the network on which to apply the NAT rule. Default: "org". *`network_type` will be required field in the next major version.*
* `external_ip` - (Required) One of the external IPs available on your Edge Gateway
* `internal_ip` - (Required) The IP or IP Range of the VM(s) to map from
* `org` - (Optional; *v2.0+*) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `vdc` - (Optional; *v2.0+*) The name of VDC to use, optional if defined at provider level
