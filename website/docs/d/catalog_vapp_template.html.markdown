---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_catalog_vapp_template"
sidebar_current: "docs-vcd-data-source-catalog-vapp_template"
description: |-
  Provides a vApp Template data source.
---

# vcd\_catalog\_vapp\_template

Provides a VMware Cloud Director vApp Template data source. A vApp Template can be used to reference an already existing
vApp Template in VCD and use its data within other resources or data sources.

Supported in provider *v3.8+*

## Example: Fetching a vApp Template from a Catalog

```hcl
data "vcd_catalog" "my-catalog" {
  org  = "my-org"
  name = "my-catalog"
}

data "vcd_catalog_vapp_template" "my-first-vapp-template" {
  org        = "my-org"
  catalog_id = data.vcd_catalog.my-catalog.id
  name       = "my-first-vapp-template"
}

resource "vcd_catalog_vapp_template" "my-second-vapp_template" {
  # Using the data source, two properties from another vApp Templates are
  # used in this resource.
  # You can read it as "use the org from vApp Template `my-first-vapp-template`"
  # and "use the catalog from vApp Template `my-first-vapp-template`"
  org        = data.vcd_catalog_vapp_template.my-first-vapp-template.org
  catalog_id = data.vcd_catalog_vapp_template.my-first-vapp-template.catalog_id

  name = "my-second-item"

  # The description uses the data source to create a dynamic text
  # The description will become "Belongs to my-cat"
  description       = "Belongs to ${data.vcd_catalog.my-catalog.name}"
  ova_path          = "/path/to/test_vapp_template.ova"
  upload_piece_size = 5

  # Assign all the metadata from the vApp template to this new one.
  dynamic "metadata_entry" {
    for_each = data.vcd_catalog_vapp_template.photon.metadata_entry
    content {
      key         = metadata_entry.value["key"]
      value       = metadata_entry.value["value"]
      type        = metadata_entry.value["type"]
      is_system   = metadata_entry.value["is_system"]
      user_access = metadata_entry.value["user_access"]
    }
  }
}
```

## Example: Fetching a vApp Template from a VDC

```hcl
data "vcd_org_vdc" "my-vdc" {
  org  = "my-org"
  name = "my-vdc"
}

data "vcd_catalog_vapp_template" "my-first-vapp-template" {
  org    = "my-org"
  vdc_id = data.vcd_org_vdc.my-vdc.id
  name   = "my-first-vapp-template"
}

resource "vcd_catalog_vapp_template" "my-second-vapp_template" {
  # Using the data source, two properties from another vApp Templates are
  # used in this resource.
  # You can read it as "use the org from vApp Template `my-first-vapp-template`"
  # and "use the catalog from vApp Template `my-first-vapp-template`"
  org        = data.vcd_catalog_vapp_template.my-first-vapp-template.org
  catalog_id = data.vcd_catalog_vapp_template.my-first-vapp-template.catalog_id

  name = "my-second-item"

  # The description uses the data source to create a dynamic text
  # The description will become "Belongs to my-vdc"
  description       = "Belongs to ${data.vcd_org_vdc.my-vdc.name}"
  ova_path          = "/path/to/test_vapp_template.ova"
  upload_piece_size = 5
  
  # Assign all the metadata from the vApp template to this new one.
  dynamic "metadata_entry" {
    for_each = data.vcd_catalog_vapp_template.photon.metadata_entry
    content {
      key         = metadata_entry.value["key"]
      value       = metadata_entry.value["value"]
      type        = metadata_entry.value["type"]
      is_system   = metadata_entry.value["is_system"]
      user_access = metadata_entry.value["user_access"]
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional, but required if not set at provider level) Org name 
* `catalog_id` - (Required) ID of the catalog containing the vApp Template. Can't be used if a specific VDC identifier is set (`vdc_id`).
* `vdc_id` - (Required) ID of the VDC to which the vApp Template belongs. Can't be used if a specific Catalog is set (`catalog_id`).
* `name` - (Required) vApp Template name (optional when `filter` is used)
* `filter` - (Optional) Retrieves the data source using one or more filter parameters

## Attribute Reference

* `description` - vApp Template description
* `vm_names` - Set of VM names within the vApp template
* `metadata` - (Deprecated) Use `metadata_entry` instead. Key/value map of metadata for the associated vApp template.
* `metadata_entry` - A set of metadata entries assigned to this vApp Template. See [Metadata](/providers/vmware/vcd/latest/docs/resources/catalog_vapp_template#metadata) section for details.

## Filter arguments

* `name_regex` (Optional) matches the name using a regular expression.
* `date` (Optional) is an expression starting with an operator (`>`, `<`, `>=`, `<=`, `==`), followed by a date, with
  optional spaces in between. For example: `> 2020-02-01 12:35:00.523Z`
  The filter recognizes several formats, but one of `yyyy-mm-dd [hh:mm[:ss[.nnnZ]]]` or `dd-MMM-yyyy [hh:mm[:ss[.nnnZ]]]`
  is recommended.
  Comparison with equality operator (`==`) need to define the date to the microseconds.
* `latest` (Optional) If `true`, retrieve the latest item among the ones matching other parameters. If no other parameters
  are set, it retrieves the newest item.
* `earliest` (Optional) If `true`, retrieve the earliest item among the ones matching other parameters. If no other parameters
  are set, it retrieves the oldest item.
* `metadata` (Optional) One or more parameters that will match metadata contents.

See [Filters reference](/providers/vmware/vcd/latest/docs/guides/data_source_filters) for details and examples.
