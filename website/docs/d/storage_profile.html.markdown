---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_storage_profile"
sidebar_current: "docs-vcd-data-source-storage-profile"
description: |-
  Provides a data source for VDC storage profile.
---

# vcd\_storage\_profile

Provides a data source for VDC storage profile.

Supported in provider *v3.1+*


## Example Usage

```hcl
data "vcd_storage_profile" "sp" {
  org  = "my-org"
  vdc  = "my-vdc"
  name = "ssd-storage-profile"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations.
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level.
* `name` - (Required) Storage profile name.

## Attribute reference
* `id` - storage profile identifier

(Supported in provider *v3.6+*)

* `limit` - Maximum number of storage bytes (scaled by 'units' field) allocated for this profile. `0` means `maximum possible`
* `used_storage` - Storage used, by the storage profile (in Megabytes)
* `default` - True if this is default storage profile for this VDC. The default storage profile is used when an object that can specify a storage profile is created with no storage profile specified
* `enabled` - True if this storage profile is enabled for use in the VDC
* `iops_allocated` - Total IOPS currently allocated to this storage profile
* `units` - Scale used to define Limit
* `iops_settings` - A block providing IOPS settings. See [IOPS settings](#iopsSettings) below for details.

<a id="iopsSettings"></a>
## IOPS settings

(Supported from VCD *v10.2*)

* `iops_limiting_enabled` - True if this storage profile is IOPS-based placement enabled
* `maximum_disk_iops` - The maximum IOPS value that this storage profile is permitted to deliver. Value of 0 means this max setting is disabled and there is no max disk IOPS restriction
* `default_disk_iops` - Value of 0 for disk IOPS means that no IOPS would be reserved or provisioned for that virtual disk
* `disk_iops_per_gb_max` - The maximum disk IOPs per GB value that this storage profile is permitted to deliver. A value of 0 means there is no per GB IOPS restriction
* `iops_limit` - Maximum number of IOPs that can be allocated for this profile. `0` means `maximum possible`