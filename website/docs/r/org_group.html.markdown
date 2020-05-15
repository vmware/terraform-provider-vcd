---
layout: "vcd"
page_title: "vCloudDirector: vcd_org_group"
sidebar_current: "docs-vcd-resource-org-group"
description: |-
  Provides a vCloud Director Organization group. This can be used to create, update, and delete organization groups.
---

# vcd\_org\_group

Provides a vCloud Director Organization group. This can be used to create, update, and delete organization groups.

Supported in provider *v2.9+*

~> **Note:** Only `System Administrator` or `Org Administrator` users can create groups.

## Example Usage

```hcl
resource "vcd_org_group" "org1" {
  org  = "org1"
  name = "Org1-AdminGroup"
  role = "Organization Administrator"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which the VDC belongs. Optional if defined at provider level.
* `name` - (Required) A unique name for the group.
* `provider_type` - (Optional) Identity provider type for this this group. Only `SAML` is supported
  at this time and it will default to it.
* `role` - (Required) The role of the user. Role names can be retrieved from the organization. Both built-in roles and
  custom built can be used. The roles normally available are:
    * `Organization Administrator`
    * `Catalog Author`
    * `vApp Author`
    * `vApp User`
    * `Console Access Only`
    * `Defer to Identity Provider`

## Attribute Reference

The following attributes are exported on this resource:

* `id` - The ID of the Organization group
* `description` - The description of Organization group

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing user can be [imported][docs-import] into this resource via supplying the full dot separated path for an
org user. For example, using this structure, representing an existing user that was **not** created using Terraform:

```hcl
resource "vcd_org_group" "my-admin-group" {
  org  = "my-org"
  name = "my-admin-group"
  role = "Organization Administrator"
}
```

You can import such user into terraform state using this command

```
terraform import vcd_org_group.my-admin-group my-org.my-admin-group
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/
