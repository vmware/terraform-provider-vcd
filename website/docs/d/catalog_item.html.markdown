---
layout: "vcd"
page_title: "vCloudDirector: vcd_catalog_item"
sidebar_current: "docs-vcd-data-source-catalog-item"
description: |-
  Provides a catalog item data source.
---

# vcd\_catalog\_item

Provides a vCloud Director Catalog item data source. A Catalog item can be used to reference a catalog item and use its 
data within other resources or data sources.

Supported in provider *v2.5+*

## Example Usage

```hcl
data "vcd_catalog_item" "my-first-item" {
  org     = "my-org"
  catalog = "my-cat"
  name    = "my-first-item"
}

resource "vcd_catalog_item" "my-second-item" {
  # Using the data source, two properties from another catalog items are
  # used in this resource.
  # You can read it as "use the org from catalog item `my-first-item`"
  # and "use the catalog from catalog item `my-first-item`"
  org     = "${data.vcd_catalog_item.my-first-item.org}"
  catalog = "${data.vcd_catalog_item.my-first-item.catalog}"

  name                 = "my-second-item"
  # The description uses the data source to create a dynamic text
  # The description will become "Belongs to my-cat"
  description          = "Belongs to ${data.vcd_catalog_item.my-first-item.catalog}"
  ova_path             = "/path/to/test_vapp_template.ova"
  upload_piece_size    = 5
  show_upload_progress = "true"
  metadata             = "${data.vcd_catalog_item.my-first-item.metadata}"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional, but required if not set at provider level) Org name 
* `catalog` - (Required) Catalog name
* `name` - (Optional) Catalog Item name (Required when `filter` is not provided)
* `filter` - (Optional; *v2.8+*) Retrieve the data source with several parameters. See below for more info

## Attribute Reference

* `description` - Catalog item description.
* `metadata` -  Key value map of metadata.

## Retrieving data source by filter

(supported in provider *v2.8+*)

A catalog item can be retrieved using a filter.
The criteria should be specfic enough to retrieve a single item. The retrieval will fail if the criteria match more than
one item.

When you don't know the name, you may get the data source using the `filter` section, and using one or more of the following criteria:

* `name_regex` (Optional) matches the name using a regular expression.
* `date` (Optional) is an expression containing an operator (`>`, `<`, `>=`, `<=`, `==`) and a date. Several formats 
  are recognized, but one of `yyyy-mm-dd [hh:mm[:ss]]` or `dd-MMM-yyyy [hh:mm[:ss]]` is recommended.
* `latest` (Optional) If `true`, retrieve the latest item among the ones matching other parameters. If no other parameters
  are set, it retrieves the newest item.
* `metadata` (Optional) One or more parameters that will match metadata contents, as defined below

### Metadata filter arguments

* `key` (Required) The name of the metadata field
* `value` (Required) The value to look for. It is treated as a regular expression.
* `is_system` (Optional) If `true`, the metadata fields will be passed as `metadata@SYSTEM:fieldName`.

## Example filter 1

```hcl
data "vcd_catalog_item" "mystery" {

  org     = "datacloud"
  catalog = "cat-datacloud"
  
  filter {
    name_regex = "^p.*11$"
  }
}

output "filtered_item" {
    value = data.vcd_catalog_item.mystery
}
```

Will find a catalog item with name staring with `p` and ending with `11`.
It fails if there are several items named as requested, such as `photon-v11`, `platform911`, or `poorName211`
Note that regular expressions are case sensitive: `photon-v11` and `Photon-v11` are two different entities.

## Example filter 2

```hcl
data "vcd_catalog_item" "mystery" {

  org     = "datacloud"
  catalog = "cat-datacloud"
  
  filter {
    name_regex = "^CentOS"
    latest     = true
  }
}
```

Will find the most recent item where the name starts by `CentOS`.

## Example filter 3

```hcl
data "vcd_catalog_item" "mystery" {

  org     = "datacloud"
  catalog = "cat-datacloud"
  
  filter {
    date = ">= 2020-03-20"
    latest     = true
  }
}
# Alternative date conditions for the same value:
# date = ">= March 23rd, 2020"
# date = ">= 22-Mar-2020"
# date = ">= 22-Mar-2020 00:00:00"
# date = ">= 2020-03-23 00:00:00"
```

Will find the most recent item created on or after March 23rd, 2020

## Example filter 4

```hcl
data "vcd_catalog_item" "mystery" {

  org     = "datacloud"
  catalog = "cat-datacloud"

  filter {
    metadata {
     key       = "ONE"
     value     = "FirstValue"  # explicit value
    }
    metadata {
     key       = "TWO"
     value     = "^S\\w+$"    # regular expression
    }
  }
}
```

Finds an item where the metadata contains a key `ONE` with value `FirstValue`, and a key `TWO` with value to be a single
alphanumeric word starting with `S`. Will fail if the criteria match more than one item.

Note that the `value` is always treated as a regular expression. For example:
`value = "cloud"` will match a metadata value `cloud`, but also one containing `on clouds` or `cloud9`.
To match only `cloud`, the value should be specified as `"^cloud$"`.
