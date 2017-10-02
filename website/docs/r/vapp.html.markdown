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
  networks      = [
    {
      "orgnetwork" = "${vcd_network.net.name}"
      "is_primary" = true
    }
  ]
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
* `networks` - (Optional) List of network adapter definitions
  - `orgnetwork` (Required) name of organization network to use
  - `ip` (Optional) see below
  - `is_primary` (Optional) A boolean value which ensures that the network adapter is
  primary
* `ip` - (Optional) The IP to assign to this vApp. Must be an IP address or
  one of dhcp, allocated or none. If given the address must be within the
  `static_ip_pool` set for the network. If left blank, and the network has
  `dhcp_pool` set with at least one available IP then this will be set with
  DHCP.
* `metadata` - (Optional) Key value map of metadata to assign to this vApp
* `ovf` - (Optional) Key value map of ovf parameters to assign to VM product section
* `power_on` - (Optional) A boolean value stating if this vApp should be powered on. Default to `true`
