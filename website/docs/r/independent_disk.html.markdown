---
layout: "vcd"
page_title: "vCloudDirector: vcd_independent_disk"
sidebar_current: "docs-vcd-independent-disk"
description: |-
  Provides a vCloud Director independent disk resource. This can be used to create and delete independent disks.
---

# vcd\_independent\_disk

Provides a vCloud Director independent disk resource. This can be used to create and delete independent disks.

Supported in provider *v2.1+*

## Example Usage

```
resource "vcd_independent_disk" "myNewIndependentDisk" {  
  vdc             = "my-vcd"
  
  name            = "logDisk"
  size            = "33000"
  bus_type        = "SCSI"
  bus_sub_type    = "VirtualSCSI"
  storage_profile = "external"
}

resource "vcd_vapp_vm" "web2" {
  vapp_name     = "${vcd_vapp.web.name}"

...
  
  disk {
    name = "${vcd_independent_disk.myNewIndependentDisk.name}"
    bus_number = 1
    unit_number = 0
  }

  depends_on = ["vcd_independent_disk.myNewIndependentDisk"]
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level
* `name` - (Required) Disk name
* `size` - (Optional; **Deprecated**) Size of disk in MB
* `size_in_bytes` - (Optional; *v2.5+*) Size in bytes
* `bus_type` - (Optional) Disk bus type. Values can be: IDE, SCSI, SATA 
* `bus_sub_type` - (Optional) Disk bus subtype. Values can be: "IDE" for IDE. buslogic, lsilogic, lsilogicsas, VirtualSCSI for SCSI and ahci for SATA
* `storage_profile` - (Optional) The name of storage profile where disk will be created

## Attribute reference

Supported in provider *v2.5+*

* `iops` - (Computed) IOPS request for the created disk
* `owner_name` - (Computed) The owner name of the disk
* `datastore_name` - (Computed) Data store name. Readable only for system user.
* `is_attached` - (Computed) True if the disk is already attached

## Importing

Supported in provider *v2.5+*

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing independent disk can be [imported][docs-import] into this resource via supplying its path.
The path for this resource is made of vdc-name.disk-name
For example, using this structure, representing a independent disk that was **not** created using Terraform:

```hcl
resource "vcd_independent_disk" "tf-myDisk" {
  vdc     = "my-vdc"
  name    = "my-disk"
}
```

You can import such independent disk into terraform state using this command

```
terraform import vcd_independent_disk.tf-myDisk vdc-name.my-disk-name
```

[docs-import]:https://www.terraform.io/docs/import/

After importing, if you run `terraform plan` you will see the rest of the values and modify the script accordingly for
further operations.