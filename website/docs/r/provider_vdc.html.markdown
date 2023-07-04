---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_provider_vdc"
sidebar_current: "docs-vcd-resource-provider-vdc"
description: |-
  Provides a Provider VDC resource.
---

# vcd\_provider\_vdc

Gives a VMware Cloud Director Provider VDC resource. This resource can be used to create, modify, and delete a Provider VDC and use its 
data within other resources or data sources.

Supported in provider *v3.10+*

~> Note: this resource requires system administrator privileges

## Example Usage 1

```hcl
data "vcd_vcenter" "vcenter1" {
  name = "vc1"
}

data "vcd_resource_pool" "rp1" {
  name       = "resource-pool-for-vcd-01"
  vcenter_id = data.vcd_vcenter.vcenter1.id
  # maximum hardware version: "vmx-18"
}

data "vcd_resource_pool" "rp2" {
  name       = "resource-pool-for-vcd-01"
  vcenter_id = data.vcd_vcenter.vcenter1.id
  # maximum hardware version: "vmx-19"
}

data "vcd_nsxt_manager" "mgr1" {
  name = "nsxManager1"
}

data "vcd_network_pool" "np1" {
  name = "NSX-T Overlay 1"
}

resource "vcd_provider_vdc" "pvdc1" {
  name                               = "myPvdc"
  description                        = "new provider VDC"
  is_enabled                         = true
  vcenter_id                         = data.vcd_vcenter.vcenter1.id
  nsxt_manager_id                    = data.vcd_nsxt_manager.mgr1.id
  network_pool_ids                   = [data.vcd_network_pool.np1.id]
  resource_pool_ids                  = [data.vcd_resource_pool.rp1.id]
  storage_profile_names              = ["Development"]
  highest_supported_hardware_version = data.vcd_resource_pool.rp1.hardware_version # vmx-18
}
```

## Example Usage 2

You can update the provider VDC in [Example Usage 1](#example-usage-1) to use a higher hardware version by adding a new
resource pool that supports such version.

```hcl
resource "vcd_provider_vdc" "pvdc1" {
  name                               = "myPvdc"
  description                        = "new provider VDC"
  is_enabled                         = true
  vcenter_id                         = data.vcd_vcenter.vcenter1.id
  nsxt_manager_id                    = data.vcd_nsxt_manager.mgr1.id
  network_pool_ids                   = [data.vcd_network_pool.np1.id]
  resource_pool_ids                  = [data.vcd_resource_pool.rp1.id, data.vcd_resource_pool.rp2.id]
  storage_profile_names              = ["Development"]
  highest_supported_hardware_version = data.vcd_resource_pool.rp2.hardware_version # vmx-19
}
```

## Argument Reference

The following arguments are supported:
 
* `name` - (Required) Provider VDC name
* `description` - (Optional) description of the Provider VDC.
* `is_enabled` - (Optional) True if this Provider VDC is enabled and can provide resources to organization VDCs. A Provider VDC is always enabled on creation.
* `nsxt_manager_id` - (Required) ID of the registered NSX-T Manager that backs networking operations for this Provider VDC.
* `highest_supported_hardware_version` - (Required) The highest virtual hardware version supported by this Provider VDC. This value cannot be changed to a lower version, and can only be updated when adding a new resource pool.
* `storage_profile_names` - (Required) Set of Storage Profile names used to create this provider VDC.
* `resource_pool_ids` - (Required) Set of IDs of the Resource Pools backing this provider VDC. (Note: only one resource pool can be set at creation)
* `vcenter_id` - (Required) ID of the vCenter Server that provides the Resource Pools and Datastores.
* `network_pool_ids` - (Required) Set IDs of the Network Pools used by this Provider VDC.

## Attribute reference

* `status` - Status of the Provider VDC, it can be -1 (creation failed), 0 (not ready), 1 (ready), 2 (unknown) or 3 (unrecognized).
* `capabilities` - Set of virtual hardware versions supported by this Provider VDC.
* `compute_capacity` - An indicator of CPU and memory capacity. See [Compute Capacity](#compute-capacity) below for details.
* `compute_provider_scope` - Represents the compute fault domain for this Provider VDC. This value is a tenant-facing tag that is shown to tenants when viewing fault domains of the child Organization VDCs (for example, a VDC Group).
* `storage_containers_ids` - Set of IDs of the vSphere Datastores backing this Provider VDC.
* `external_network_ids` - Set of IDs of External Networks.
* `storage_profile_ids` - Set of IDs to the Storage Profiles available to this Provider VDC.
* `universal_network_pool_id` - ID of the universal network reference.
* `host_ids` - Set with all the hosts which are connected to VC server.
<!--
// metadata to be added soon
* `metadata` - (Deprecated) Use `metadata_entry` instead. Key and value pairs for Provider VDC Metadata.
* `metadata_entry` - A set of metadata entries assigned to the Provider VDC. See [Metadata](#metadata) section for details.
-->

<a id="compute-capacity"></a>
## Compute Capacity

The `compute_capacity` attribute is a list with a single item which has the following nested attributes:

* `cpu` - An indicator of CPU. See [CPU and memory](#cpu-and-memory) below.
* `memory` - An indicator of memory. See [CPU and memory](#cpu-and-memory) below.
* `is_elastic` -  True if compute capacity can grow or shrink based on demand.
* `is_ha` - True if compute capacity is highly available.

<a id="cpu-and-memory"></a>
### CPU and memory

The `cpu` and `memory` indicators have the following nested attributes:

* `allocation` - Allocated CPU/Memory for this Provider VDC.
* `overhead` - CPU/Memory overhead for this Provider VDC.
* `reserved` - Reserved CPU/Memory for this Provider VDC.
* `total` - Total CPU/Memory for this Provider VDC.
* `units` - Units for the CPU/Memory of this Provider VDC.
* `used` - Used CPU/Memory in this Provider VDC.

<a id="metadata"></a>
## Metadata

The `metadata_entry` is a set of metadata entries that have the following structure:

* `key` - Key of this metadata entry.
* `value` - Value of this metadata entry.
* `type` - Type of this metadata entry. One of: `MetadataStringValue`, `MetadataNumberValue`, `MetadataDateTimeValue`, `MetadataBooleanValue`.
* `user_access` - User access level for this metadata entry. One of: `PRIVATE` (hidden), `READONLY` (read only), `READWRITE` (read/write).
* `is_system` - Domain for this metadata entry. true if it belongs to `SYSTEM`, false if it belongs to `GENERAL`.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing provider VDC configuration can be [imported][docs-import] into this resource via supplying the path for the provider VDC.
Since the provider VDC is at the top of the VCD hierarchy, the path corresponds to the provider VDC name.
For example, using the structure in [example usage](#example-usage-1), representing an existing provider VDC configuration
that was **not** created using Terraform:

You can import such provider VDC configuration into terraform state using one of the following commands

```
terraform import vcd_provider_vdc.pvdc1 myPvdc
# OR
terraform import vcd_provider_vdc.pvdc1 provider-vdc-ID
```

After that, you may need to edit the configuration file before you can either update or delete the provider VDC configuration.
Running `terraform plan` at this stage will show the difference between the minimal configuration file and the stored properties.

One important point: if the NSX-T manager has more than one network pools attached, all of them will end up in the provider VDC
configuration, and the plan will show such difference. The discrepancy will not appear if you created the provider VDC
with Terraform, but it will if you import it. 