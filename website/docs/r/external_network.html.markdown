---
layout: "vcd"
page_title: "vCloudDirector: vcd_external_network"
sidebar_current: "docs-vcd-resource-external-network"
description: |-
  Provides a vCloud Director external network resource.  This can be used to create and delete external networks.
---

# vcd\_external\_network

Provides a vCloud Director external network resource.  This can be used to create and delete external networks.

Supported in provider *v2.2+*

## Example Usage

```hcl
resource "vcd_external_network" "net" {
  name        = "my-ext-net"
  description = "Reference for VDC direct network"

  ip_scope {
    is_inherited = "false"
    gateway      = "192.168.30.49"
    netmask      = "255.255.255.240"
    dns1         = "192.168.0.164"
    dns2         = "192.168.0.196"
    dns_suffix   = "mybiz.biz"

    static_ip_pool {
      start_address = "192.168.30.51"
      end_address   = "192.168.30.62"
    }
  }

  vsphere_networks {
    vcenter         = "vC1"
    vsphere_network = "myNetwork"
    type            = "DV_PORTGROUP"
  }

  fence_mode                         = "isolated"
  retain_net_info_across_deployments = "false"
}

resource "vcd_network_direct" "net" {
  name             = "my-net"
  external_network = "${vcd_external_network.net.name}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the network
* `description` - (Optional) Network friendly description
* `ip_scope` - (Required) A list of IP scopes for the network.  See [IP Scope](#ipscope) below for details.
* `vsphere_networks` - (Required) A list of DV_PORTGROUP or NETWORK objects names that back this network. Each referenced DV_PORTGROUP or NETWORK must exist on a vCenter server registered with the system.  See [vSphere Networks](#vspherenetworks) below for details.
* `fence_mode` - (Optional) Isolation type of the network. If ParentNetwork is specified, this property controls connectivity to the parent. One of: `bridged` (connected directly to the ParentNetwork), `isolated` (not connected to any other network), `natRouted` (connected to the ParentNetwork via a NAT service) 
* `retain_net_info_across_deployments` - (Optional) Specifies whether the network resources such as IP/MAC of router will be retained across deployments. Default is false.
* `parent_network` - (Optional) A name of of parent network


<a id="ipscope"></a>
## IP Scope

* `is_inherited` - (Optional) True if the IP scope is inherit from parent network. Default is false.
* `gateway` - (Required) Gateway of the network
* `netmask` - (Required) Network mask
* `dns1` - (Required) Primary DNS server
* `dns2` - (Required) Secondary DNS server
* `dns_suffix` (Optional)
* `static_ip_pool` - (Required) IP ranges used for static pool allocation in the network.  See [IP Pools](#ip-pools) below for details.

<a id="ip-pools"></a>
## IP Pools

* `start_address` - (Required) Start address of the IP range
* `end_address` - (Required) End address of the IP range

<a id="vspherenetworks"></a>
## vShere Networks

* `vcenter` - (Required) The vCenter server reference
* `vsphere_network` - (Required) Managed object reference of the object
* `type` - (Required) The vSphere type of the object.  One of: DV_PORTGROUP (distributed virtual portgroup), NETWORK
