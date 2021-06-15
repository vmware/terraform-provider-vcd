---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_global_role"
sidebar_current: "docs-vcd-data-source-global-role"
description: |-
 Provides a VMware Cloud Director global role data source . This can be used to read global roles.
---

# vcd\_global\_role

Provides a VMware Cloud Director global role data source. This can be used to read global roles.

Supported in provider *v3.3+*

## Example Usage

```hcl
data "vcd_global_role" "vapp-author" {
  name = "vApp Author"
}
```

```
Sample output:

global-role-vapp = {
  "bundle_key" = "ROLE_VAPP_AUTHOR"
  "description" = "Rights given to a user who uses catalogs and creates vApps"
  "id" = "urn:vcloud:globalRole:1bf4457f-a253-3cf1-b163-f319f1a31802"
  "name" = "vApp Author"
  "publish_to_all_tenants" = true
  "read_only" = false
  "rights" = toset([
    "Catalog: Add vApp from My Cloud",
    "Catalog: View Private and Shared Catalogs",
    "Organization vDC Compute Policy: View",
    "Organization vDC Named Disk: Create",
    "Organization vDC Named Disk: Delete",
    "Organization vDC Named Disk: Edit Properties",
    "Organization vDC Named Disk: View Encryption Status",
    "Organization vDC Named Disk: View Properties",
    "Organization vDC Network: View Properties",
    "Organization vDC: VM-VM Affinity Edit",
    "Organization: View",
    "UI Plugins: View",
    "VAPP_VM_METADATA_TO_VCENTER",
    "vApp Template / Media: Copy",
    "vApp Template / Media: Edit",
    "vApp Template / Media: View",
    "vApp Template: Checkout",
    "vApp: Copy",
    "vApp: Create / Reconfigure",
    "vApp: Delete",
    "vApp: Download",
    "vApp: Edit Properties",
    "vApp: Edit VM CPU",
    "vApp: Edit VM Compute Policy",
    "vApp: Edit VM Hard Disk",
    "vApp: Edit VM Memory",
    "vApp: Edit VM Network",
    "vApp: Edit VM Properties",
    "vApp: Manage VM Password Settings",
    "vApp: Power Operations",
    "vApp: Sharing",
    "vApp: Snapshot Operations",
    "vApp: Upload",
    "vApp: Use Console",
    "vApp: VM Boot Options",
    "vApp: View ACL",
    "vApp: View VM and VM's Disks Encryption Status",
    "vApp: View VM metrics",
  ])
  "tenants" = toset([
    "org1",
    "org2",
  ])
}
```


## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the global role.

## Attribute Reference

* `description` - A description of the global role
* `bundle_key` - Key used for rights bundles. Default "com.vmware.vcloud.undefined.key"
* `rights` - List of rights assigned to this role
* `publish_to_all_tenants` - When true, publishes the global role to all tenants
* `tenants` - List of tenants to which this global role gets published. Ignored if `publish_to_all_tenants` is true.
* `read_only` - Whether this global role is read-only

## More information

See [Roles management](/docs/providers/vcd/guides/roles_management.html) for a broader description of how global roles and
rights work together.
