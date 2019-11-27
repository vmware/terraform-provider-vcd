---
layout: "vcd"
page_title: "vCloudDirector: vcd_ipset"
sidebar_current: "docs-vcd-resource-ipset"
description: |-
  Provides an IP set resource.
---

# vcd\_ipset

Provides a vCloud Director IP set resource. An IP set is a group of IP addresses that you can add as
  the source or destination in a firewall rule or in DHCP relay configuration.


Supported in provider *v2.6+*

## Example Usage 1

```hcl
resource "vcd_ipset" "test-ipset" {
  org          = "my-org"
  vdc          = "my-org-vdc"

  name                   = "TestAccVcdIpSet-changed2"
  is_inheritance_allowed = false
  description            = "test-ip-set-changed-description"
  ip_addresses           = ["1.1.1.1/24","10.10.10.100-10.10.10.110"]
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level
* `name` - (Required) Unique IP set name.
* `description` - (Optional) An optional description for IP set.
* `ip_addresses` - (Required) A set of IP addresses, CIDRs and ranges as strings.
* `is_inheritance_allowed` (Optional) Toggle to enable inheritance to allow visibility at underlying scopes. (Default `true`)

## Attribute Reference

The following attributes are exported on this resource:

* `id` - ID of IP set

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing load balancer application rule can be [imported][docs-import] into this resource
via supplying the full dot separated path for load balancer application rule. An example is
below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_ipset.imported org-name.vdc-name.ipset-name
```

The above would import the IP set named `ipset-name` that is defined in org named `org-name` and vDC
named `vdc-name`.
