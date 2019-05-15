---
layout: "vcd"
page_title: "vCloudDirector: vcd_org_vdc"
sidebar_current: "docs-vcd-resource-org-vdc"
description: |-
  Provides a vCloud Director Organization VDC resource. This can be used to create and delete a Organization VDC.
---

# vcd\_org\_vdc

Provides a vCloud Director Organization VDC resource. This can be used to create and delete a Organization VDC.
Requires system administrator privileges.

Supported in provider *v2.2+*

## Example Usage

```hcl
provider "vcd" {
  user     = "${var.admin_user}"
  password = "${var.admin_password}"
  org      = "System"
  url      = "https://AcmeVcd/api"
}

resource "vcd_org_vdc" "my-vdc" {
  name        = "my-vdc"
  description = "The pride of my work"
  org         = "my-org"

  allocation_model = "ReservationPool"
  network_pool_name = "vDC1-VXLAN-NP"
  provider_vdc_name = "vDC1"

  compute_capacity {
    cpu {
      allocated = 2048
      limit     = 2048
      reserved  = 2048
    }

    memory {
      allocated = 2048
      limit     = 2048
      reserved  = 2048
    }
  }

  storage_profile {
    name     = "storage-name"
    limit    = 10240
    default  = true    
  }

  enabled                  = true
  enable_thin_provisioning = true
  enable_fast_provisioning = true
  delete_force             = true
  delete_recursive         = true
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) Organization to create the VDC in, optional if defined at provider level
* `name` - (Required) VDC name
* `description` - (Optional) VDC friendly description
* `provider_vdc_name` - (Required) A name of the Provider VDC from which this organization VDC is provisioned.
* `allocation_model` - (Required) The allocation model used by this VDC; must be one of {AllocationVApp ("Pay as you go"), AllocationPool ("Allocation pool"), ReservationPool ("Reservation pool")}
* `compute_capacity` - (Required) The compute capacity allocated to this VDC.  See [Compute Capacity](#computecapacity) below for details.
* `nic_quota` - (Optional) Maximum number of virtual NICs allowed in this VDC. Defaults to 0, which specifies an unlimited number.
* `network_quota` - (Optional) Maximum number of network objects that can be deployed in this VDC. Defaults to 0, which means no networks can be deployed.
* `vm_quota` - (Optional) The maximum number of VMs that can be created in this VDC. Includes deployed and undeployed VMs in vApps and vApp templates. Defaults to 0, which specifies an unlimited number.
* `enabled` - (Optional) True if this VDC is enabled for use by the organization VDCs. Default is true.
* `storage_profile` - (Required) Storage profiles supported by this VDC.  See [Storage Profile](#storageprofile) below for details.
* `memory_guaranteed` - (Optional) Percentage of allocated memory resources guaranteed to vApps deployed in this VDC. For example, if this value is 0.75, then 75% of allocated resources are guaranteed. Required when AllocationModel is AllocationVApp or AllocationPool. Value defaults to 1.0 if the element is empty.
* `cpu_guaranteed` - (Optional) Percentage of allocated CPU resources guaranteed to vApps deployed in this VDC. For example, if this value is 0.75, then 75% of allocated resources are guaranteed. Required when AllocationModel is AllocationVApp or AllocationPool. Value defaults to 1.0 if the element is empty.
* `cpu_speed` - (Optional) Specifies the clock frequency, in Megahertz, for any virtual CPU that is allocated to a VM. A VM with 2 vCPUs will consume twice as much of this value. Ignored for ReservationPool. Required when AllocationModel is AllocationVApp or AllocationPool, and may not be less than 256 MHz. Defaults to 1000 MHz if value isn't provided.
* `enable_thin_provisioning` - (Optional) Boolean to request thin provisioning. Request will be honored only if the underlying data store supports it. Thin provisioning saves storage space by committing it on demand. This allows over-allocation of storage.
* `enable_fast_provisioning` - (Optional) Request fast provisioning. Request will be honored only if the underlying datastore supports it. Fast provisioning can reduce the time it takes to create virtual machines by using vSphere linked clones. If you disable fast provisioning, all provisioning operations will result in full clones.
* `network_pool_name` - (Optional) Reference to a network pool in the Provider VDC. Required if this VDC will contain routed or isolated networks.
* `allow_over_commit` - (Optional) Set to false to disallow creation of the VDC if the AllocationModel is AllocationPool or ReservationPool and the ComputeCapacity you specified is greater than what the backing Provider VDC can supply. Default is true.
* `enable_vm_discovery` - (Optional) If true, discovery of vCenter VMs is enabled for resource pools backing this VDC. If false, discovery is disabled. If left unspecified, the actual behaviour depends on enablement at the organization level and at the system level.
* `delete_force` - (Required) When destroying use `delete_force=True` to remove a VDC and any objects it contains, regardless of their state.
* `delete_recursive` - (Required) When destroying use `delete_recursive=True` to remove the VDC and any objects it contains that are in a state that normally allows removal.


<a id="storageprofile"></a>
## Storage Profile

* `name` - (Required) Name of Provider VDC storage profile.
* `enabled` - (Optional) True if this storage profile is enabled for use in the VDC. Default is true.
* `limit` - (Required) Maximum number of MB allocated for this storage profile. A value of 0 specifies unlimited MB.
* `default` - (Required) True if this is default storage profile for this VDC. The default storage profile is used when an object that can specify a storage profile is created with no storage profile specified.

<a id="computecapacity"></a>
## Compute Capacity

Capacity must be specified twice, once for `memory` and another for `cpu`.  Each has the same structure:

* `allocated` - (Required) Capacity that is committed to be available. Value in MB or MHz.
* `limit` - (Optional) Capacity limit relative to the value specified for Allocation. It must not be less than that value. If it is greater than that value, it implies over provisioning. A value of 0 specifies unlimited Units.
* `reserved` - (Optional) Capacity reserved