---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_solution_add_on_instance_publish"
sidebar_current: "docs-vcd-resource-solution-add-on-instance-publish"
description: |-
  Provides a resource to manage publishing configuration of Solution Add-Ons in Cloud Director.

---

# vcd\_solution\_add\_on\_instance\_publish

Supported in provider *v3.13+* and VCD 10.4.1+.

Provides a resource to manage publishing configuration of Solution Add-Ons in Cloud Director.

~> Only `System Administrator` can create this resource.

## Example Usage (Uploading an image and create a Solution Add-On entry)

```hcl
data "vcd_org" "recipient" {
  name = "tenant_org"
}

resource "vcd_solution_add_on_instance_publish" "public" {
  add_on_instance_id     = vcd_solution_add_on_instance.dse14.id
  org_ids                = [data.vcd_org.recipient.id]
  publish_to_all_tenants = false
}

data "vcd_catalog_media" "dse14" {
  org        = "solutions_org"
  catalog_id = data.vcd_catalog.nsxt.id

  name = "vmware-vcd-ds-1.4.0-23376809.iso"
}

resource "vcd_solution_add_on" "dse14" {
  catalog_item_id   = data.vcd_catalog_media.dse14.catalog_item_id
  addon_path        = "vmware-vcd-ds-1.4.0-23376809.iso"
  trust_certificate = true
}

resource "vcd_solution_add_on_instance" "dse14" {
  add_on_id     = vcd_solution_add_on.dse14.id
  accept_eula   = true
  name = "MyDseInstanceName"
  input = {
    input-delete-previous-uiplugin-versions = false
  }
}
```

## Argument Reference

The following arguments are supported:

* `add_on_instance_id` - (Required) Solution Add-On instance ID 
* `org_ids` - (Optional) Recipient Organization IDs
* `publish_to_all_tenants` - (Optional) Set to `true` to publish to everyone

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing Solution Add-On Instance Publishing configuration can be [imported][docs-import] into
this resource via supplying path for it. 


```
terraform import vcd_solution_add_on_instance_publish.public MyDseInstanceName
```

[docs-import]: https://www.terraform.io/docs/import/
