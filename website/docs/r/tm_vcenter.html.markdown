---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_tm_vcenter"
sidebar_current: "docs-vcd-resource-tm-vcenter"
description: |-
  Provides a resource to manage vCenters.
---

# vcd\_tm\_vcenter

Provides a resource to manage vCenters.

~> Only `System Administrator` can create this resource.

## Example Usage

```hcl
resource "vcd_tm_vcenter" "test" {
  name                    = "TestAccVcdTmVcenter-rename"
  url                     = "https://host:443"
  auto_trust_certificate  = true
  refresh_vcenter_on_read = true
  username                = "admim@vsphere.local"
  password                = "CHANGE-ME"
  is_enabled              = true
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for vCenter server
* `description` - (Optional) An optional description for vCenter server
* `username` - (Required) A username for authenticating to vCenter server
* `password` - (Required) A password for authenticating to vCenter server
* `refresh_vcenter_on_read` - (Optional) An optional flag to trigger refresh operation on the
  underlying vCenter on every read. This might take some time, but can help to load up new artifacts
  from vCenter (e.g. Supervisors). This operation is visible as a new task in UI. Update is a no-op.
  It may be useful after adding vCenter or if new infrastructure is added to vCenter. Default
  `false`.
* `refresh_policies_on_read` - (Optional) An optional flag to trigger policy refresh operation on
  the underlying vCenter on every read. This might take some time, but can help to load up new
  artifacts from vCenter (e.g. Storage Policies). Update is a no-op. This operation is visible as a
  new task in UI. It may be useful after adding vCenter or if new infrastructure is added to
  vCenter. Default `false`. 
* `url` - (Required) An URL of vCenter server
* `auto_trust_certificate` - (Required) Defines if the certificate of a given vCenter server should
  automatically be added to trusted certificate store. **Note:** not having the certificate trusted
  will cause malfunction.
* `is_enabled` - (Optional) Defines if the vCenter is enabled. Default `true`. The vCenter must
  always be disabled before removal (this resource will disable it automatically on destroy).


## Attribute Reference

The following attributes are exported on this resource:

* `has_proxy` - Indicates that a proxy exists within vCloud Director that proxies this vCenter
  server for access by authorized end-users
* `is_connected` - Defines if the vCenter server is connected.
* `mode` - One of `NONE`, `IAAS` (scoped to the provider), `SDDC` (scoped to tenants), `MIXED` (both
  uses are possible)
* `connection_status` - `INITIAL`, `INVALID_SETTINGS`, `UNSUPPORTED`, `DISCONNECTED`, `CONNECTING`,
  `CONNECTED_SYNCING`, `CONNECTED`, `STOP_REQ`, `STOP_AND_PURGE_REQ`, `STOP_ACK`
* `cluster_health_status` - Cluster health status. One of `GRAY` , `RED` , `YELLOW` , `GREEN`
* `version` - vCenter version
* `uuid` - UUID of vCenter
* `vcenter_host` - Host of Vcenter server
* `status` - Status can be `READY` or `NOT_READY`. It is a derivative field of `is_connected` and
  `connection_status` so relying on those fields could be more precise.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing vCenter configuration can be [imported][docs-import] into this resource via supplying
path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_tm_vcenter.imported my-vcenter
```

The above would import the `my-vcenter` vCenter settings that are defined at provider level.
