---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_org_ldap"
sidebar_current: "docs-vcd-datasource-org-ldap"
description: |-
  Provides a data source to read LDAP configuration for an organization.
---

# vcd\_org\_ldap

Supported in provider *v3.8+*.

Provides a data source to read LDAP configuration for an organization.

## Example Usage

```hcl
data "vcd_org_ldap" "first" {
  org_name = "my-org"
}
```

## Argument Reference

The following arguments are supported:

* `org_name` - (Required)  - Name of the organization containing the LDAP settings

## Attribute Reference

All the arguments and attributes defined in
[`vcd_org_ldap`](/providers/vmware/vcd/latest/docs/resources/org_ldap) resource are available.
