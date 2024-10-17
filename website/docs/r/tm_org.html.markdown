---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_tm_org"
sidebar_current: "docs-vcd-resource-tm-org"
description: |-
  Provides a resource to manage VMware Cloud Foundation Tenant Manager Organization.
---

# vcd\_nsxt\_tm\_org

Provides a resource to manage VMware Cloud Foundation Tenant Manager Organization.

~> Only `System Administrator` can create this resource.

## Example Usage

```hcl
resource "vcd_tm_org" "test" {
  name         = "terraform-org"
  display_name = "Terraform Organization"
  description  = "Created with Terraform"
  is_enabled   = true
}
```
## Example Usage (sub-provider)

```hcl
resource "vcd_tm_org" "test" {
  name           = "subprovider-org"
  display_name   = "Sub-provider Organization"
  is_enabled     = true
  is_subprovider = true
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for Organization with which users log in to this organization
* `display_name` - (Required) A human readable name for Organization
* `description` - (Optional) An optional description for Organization
* `is_enabled` - (Optional) Defines if Organization is enabled. Default `true`. **Note:**
  Organization has to be disabled before removal and this resource will automatically perform it
* `is_subprovider` - (Optional) Enables this Organization to manage other Organizations. **Note**:
  This value cannot updated as there may be any number of Rights Bundles granting sub-provider rights
  to this Org. Instead, unpublish any rights bundles that have the Org Traverse right from this org.
  This can be toggled to true to automatically perform the following steps:
 * Publish the Default Sub-Provider Entitlement Rights Bundle to the Organization
 * Publish the Sub-Provider Administrator global role (if it exists) to the Organization
 * Create a Default Rights Bundle in the Organization containing all publishable rights that are
   currently published to the Organization and mark that Rights Bundle as publish all.
 * Clone all default roles currently published to the Organization into Global Roles in the
   Organization and marks them all publish all.

## Attribute Reference

The following attributes are exported on this resource:

* `org_vdc_count` - Number of VDCs belonging to this Organization
* `catalog_count` - Number of catalogs belonging to this Organization
* `vapp_count` - Number of vApps belonging to this Organization
* `running_vm_count` - Number of running VMs belonging to this Organization
* `user_count` - Number of users belonging to this Organization
* `disk_count` - Number of disks belonging to this Organization
* `can_publish` - Defines if this Organization can publish catalogs externally
* `directly_managed_org_count` - Number of directly managed Organizations
* `is_classic_tenant` - Defines if this Organization is a classic VRA style tenant

## Importing

~> The current implementation of Terraform import can only import resources into the state. It does
not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing Org configuration can be [imported][docs-import] into this resource via supplying path
for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_tm_org.imported my-org-name
```

The above would import the `my-org-name` Organization settings.