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

~> **Note:** To make sure resources are created in the right order and both plan apply and destroy succeeds, use the `depends_on` clause (see example below)


## Example Usage

```hcl
resource "vcd_network_direct" "net" {
  name             = "net"
  external_network = "corp-network"
}

resource "vcd_vapp" "web" {
  name = "web"

  depends_on = ["vcd_network_direct.net"]
}

resource "vcd_vapp_vm" "web1" {
  vapp_name     = "${vcd_vapp.web.name}"
  name          = "web1"
  catalog_name  = "Boxes"
  template_name = "lampstack-1.10.1-ubuntu-10.04"
  memory        = 2048
  cpus          = 2
  cpu_cores     = 1

  metadata {
    role    = "web"
    env     = "staging"
    version = "v1"
    bla     = "foo"
  }

  networks = [{
    network_type       = "org"
    network_name       = "net"
    ip                 = "10.10.104.161"
    ip_allocation_mode = "MANUAL"
    is_primary         = true
  }]

  depends_on = ["vcd_vapp.web"]
}

resource "vcd_vapp_vm" "web2" {
  vapp_name     = "${vcd_vapp.web.name}"
  name          = "web2"
  catalog_name  = "Boxes"
  template_name = "lampstack-1.10.1-ubuntu-10.04"
  memory        = 2048
  cpus          = 1

  metadata {
    role         = "web"
    env          = "staging"
    version      = "v2"
    my_extra_key = "My extra value"
  }

  networks = [{
    network_type       = "org"
    network_name       = "net"
    ip                 = "10.10.104.162"
    ip_allocation_mode = "MANUAL"
    is_primary         = true
  },
  {
    network_type       = "vapp"
    ip_allocation_mode = "POOL"
  },
  {
    network_type       = "none"
    ip_allocation_mode = "NONE"
  }]

  disk {
    name = "logDisk1"
    bus_number = 1
    unit_number = 0
  }

  disk {
    name = "logDisk2"
    bus_number = 1
    unit_number = 1
  }

  depends_on = ["vcd_vapp.web"]
}
```

## Argument Reference

The following arguments are supported:

* `vapp_name` - (Required) The vApp this VM should belong to.
* `name` - (Required) A unique name for the VM
* `catalog_name` - (Required) The catalog name in which to find the given vApp Template
* `template_name` - (Required) The name of the vApp Template to use
* `memory` - (Optional) The amount of RAM (in MB) to allocate to the VM
* `cpus` - (Optional) The number of virtual CPUs to allocate to the VM. Socket count is a result of: virtual logical processors/cores per socket
* `cpu_cores` - (Optional; *v2.1+*) The number of cores per socket
* `metadata` - (Optional; *v2.2+*) Key value map of metadata to assign to this VM
* `initscript` (Optional) A script to be run only on initial boot
* `network_name` - (**Deprecated**) Name of the network this VM should connect to. **Conflicts with** with networks.
* `vapp_network_name` - (**Deprecated**; *v2.1+*) Name of the vApp network this VM should connect to
* `ip` - (**Deprecated**) The IP to assign to this vApp. Must be an IP address or
one of dhcp, allocated or none. If given the address must be within the
  `static_ip_pool` set for the network. If left blank, and the network has
  `dhcp_pool` set with at least one available IP then this will be set with
DHCP. **Conflicts with** with networks.
* `networks` - (Optional; *v2.2+*) List of network interfaces to attach to vm. **Conflicts with**: `networks`, `ip`, `vapp_network_name`. **Note**: all params of this parameter and itself do force recreation of vms!
  * `network_type` (Required) Network type, one of: `none`, `vapp` or `vdc`. `none` creates a NIC with no network attached. `vapp` attaches a vApp network, while `vdc` attaches organization VDC network.
  * `network_name` (Optional) Name of the network this VM should connect to. Always required except for `network_type` `NONE`.
  * `ip` (Computed, Optional) Empty or valid IP address if `ip_allocation_mode` is `MANUAL`. IP address will be set in case of `ip_allocation_mode` is `POOL` and may be set when it is `DHCP`.
  * `is_primary` (Optional) Set to true if network interface should be primary. First network card in the list will be primary by default.
  * `mac` - (Computed) Mac address of network interface.
  * `ip_allocation_mode` (Required) IP address allocation mode. One of `POOL`, `DHCP`, `MANUAL`, `NONE`:  
    
      **`POOL`** - Static IP address is allocated automatically from defined static pool in network.
      
      **`DHCP`** - IP address is obtained from a DHCP service. Ip field is not guaranteed to be populated. Because of this may appear after
      after multiple `terraform refresh` operations.
      
      **`MANUAL`** - IP address is assigned manually in the `ip` field. Must be valid IP address from static pool.
      
      **`NONE`** - No IP address will be set because VM will have a NIC without network.
* `power_on` - (Optional) A boolean value stating if this vApp should be powered on. Default is `true`
* `accept_all_eulas` - (Optional; *v2.0+*) Automatically accept EULA if OVA has it. Default is `true`
* `org` - (Optional; *v2.0+*) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `vdc` - (Optional; *v2.0+*) The name of VDC to use, optional if defined at provider level
* `disk` - (Optional; *v2.1+*) Independent disk attachment configuration. Details below
* `expose_hardware_virtualization` - (Optional; *v2.2+*) Boolean for exposing full CPU virtualization to the
guest operating system so that applications that require hardware virtualization can run on virtual machines without binary
translation or paravirtualization. Useful for hypervisor nesting provided underlying hardware supports it. Default is `false`.

Independent disk support the following attributes:

* `name` - (Required) Independent disk name
* `bus_number` - (Required) Bus number on which to place the disk controller
* `unit_number` - (Required) Unit number (slot) on the bus specified by BusNumber.
