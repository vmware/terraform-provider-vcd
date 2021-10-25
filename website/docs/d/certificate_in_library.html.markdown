**---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_certificate_in_library"
sidebar_current: "docs-vcd-data-source-certificate-in-library"
description: |-
Provides a data source to read certificate in System or Org library.
---

# vcd\_certificate\_in\_library
Supported in provider *v3.5+* and VCD 10.2+.

Provides a data source to read certificate in System or Org library and reference in other resources.
~> Only `System Administrator` can access System certificates using this data source.

## Example Usage

```hcl
data "vcd_certificate_in_library" "certificate1" {
  org   = "myOrg"
  alias = "SAML Encryption"
}
```

## Argument Reference

The following arguments are supported:

* `alias` - (Optional)  - alias(name) of certificate
* `id` - (Optional)  - id of certificate

`alias` or `id` is required field.

## Attribute Reference

All the arguments and attributes defined in
[`vcd_certificate_in_library`](/providers/vmware/vcd/latest/docs/resources/vcd_certificate_in_library) resource are available.