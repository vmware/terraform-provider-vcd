**---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_certificate_in_library"
sidebar_current: "docs-vcd-resource-certificate-in-library"
description: |-
Provides a certificate in System or Org library resource.
---

# vcd\_certificate\_in\_library
Supported in provider *v3.5+* and VCD 10.2+.

Provides a resource to manage certificate in System or Org library.
~> Only `System Administrator` can create this resource.

## Example Usage

```hcl
resource "vcd_certificate_in_library" "new-certificate" {
  org  = "myOrg"
  alias = "SAML certificate"
  description = "my description"
  certificate = file("/home/user/cert.pem")
  private_key = file("/home/user/key.pem")
  private_key_passphrase = "passphrase"
}
```

## Argument Reference

The following arguments are supported:

* `alias` - (Required)  - Alias(name) of certificate
* `description` - (Optional)  - Certificate description
* `certificate` - (Optional)  - Content of Certificate
* `private_key` - (Optional)  - Content of private key
* `private_key_passphrase` - (Optional)  - private key pass phrase 

## Attribute Reference

The following attributes are exported on this resource:

* `id` - The added to library certificate ID

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing certificate from library can be [imported][docs-import] into this resource
via supplying the full dot separated path certificate in library. An example is
below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_certificate_in_library.imported my-org.my-certificate-alias
```

The above would import the certificate named `my-certificate-alias` which is configured in organization named `my-org`.
