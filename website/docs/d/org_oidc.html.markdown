---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_org_oidc"
sidebar_current: "docs-vcd-data-source-org-oidc"
description: |-
  Provides a data source to read OIDC configuration for an Organization.
---

# vcd\_org\_oidc

Supported in provider *v3.13+*.

Provides a data source to read OIDC configuration for an Organization.

## Example Usage

```hcl
data "vcd_org" "my_org" {
  name = "my-org"
}

data "vcd_org_oidc" "oidc_settings" {
  org_id = data.vcd_org.my_org.id
}
```

## Argument Reference

The following arguments are supported:

* `org_id` - (Required)  - ID of the organization containing the OIDC settings

## Attribute Reference

* `foo` - TBD
