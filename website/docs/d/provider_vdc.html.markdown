---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_provider_vdc"
sidebar_current: "docs-vcd-data-source-provider-vdc"
description: |-
  Provides a Provider VDC data source.
---

# vcd\_provider\_vdc

Gives a VMware Cloud Director Provider VDC data source. This data source can be used to reference a Provider VDC and use its 
data within other resources or data sources.

Supported in provider *v3.8+*

## Example Usage

```hcl
data "vcd_provider_vdc" "my-pvdc" {
  name = "my-pvdc"
}

output "provider_vdc" {
  value = data.vcd_provider_vdc.my-pvdc.id
}

```

## Argument Reference

The following arguments are supported:
 
* `name` - (Required) Provider VDC name

## Attribute reference

* `description` - Optional description of the Provider VDC.
* `status` - Status of the Provider VDC, it can be -1 (creation failed), 0 (not ready), 1 (ready), 2 (unknown) or 3 (unrecognized).
* `is_enabled` - True if this Provider VDC is enabled and can provide resources to organization VDCs. A Provider VDC is always enabled on creation.
* `capabilities` - Set of virtual hardware versions supported by this Provider VDC.
* `compute_capacity` - An indicator of CPU and memory capacity. See [Compute Capacity](#compute-capacity) below for details.
* `compute_provider_scope` - Represents the compute fault domain for this Provider VDC. This value is a tenant-facing tag that is shown to tenants when viewing fault domains of the child Organization VDCs (for example, a VDC Group).
* `highest_supported_hardware_version` - The highest virtual hardware version supported by this Provider VDC.
* `nsxt_manager_id` - ID of the registered NSX-T Manager that backs networking operations for this Provider VDC.
* `storage_containers_ids` - Set of IDs of the vSphere Datastores backing this Provider VDC.
* `external_network_ids` - Set of IDs of External Networks.
* `storage_profile_ids` - Set of IDs to the Storage Profiles available to this Provider VDC.
* `resource_pool_ids` - Set of IDs of the Resource Pools backing this provider VDC.
* `network_pool_ids` - Set IDs of the Network Pools used by this Provider VDC.
* `universal_network_pool_id` - ID of the universal network reference.
* `host_ids` - Set with all the hosts which are connected to VC server.
* `vcenter_id` - ID of the vCenter Server that provides the Resource Pools and Datastores.
* `metadata` - (Deprecated) Use `metadata_entry` instead. Key and value pairs for Provider VDC Metadata.
* `metadata_entry` - A set of metadata entries assigned to the Provider VDC. See [Metadata](#metadata) section for details.

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

The `metadata_entry` (*v3.8+*) is a set of metadata entries that have the following structure:

* `key` - Key of this metadata entry.
* `value` - Value of this metadata entry.
* `type` - Type of this metadata entry. One of: `MetadataStringValue`, `MetadataNumberValue`, `MetadataDateTimeValue`, `MetadataBooleanValue`.
* `user_access` - User access level for this metadata entry. One of: `PRIVATE` (hidden), `READONLY` (read only), `READWRITE` (read/write).
* `is_system` - Domain for this metadata entry. true if it belongs to `SYSTEM`, false if it belongs to `GENERAL`.
