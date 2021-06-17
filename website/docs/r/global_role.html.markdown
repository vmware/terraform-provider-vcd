---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_global_role"
sidebar_current: "docs-vcd-resource-global-role"
description: |-
 Provides a VMware Cloud Director global role. This can be used to create, modify, and delete global roles.
---

# vcd\_global\_role

Provides a VMware Cloud Director global role. This can be used to create, modify, and delete global roles.

Supported in provider *v3.3+*

## Example Usage

```hcl
resource "vcd_global_role" "new-global-role" {
  name        = "new-global-role"
  description = "new global role from CLI"
  rights = [
    "Catalog: Add vApp from My Cloud",
    "Catalog: Edit Properties",
    "Catalog: View Private and Shared Catalogs",
    "Organization vDC Compute Policy: View",
    "vApp Template / Media: Edit",
    "vApp Template / Media: View",
  ]
  publish_to_all_tenants = false
  tenants = [
    "org1",
    "org2",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the global role.
* `description` - (Required) A description of the global role
* `rights` - (Optional) List of rights assigned to this role
* `publish_to_all_tenants` - (Required) When true, publishes the global role to all tenants
* `tenants` - (Optional) List of tenants to which this global role gets published. Ignored if `publish_to_all_tenants` is true.

## Attribute Reference

* `read_only` - Whether this global role is read-only
* `bundle_key` - Key used for internationalization

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing global role can be [imported][docs-import] into this resource via supplying the global role name (the global
role is at the top of the entity hierarchy).
For example, using this structure, representing an existing global role that was **not** created using Terraform:

```hcl
resource "vcd_global_role" "catalog-author" {
  name = "Catalog Author"
}
```

You can import such global role into terraform state using this command

```
terraform import vcd_global_role.catalog-author "Catalog Author"
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the global role as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the global role's stored properties.

## More information

See [Roles management](/docs/providers/vcd/guides/roles_management.html) for a broader description of how global roles and
rights work together.
