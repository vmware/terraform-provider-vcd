---
layout: "vcd"
page_title: "vCloudDirector: vcd_vapp_vm"
sidebar_current: "docs-vcd-resource-vapp-vm"
description: |-
  Provides a vCloud Director VM resource. This can be used to create, modify, and delete VMs within a vApp.
---

# vcd\_vapp\_vm

Provides a vCloud Director VM resource. This can be used to create,
modify, and delete VMs within a vApp.

~> **Note:** There is known bug with this implementation, that to use the vcd_vapp_vm resource, you must set the paralellism parameter to 1. [We are working on this.](https://github.com/terraform-providers/terraform-provider-vcd/issues/27)


## Example Usage

```hcl
resource "vcd_network" "net" {
  # ...
}

resource "vcd_vapp" "web" {
  name          = "web"
}

resource "vcd_vapp_vm" "web2" {
  vapp_name     = "${vcd_vapp.web.name}"
  name          = "web2"
  catalog_name  = "Boxes"
  template_name = "lampstack-1.10.1-ubuntu-10.04"
  memory        = 2048
  cpus          = 1

  ip           = "10.10.104.161"
}

resource "vcd_vapp_vm" "web3" {
  vapp_name     = "${vcd_vapp.web.name}"
  name          = "web3"
  catalog_name  = "Boxes"
  template_name = "lampstack-1.10.1-ubuntu-10.04"
  memory        = 2048
  cpus          = 1

  ip           = "10.10.104.162"
}
```

## Argument Reference

The following arguments are supported:

* `vapp_name` - (Required) The vApp this VM should belong to.
* `name` - (Required) A unique name for the vApp
* `catalog_name` - (Required) The catalog name in which to find the given vApp Template
* `template_name` - (Required) The name of the vApp Template to use
* `memory` - (Optional) The amount of RAM (in MB) to allocate to the vApp
* `cpus` - (Optional) The number of virtual CPUs to allocate to the vApp
* `initscript` (Optional) A script to be run only on initial boot
* `ip_allocation_mode` - (Optional) Instruct how the virtual machine must acquire an IP address if not a static one is provided. Can be either `allocated` for vCloud to provide an IP from the `static_ip_pool` or `dhcp` for DHCP acquistition.
* `ip` - (Optional) The IP to assign to this vApp. Must be an IP address or blank. If given the address must be within the
  `static_ip_pool` set for the network. If left blank, `ip_allocation_mode` should be set, or DHCP is used.
* `power_on` - (Optional) A boolean value stating if this vApp should be powered on. Default to `true`
