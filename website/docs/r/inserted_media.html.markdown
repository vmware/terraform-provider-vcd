---
layout: "vcd"
page_title: "vCloudDirector: vcd_inserted_media"
sidebar_current: "docs-vcd-inserted-media"
description: |-
  Provides a vCloud Director resource for inserting or ejecting media (ISO) file for the VM. Create this resource for inserting the media, and destroy it for ejecting.
---

# vcd\_catalog\_media

Provides a vCloud Director resource for inserting or ejecting media (ISO) file for the VM. Create this resource for inserting the media, and destroy it for ejecting.

Supported in provider *v2.0+*

## Example Usage

```
resource "vcd_inserted_media" "myInsertedMedia" {
  org = "my-org"
  vdc = "my-vcd"
  catalog = "my-catalog" 
  name = "my-iso"
  
  vapp_name = "my-vApp"
  vm_name = "my-VM"
}
```

## Argument Reference

The following arguments are supported:
* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level
* `catalog` - (Required) The name of the catalog where to find media file
* `name` - (Required) Media file name in catalog which will be inserted to VM
* `vapp_name` - (Required) - The name of vApp to find
* `vm_name` - (Required) - The name of VM to be used to insert media file