---
layout: "vcd"
page_title: "vCloudDirector: vcd_vm"
sidebar_current: "docs-vcd-resource-vm"
description: |-
  Provides a vCloud Director VM resource. This can be used to create, modify, and delete VMs within a vApp.
---

# vcd\_vapp\_vm

Provides a vCloud Director VM resource. This can be used to create,
modify, and delete VMs within a vApp.

Please note that a vApp has to be created and passed to a VM before creation.


## Example Usage

```hcl
resource "vcd_vapp" "test-vapp" {
  name     = "test-vapp"

  organization_network = [
    "service-network",
  ]

  vapp_network {
     name = "test"
     description = ""
     gateway = "192.168.2.1"
     netmask = "255.255.255.0"
     dns1 = "8.8.8.8"
     dns2 = "8.8.4.4"
     start = "192.168.2.100"
     end = "192.168.2.199"
     nat = false
     parent = "service-network"
     dhcp = false
  }
}

resource "vcd_vm" "test-vm"    {
  name          = "test"
  vapp_href     = "${vcd_vapp.test-vapp.id}"
  catalog_name  = "BETA_PUBLIC_IT_DEPARTMENT"
  template_name = "Ubuntu_Server_16.04"
  memory        = 512
  cpus          = 1
  power_on      = false
  storage_profile = "Silver"
  nested_hypervisor_enabled = false

  network = {
    name               = "service-network"
    ip_allocation_mode = "DHCP"
    adapter_type       = "VMXNET3"
  }
  network = {
    name               = "test"
    ip_allocation_mode = "POOL"
    adapter_type       = "E1000"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the vApp
* `vapp_href` - (Required) The vApp this VM must belong to. It is important to use the reference as shown in the example to make sure the vApp is created before the VM.
* `description` - (Optional) Description of VM.
* `catalog_name` - (Required) The catalog name in which to find the given vApp Template
* `template_name` - (Required) The name of the vApp Template to use
* `memory` - (Optional) The amount of RAM (in MB) to allocate to the vApp
* `cpus` - (Optional) The number of virtual CPUs to allocate to the vApp
* `initscript` (Optional) A script to be run only on initial boot
* `power_on` - (Optional) A boolean value stating if this vApp should be powered on. Default to `true`
* `network` - (Optional) List of networks (and nics) to attach to the VM.
* `nested_hypervisor_enabled` - (Optional) Exposes CPU virtualization to the VM.
* `storage_profile` - (Optional) Set the storage profile for the VMs storage.
* `admin_password_auto` - (Optional) Bool to automatically set the admin password of the VM.
* `admin_password` - (Optional) Set the admin password for the VM. Requires `admin_password_auto` to be `false`.

`network` supports the following arguments:

* `name` - (Required) Name of the network to attach the network/nic to.
* `ip` - (Optional) Set a static IP for the virtual machine. Must be within the network pool available and requires `ip_allocation_mode` to be `MANUAL`.
* `ip_allocation_mode` - (Required) Defines how the VM acquires an IP address. Available modes:
    - `MANUAL` - Do not acquire IP
    - `DHCP` - Acquire IP by DHCP
    - `POOL` - Let VCD set a static IP for the VM

* `is_primary` - (Optional) Set primary network
* `is_connected` - (Optional) Set connection status of network
* `adapter_type` - (Optional) Set the adapter type for the nic. Available:
    - `VMXNET3`
    - `E1000`
    - `E1000E`
    
