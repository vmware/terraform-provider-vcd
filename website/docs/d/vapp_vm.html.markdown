---
layout: "vcd"
page_title: "vCloudDirector: vcd_vapp_vm"
sidebar_current: "docs-vcd-datasource-vapp-vm"
description: |-
  Provides a vCloud Director VM data source. This can be used to access VMs within a vApp.
---

# vcd\_vapp\_vm

Provides a vCloud Director VM data source. This can be used to access VMs within a vApp.

Supported in provider *v2.6+*

## Example Usage

```hcl

data "vcd_vapp" "web" {
  name= "web"
}

data "vcd_vapp_vm" "web1" {
  vapp_name     = "${vcd_vapp.web.name}"
  name          = "web1"
}

output "vm" {
  value = data.vcd_vapp_vm.web1
}
```

Sample output:
```
vm = {
  "computer_name" = "TestVM"
  "cpu_cores" = 1
  "cpus" = 2
  "description" = "This OVA provides a minimal installed profile of PhotonOS. Default password for root user is changeme"
  "disk" = []
  "guest_properties" = {}
  "href" = "https://my-vcd.org/api/vApp/vm-ecb449a2-0b11-494d-bbc7-6ae2f2ff9b82"
  "id" = "urn:vcloud:vm:ecb449a2-0b11-494d-bbc7-6ae2f2ff9b82"
  "memory" = 1024
  "metadata" = {
    "vm_metadata" = "VM Metadata."
  }
  "name" = "vm-datacloud"
  "network" = [
    {
      "ip" = "192.168.2.10"
      "ip_allocation_mode" = "MANUAL"
      "is_primary" = true
      "mac" = "00:50:56:29:08:89"
      "name" = "net-datacloud-r"
      "type" = "org"
    },
  ]
  "org" = "datacloud"
  "storage_profile" = "*"
  "vapp_name" = "vapp-datacloud"
  "vdc" = "vdc-datacloud"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level
* `vapp_name` - (Required) The vApp this VM belongs to.
* `name` - (Required) A name for the VM, unique within the vApp 

## Attribute reference

* `computer_name` -  Computer name to assign to this virtual machine. 
* `catalog_name` -  The catalog name in which to find the given vApp Template
* `template_name` -  The name of the vApp Template to use
* `memory` -  The amount of RAM (in MB) allocated to the VM
* `cpus` -  The number of virtual CPUs allocated to the VM
* `cpu_cores` -  The number of cores per socket
* `metadata` -  Key value map of metadata assigned to this VM
* `disk` -  Independent disk attachment configuration.
* `network` -  A block defining a network interface. Multiple can be used.
* `guest_properties` -  Key value map of guest properties
* `description`  -  The VM description. Note: description is read only. Currently, this field has
  the description of the OVA used to create the VM

See [VM resource](/docs/providers/vcd/r/vapp_vm.html#attribute-reference) for more info about VM attributes.
