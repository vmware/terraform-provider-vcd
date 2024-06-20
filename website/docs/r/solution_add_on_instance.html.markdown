---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_solution_add_on_instance"
sidebar_current: "docs-vcd-resource-solution-add-on-instance"
description: |-
  Provides a resource to manage Solution Add-On Instances in Cloud Director. A Solution Add-On Instance
  is created from an existing Solution Add-On by supplying configuration values of that particular instance.
---

# vcd\_solution\_add\_on\_instance

Supported in provider *v3.13+* and VCD 10.4.1+.

Provides a resource to manage Solution Add-Ons Instances in Cloud Director. 

~> Only `System Administrator` can create this resource.

## Example Usage (Uploading an image and create a Solution Add-On entry)

```hcl
resource "vcd_solution_add_on_instance" "dse14" {
  add_on_id                     = vcd_solution_add_on.dse14.id
  accept_eula                   = true
  name                          = "MyDseInstance"
  validate_only_required_inputs = true

  input = {
    delete-previous-uiplugin-versions = true
  }

  delete_input = {
    force-delete = true
  }
}

resource "vcd_solution_add_on" "dse14" {
  catalog_item_id   = data.vcd_catalog_media.dse14.catalog_item_id
  addon_path        = "vmware-vcd-ds-1.4.0-23376809.iso"
  trust_certificate = true

  depends_on = [vcd_solution_landing_zone.slz]
}
```


## Argument Reference

The following arguments are supported:

* `add_on_id` - (Required) Existing Solution Add-On ID
  [`vcd_solution_add_on`](/providers/vmware/vcd/latest/docs/resources/solution_add_on)
* `accept_eula` - (Required) Solution Add-On Instance cannot be create if EULA is not accepted.
  Supplying a `false` value will print EULA.
* `name` - (Required) Name of Solution Add-On Instance
* `validate_only_required_inputs` - (Optional) By default (`false`) will check that all fields are
defined in `input` and `delete_input` fields. It will only validate fields that are marked as
required when set to `true`. Update is a noop that will affect further operation.
* `input` - (Required) A map of keys and values as required for a particular Solution Add-On
Instance. It will require all values that are specified in a particular Add-On schema unless
`validate_only_required_inputs=true` is set. Missing a value will print an error message with all
field descriptions and missing value.
* `delete_input` - (Required) Just like `input` field for creation, it is a map of keys and values
as required for removal of a particular Solution Add-On. It will require all values that are
specified in a particular Add-On schema unless `validate_only_required_inputs=true` is set. Missing
a value will print an error message with all field descriptions and missing value. Update is a no-op
operation 


## Attribute Reference

The following attributes are exported on this resource:

* `rde_state` - reports the state of parent [Runtime Defined
  Entity](/providers/vmware/vcd/latest/docs/resources/rde)

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing Solution Add-On Instance configuration can be [imported][docs-import] into this resource
via supplying path for it. 


```
terraform import vcd_solution_add_on_instance.dse14 MyDseInstance
```

[docs-import]: https://www.terraform.io/docs/import/
