---
layout: "vcd"
page_title: "vCloudDirector: vcd_org"
sidebar_current: "docs-vcd-resource-org"
description: |-
Provides a vCloud Director Organization resource. This can be used to create and delete an organization.
---

# vcd\_org

Provides a vCloud Director Org resource. This can be used to create and delete an organization.
Requires system administrator privileges.

Supported in provider *v2.0+*

## Example Usage

```
provider "vcd" {
  user                 = "${var.admin_user}"
  password             = "${var.admin_password}"
  org                  = "System"
  url                  = "https://AcmeVcd/api"
}
resource "vcd_org" "my-org" {
  name             = "my-org"
  full_name        = "My organization"
  description      = "The pride of my work"
  is_enabled       = "true"
  delete_recursive = "true"
  delete_force     = "true"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Org name
* `full_name` - (Required) Org full name
* `delete_recursive` - (Required) - pass `delete_recursive`=true as query parameter to remove an organization or VDC and any objects it contains that are in a state that normally allows removal.
* `delete_force` - (Required) - pass `delete_force=true` and `delete_recursive=true` to remove an organization or VDC and any objects it contains, regardless of their state.
* `is_enabled` - (Optional) - True if this organization is enabled (allows login and all other operations). Default is `true`.
* `description` - (Optional) - Org description. Default is empty.
* `deployed_vm_quota` - (Optional) - Maximum number of virtual machines that can be deployed simultaneously by a member of this organization. Default is unlimited (-1)
* `stored_vm_quota` - (Optional) - Maximum number of virtual machines in vApps or vApp templates that can be stored in an undeployed state by a member of this organization. Default is unlimited (-1)
* `can_publish_catalogs` - (Optional) - True if this organization is allowed to share catalogs. Default is `true`.
* `delay_after_power_on_seconds` - (Optional) - Specifies this organization's default for virtual machine boot delay after power on. Default is `0`.

## Sources

* [OrgType](https://code.vmware.com/apis/287/vcloud#/doc/doc/types/OrgType.html)
* [ReferenceType](https://code.vmware.com/apis/287/vcloud#/doc/doc/types/ReferenceType.html)
* [Org deletion](https://code.vmware.com/apis/287/vcloud#/doc/doc/operations/DELETE-Organization.html)
