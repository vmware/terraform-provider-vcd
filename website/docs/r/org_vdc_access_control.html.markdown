---
layout: "vcd"
page_title: "VMware Cloud Director: org_vdc_access_control"
sidebar_current: "docs-vcd-resource-org-vdc-access-control"
description: |-
Provides a VMware Cloud Director Org Vdc access control resource. This can be
used to share VDC across users or groups.
---

# vcd\_org\_vdc\_access\_control

Provides a VMware Cloud Director Org Vdc access control resource. This can be
used to share VDC across users and/or groups.

Supported in provider *v3.7+*

-> **Note:** This resource requires either system or org administrator privileges.

## Example Usage

```hcl
resource "vcd_org_vdc_access_control" "my_access_control" {
  org                   = "my-org"
  vdc                   = "my-vdc"
  shared_with_everyone  = false

  shared_with {
    user_id             = vcd_org_user.my-user.id
    access_level        = "ReadOnly"
  }

  shared_with {
    user_id             = vcd_org_user.my-user2.id
    access_level        = "ReadOnly"
  }
}
```

```hcl
resource "vcd_org_vdc_access_control" "my_access_control" {
  org                   = "my-org"
  vdc                   = "my-vdc"
  shared_with_everyone  = true
  everyone_access_level = "ReadOnly"
}
```
## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organizations.
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level.
* `shared_with_everyone` - (Required) Whether the VDC is shared with everyone.
* `everyone_access_level` - (Optional) Access level when the VDC is shared with everyone (only ReadOnly is available). Required when shared_with_everyone is set.

-> The ID of `vcd_security_tag` is set to its name since VCD behind the scenes doesn't create an ID.

# Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing security tag can be [imported][docs-import] into this resource via supplying its path.
The path for this resource is made of org-name.security-tag-name
An example is below:

```
terraform import vcd_security_tag.my-tag my-org.my-security-tag-name
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR


[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the security tag as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the security tag stored properties.
