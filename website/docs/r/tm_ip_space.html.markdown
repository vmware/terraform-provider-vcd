---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_tm_ip_space"
sidebar_current: "docs-vcd-resource-tm-ip-space"
description: |-
  Provides a VMware Cloud Foundation Tenant Manager IP Space resource.
---

# vcd\_tm\_ip\_space

Provides a VMware Cloud Foundation Tenant Manager IP Space resource.


## Example Usage

```hcl
data "vcd_tm_region" "demo" {
  name = "demo-region"
}

resource "vcd_tm_ip_space" "demo" {
  name                          = "demo-ip-space"
  description                   = "description"
  region_id                     = data.vcd_tm_region.demo.id
  external_scope                = "12.12.0.0/30"
  default_quota_max_subnet_size = 24
  default_quota_max_cidr_count  = 1
  default_quota_max_ip_count    = 1

  internal_scope {
    name = "scope1"
    cidr = "10.0.0.0/24"
  }

  internal_scope {
    name = "scope2"
    cidr = "20.0.0.0/24"
  }

  internal_scope {
    name = "scope3"
    cidr = "30.0.0.0/24"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A tenant facing name for IP Space
* `description` - (Optional) An optional description
* `region_id` - (Required) A region ID. Can be looked up using
  [region data source](/providers/vmware/vcd/latest/docs/data-sources/tm_region)
* `external_scope` - (Required) A CIDR (e.g. 10.0.0.0/8) for External Reachability. It represents
  the IPs used outside the datacenter, north of the Provider Gateway
* `default_quota_max_subnet_size` - (Required) Maximum subnet size that can be allocated (e.g. 24)
* `default_quota_max_cidr_count` - (Required) Maximum number of CIDRs that can be allocated (`-1`
  for unlimited)
* `default_quota_max_ip_count` - (Required) Maximum number of floating IPs that can be allocated
  (`-1` for unlimited)
* `internal_scope` - (Required) A set of IP Blocks that represent IPs used in this local datacenter,
  south of the Provider Gateway. IPs within this scope are used for configuring services and
  networks. [internal_scope](#internal-scope)


<a id="internal-scope"></a>

## internal_scope block

* `cidr` - (Required) CIDR for IP block (e.g. 10.0.0.0/16)
* `name` - (Optional) An optional friendly name for this block


## Attribute Reference

The following attributes are exported on this resource:

* `status` - One of:
 * `PENDING` - Desired entity configuration has been received by system and is pending realization
 * `CONFIGURING` - The system is in process of realizing the entity
 * `REALIZED` - The entity is successfully realized in the system
 * `REALIZATION_FAILED` - There are some issues and the system is not able to realize the entity
 * `UNKNOWN` - Current state of entity is unknown


## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing IP Space configuration can be [imported][docs-import] into this resource via supplying
path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_tm_ip_space.imported my-ip-space-name
```

The above would import the `my-ip-space-name` IP Space.