---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_multisite_org_association"
sidebar_current: "docs-vcd-data-source-multisite-org-association"
description: |-
  Provides a data source to read a VMware Cloud Director Org associated with the current Org.
---

# vcd\_multisite\_org\_association

Provides a data source to read a VMware Cloud Director Org association information.

Supported in provider *v3.13+*

## Example Usage 1

Retrieving an Org association using the associated Org ID.

```hcl
data "vcd_org" "my-org" {
  name = "my-org"
}

data "vcd_multisite_org_association" "org1-org2" {
  org_id            = data.vcd_org.my-org.id
  associated_org_id = "urn:vcloud:org:3901d87d-1596-4a5a-a74b-57a7313737cf"
}
```

## Example Usage 2

Retrieving an Org association using the association data file.

```hcl
data "vcd_org" "my-org" {
  name = "my-org"
}

data "vcd_multisite_org_association" "org1-org2" {
  org_id                = data.vcd_org.my-org.id
  association_data_file = "remote-org.xml"
}
```

## Argument Reference

* `org_id` - (Required) The ID of the organization for which we need to collect the data.
* `association_data_file` - (Optional) Name of the file containing the data used to associate this Org to another one.
  (Used when `associated_org_id` is not known)
* `associated_org_id` - (Optional) ID of the remote organization associated with the current one. (Used in alternative to
  `associated_data_file`)


## Attribute Reference

* `associated_org_name` - The name of the associated Org.
* `associated_site_id` - The ID of the associated site.
* `status` - The status of the association (one of `ASYMMETRIC`, `ACTIVE`, `UNREACHABLE`, `ERROR`)

## More information

See [Site and Org association](/providers/vmware/vcd/latest/docs/guides/site_org_association) for a broader description
of association workflows.