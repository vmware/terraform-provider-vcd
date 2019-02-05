---
layout: "vcd"
page_title: "vCloudDirector: vcd_vapp_network"
sidebar_current: "docs-vcd-resource-vapp-network"
description: |-
  Provides a vCloud Director vApp isolated Network. This can be used to create and delete internal networks for vApps to connect.
---

# vcd\_vapp\_network

 Provides a vCloud Director vApp isolated Network. This can be used to create and delete internal networks for vApps to connect.
 This network is not attached to external networks or routers.

Supported in provider *v2.0+*

## Example Usage

```hcl
resource "vcd_vapp_network" "vappNet" {
  org = "my-org" #Optional
  vdc = "my-vdc" #Optional

  name               = "my-net"
  gateway            = "192.168.2.1"
  netmask            = "255.255.255.0"
  dns1               = "192.168.2.1"
  dns2               = "192.168.2.2"
  dns_suffix         = "mybiz.biz"
  guest_vlan_allowed = true

  static_ip_pool {
    start_address = "192.168.2.51"
    end_address   = "192.168.2.100"
  }
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional; *v2.0+*) The name of organization to use, optional if defined at provider level. Useful when 
  connected as sysadmin working across different organisations
* `vdc` - (Optional; *v2.0+*) The name of VDC to use, optional if defined at provider level
* `name` - (Required) A unique name for the network
* `netmask` - (Optional) The netmask for the new network. Defaults to `255.255.255.0`
* `gateway` (Optional) The gateway for this network
* `dns1` - (Optional) First DNS server to use. Default is `8.8.8.8`
* `dns2` - (Optional) Second DNS server to use. Default is `8.8.4.4`
* `dns_suffix` - (Optional) A FQDN for the virtual machines on this network
* `guest_vlan_allowed` (Optional) True if Network allows guest VLAN tagging. Default is false. This value supported from vCD version 9.0
* `static_ip_pool` - (Optional) A range of IPs permitted to be used as static IPs for
  virtual machines; see [IP Pools](#ip-pools) below for details.

<a id="ip-pools"></a>
## IP Pools

Static IP Pools support the following attributes:

* `start_address` - (Required) The first address in the IP Range
* `end_address` - (Required) The final address in the IP Range