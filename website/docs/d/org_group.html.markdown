---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_org_group"
sidebar_current: "docs-vcd-datasource-org-group"
description: |-
  Provides a data source for VMware Cloud Director Organization groups.
---

# vcd\_org\_group

Provides a data source for VMware Cloud Director Organization groups. This can be used to fetch organization groups already defined in `SAML` or `LDAP`.

Supported in provider *v3.6+*

~> **Note:** This operation requires the rights included in the predefined `Organization
Administrator` role or an equivalent set of rights. `SAML` or `LDAP` must be configured as vCD
does not support local groups and will return HTTP error 403 "This operation is denied." if selected
`provider_type` is not configured.

## Example Usage to fetch an Organization group

```hcl
datasource "vcd_org_group" "org1" {
  org  = "org1"
  name = "Org1-AdminGroup"
}

output "group_role" {
  value = data.vcd_org_group.org1.role
}
```


## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which the VDC belongs. Optional if defined at provider level.
* `name` - (Required) A unique name for the group.

## Attribute reference

All attributes defined in [org_group](/providers/vmware/vcd/latest/docs/resources/org_group#attribute-reference) are supported.