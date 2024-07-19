---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_api_filter"
sidebar_current: "docs-vcd-data-source-api-filter"
description: |-
  Provides a data source to read API Filters in Cloud Director. An API Filter allows to extend VCD API with customised URLs
  that can be redirected to an External Endpoint.
---

# vcd\_api\_filter

Supported in provider *v3.14+* and VCD 10.4.3+.

Provides a data source to read API Filters in Cloud Director. An API Filter allows to extend VCD API with customised URLs
that can be redirected to an [`vcd_external_endpoint`](/providers/vmware/vcd/latest/docs/resources/external_endpoint).

~> Only `System Administrator` can use this data source.

## Example Usage

```hcl
data "vcd_api_filter" "api_filter1" {
  api_filter_id = "urn:vcloud:apiFilter:4252ab09-eed8-4bc6-86d7-6019090273f5"
}
```

## Argument Reference

The following arguments are supported:

* `api_filter_id` - (Required) ID of the API Filter. This is the only way of unequivocally identify an API Filter. A list of
available API Filters can be obtained by using the `list@` option of the import mechanism of the [resource](/providers/vmware/vcd/latest/docs/r/api_filter#importing)

## Attribute Reference

All the arguments from [the `vcd_api_filter` resource](/providers/vmware/vcd/latest/docs/resources/api_filter)
are available as read-only.
