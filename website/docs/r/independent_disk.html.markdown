---
layout: "vcd"
page_title: "vCloudDirector: vcd_independent_disk"
sidebar_current: "docs-vcd-independent-disk"
description: |-
  Provides a vCloud Director independent disk resource. This can be used to create and delete independent disks.
---

# vcd\_inserted\_media

Provides a vCloud Director independent disk resource. This can be used to create and delete independent disks.

Supported in provider *v2.1+*

## Example Usage

```
resource "vcd_independent_disk" "myNewIndependentDisk" {
  org             = "my-org"
  vdc             = "my-vcd"
  
  name            = "logDisk"
  size            = "33"
  bus_type        = "SCSI"
  bus_sub_type    = "VirtualSCSI"
  storage_profile = "external"
}

resource "vcd_vapp_vm" "web2" {
  vapp_name     = "${vcd_vapp.web.name}"

...
  
  disk {
    name = "${vcd_independent_disk.Disk_1.name}"
    bus_number = 1
    unit_number = 0
  }

  depends_on = ["vcd_independent_disk.myNewIndependentDisk"]
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional; *v2.0+*) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `vdc` - (Optional; *v2.0+*) The name of VDC to use, optional if defined at provider level
* `name` - (Required) Disk name
* `size` - (Required) - Size of disk in GB
* `bus_type` - (Optional) - Disk bus type. Values can be: IDE, SCSI, SATA 
* `bus_sub_type` - (Optional) - Disk bus subtype. Values can be: "IDE" for IDE. buslogic, lsilogic, lsilogicsas, VirtualSCSI for SCSI and ahci for SATA
* `storage_profile` - (Optional) - The name of storage profile where disk will be created