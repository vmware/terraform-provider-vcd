---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_org_oidc"
sidebar_current: "docs-vcd-data-source-org-oidc"
description: |-
  Provides a data source to read the OpenID Connect (OIDC) configuration of an Organization in VMware Cloud Director.
---

# vcd\_org\_oidc

Provides a data source to read the OpenID Connect (OIDC) configuration of an Organization in VMware Cloud Director.

Supported in provider *v3.13+*.

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

* `org_id` - (Required) - ID of the organization containing the OIDC settings

## Attribute Reference

All the arguments and attributes from [the `vcd_org_oidc` resource](/providers/vmware/vcd/latest/docs/resources/org_oidc) are available as read-only.
