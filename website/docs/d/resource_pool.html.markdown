---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_resource_pool"
sidebar_current: "docs-vcd-data-source-resource-pool"
description: |-
  Provides a data source for a resource pool attached to a vCenter.
---

# vcd\_resource\_pool

Provides a data source for a resource pool attached to a vCenter. A resource pool is an essential component of a Provider VDC.


~> Note 1: this data source requires System Administrator privileges

~> Note 2: you can create or modify a resource pool using [vSphere provider](https://registry.terraform.io/providers/hashicorp/vsphere/latest/docs/resources/resource_pool)

Supported in provider *v3.10+*


## Example Usage 1

```hcl
data "vcd_vcenter" "vcenter1" {
  name = "vc1"
}

data "vcd_resource_pool" "rp1" {
  name       = "resource-pool-for-vcd-01"
  vcenter_id = data.vcd_vcenter.vcenter1.id
}
```

## Example Usage 2

```hcl
data "vcd_resource_pool" "rp1" {
  name       = "common-name"
  vcenter_id = data.vcd_vcenter.vcenter1.id
}

# Error: could not find resource pool by name 'common-name': more than one resource pool was found with name common-name - 
# use resource pool ID instead - [resgroup-241 resgroup-239]
```

When you receive such error, you can run the script again, but using the resource pool ID instead of the name.

```hcl
data "vcd_resource_pool" "rp1" {
  name       = "resgroup-241"
  vcenter_id = data.vcd_vcenter.vcenter1.id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) resource pool name. The name may not be unique within the vCenter. If that happens, you will get an
   error message with the list of IDs for the pools with the same name, and can subsequently enter the resource pool ID instead of the name.
  (See [Example Usage 2](#example-usage-2))
* `vcenter_id` - (Required) ID of the vCenter to which this resource pool belongs.

## Attribute reference

* `hardware_version` - default hardware version available to this resource pool.
* `cluster_moref` - managed object reference of the vCenter cluster that this resource pool is hosted on.
