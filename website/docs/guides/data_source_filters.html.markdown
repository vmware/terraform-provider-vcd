---
layout: "vcd"
page_title: "vCloudDirector: data source filters"
sidebar_current: "docs-vcd-guides-filters"
description: |-
  Provides guidance on filters.
---

(supported in provider *v2.9+*)

## Retrieving a data source by filter

Several data sources can be retrieved using a filter.
The criteria should be specific enough to retrieve a single item. The retrieval will fail if the criteria match more than
one item.

When you don't know the name, you may get the data source using the `filter` section, and using one or more of the following criteria:

* `name_regex` (Optional) matches the name using a regular expression.
* `date` (Optional) is an expression containing an operator (`>`, `<`, `>=`, `<=`, `==`) and a date. Several formats 
  are recognized, but one of `yyyy-mm-dd [hh:mm[:ss]]` or `dd-MMM-yyyy [hh:mm[:ss]]` is recommended.
* `latest` (Optional) If `true`, retrieve the latest item among the ones matching other parameters. If no other parameters
  are set, it retrieves the newest item.
* `earliest` (Optional) If `true`, retrieve the earliest item among the ones matching other parameters. If no other parameters
  are set, it retrieves the oldest item.
* `ip` (Optional) matches the IP of the resource using a regular expression.
* `metadata` (Optional) One or more parameters that will match metadata contents, as defined below

### Metadata filter arguments

* `key` (Required) The name of the metadata field
* `value` (Required) The value to look for. It is treated as a regular expression if `use_api_search` is not set.
* `is_system` (Optional) If `true`, the metadata fields will be passed as `metadata@SYSTEM:fieldName`.
* `use_api_search` (Optional) If true, the search happens using the API query for Metadata, without using regular
   expressions. It is slightly faster than using the search by regular expression, but when this is set, the type
   field is mandatory.
* `type` (Optional) One of `STRING`, `INT`, `BOOL`. It is required when `use_api_search` is set.

### Availability of filters

Not all the data sources support filters, and when they do, they may not support all the search fields. For example,
and edge gateway only supports `name_regex`, while the `vcd_network_*` support the full set.

### About metadata search

Metadata can be searched even for those data sources that don't expose metadata in their interface. If the `filter`
section lists `metadata` among the available criteria, you can search the metadata and get the results accordingly,
although the vcd provider may not show the metadata for the found item.
Note that the names of the metadata fields are case sensitive.

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

Will find a catalog item with name starting with `p` and ending with `11`.
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
    date   = ">= 2020-03-20"
    latest = true
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
     key   = "Code"
     value = "BlackEagle"
    }
  }
}
```

Finds an item where the metadata contains a key `Code` with value `BlackEafle`.
Will fail if the criteria match more than one item.

## Example filter 5

```hcl
data "vcd_catalog_item" "mystery" {

  org     = "datacloud"
  catalog = "cat-datacloud"

  filter {
    metadata {
     key            = "ONE"
     value          = "FirstValue"  # explicit value
     use_api_search = "true"
     type           = "STRING"
    }
    metadata {
     key            = "TWO"
     value          = "SecondValue" # explicit value
     use_api_search = "true"
     type           = "STRING"
    }
  }
}
```

You can use several `metadata` blocks. This example finds an item where the metadata contains a key `ONE` with
value `FirstValue`, and a key `TWO` with value `SecondValue`.
Will fail if the criteria match more than one item.

## Example filter 6

```hcl
data "vcd_catalog_item" "mystery" {

  org     = "datacloud"
  catalog = "cat-datacloud"

  filter {
    metadata {
     key   = "ONE"
     value = "^First"  # regular expression
    }
    metadata {
     key   = "TWO"
     value = "^S\\w+$"    # regular expression
    }
  }
}
```

Will perform the same search of example 5, using regular expressions instead of exact values.


Note that the `value` is treated as a regular expression when `use_api_search` is false. For example:
`value = "cloud"` will match a metadata value `cloud`, but also one containing `on clouds` or `cloud9`.
To match only `cloud`, the value should be specified as `"^cloud$"`.


## Example filter 7


Several data sources with a quick search

```hcl
# Finds the oldest catalog created after April 2nd, 2020
data "vcd_catalog" "mystery" {
  org = "datacloud"

  filter {
    date     = ">= 2020-04-02 10:00"
    earliest = "true"
  }
}

# Finds an isolated network with gateway IP starting with `192.168.3`
data "vcd_network_isolated" "mystery" {
  org = "datacloud"
  vdc = "vdc-datacloud"

  filter {
    ip = "^192.168.3"
  }
}

# Finds an edge gateway with name starting with `gw` and ending with `191`
data "vcd_edgegateway" "mystery" {
  org = "datacloud"
  vdc = "vdc-datacloud"

  filter {
    name_regex = "^gw.+191"
  }
}

# Finds the newest media item created after March 1st, 2020
data "vcd_catalog_media" "mystery" {
  org     = "datacloud"
  catalog = "cat-datacloud"

  filter {
    date   = "> 2020-03-01"
    latest = true
  }
}
```