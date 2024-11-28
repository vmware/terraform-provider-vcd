---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_tm_org_vdc"
sidebar_current: "docs-vcd-datasource-tm-org-vdc"
description: |-
  Provides a data source to manage VMware Cloud Foundation Tenant Manager Organization VDC.
---

# vcd\_tm\_org\_vdc

Provides a data source to manage VMware Cloud Foundation Tenant Manager Organization VDC.

## Example Usage

```hcl
data "vcd_tm_org_vdc" "test" {
  name = "my-tm-org-vdc"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for the existing Org VDC
* `org_id` - (Required) An ID for the parent Org

## Attribute Reference

All the arguments and attributes defined in
[`vcd_tm_org_vdc`](/providers/vmware/vcd/latest/docs/resources/tm_org_vdc) resource are available.
