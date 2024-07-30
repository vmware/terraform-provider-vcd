---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_external_endpoint"
sidebar_current: "docs-vcd-resource-external-endpoint"
description: |-
  Provides a resource to manage External Endpoints in VMware Cloud Director. An External Endpoint holds information for the
  HTTPS endpoint which requests will be proxied to when using an API Filter.
---

# vcd\_external\_endpoint

Supported in provider *v3.14+* and VCD 10.4.3+.

Provides a resource to manage External Endpoints in VMware Cloud Director. An External Endpoint holds information for the
HTTPS endpoint which requests will be proxied to when using a [`vcd_api_filter`](/providers/vmware/vcd/latest/docs/resources/api_filter).

~> Only `System Administrator` can create this resource.

## Example Usage

```hcl
resource "vcd_external_endpoint" "external_endpoint1" {
  vendor      = "vmware"
  name        = "my-endpoint"
  version     = "1.0.0"
  enabled     = true
  description = "A simple external endpoint example"
  root_url    = "https://www.vmware.com"

  disable_on_removal = true # Will disable the endpoint and then remove it when this resource is destroyed
}
```

## Argument Reference

The following arguments are supported:

* `vendor` - (Required) The vendor name of the External Endpoint. The combination of `vendor` + `name` + `version` must be unique. Can't be modified after creation
* `name` - (Required) The name of the External Endpoint. The combination of `vendor` + `name` + `version` must be unique. Can't be modified after creation
* `version` - (Required) The version of the External Endpoint. The combination of `vendor` + `name` + `version` must be unique. Can't be modified after creation
* `enabled` - (Required) Whether the External Endpoint is enabled or not. **Must be `false` before removing this resource**, otherwise deletion will fail.
  To disable it automatically on removal, set `disable_on_removal=true` (see below)
* `disable_on_removal` - (Optional) Whether the External Endpoint should be disabled before a delete operation, to flawlessly remove it even if it is enabled.
  It is `false` by default
* `description` - (Optional) Description of the External Endpoint
* `root_url` - (Required) The endpoint which requests will be redirected to. Must use HTTPS protocol

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing External Endpoint configuration can be [imported][docs-import] into this resource via
supplying path for it. It can be imported by providing the unique combination of `vendor` + `name` + `version`:

```shell
terraform import vcd_external_endpoint.ep1 vmware.my-endpoint.1.0.0
```

```shell
VCD_IMPORT_SEPARATOR='%' terraform import vcd_external_endpoint.ep1 vmware%my-endpoint%1.0.0
```

[docs-import]: https://www.terraform.io/docs/import/