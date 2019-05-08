---
layout: "vcd"
page_title: "vCloudDirector: vcd_external_network"
sidebar_current: "docs-vcd-resource-external-network"
description: |-
  Provides a vCloud Director external network resource.  This can be used to create and delete external networks.
---

# vcd\_external\_network

Provides a vCloud Director external network resource.  This can be used to create and delete external networks.
Requires system administrator privileges.

Supported in provider *v2.2+*

## Example Usage

```hcl
provider "vcd" {
  user     = "${var.admin_user}"
  password = "${var.admin_password}"
  org      = "System"
  url      = "https://Vcd/api"
}

resource "vcd_external_network" "net" {
  name        = "my-ext-net"
  description = "Reference for vCD external network"

  ip_scope {
    gateway    = "192.168.30.49"
    netmask    = "255.255.255.240"
    dns1       = "192.168.0.164"
    dns2       = "192.168.0.196"
    dns_suffix = "mybiz.biz"

    static_ip_pool {
      start_address = "192.168.30.51"
      end_address   = "192.168.30.62"
    }
  }

  ip_scope {
    gateway      = "192.168.31.49"
    netmask      = "255.255.255.240"
    dns1         = "192.168.1.164"
    dns2         = "192.168.1.196"
    dns_suffix   = "my.biz"

    static_ip_pool {
      start_address = "192.168.31.51"
      end_address   = "192.168.31.60"
    }

    static_ip_pool {
      start_address = "192.168.31.31"
      end_address   = "192.168.31.40"
    }
  }

  vsphere_networks {
    vcenter         = "vcenter-name"
    vsphere_network = "myNetwork"
    type            = "DV_PORTGROUP"
  }

  retain_net_info_across_deployments = "false"
}

resource "vcd_network_direct" "net" {
  org              = "my-org"
  vdc              = "my-vdc"
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
* `retain_net_info_across_deployments` - (Optional) Specifies whether the network resources such as IP/MAC of router will be retained across deployments. Default is false.

<a id="ipscope"></a>
## IP Scope

* `gateway` - (Required) Gateway of the network
* `netmask` - (Required) Network mask
* `dns1` - (Optional) Primary DNS server
* `dns2` - (Optional) Secondary DNS server
* `dns_suffix` (Optional) A FQDN for the virtual machines on this network.
* `static_ip_pool` - (Required) IP ranges used for static pool allocation in the network.  See [IP Pools](#ip-pools) below for details.

<a id="ip-pools"></a>
## IP Pools

* `start_address` - (Required) Start address of the IP range
* `end_address` - (Required) End address of the IP range

<a id="vspherenetworks"></a>
## vSphere Networks

* `vcenter` - (Required) The vCenter server name
* `vsphere_network` - (Required) Port group name
* `type` - (Required) The vSphere type of the object. One of: DV_PORTGROUP (distributed virtual port group), NETWORK (standard switch port group)
