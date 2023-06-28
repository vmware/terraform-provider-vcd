---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_api_token"
sidebar_current: "docs-vcd-resource-api_token"
description: |-
  Provides a resource to manage API tokens. API tokens are an easy way to authenticate to VCD. 
  They are user-based and have the same role as the user.
---

# vcd\_api\_token 

Provides a resource to manage API tokens. API tokens are an easy way to authenticate to VCD. 
They are user-based and have the same role as the user. Explained in more detail [here][api-tokens].

Supported in provider *v3.10+* and VCD 10.3.1+.

## Example usage

```hcl
resource "vcd_api_token" "example_token" {
  name             = "example_token"
  file_name        = "example_token.json"
  allow_token_file = true
}
```

## Argument reference

The following arguments are supported:

* `name` - (Required) The unique name of the API token for a specific user.
* `file_name` - (Required) The name of the file which will be created containing the API token
* `allow-token-file` - (Required) An additional check that the user is aware that the file contains
  SENSITIVE information. Must be set to `true` or it will return a validation error.

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.][docs-import]

An existing API token can be [imported][docs-import] into this resource via supplying
the full dot separated path. An example is below:

```
terraform import vcd_api_token.example_token example_token
```

[api-tokens]: [https://blogs.vmware.com/cloudprovider/2022/03/cloud-director-api-token.html]
[docs-import]: https://www.terraform.io/docs/import/
