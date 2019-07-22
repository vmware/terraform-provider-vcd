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

!> **Warning:** When advanced edge gateway is used and the rule is updated using UI, then Id mapping will be lost and terraform won't find the rule anymore and remove it from state.

## Example Usage

```hcl
resource "vcd_snat" "outbound" {
  edge_gateway = "Edge Gateway Name"
  network_name = "my-org-vdc-network"
  network_type = "org"
  external_ip  = "78.101.10.20"
  internal_ip  = "10.10.0.0/24"
}
```

## Argument Reference

The following arguments are supported:

* `edge_gateway` - (Required) The name of the edge gateway on which to apply the SNAT
* `external_ip` - (Required) One of the external IPs available on your Edge Gateway
* `internal_ip` - (Required) The IP or IP Range of the VM(s) to map from
* `network_type` - (Optional; *v2.4+*) Type of the network on which to apply the NAT rule. Possible values org or ext. *`network_type` will be a required field in the next major version.*
* `network_name` - (Optional; *v2.4+*) The name of the network on which to apply the SNAT. *`network_name` will be a required field in the next major version.*
* `org` - (Optional; *v2.0+*) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `vdc` - (Optional; *v2.0+*) The name of VDC to use, optional if defined at provider level