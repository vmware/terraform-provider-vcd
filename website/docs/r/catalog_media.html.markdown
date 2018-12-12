---
layout: "vcd"
page_title: "vCloudDirector: vcd_catalog_media"
sidebar_current: "docs-vcd-resource-catalog-media"
description: |-
  Provides a vCloud Director media resource. This can be used to upload and delete media (ISO) file inside a catalog.
---

# vcd\_catalog\_media

Provides a vCloud Director media resource. This can be used to upload media to catalog and delete it.

Supported in provider *v2.0+*

## Example Usage

```
resource "vcd_catalog_media" "myNewMedia" {
  org = "my-org"
  catalog = "my-catalog" 

  name = "my iso"
  description = "new os versions"
  media_path = "/home/user/file.iso"
  upload_piece_size = 10 
  show_upload_progress = true
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `catalog` - (Required) The name of the catalog where to upload media file
* `name` - (Required) Media file name in catalog
* `description` - (Optional) - Description of media file
* `media_path` - (Required) - Absolute or relative path to file to upload
* `upload_piece_size` - (Optional) - size in MB for splitting upload size. It can possibly impact upload performance. Default 1MB.
* `show_upload_progress` - (Optional) - Default false. Allows to see upload progress