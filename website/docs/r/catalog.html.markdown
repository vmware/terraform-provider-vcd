---
layout: "vcd"
page_title: "vCloudDirector: vcd_catalog"
sidebar_current: "docs-vcd-resource-catalog"
description: |-
Provides a vCloud Director catalog resource. This can be used to create and delete a catalog.
---

# vcd\_catalog

Provides a vCloud Director catalog resource. This can be used to create and delete a catalog.

Supported in provider *v2.0+*

## Example Usage

```
resource "vcd_catalog" "myNewCatalog" {
  org = "my-org"

  name = "my-catalog"
  description = "catalog for files"  
  delete_recursive  = "true"  
  delete_force      = "true"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `name` - (Required) Catalog name
* `description` - (Optional) - Description of catalog
* `delete_recursive` - (Required) - When destroying use delete_recursive=True to remove the catalog and any objects it contains that are in a state that normally allows removal
* `delete_force` -(Required) - When destroying use delete_force=True with delete_recursive=True to remove a catalog and any objects it contains, regardless of their state