---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_org_saml"
sidebar_current: "docs-vcd-data-source-org-saml"
description: |-
  Provides a data source to read SAML configuration for an organization.
---

# vcd\_org\_saml

Supported in provider *v3.10+*.

Provides a data source to read SAML configuration for an organization.

-> **Note:** This data source requires system administrator privileges.

## Example Usage

```hcl
data "vcd_org" "my-org" {
  name = "my-org"
}

data "vcd_org_saml" "first" {
  org_id = data.vcd_org.my-org.id
}
```

## Argument Reference

The following arguments are supported:

* `org_id` - (Required)  - ID of the organization containing the SAML settings

## Attribute Reference

* `enabled` - Shows whether the SAML identity service is used for authentication
* `entity_id` - Your service provider entity ID
* `group` - The name of the SAML attribute that returns the identifiers of all the groups of which the user is a member
* `role` - The name of the SAML attribute that returns the identifiers of all roles of the user
* `email` - The name of the SAML attribute that returns the email address of the user
* `first_name` - The name of the SAML attribute that returns the first name of the user
* `surname` - The name of the SAML attribute that returns the surname of the user
* `full_name` - The name of the SAML attribute that returns the full name of the user
* `user_name` - The name of the SAML attribute that returns the username of the user
