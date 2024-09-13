---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_api_filter"
sidebar_current: "docs-vcd-resource-api-filter"
description: |-
  Provides a resource to manage API Filters in VMware Cloud Director. An API Filter allows to extend VCD API with customised URLs
  that can be redirected to an External Endpoint.
---

# vcd\_api\_filter

Supported in provider *v3.14+* and VCD 10.5.1+.

Provides a resource to manage API Filters in VMware Cloud Director. An API Filter allows to extend VCD API with customised URLs
that can be redirected to an [`vcd_external_endpoint`](/providers/vmware/vcd/latest/docs/resources/external_endpoint).

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
}

# The requests to '<vcd-url>/ext-api/custom/...' will be redirected to the External Endpoint
resource "vcd_api_filter" "af" {
  external_endpoint_id = vcd_external_endpoint.external_endpoint1.id
  url_matcher_pattern  = "/custom/.*"
  url_matcher_scope    = "EXT_API"
}

```

## Argument Reference

The following arguments are supported:

* `external_endpoint_id` - (Required) ID of the [External Endpoint](/providers/vmware/vcd/latest/docs/resources/external_endpoint) where this API Filter will process the requests to
* `url_matcher_pattern` - (Required) Request URL pattern, written as a regular expression. This argument cannot exceed 1024 characters.
  In most cases, it should end with `.*` (it is like a suffix) which specifies that all the parts of the URL coming after (like parameters) will be redirected to an external endpoint.
  It is important to note that in the case of `url_matcher_scope=EXT_UI_TENANT`, the tenant name is not part of the pattern, it will match the request after the tenant name - if request
  is *"/ext-ui/tenant/testOrg/custom/test"*, the pattern will match against */custom/test*
* `url_matcher_scope` - (Required) Allowed values are `EXT_API`, `EXT_UI_PROVIDER`, `EXT_UI_TENANT` corresponding to
 */ext-api*, */ext-ui/provider*, */ext-ui/tenant/<tenant-name>*

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing API Filter configuration can be [imported][docs-import] into this resource via
supplying path for it. It must be imported by ID as there's no other unequivocally way to identify an API Filter.
This might not be trivial to lookup therefore there is a helper for listing available items:

```
terraform import vcd_api_filter.my_api_filter list@vmware.my-external-endpoint.1.0.0
vcd_api_filter.my_api_filter: Importing from ID "list@vmware.my-external-endpoint.1.0.0"...
╷
│ Error: resource was not imported! resource id must be specified in one of these formats:
│ 'api-filter-id' to import by API Filter ID
│ 'list@vendor.name.version' to get a list of API Filters related to the External Endpoint identified by vendor, name and version
│ Retrieving all API Filters that use urn:vcloud:extensionEndpoint:vmware.my-external-endpoint.1.0.0 as External Endpoint
│ No	ID								Scope		Pattern
│ --	--								-----		-------
│ 1	urn:vcloud:apiFilter:4252ab09-eed8-4bc6-86d7-6019090273f5	EXT_UI_PROVIDER	/custom/.*
```

The argument of `list@` corresponds to an External Endpoint, so it will list all API Filters of that External Endpoint.

An import then can be done by ID

```
terraform import vcd_api_filter.my_api_filter urn:vcloud:apiFilter:4252ab09-eed8-4bc6-86d7-6019090273f5
```

[docs-import]: https://www.terraform.io/docs/import/