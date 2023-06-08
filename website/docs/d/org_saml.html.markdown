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
