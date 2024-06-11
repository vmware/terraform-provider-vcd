---
layout: "vcd"
page_title: "VMware Cloud Director: Site and Org Association"
sidebar_current: "docs-vcd-guides-associations"
description: |-
 Provides guidance to VMware Cloud Site and Org Associations
---

WIP

# Site and Org association

Supported in provider *v3.13+*.

-> In this document, when we mention **tenants**, the term can be substituted with **organizations**.

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

See a full example for this workflow at https://github.com/dataclouder/terraform-provider-vcd/tree/site-org-associations/examples/multi-site/site-all-at-once
<!-- TODO: After merge, change to https://github.com/vmware/terraform-provider-vcd/tree/main/examples/multi-site/site-all-at-once -->

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

See a full example for this workflow at https://github.com/dataclouder/terraform-provider-vcd/tree/site-org-associations/examples/multi-site/org-all-at-once
<!-- TODO: After merge, change to https://github.com/vmware/terraform-provider-vcd/tree/main/examples/multi-site/org-all-at-once -->

## Data collection

### Data collection for a Site 

Each VCD has only one site. No names or ID are needed to identify it. The data source that performs the data collection
is `vcd_multisite_site_data`:

```hcl
data "vcd_multisite_site_data" "site1" {
  download_to_file = "site1.xml"
}
```
