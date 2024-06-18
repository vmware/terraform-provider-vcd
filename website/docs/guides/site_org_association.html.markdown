---
layout: "vcd"
page_title: "VMware Cloud Director: Site and Org Association"
sidebar_current: "docs-vcd-guides-associations"
description: |-
 Provides guidance to VMware Cloud Director Site and Org Associations
---

# Site and Org association

Supported in provider *v3.13+*.

## Overview

Association between sites and between organizations follows the same pattern:

1. **Data collection** The administrator of each site or organization produces an **association file** (in XML format).
2. **File exchange** The file is given to the administrator on the other side.
3. **Association creation** Each administrator takes the received file and establishes an association.
4. When both sides have completed the above three steps, the association is complete.

VMware Cloud Director provider v3.13+ supplies data sources and resources to run operation 1 (data collection) and 3 
(association creation). The file exchange is not an operation directly supported by the provider, although in some
scenarios, when the same user has access to both sites or both organizations, all operations can run through Terraform
resources.
The Org association can be established between organizations in the same VCD or in a different one. When the associated
organization is remote (belongs to a different VCD), we need to establish an association between the two VCD (sites)
before we can associate the organizations.

## Associations are binary

Each association can only have two members: local site and remote site (or local organization and remote organization).
However, each site or organization can establish several associations, each of which will only have two members.
For example, if we want to associate Organization A with organizations B, C, and D, we need to establish three associations:
A-B, A-C, A-D. In this scenario, organization A will share resources with B, C, and D, but B, C, and D will not be associated
among each other, unless they establish their own associations.

## Workflow multi-admin for a site association

In this example, we will see the steps to associate `site1` with `site2`, assuming that the two sites are administered by
different persons.

1. _Site1_ admin produces the data using `vcd_multisite_site_data`, resulting in `site1.xml`.
2. _Site2_ admin produces the data using `vcd_multisite_site_data`, resulting in `site2.xml`.
3. _Site1_ admin transfers `site1.xml` to _Site2_ admin.
4. _Site2_ admin transfers `site2.xml` to _Site1_ admin.
5. _Site1_ admin runs resource `vcd_multisite_site_association` using `site2.xml` as input.
6. _Site2_ admin runs resource `vcd_multisite_site_association` using `site1.xml` as input.

## Workflow single-admin for a site association

In this example, we will see the steps to associate `site1` with `site2`, assuming that the two sites are administered by
the same person.

1. Common admin produces the data using `vcd_multisite_site_data`, resulting in `site1.xml`.
2. Common admin produces the data using `vcd_multisite_site_data`, resulting in `site2.xml`.
3. `site1.xml` and `site2.xml` are converted to Terraform `local_file` data sources.
4. Common admin runs resource `vcd_multisite_site_association` using `site2.xml` `local_file` data source as input.
5. Common admin runs resource `vcd_multisite_site_association` using `site1.xml` `local_file` data source as input.

See a full example for this workflow at [examples/site-all-at-once][site-all-at-once]

## Workflow multi-admin for an organization association

In this example, we will see the steps to associate `org1` with `org2`, assuming that the two organizations are administered by
different persons.

1. _Org1_ admin produces the data using `vcd_multisite_org_data`, resulting in `org1.xml`.
2. _Org2_ admin produces the data using `vcd_multisite_org_data`, resulting in `org2.xml`.
3. _Org1_ admin transfers `org1.xml` to _Org2_ admin.
4. _Org2_ admin transfers `org2.xml` to _Org1_ admin.
5. _Org1_ admin runs resource `vcd_multisite_org_association` using `org2.xml` as input.
6. _Org2_ admin runs resource `vcd_multisite_org_association` using `org1.xml` as input.

## Workflow single-admin for an organization association

In this example, we will see the steps to associate `org1` with `org2`, assuming that the two sites are administered by
the same person (the system administrator).

1. Common admin produces the data using `vcd_multisite_org_data`.
2. Common admin produces the data using `vcd_multisite_org_data`.
3. Common admin runs resource `vcd_multisite_org_association` using the field `association_data` from data source `vcd_multisite_site_data.org2` as input.
4. Common admin runs resource `vcd_multisite_org_association` using the field `association_data` from data source `vcd_multisite_site_data.org2` as input.

See a full example for this workflow at [examples/org-all-at-once][org-all-at-once]

## Data collection

### Data collection for a Site

Each VCD has only one site. No names or ID are needed to identify it. The data source that performs the data collection
is `vcd_multisite_site_data`:

```hcl
data "vcd_multisite_site_data" "site1" {
  download_to_file = "site1.xml"
}
```

Even if `download_to_file` is not specified, the data is still available in the field `association_data`. 

### Data collection for an organization

Each organization can have only one data collection entity. The only identification needed is the organization ID.

```hcl
data "vcd_org" "my_org" {
  name = "my-org"
}

data "vcd_multisite_org_data" "org1" {
  org_id           = data.vcd_org.my_org.id
  download_to_file = "org1.xml"
}
```

Even if `download_to_file` is not specified, the data is still available in the field `association_data`.

