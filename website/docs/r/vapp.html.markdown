---
layout: "vcd"
page_title: "vCloudDirector: vcd_vapp"
sidebar_current: "docs-vcd-resource-vapp"
description: |-
  Provides a vCloud Director vApp resource. This can be used to create, modify, and delete vApps.
---

# vcd\_vapp

Provides a vCloud Director vApp resource. This can be used to create,
modify, and delete vApps.

## Example Usage

Example with more than one VM under a vApp.

```hcl
resource "vcd_network_direct" "net" {
  name             = "net"
  external_network = "corp-network"
}

resource "vcd_vapp" "web" {
  name = "web"

  metadata = {
    CostAccount = "Marketing Department"
  }

  depends_on = ["vcd_network_direct.net"]
}

resource "vcd_vapp_vm" "web1" {
  vapp_name     = "${vcd_vapp.web.name}"
  name          = "web1"
  catalog_name  = "Boxes"
  template_name = "lampstack-1.10.1-ubuntu-10.04"
  memory        = 2048
  cpus          = 1

  network_name = "net"
  ip           = "10.10.104.161"

  properties = {
    "vapp.property1"   = "value1"
    "vapp.property2"   = "value2"
  }

  depends_on = ["vcd_vapp.web"]
}

resource "vcd_vapp_vm" "web2" {
  vapp_name     = "${vcd_vapp.web.name}"
  name          = "web2"
  catalog_name  = "Boxes"
  template_name = "lampstack-1.10.1-ubuntu-10.04"
  memory        = 2048
  cpus          = 1

  network_name = "net"
  ip           = "10.10.104.162"

  depends_on = ["vcd_vapp.web"]
}
```

## Example of vApp with single VM

**Not recommended in v2.0+** : in the earlier version of the provider it was possible to define a vApp with a single VM in one resource, but it is not recommended as of *v2.0+* provider. Please define vApp and VM in separate resources instead.

```hcl
resource "vcd_network_routed" "net" {
  # ...
}

resource "vcd_vapp" "web" {
  name          = "web"
  catalog_name  = "Boxes"
  template_name = "lampstack-1.10.1-ubuntu-10.04"
  memory        = 2048
  cpus          = 1

  network_name = "${vcd_network.net.name}"
  ip           = "10.10.104.160"

  metadata = {
    role    = "web"
    env     = "staging"
    version = "v1"
  }

  ovf {
    hostname = "web"
  }

  depends_on = ["vcd_network_routed.net"]
}
```

## Example of Empty vApp with no VMs

```hcl
resource "vcd_network_routed" "net" {
  # ...
}

resource "vcd_vapp" "web" {
  name = "web"

  metadata = {
    boss = "Why is this vApp empty?"
    john = "I don't really know. Maybe somebody did forget to clean it up."
  }

  depends_on = ["vcd_network_routed.net"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the vApp
* `catalog_name` - (Optional) The catalog name in which to find the given vApp Template
* `template_name` - (Optional) The name of the vApp Template to use
* `memory` - (Optional) The amount of RAM (in MB) to allocate to the vApp
* `cpus` - (Optional) The number of virtual CPUs to allocate to the vApp
* `initscript` (Optional) A script to be run only on initial boot
* `network_name` - (Optional) Name of the network this vApp should join
* `network_href` - (Optional; **Deprecated**) The vCloud Director generated href of the network this vApp
  should join. If empty it will use the network name and query vCloud Director to discover
  this
* `ip` - (Optional) The IP to assign to this vApp. Must be an IP address or
  one of dhcp, allocated or none. If given the address must be within the
  `static_ip_pool` set for the network. If left blank, and the network has
  `dhcp_pool` set with at least one available IP then this will be set with
  DHCP.
* `metadata` - (Optional) Key value map of metadata to assign to this vApp. Key and value can be any string. (Since *v2.2+* metadata is added directly to vApp instead of first VM in vApp)
* `ovf` - (Optional) Key value map of ovf parameters to assign to VM product section
* `power_on` - (Optional) A boolean value stating if this vApp should be powered on. Default is `true`
* `org` - (Optional; *v2.0+*) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `vdc` - (Optional; *v2.0+*) The name of VDC to use, optional if defined at provider level
* `accept_all_eulas` - (Optional; *v2.0+*) Automatically accept EULA if OVA has it. Default is `true`
* `properties` - (Optional; *v2.5+*) Key value map of vApp guest properties **Note** `ovf` attribute
sets guest properties on the first VM using a legacy ability of this resource to spawn 1 VM. Please
use resources `vcd_vapp_vm` to provision VMs.
