---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_org"
sidebar_current: "docs-vcd-data-source-org"
description: |-
  Provides an organization data source.
---

# vcd\_org

Provides a VMware Cloud Director Org data source. An organization can be used to manage catalogs, virtual
data centers, and users.

Supported in provider *v2.5+*

## Example Usage

```hcl
data "vcd_org" "my-org" {
  name = "my-org"
}

resource "vcd_org" "my-org-clone" {
  name                 = "my-org-clone"
  full_name            = data.vcd_org.my-org.full_name
  can_publish_catalogs = data.vcd_org.my-org.can_publish_catalogs
  deployed_vm_quota    = data.vcd_org.my-org.deployed_vm_quota
  stored_vm_quota      = data.vcd_org.my-org.stored_vm_quota
  is_enabled           = data.vcd_org.my-org.is_enabled
  delete_force         = true
  delete_recursive     = true
  vapp_lease {
    maximum_runtime_lease_in_sec          = data.vcd_org.my-org.vapp_lease.0.maximum_runtime_lease_in_sec
    power_off_on_runtime_lease_expiration = data.vcd_org.my-org.vapp_lease.0.power_off_on_runtime_lease_expiration
    maximum_storage_lease_in_sec          = data.vcd_org.my-org.vapp_lease.0.maximum_storage_lease_in_sec
    delete_on_storage_lease_expiration    = data.vcd_org.my-org.vapp_lease.0.delete_on_storage_lease_expiration
  }
  vapp_template_lease {
    maximum_storage_lease_in_sec       = data.vcd_org.my-org.vapp_template_lease.0.maximum_storage_lease_in_sec
    delete_on_storage_lease_expiration = data.vcd_org.my-org.vapp_template_lease.0.delete_on_storage_lease_expiration
  }
}

```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Org name

## Attribute Reference

* `full_name` - Org full name
* `is_enabled` - True if this organization is enabled (allows login and all other operations).
* `description` - Org description.
* `deployed_vm_quota` - Maximum number of virtual machines that can be deployed simultaneously by a member of this organization.
* `stored_vm_quota` - Maximum number of virtual machines in vApps or vApp templates that can be stored in an undeployed state by a member of this organization.
* `can_publish_catalogs` - True if this organization is allowed to share catalogs.
* `can_publish_external_catalogs` - (*v3.6+*) True if this organization is allowed to publish external catalogs.
* `can_subscribe_external_catalogs` - (*v3.6+*) True if this organization is allowed to subscribe to external catalogs.
* `delay_after_power_on_seconds` - Specifies this organization's default for virtual machine boot delay after power on.
* `metadata` - (Deprecated; *v3.6+*) Use `metadata_entry` instead. Key value map of metadata assigned to this organization.
* `metadata_entry` - A set of metadata entries assigned to the organization. See [Metadata](#metadata) section for details.
* `vapp_lease` - (*v2.7+*) Defines lease parameters for vApps created in this organization. See [vApp Lease](#vapp-lease) below for details. 
* `vapp_template_lease` - (*v2.7+*) Defines lease parameters for vApp templates created in this organization. See [vApp Template Lease](#vapp-template-lease) below for details.
* `number_of_catalogs` - (*v3.11+*) Number of catalogs owned or shared, available to this organization.
* `list_of_catalogs` - (*v3.11+*) List of catalogs (sorted alphabetically), owned or shared, available to this organization.
* `number_of_vdcs` - (*v3.11+*) Number of VDCs owned or shared, available to this organization.
* `list_of_vdcs` - (*v3.11+*) List of VDCs (sorted alphabetically), owned or shared, available to this organization.
* `account_lockout` - (*v3.14+*) Contains the account lockout properties of the read organization:
    * `enabled` - (*v3.14+*) Whether account lockout is enabled or not
    * `invalid_logins_before_lockout` - (*v3.14+*) Number of login attempts that will trigger an account lockout for the given user
    * `lockout_interval_minutes` - (*v3.14+*) Once a user is locked out, they will not be able to log back in for this time period

<a id="vapp-lease"></a>
## vApp Lease

The `vapp_lease` section contains lease parameters for vApps created in the current organization, as defined below:

* `maximum_runtime_lease_in_sec` - How long vApps can run before they are automatically stopped (in seconds)
* `power_off_on_runtime_lease_expiration` - When true, vApps are powered off when the runtime lease expires. When false, vApps are suspended when the runtime lease expires.
* `maximum_storage_lease_in_sec` - How long stopped vApps are available before being automatically cleaned up (in seconds)
* `delete_on_storage_lease_expiration` - If true, storage for a vApp is deleted when the vApp's lease expires. If false, the storage is flagged for deletion, but not deleted.

<a id="vapp-template-lease"></a>
## vApp Template Lease

The `vapp_template_lease` section contains lease parameters for vApp templates created in the current organization, as defined below:

* `maximum_storage_lease_in_sec` - How long vApp templates are available before being automatically cleaned up (in seconds)
* `delete_on_storage_lease_expiration` - If true, storage for a vAppTemplate is deleted when the vAppTemplate lease expires. If false, the storage is flagged for deletion, but not deleted

<a id="metadata"></a>
## Metadata

The `metadata_entry` (*v3.8+*) is a set of metadata entries that have the following structure:

* `key` - Key of this metadata entry.
* `value` - Value of this metadata entry.
* `type` - Type of this metadata entry. One of: `MetadataStringValue`, `MetadataNumberValue`, `MetadataDateTimeValue`, `MetadataBooleanValue`.
* `user_access` - User access level for this metadata entry. One of: `PRIVATE` (hidden), `READONLY` (read only), `READWRITE` (read/write).
* `is_system` - Domain for this metadata entry. true if it belongs to `SYSTEM`, false if it belongs to `GENERAL`.
