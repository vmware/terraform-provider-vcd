---
layout: "vcd"
page_title: "vCloudDirector: vcd_catalog_item"
sidebar_current: "docs-vcd-resource-catalog-item"
description: |-
  Provides a vCloud Director catalog item resource. This can be used to upload and delete OVA file inside a catalog.
---

# vcd\_catalog\_item

Provides a vCloud Director catalog item resource. This can be used to upload OVA to catalog and delete it.

Supported in provider *v2.0+*

## Example Usage

```
resource "vcd_catalog_item" "myNewCatalogItem" {
  org = "my-org"
  catalog = "my-catalog" 

  name = "my ova"
  description = "new vapp template"
  ova_path = "/home/user/file.ova"
  upload_piece_size = 10 
  show_upload_progress = true
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `catalog` - (Required) The name of the catalog where to upload OVA file
* `name` - (Required) Item name in catalog
* `description` - (Optional) - Description of item
* `ova_path` - (Required) - Absolute or relative path to file to upload
* `upload_piece_size` - (Optional) - size in MB for splitting upload size. It can possibly impact upload performance. Default 1MB.
* `show_upload_progress` - (Optional) - Default false. Allows to see upload progress