---
layout: "vcd"
page_title: "vCloudDirector: vcd_vapp"
sidebar_current: "docs-vcd-resource-vapp"
description: |-
  Provides a vCloud Director vApp resource. This can be used to create, modify, and delete vApps.
---

# vcd\_vapp

Provides a vCloud Director vApp resource. This can be used to create,
modify, and delete vApps. A vApp is a container for VMs, it is created without any VMs. Networks available to the VMs, both vApp specific and public must be assigned to the vApp.

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
```


## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the vApp
* `organization_network` - (Optional) List of organization networks by name available in the virtual datacenter.
* `vapp_network` - (Optional) List of internal network definitions only available to virtual machines within this vApp. 

`vapp_network` supports the following arguments:

* `name` - (Required) Name of the vApp network, must be unique within the vApp resource.
* `description` - (Optional) Description of the vApp network.
* `gateway` - (Required) Gateway address of the vApp network.
* `netmask` - (Required) Netmask address of the vApp network.
* `dns1` - (Required) First DNS server for vApp network.
* `dns2` - (Required) Second DNS server for vApp network.
* `start` - (Required) Start of IP range given to VMs with DHCP or POOL. Must correspond with gateway and netmask.
* `end` - (Required) End of IP range given to VMs with DHCP or POOL. Must correspond with gateway and netmask.
* `parent` - (Required) An `orginzation_network` to connect the internal network to.
* `nat` - (Required) Make the `organization_network` set in parent available by NAT.
* `dhcp` - (Required) Set up a DHCP server on the internal network.



