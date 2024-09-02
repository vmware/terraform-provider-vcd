---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_catalog_access_control"
sidebar_current: "docs-vcd-data-source-catalog-access-control"
description: |-
  Provides a data source to read Access Control details from a Catalog in VMware Cloud Director.
---

# vcd\_catalog\_access\_control

Provides a data source to read Access Control details from a Catalog in VMware Cloud Director.

-> **Note:** Access control reads run in tenant context, meaning that, even if the user is a system administrator,
in every request it uses headers items that define the tenant context as restricted to the organization to which the Catalog belongs.

Supported in provider *v3.14+*

## Example Usage

```hcl
data "vcd_catalog" "catalog" {
  name = "my-catalog"
}

data "vcd_catalog_access_control" "ac" {
  catalog_id = data.vcd_catalog.catalog.id
}

output "shared_with" {
  value = data.vcd_catalog_access_control.ac.shared_with
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which the Catalog belongs. Optional if defined at provider level.
* `catalog_id` - (Required) A unique identifier for the Catalog.

## Attribute Reference

All the arguments from [the `vcd_catalog_access_control` resource](/providers/vmware/vcd/latest/docs/resources/catalog_access_control)
are available as read-only.