## Association scenarios

### Associating two sites

To associate your site with another site, you need to have the association data available, either as an XML file or as
a text string. If you don't own both sites, you should receive the data from the other site administrator, who will
get the data using [data collection](#data-collection-for-a-site).

Using the XML file (`site2.xml`) received from the other administrator, you can run the following:

```hcl
# As administrator of site1
resource "vcd_multisite_site_association" "site1-site2" {
  association_data_file   = "site2.xml"
  connection_timeout_mins = 2
}
```

This operation will establish one side of the association between site1 and site2. At this stage, if no other operation
is performed, the association is in state `ASYMMETRIC`, meaning that it has been initiated, but its counterpart has not been
received yet.

For the association to be completed, the other administrator must perform the same operation, using the XML data file for site1.

```hcl
# As administrator of site2
resource "vcd_multisite_site_association" "site2-site1" {
  association_data_file   = "site1.xml"
  connection_timeout_mins = 2
}
```

### Associating two organizations

To associate your organization with another one, you need to have the association data available, either as an XML file or as
a text string. If you don't own both organizations, you should receive the data from the other organization administrator, who will
get the data using [data collection](#data-collection-for-an-organization).

Using the XML file (`org2.xml`) received from the other administrator, you can run the following:

```hcl
# As administrator of org1
resource "vcd_multisite_org_association" "org1-org2" {
  association_data_file   = "org2.xml"
  connection_timeout_mins = 2
}
```

This operation will establish one side of the association between org1 and org2. At this stage, if no other operation
is performed, the association is in state `ASYMMETRIC`, meaning that it has been initiated, but its counterpart has not been
received yet.

For the association to be completed, the other administrator must perform the same operation, using the XML data file for org1.

```hcl
# As administrator of org2
resource "vcd_multisite_org_association" "org2-org1" {
  association_data_file   = "org1.xml"
  connection_timeout_mins = 2
}
```

### Checking the association completion

When both sides of the association operation have been performed (in both `vcd_multisite_site_association` or 
`vcd_multisite_org_association`), you can run `terraform apply` once more. This _update_ operation will use the
field `connection_timeout_mins` to check the association status. 
The expression `connection_timeout_mins = 2` means "check for up to
two minutes whether the status of the connection is `ACTIVE`". If the association reaches the desired status within the
intended timeout, all is well, and the association is ready to be used. If it fails, it means that either the connection
at the other side was not executed, or that there is some communication problem.

Note about `connection_timeout_mins`: 
1. You must not run this check before both sides of the association have run. If you run it with only one side, it will
  fail, as the status cannot be `ACTIVE`.
2. The property `connection_timeout_mins` is only evaluated during an _update_. It is safe to have it in the script at
  creation, as it will be ignored during that operation. 

## Listing associations

### Listing site associations

You can use one of the two methods below:

1. Run data source `vcd_resource_list` with `resource_type = "vcd_multisite_site_association"`
2. Run data source `vcd_multisite_site`: it will show the number and the name of site associations.

```hcl
data "vcd_multisite_site" "sites" {
}

data "vcd_resource_list" "sites" {
  name          = "sites"
  resource_type = "vcd_multisite_site_association"
}
```

### Listing organization associations

You can use one of the two methods below:

1. Run data source `vcd_resource_list` with `resource_type = "vcd_multisite_org_association"`
2. Run data source `vcd_multisite_org_data`: it will show the number and the name of org associations.

```hcl
data "vcd_org" "my_org" {
  name = "my-org"
}

data "vcd_multisite_org_data" "orgs" {
  org_id = data.vcd_org.my_org.id
}

data "vcd_resource_list" "orgs" {
  name          = "orgs"
  resource_type = "vcd_multisite_org_association"
}
```

### Using site and organization association data sources

To read an association data source (`vcd_multisite_site_association` or `vcd_multisite_org_association`) you need to
provide the ID of the remote entity (`associated_site_id` or `associated_org_id`). This information is usually found in
the association data file used to create the association (e.g. `site2.xml`).  To make the data source retrieval easier,
you can either supply the associated entity ID, or the XML file used to create the association.
In the examples below, the two data sources are equivalent:

```hcl
data "vcd_multisite_site_association" "site1-site2a" {
  associated_site_id = "urn:vcloud:site:deadbeef-fcf3-414a-be95-a3e26cf1296b"
}

data "vcd_multisite_site_association" "site1-site2b" {
  association_data_file = "site2.xml"
}
```


[site-all-at-once]:https://github.com/dataclouder/terraform-provider-vcd/tree/site-org-associations/examples/multi-site/site-all-at-once
<!-- TODO: After merge, change to https://github.com/vmware/terraform-provider-vcd/tree/main/examples/multi-site/site-all-at-once -->
[org-all-at-once]:https://github.com/dataclouder/terraform-provider-vcd/tree/site-org-associations/examples/multi-site/org-all-at-once
<!-- TODO: After merge, change to https://github.com/vmware/terraform-provider-vcd/tree/main/examples/multi-site/org-all-at-once -->
