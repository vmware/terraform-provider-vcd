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