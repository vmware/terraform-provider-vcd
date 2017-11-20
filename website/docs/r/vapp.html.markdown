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

```hcl
resource "vcd_network" "net" {
  # ...
}

resource "vcd_vapp" "web" {
  name          = "web"
  catalog_name  = "Boxes"
  template_name = "lampstack-1.10.1-ubuntu-10.04"
  memory        = 2048
  cpus          = 1

  network_name = "${vcd_network.net.name}"
  network_href = "${vcd_network.net.href}"
  ip           = "10.10.104.160"

  metadata {
    role    = "web"
    env     = "staging"
    version = "v1"
  }

  ovf {
    hostname = "web"
  }
}
```

## Example RAW vApp with No VMS

```hcl
resource "vcd_network" "net" {
  # ...
}

resource "vcd_vapp" "web" {
  name          = "web"
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
* `network_href` - (Deprecated) The vCloud Director generated href of the network this vApp
  should join. If empty it will use the network name and query vCloud Director to discover
  this
* `ip_allocation_mode` - (Optional) Instruct how the virtual machine must acquire an IP address if not a static one is provided. Can be either `allocated` for vCloud to provide an IP from the `static_ip_pool` or `dhcp` for DHCP acquistition.
* `ip` - (Optional) The IP to assign to this vApp. Must be an IP address or blank. If given the address must be within the
  `static_ip_pool` set for the network. If left blank, `ip_allocation_mode` should be set, or DHCP is used.
* `metadata` - (Optional) Key value map of metadata to assign to this vApp
* `ovf` - (Optional) Key value map of ovf parameters to assign to VM product section
* `power_on` - (Optional) A boolean value stating if this vApp should be powered on. Default to `true`
