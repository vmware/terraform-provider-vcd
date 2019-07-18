---
layout: "vcd"
page_title: "vCloudDirector: vcd_org"
sidebar_current: "docs-vcd-data-source-org"
description: |-
  Provides an organization data source.
---

# vcd\_org

Provides a vCloud Director Org data source. An organization can be used to manage catalogs, virtual
data centers, and users.

Supported in provider *v2.5+*

## Example Usage

```hcl
data "vcd_org" "my-org" {
  name   = "my-org"
}

resource "vcd_org" "my-org-clone" {
  name                 = "my-org-clone"
  full_name            = "${data.vcd_org.my-org.full_name}"
  can_publish_catalogs = "${data.vcd_org.my-org.can_publish_catalogs}"
  deployed_vm_quota    = "${data.vcd_org.my-org.deployed_vm_quota}"
  stored_vm_quota      = "${data.vcd_org.my-org.stored_vm_quota}"
  is_enabled           = "${data.vcd_org.my-org.is_enabled}"
  delete_force         = "true"
  delete_recursive     = "true"
}

```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Org name
* `full_name` - (Optional) Org full name
* `is_enabled` - (Optional) - True if this organization is enabled (allows login and all other operations). Default is `true`.
* `description` - (Optional) - Org description. Default is empty.
* `deployed_vm_quota` - (Optional) - Maximum number of virtual machines that can be deployed simultaneously by a member of this organization. Default is unlimited (0)
* `stored_vm_quota` - (Optional) - Maximum number of virtual machines in vApps or vApp templates that can be stored in an undeployed state by a member of this organization. Default is unlimited (0)
* `can_publish_catalogs` - (Optional) - True if this organization is allowed to share catalogs. Default is `true`.
* `delay_after_power_on_seconds` - (Optional) - Specifies this organization's default for virtual machine boot delay after power on. Default is `0`.
