---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_cloned_vapp"
sidebar_current: "docs-vcd-resource-cloned-vapp"
description: |-
  Provides a VMware Cloud Director Cloned vApp resource. This can be used to create vApps from either a vApp template or another vApp.
---

# vcd\_cloned\_vapp

Provides a VMware Cloud Director Cloned vApp resource. This can be used to create vApps from either a vApp template or another vApp.
This resource should be used only on creation, although deletion also works. The result of using this resource is a
regular vApp ([`vcd_vapp`](/providers/vmware/vcd/latest/docs/resources/vapp)), with all its contents derived by either a vApp template or another vApp.
As of this first implementation, no configuration is available: the vApp is simply cloned from the source vApp template
or vApp.

This resource is useful in two scenarios:

* When users want to create a vApp from a vApp template containing several VMs.
* When users want to move or copy a vApp. The "move" means using this resource with `delete_source = true`, and it is in
  fact a copy followed by a deletion.

Supported in provider *v3.10+*

## Example of creation from vApp template

```hcl
data "vcd_catalog" "cat" {
  name = "my-catalog"
}

data "vcd_catalog_vapp_template" "tmpl" {
  catalog_id = data.vcd_catalog.cat.id
  name       = "3VM"
}

resource "vcd_cloned_vapp" "vapp_from_template" {
  name        = "VappFromTemplate"
  description = "vApp from template"
  power_on    = true
  source_id   = data.vcd_catalog_vapp_template.tmpl.id
  source_type = "template"
}
```

## Example of creation from vApp

```hcl
data "vcd_vapp" "source_vapp" {
  name = "source_vapp"
}

resource "vcd_cloned_vapp" "vapp_from_vapp" {
  name          = "VappFromVapp"
  description   = "vApp from vApp"
  power_on      = true
  source_id     = data.vcd_vapp.source_vapp.id
  source_type   = "vapp"
  delete_source = false
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the vApp
* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level
* `description` (Optional) An optional description for the vApp, up to 256 characters.
* `power_on` - (Optional) A boolean value stating if this vApp should be powered on. Default is `false`.
* `source_type` - (Required) The type of the source to use: one of `template` or `vapp`.
* `source_id` - (Required) The ID of the source to use.
* `delete_source` - (Optional) A boolean value of `true` or `false` stating if the source entity should be deleted after creation.
  A source vApp can only be deleted if it is fully powered off.

## Attribute reference

* `href` - (Computed) The vApp Hyper Reference.
* `status` - (Computed) The vApp status as a numeric code.
* `status_text` - (Computed) The vApp status as text.
* `vm_list` - (Computed) The list of VM names included in this vApp, in alphabetic order.

## Importing

There is no importing for this resource, as it should be used only on creation. A vApp can be imported using `vcd_vapp`.
See [Importing resources](https://registry.terraform.io/providers/vmware/vcd/3.10.0/docs/guides/importing_resources) for
the theory and some [examples](https://github.com/vmware/terraform-provider-vcd/tree/main/examples/importing/vapp-vm) in
`terraform-provider-vcd` repository.
