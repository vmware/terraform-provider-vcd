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

  metadata = {
    role    = "web"
    env     = "staging"
    version = "v1"
    my_key  = "my value"
  }

  guest_properties = {
    "guest.hostname"   = "my-host"
    "another.var.name" = "var-value"
  }

  network {
    type               = "org"
    name               = "net"
    ip                 = "10.10.104.161"
    ip_allocation_mode = "MANUAL"
    is_primary         = true
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

  metadata = {
    role    = "web"
    env     = "staging"
    version = "v1"
    my_key  = "my value"
  }

  network {
    type               = "org"
    name               = "net"
    ip                 = "10.10.104.162"
    ip_allocation_mode = "MANUAL"
    is_primary         = true
  }

  network {
    type               = "vapp"
    name               = "vapp-network"
    ip_allocation_mode = "POOL"
  }

  network {
    type               = "none"
    ip_allocation_mode = "NONE"
  }

  disk {
    name        = "logDisk1"
    bus_number  = 1
    unit_number = 0
  }

  disk {
    name        = "logDisk2"
    bus_number  = 1
    unit_number = 1
  }

  guest_properties = {
    "guest.hostname" = "my-hostname"
    "guest.other"    = "another-setting"
  }

  depends_on = ["vcd_vapp.web"]
}

resource "vcd_vapp_vm" "internalDiskOverride" {
  vapp_name     = "${vcd_vapp.web.name}"
  name          = "internalDiskOverride"
  catalog_name  = "Boxes"
  template_name = "lampstack-1.10.1-ubuntu-10.04"
  memory        = 2048
  cpus          = 2
  cpu_cores     = 1

  override_template_disk {
    bus_type         = "paravirtual"
    size_in_mb       = "22384"
    bus_number       = 0
    unit_number      = 0
    iops             = 0
    thin_provisioned = true
    storage_profile  = "*"
  }

  depends_on = ["vcd_vapp.web"]
}

