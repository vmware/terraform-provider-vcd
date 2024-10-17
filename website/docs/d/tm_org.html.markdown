---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_tm_org"
sidebar_current: "docs-vcd-datasource-tm-org"
description: |-
  Provides a data source to read VMware Cloud Foundation Tenant Manager Organization.
---

# vcd\_nsxt\_tm\_org

Provides a data source to read VMware Cloud Foundation Tenant Manager Organization.

## Example Usage

```hcl
data "vcd_tm_org" "existing" {
  name = "my-org-name"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of organization.

## Attribute Reference

All the arguments and attributes defined in
[`vcd_tm_org`](/providers/vmware/vcd/latest/docs/resources/tm_org) resource are available.