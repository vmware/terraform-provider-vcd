---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_multisite_site_association"
sidebar_current: "docs-vcd-data-source-multisite-site-association"
description: |-
  Provides a data source to read a VMware Cloud Director site association with the current site.
---

# vcd\_multisite\_site\_association

Provides a data source to read a VMware Cloud Director site association information.

~> Note: this data source requires System Administrator privileges

Supported in provider *v3.13+*

## Example Usage 1

Retrieving a site association using the associated site ID.

```hcl
data "vcd_multisite_site_association" "site1-site2" {
  associated_site_id = "urn:vcloud:site:dca02216-fcf3-414a-be95-a3e26cf1296b"
}
```

## Example Usage 2

Retrieving a site association using the association data file.

```hcl
data "vcd_multisite_site_association" "site1-site2" {
  association_data_file = "remote-site.xml"
}
```

## Argument Reference

* `association_data_file` - (Optional) Name of the file containing the data used to associate this site to another one.
  (Used when `associated_site_id` is not known)
* `associated_site_id` - (Optional) ID of the remote site associated with the current one. (Used in alternative to
  `associated_data_file`)


## Attribute Reference

* `associated_site_name` - The name of the associated site.
* `associated_site_href` - The URL of the associated site.
* `status` - The status of the association (one of `ASYMMETRIC`, `ACTIVE`, `UNREACHABLE`, `ERROR`)
