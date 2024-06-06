---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_multisite_site_data"
sidebar_current: "docs-vcd-data-source-multisite-site-data"
description: |-
  Provides a data source to read a VMware Cloud Director Site association data to be used for association with another site.
---

# vcd\_multisite\_site\_data

Provides a data source to read a VMware Cloud Director Site association data to be used for association with another site.

Supported in provider *v3.13+*

## Example Usage 

Note: there is only one site available for each VCD. No ID or name is necessary to identify it.

~> Note: this data source requires System Administrator privileges

```hcl
data "vcd_multisite_site_data" "current_site" {
  download_to_file = "filename.xml"
}
```

## Argument Reference

* `download_to_file` - (Optional) Name of the file that will contain the data needed to associate this site to a remote one.
  Contains the same data returned in `association_data`.

## Attribute Reference

* `association_data` - The data needed to associate this site to another one. Contains the same data that would be saved into
  the file defined in `download_to_file`.