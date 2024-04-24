---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_solution_add_on"
sidebar_current: "docs-vcd-resource-solution-add-on"
description: |-
  Provides a resource to manage Solution Add-Ons in Cloud Director. A solution add-on is the
  representation of a solution that is custom built for VMware Cloud Director in the VMware Cloud
  Director extensibility ecosystem. A solution add-on can encapsulate UI and API VMware Cloud
  Director extensions together with their backend services and lifecycle management. Solution
  аdd-оns are distributed as .iso files. A solution add-on can contain numerous
  elements: UI plugins, vApps, users, roles, runtime defined entities, and more.
---

# vcd\_solution\_add\_on

Supported in provider *v3.13+* and VCD 10.4.1+.

Provides a resource to manage Solution Add-Ons in Cloud Director. A solution add-on is the
representation of a solution that is custom built for VMware Cloud Director in the VMware Cloud
Director extensibility ecosystem. A solution add-on can encapsulate UI and API VMware Cloud Director
extensions together with their backend services and lifecycle management. Solution аdd-оns are
distributed as .iso files. A solution add-on can contain numerous elements: UI plugins, vApps,
users, roles, runtime defined entities, and more.

~> Only `System Administrator` can create this resource.

## Example Usage (Uploading an image and create a Solution Add-On entry)

```hcl
data "vcd_solution_landing_zone" "slz" {
    
}

resource "vcd_catalog_media" "dse14" {
  org        = "solution_org"
  catalog_id = tolist(data.vcd_solution_landing_zone.slz.catalog)[0].id

  name              = basename("/Users/demo/Downloads/vmware-vcd-ds-1.4.0-23376809.iso")
  description       = "new os versions"
  media_path        = "/Users/demo/Downloads/vmware-vcd-ds-1.4.0-23376809.iso"
  upload_any_file   = false # Add-ons are packaged in '.iso' files
  upload_piece_size = 10
}

resource "vcd_solution_add_on" "dse14" {
  org               = "solution_org"
  catalog_item_id   = data.vcd_catalog_media.dse14.catalog_item_id
  addon_path        = "/Users/demo/Downloads/vmware-vcd-ds-1.4.0-23376809.iso"
  trust_certificate = true
  accept_eula       = true
}

```

## Example usage (Using already uploaded image)
```hcl
data "vcd_solution_landing_zone" "slz" {
    
}


data "vcd_catalog_media" "dse14" {
  org        = "solution_org"
  catalog_id = tolist(data.vcd_solution_landing_zone.slz.catalog)[0].id

  name = basename("/Users/demo/Downloads/vmware-vcd-ds-1.4.0-23376809.iso")
  #description       = "new os versions"
  #media_path        = "/Users/demo/Downloads/vmware-vcd-ds-1.4.0-23376809.iso"
  #upload_any_file   = false # Add-ons are packaged in '.iso' files
  #upload_piece_size = 10
}

#resource "vcd_catalog_media" "dse14" {
#  org        = "solution_org"
#  catalog_id = data.vcd_catalog.nsxt.id
#
#  name              = basename("/Users/demo/Downloads/vmware-vcd-ds-1.4.0-23376809.iso")
#  description       = "new os versions"
#  media_path        = "/Users/demo/Downloads/vmware-vcd-ds-1.4.0-23376809.iso"
#  upload_any_file   = false # Add-ons are packaged in '.iso' files
#  upload_piece_size = 10
#}

resource "vcd_solution_add_on" "dse14" {
  org               = "solution_org"
  catalog_item_id   = data.vcd_catalog_media.dse14.catalog_item_id
  addon_path        = "/Users/demo/Downloads/vmware-vcd-ds-1.4.0-23376809.iso"
  trust_certificate = true
  accept_eula       = true
}

```

## Argument Reference

The following arguments are supported:

* `org` - (Required) Solution Organization (the one that is used in Landing Zone)
* `catalog_item_id` - (Required) The catalog item ID of Solution Add-on
* `addon_path` - (Required) Local filesystem path of Solution Add-on
* `trust_certificate` - (Required) Plugin can automatically trust the certificate of Solution
  Add-On. This is required for Solution Add-Ons to work and one will have to do it manually before
  using the Solution Add-On.
* `accept_eula` - (Required) A mandatory field to accept EULA of the solution Add-On. EULA will be
  printed to screen when `false`.


## Attribute Reference

The following attributes are exported on this resource:

* `state` - reports the state of parent [Runtime Defined
  Entity](/providers/vmware/vcd/latest/docs/resources/rde)

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing Solution Add-On configuration can be [imported][docs-import] into this resource via
supplying path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_solution_add_on.imported ????????????????
```

The above would import the `??????????????` Solution Add-On.