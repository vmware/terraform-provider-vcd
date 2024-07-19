---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_external_endpoint"
sidebar_current: "docs-vcd-data-source-external-endpoint"
description: |-
  Provides a data source to read External Endpoints in Cloud Director. An External Endpoint holds information for the
  HTTPS endpoint which requests will be proxied to when using an API Filter.
---

# vcd\_external\_endpoint

Supported in provider *v3.14+* and VCD 10.4.3+.

Provides a data source to read External Endpoints in VMware Cloud Director. An External Endpoint holds information for the
HTTPS endpoint which requests will be proxied to when using a [`vcd_api_filter`](/providers/vmware/vcd/latest/docs/data sources/api_filter).

~> Only `System Administrator` can use this data source.

## Example Usage

```hcl
data "vcd_external_endpoint" "external_endpoint1" {
  vendor      = "vmware"
  name        = "my-endpoint"
  version     = "1.0.0"
}
```

## Argument Reference

* `vendor` - (Required) The vendor name of the External Endpoint
* `name` - (Required) The name of the External Endpoint
* `version` - (Required) The version of the External Endpoint

## Attribute Reference

All the remaining arguments from [the `vcd_external_endpoint` resource](/providers/vmware/vcd/latest/docs/resources/external_endpoint)
are available as read-only.
