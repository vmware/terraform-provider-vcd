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

    ip_range {
      start = "192.168.30.51"
      end   = "192.168.30.62"
    }
  }

  vim_port_group {
    vim_server      = "<uuid>"
    mo_ref          = "dvportgroup-0000"
    vim_object_type = "DV_PORTGROUP"
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
* `vim_port_group` - (Required) A list of DV_PORTGROUP or NETWORK objects that back this network. Each referenced DV_PORTGROUP or NETWORK must exist on a vCenter server registered with the system.  See [Vim Port Group](#vimportgroup) below for details.
* `fence_mode` - (Optional) Isolation type of the network. If ParentNetwork is specified, this property controls connectivity to the parent. One of: bridged (connected directly to the ParentNetwork), isolated (not connected to any other network), natRouted (connected to the ParentNetwork via a NAT service) 
* `retain_net_info_across_deployments` - (Optional)  Specifies whether the network resources such as IP/MAC of router will be retained across deployments. Default is false.
* `parent_network` - (Optional) Contains reference to parent network


<a id="ipscope"></a>
## IP Scope

* `is_inherited` - (Required) True if the IP scope is inherit from parent network
* `gateway` - (Required) Gateway of the network
* `netmask` - (Required) Network mask
* `dns1` - (Required) Primary DNS server
* `dns2` - (Required) Secondary DNS server
* `ip_range` - (Required) IP ranges used for static pool allocation in the network.  See [IP Range](#iprange) below for details.

<a id="iprange"></a>
## IP Range

* `start` - (Required) Start address of the IP range
* `end` - (Required) End address of the IP range

<a id="vimporrtgroup"></a>
## Vim Port Group

* `vim_server` - (Required) The vCenter server reference
* `mo_ref` - (Required) Managed object reference of the object
* `vim_object_type` - (Required) The vSphere type of the object.  One of: DV_PORTGROUP (distributed virtual portgroup), NETWORK
