---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_catalog_access_control"
sidebar_current: "docs-vcd-resource-catalog-access-control"
description: |-
  Provides a VMware Cloud Director Access Control structure for a Catalog.
---

# vcd\_catalog\_access\_control

Provides a VMware Cloud Director Access Control structure for a Catalog. This can be used to create, update, and delete access control structures for a Catalog.

~> **Warning:** The access control info is tied to a Catalog. Thus, there could be only one instance per catalog. Using a different
definition for the same Catalog ID will result in a previous instance to be overwritten.

-> **Note:** Access control operations run in tenant context, meaning that, even if the user is a system administrator,
in every request it uses headers items that define the tenant context as restricted to the organization to which the Catalog belongs.

Supported in provider *v3.8+*

## Example Usage 1

```hcl
data "vcd_org" "another-org" {
  name = "another-org"
}

data "vcd_org_user" "ac-admin1" {
  org  = "this-org"
  name = "ac-admin1"
}

data "vcd_org_user" "ac-vapp-creator2" {
  org  = "this-org"
  name = "ac-vapp-creator2"
}

data "vcd_catalog" "Catalog-AC-0" {
  name = "Catalog-AC-0"
}

data "vcd_catalog" "Catalog-AC-1" {
  name = "Catalog-AC-1"
}

data "vcd_catalog" "Catalog-AC-2" {
  name = "Catalog-AC-2"
}

resource "vcd_catalog_access_control" "AC-not-shared" {
  catalog_id = data.vcd_catalog.Catalog-AC-0.id

  shared_with_everyone = false
}

resource "vcd_catalog_access_control" "AC-global" {
  catalog_id = data.vcd_catalog.Catalog-AC-1.id

  shared_with_everyone  = true
  everyone_access_level = "ReadOnly"
}

resource "vcd_catalog_access_control" "AC-users-and-orgs" {
  catalog_id = data.vcd_catalog.Catalog-AC-2.id

  shared_with_everyone = false

  shared_with {
    user_id      = data.vcd_org_user.ac-admin1.id
    access_level = "FullControl"
  }
  shared_with {
    user_id      = data.vcd_org_user.ac-vapp-creator2.id
    access_level = "Change"
  }
  shared_with {
    org_id       = data.vcd_org.another-org.id
    access_level = "ReadOnly"
  }
}
```

## Example Usage 2

```hcl
resource "vcd_catalog_access_control" "ac-other-orgs" {
  org        = "datacloud"
  catalog_id = vcd_catalog.Test-Catalog-AC-5.id

  shared_with_everyone           = false
  read_only_shared_with_all_orgs = true
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which the Catalog belongs. Optional if defined at provider level.
* `catalog_id` - (Required) A unique identifier for the Catalog.
* `shared_with_everyone` - (Required) Whether the Catalog is shared with everyone. If any `shared_with` blocks are included,
  this property must be set to `false`.
* `everyone_access_level` - (Optional) Access level when the Catalog is shared with everyone (it can only be set to
  `ReadOnly`). Required if `shared_with_everyone` is set.
* `read_only_shared_with_all_orgs` - (Optional; *v3.11+*) If true, the catalog is shared as read-only with all organizations
* `shared_with` - (Optional) one or more blocks defining a subject (one of Organization, User, or Group) to which we are sharing. 
   See [shared_with](#shared_with) below for detail. It cannot be used if `shared_with_everyone` is true.


## shared_with

* `user_id` - (Optional) The ID of a user with which we are sharing. Required if `group_id` or `org_id` is not set.
* `group_id` - (Optional) The ID of a group with which we are sharing. Required if `user_id` or `org_id` is not set.
* `org_id` - (Optional) The ID of a group with which we are sharing. Required if `user_id` or `group_id` is not set.
* `access_level` - (Required) The access level for the user or group to which we are sharing. (One of `ReadOnly`, 
  `Change`, `FullControl`, but it can only be `ReadOnly` when we share to an Organization)
* `subject_name` - (Computed) the name of the subject (Org, group, or user) with which we are sharing.


## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing `vcd_catalog_access_control` can be [imported][docs-import] into this resource via supplying its full dot separated path.
For example, using this structure, representing an existing access control structure that was **not** created using Terraform:

```hcl
data "vcd_catalog" "my-cat" {
  org  = "my-org"
  name = "my-catalog"
}

resource "vcd_catalog_access_control" "my-ac" {
  org        = "my-org"
  catalog_id = data.vcd_catalog.my-cat.id
}
```

You can import such structure into terraform state using one of these commands

```
terraform import vcd_catalog_access_control.my-ac my-org.catalog-id
terraform import vcd_catalog_access_control.my-ac my-org.catalog-name
```

terraform will import the structure using either the catalog name or its ID.


NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the access control structure as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the access control stored properties.
