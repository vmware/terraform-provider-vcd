---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_solution_add_on"
sidebar_current: "docs-vcd-resource-solution-add-on"
description: |-
  Provides a resource to manage Solution Add-Ons in Cloud Director. A Solution Add-On is the
  representation of a solution that is custom built for Cloud Director in the Cloud
  Director extensibility ecosystem. A Solution Add-On can encapsulate UI and API Cloud
  Director extensions together with their backend services and lifecycle management. Solution
  Add-Ons are distributed as .iso files. A Solution Add-on can contain numerous
  elements: UI plugins, vApps, users, roles, runtime defined entities, and more.
---

# vcd\_solution\_add\_on

Supported in provider *v3.13+* and VCD 10.4.1+.

Provides a resource to manage Solution Add-Ons in Cloud Director. A Solution Add-On is the
representation of a solution that is custom built for Cloud Director in the Cloud
Director extensibility ecosystem. A Solution Add-On can encapsulate UI and API Cloud Director
extensions together with their backend services and lifecycle management. Solution аdd-оns are
distributed as .iso files. A Solution Add-On can contain numerous elements: UI plugins, vApps,
users, roles, runtime defined entities, and more.

~> Only `System Administrator` can create this resource.

## Example Usage (Uploading an image and creating a Solution Add-On entry)

```hcl
data "vcd_solution_landing_zone" "slz" {}

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
  catalog_item_id        = data.vcd_catalog_media.dse14.catalog_item_id
  add_on_path            = "/Users/demo/Downloads/vmware-vcd-ds-1.4.0-23376809.iso"
  auto_trust_certificate = true
}

```

## Example usage (Using already uploaded image)
```hcl
data "vcd_solution_landing_zone" "slz" {}

data "vcd_catalog_media" "dse14" {
  org        = "solution_org"
  catalog_id = tolist(data.vcd_solution_landing_zone.slz.catalog)[0].id

  name = basename("/Users/demo/Downloads/vmware-vcd-ds-1.4.0-23376809.iso")
}

resource "vcd_solution_add_on" "dse14" {
  org                    = "solution_org"
  catalog_item_id        = data.vcd_catalog_media.dse14.catalog_item_id
  add_on_path            = "/Users/demo/Downloads/vmware-vcd-ds-1.4.0-23376809.iso"
  auto_trust_certificate = true
}
```

## Argument Reference

The following arguments are supported:

* `catalog_item_id` - (Required) The catalog item ID of Solution Add-On. It can be uploaded using
  [`vcd_catalog_media`](/providers/vmware/vcd/latest/docs/resources/catalog_media)
* `add_on_path` - (Required) Local filesystem path of Solution Add-On. The Add-On image is required
locally as some of the metadata must be extracted for creating the Solution Add-On entry in Cloud
Director.
* `auto_trust_certificate` - (Optional) Plugin can automatically trust the certificate of Solution
  Add-On. This is required for Solution Add-Ons to work and one will have to do it manually before
  using the Solution Add-On if not enabled here.


## Attribute Reference

The following attributes are exported on this resource:

* `name` - reports the name of Solution Add-On as defined in the system. This name can be also used
  in data source
  [`vcd_solution_add_on`](/providers/vmware/vcd/latest/docs/data-sources/solution_add_on)
* `state` - reports the state of parent [Runtime Defined
  Entity](/providers/vmware/vcd/latest/docs/resources/rde)

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing Solution Add-On configuration can be [imported][docs-import] into this resource via
supplying path for it. It can be imported either by ID or by Name. This might not be trivial to
lookup therefore there is a helper for listing available items:

```
terraform import vcd_solution_add_on.dse14 list@
vcd_solution_add_on.dse14: Importing from ID "list@"...
data.vcd_solution_add_on.dse14: Reading...
data.vcd_solution_add_on.dse14: Read complete after 1s
╷
│ Error: resource was not imported! 
│ No    ID                                                                              Name                            Status  Extension Name  Version
│ --    --                                                                              -------                         ------  ------          ------
│ 1     urn:vcloud:entity:vmware:solutions_add_on:45ce689b-acf7-458f-85af-953871aa1f2e  vmware.ds-1.4.0-23376809        READY   ds              1.4.0-23376809
│ 2     urn:vcloud:entity:vmware:solutions_add_on:26818d72-b2bc-41a9-9f75-898e8d551491  vmware.ds-1.3.0-22829404        READY   ds              1.3.0-22829404
│ 3     urn:vcloud:entity:vmware:solutions_add_on:1a38eb2d-75f5-4651-bbdc-cb80f489eca0  vmware.ose-3.0.0-23443325       READY   ose             3.0.0-23443325
```



An import then can be done either by ID

```
terraform import vcd_solution_add_on.dse14 urn:vcloud:entity:vmware:solutions_add_on:45ce689b-acf7-458f-85af-953871aa1f2e
```

Or by Name

```
terraform import vcd_solution_add_on.dse14 vmware.ds-1.4.0-23376809
```

[docs-import]: https://www.terraform.io/docs/import/