---
layout: "vcd"
page_title: "vCloudDirector: vcd_catalog_media"
sidebar_current: "docs-vcd-resource-catalog-media"
description: |-
  Provides a vCloud Director Media resource. This can be used to upload and delete media file in catalog.
---

# vcd\_catalog\_media

Provides a vCloud Director Media resource. This can be used to upload media to catalog and delete it.

## Example Usage

```
resource "vcd_catalog_media" "myNewMedia" {
  org = "my-org"
  catalog = "my-catalog" 

  name = "my iso"
  description = "new os versions 1.2"
  media_path = "/home/user/file.iso"
  upload_piece_size = 10 
  show_upload_progress = true
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, isn't needed if defined/match at provider level
* `catalog` - (Required) The name catalog where upload media file
* `name` - (Required) Media file name in catalog
* `description` - (Optional) - Description of media file
* `media_path` - (Required) - Absolute ir relative path to file to upload
* `upload_piece_size` - (Optional) - size in MB for dividing upload size. Possibility to impact upload performance. Default 1MB.
* `show_upload_progress` - (Optional) - Default false. Allows to see upload progress
t