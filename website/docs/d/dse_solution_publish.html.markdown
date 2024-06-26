---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_dse_solution_publish"
sidebar_current: "docs-vcd-data-source-dse-solution-publish"
description: |-
  Provides a data source to read Data Solution publishing settings for a particular tenant.
---

# vcd\_dse\_solution\_publish

Supported in provider *v3.13+* and VCD 10.5.0+ with Data Solution Extension.

Provides a data source to read Data Solution publishing settings for a particular tenant.

## Example Usage

```hcl
data "vcd_dse_solution_publish" "mongodb-community" {
  data_solution_id = vcd_dse_registry_configuration.mongodb-community.id
  org_id           = data.vcd_org.tenant-org.id
}

data "vcd_org" "tenant-org" {
  name = "tenant_org"
}
```

## Argument Reference

The following arguments are supported:

* `data_solution_id` - (Required) ID of Data Solution
* `org_id` - (Required) Organization ID

## Attribute Reference

All the arguments and attributes defined in
[`vcd_dse_solution_publish`](/providers/vmware/vcd/latest/docs/resources/dse_solution_publish)
resource are available.