```

## Argument Reference

The following arguments are supported:

* `org` - (Optional; *v2.0+*) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `vdc` - (Optional; *v2.0+*) The name of VDC to use, optional if defined at provider level
* `vapp_name` - (Required) The vApp this VM belongs to.
* `name` - (Required) A name for the VM, unique within the vApp 
* `computer_name` - (Optional; *v2.5+*) Computer name to assign to this virtual machine. 
* `catalog_name` - (Required) The catalog name in which to find the given vApp Template
* `template_name` - (Required) The name of the vApp Template to use
* `memory` - (Optional) The amount of RAM (in MB) to allocate to the VM
* `cpus` - (Optional) The number of virtual CPUs to allocate to the VM. Socket count is a result of: virtual logical processors/cores per socket. The default is 1
* `cpu_cores` - (Optional; *v2.1+*) The number of cores per socket. The default is 1
* `metadata` - (Optional; *v2.2+*) Key value map of metadata to assign to this VM
* `initscript` (Optional) Script to run on initial boot or with customization.force=true set
* `storage_profile` (Optional; *v2.6+*) Storage profile to override the default one
* `network_name` - (Optional; **Deprecated** by `network`) Name of the network this VM should connect to.
* `vapp_network_name` - (Optional; v2.1+; **Deprecated** by `network`) Name of the vApp network this VM should connect to.
* `ip` - (Optional; **Deprecated** by `network`) The IP to assign to this vApp. Must be an IP address or
one of `dhcp`, `allocated`, or `none`. If given the address must be within the
  `static_ip_pool` set for the network. If left blank, and the network has
  `dhcp_pool` set with at least one available IP then this will be set with
DHCP.
* `power_on` - (Optional) A boolean value stating if this VM should be powered on. Default is `true`
* `accept_all_eulas` - (Optional; *v2.0+*) Automatically accept EULA if OVA has it. Default is `true`
* `disk` - (Optional; *v2.1+*) Independent disk attachment configuration. See [Disk](#disk) below for details.
* `expose_hardware_virtualization` - (Optional; *v2.2+*) Boolean for exposing full CPU virtualization to the
guest operating system so that applications that require hardware virtualization can run on virtual machines without binary
translation or paravirtualization. Useful for hypervisor nesting provided underlying hardware supports it. Default is `false`.
* `network` - (Optional; *v2.2+*) A block to define network interface. Multiple can be used. See [Network](#network) and 
example for usage details. **Deprecates**: `network_name`, `ip`, `vapp_network_name`.
* `customization` - (Optional; *v2.5+*) A block to define for guest customization options. See [Customization](#customization)
* `guest_properties` - (Optional; *v2.5+*) Key value map of guest properties
* `description`  - (Computed; *v2.6+*) The VM description. Note: description is read only. Currently, this field has
  the description of the OVA used to create the VM
* `override_template_disk` - (Optional; *v2.6+*) Allows to update internal disk in template before first VM boot. Disk are matched by `bus_type`, `bus_number` and `unit_number`. See [Override template Disk](#override_template_disk) below for details.

<a id="disk"></a>
## Disk

* `name` - (Required) Independent disk name
* `bus_number` - (Required) Bus number on which to place the disk controller
* `unit_number` - (Required) Unit number (slot) on the bus specified by BusNumber.


<a id="network"></a>
## Network

* `type` (Required) Network type, one of: `none`, `vapp` or `org`. `none` creates a NIC with no network attached, `vapp` attaches a vApp network, while `org` attaches organization VDC network.
* `name` (Optional) Name of the network this VM should connect to. Always required except for `type` `NONE`.
* `is_primary` (Optional) Set to true if network interface should be primary. First network card in the list will be primary by default.
* `mac` - (Computed) Mac address of network interface.
* `ip_allocation_mode` (Required) IP address allocation mode. One of `POOL`, `DHCP`, `MANUAL`, `NONE`:  

  * `POOL` - Static IP address is allocated automatically from defined static pool in network.
  
  * `DHCP` - IP address is obtained from a DHCP service. Field `ip` is not guaranteed to be populated. Because of this it may appear
  after multiple `terraform refresh` operations.
  
  * `MANUAL` - IP address is assigned manually in the `ip` field. Must be valid IP address from static pool.
  
  * `NONE` - No IP address will be set because VM will have a NIC without network.

* `ip` (Optional, Computed) Settings depend on `ip_allocation_mode`. Field requirements for each `ip_allocation_mode` are listed below:

  * `ip_allocation_mode=POOL` - **`ip`** value must be omitted or empty string "". Empty string may be useful when doing HCL
  variable interpolation. Field `ip` will be populated with an assigned IP from static pool after run.
  
  * `ip_allocation_mode=DHCP` - **`ip`** value must be omitted or empty string "". Field `ip` is not guaranteed to be populated
  after run due to the VM lacking VMware tools or not working properly with DHCP. Because of this `ip` may also appear after multiple `terraform refresh` operations when is reported back to vCD.

  * `ip_allocation_mode=MANUAL` - **`ip`** value must be valid IP address from a subnet defined in `static pool` for network.

  * `ip_allocation_mode=NONE` - **`ip`** field can be omitted or set to an empty string "". Empty string may be useful when doing HCL variable interpolation.

<a id="override_template_disk"></a>
## Override template disk

* `bus_type` - (Required) The type of disk controller. Possible values: `ide`, `parallel`( LSI Logic Parallel SCSI), `sas`(LSI Logic SAS (SCSI)), `paravirtual`(Paravirtual (SCSI)), `sata`. 
* `size_in_mb` - (Required) The size of the disk in MB. 
* `bus_number` - (Required) The number of the SCSI or IDE controller itself.
* `unit_number` - (Required) The device number on the SCSI or IDE controller of the disk.
* `thin_provisioned` - (Optional) Specifies whether the disk storage is pre-allocated or allocated on demand.
* `iops` - (Optional) Specifies the IOPS for the disk. Default - 0.
* `storage_profile` - (Optional) Storage profile which overrides the VM default one.


<a id="customization"></a>
## Customization

* `force` (Optional) **Warning.** `true` value will cause the VM to reboot on every `apply` operation.
This field works as a flag and triggers force customization when `true` during an update 
(`terraform apply`) every time. It never complains about a change in statefile. Can be used when guest customization
is needed after VM configuration (e.g. NIC change, customization options change, etc.) and then set back to `false`.
**Note.** It will not have effect when `power_on` field is set to `false`. See [example workflow below](#example-forced-customization-workflow).

## Example forced customization workflow

Step 1 - Setup VM:

```hcl
resource "vcd_vapp_vm" "web2" {
  vapp_name     = "${vcd_vapp.web.name}"
  name          = "web2"
  catalog_name  = "Boxes"
  template_name = "lampstack-1.10.1-ubuntu-10.04"
  memory        = 2048
  cpus          = 1

  network {
    type               = "org"
    name               = "net"
    ip                 = "10.10.104.162"
    ip_allocation_mode = "MANUAL"
  }
}
```

Step 2 - Change VM configuration and force customization (VM will be rebooted during
`terraform apply`):

```hcl
resource "vcd_vapp_vm" "web2" {
//...
  network {
    type               = "org"
    name               = "net"
    ip_allocation_mode = "DHCP"
  }

  customization {
    force = true
  }
}
```

Step 3 - Once customization is done, set the force customization flag to false (or remove it) to
prevent forcing customization on every `terraform apply` command:

```hcl
resource "vcd_vapp_vm" "web2" {
//...
  network {
    type               = "org"
    name               = "net"
    ip_allocation_mode = "DHCP"
  }

  customization {
    force = false
  }
}
```

## Attribute Reference

The following additional attributes are exported:

* `internal_disk` - (*v2.6+*) A block provides internal disk of VM details. See [Internal Disk](#internalDisk) below for details.

<a id="internalDisk"></a>
## Internal disk

* `disk_id` - (*v2.6+*) Specifies a unique identifier for this disk in the scope of the corresponding VM.
* `bus_type` - (*v2.6+*) The type of disk controller. Possible values: `ide`, `parallel`( LSI Logic Parallel SCSI), `sas`(LSI Logic SAS (SCSI)), `paravirtual`(Paravirtual (SCSI)), `sata`. 
* `size_in_mb` - (*v2.6+*) The size of the disk in MB. 
* `bus_number` - (*v2.6+*) The number of the SCSI or IDE controller itself.
* `unit_number` - (*v2.6+*) The device number on the SCSI or IDE controller of the disk.
* `thin_provisioned` - (*v2.6+*) Specifies whether the disk storage is pre-allocated or allocated on demand.
* `iops` - (*v2.6+*) Specifies the IOPS for the disk. Default - 0.
* `storage_profile` - (*v2.6+*) Storage profile which overrides the VM default one.


## Importing

Supported in provider *v2.6+*

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing VM can be [imported][docs-import] into this resource via supplying its path.
The path for this resource is made of org-name.vdc-name.vapp-name.vm-name
For example, using this structure, representing a VM that was **not** created using Terraform:

```hcl
resource "vcd_vapp_vm" "tf-vm" {
  name              = "my-vm"
  org               = "my-org"
  vdc               = "my-vdc"
  vapp_name         = "my-vapp"
}
```

You can import such vapp into terraform state using this command

```
terraform import vcd_vapp_vm.tf-vm my-org.my-vdc.my-vapp.my-vm
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After importing, the data for this VM will be in the state file (`terraform.tfstate`). If you want to use this
resource for further operations, you will need to integrate it with data from the state file, and with some data that
is used to create the VM, such as `catalog_name`, `template_name`.
