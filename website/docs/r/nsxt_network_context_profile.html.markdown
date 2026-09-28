---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_network_context_profile"
sidebar_current: "docs-vcd-resource-nsxt-network-context-profile"
description: |-
  Provides a resource to manage custom NSX-T Network Context Profiles. Network Context Profiles
  carry Layer 7 attributes (App IDs, FQDNs) and can be referenced in Distributed Firewall rules.
  In addition to the default SYSTEM profiles synchronized from NSX-T, custom profiles can be
  defined with PROVIDER or TENANT scope.
---

# vcd\_nsxt\_network\_context\_profile

Supported in provider *v4.0+* and VCD 10.2+ with NSX-T backed VDC Groups.

Provides a resource to manage custom NSX-T Network Context Profiles. Network Context Profiles
carry Layer 7 attributes (App IDs, FQDNs) and can be referenced in Distributed Firewall rules
(`vcd_nsxt_distributed_firewall`). In addition to the default `SYSTEM` profiles synchronized from
NSX-T, custom profiles can be defined with `PROVIDER` or `TENANT` scope.

-> `SYSTEM` scoped profiles are built-in and cannot be created or modified. Use the
[`vcd_nsxt_network_context_profile`](/providers/vmware/vcd/latest/docs/data-sources/nsxt_network_context_profile)
data source to reference them.

## Example Usage (FQDN based profile for a VDC Group)

```hcl
data "vcd_vdc_group" "g1" {
  org  = "my-org"
  name = "my-vdc-group"
}

resource "vcd_nsxt_network_context_profile" "fqdn-filter" {
  org        = "my-org"
  context_id = data.vcd_vdc_group.g1.id

  name        = "office-365-endpoints"
  description = "Matches Microsoft 365 FQDNs known to NSX-T"
  scope       = "TENANT"

  attribute {
    type   = "DOMAIN_NAME"
    values = ["*.outlook.office.com", "smtp.office365.com"]
  }
}
```

## Example Usage (App ID profile with TLS sub-attributes)

```hcl
resource "vcd_nsxt_network_context_profile" "modern-tls" {
  org        = "my-org"
  context_id = data.vcd_vdc_group.g1.id

  name  = "ssl-modern-versions"
  scope = "TENANT"

  attribute {
    type   = "APP_ID"
    values = ["SSL"]

    sub_attribute {
      type   = "TLS_VERSION"
      values = ["TLS_V12", "TLS_V13"]
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level.
* `context_id` - (Required) ID of VDC, VDC Group, or NSX-T Manager. `TENANT` scoped profiles use a
  VDC or VDC Group ID, `PROVIDER` scoped ones an NSX-T Manager ID.
* `name` - (Required) A unique name for the Network Context Profile.
* `description` - (Optional) Description of the Network Context Profile.
* `scope` - (Optional) `TENANT` (default) or `PROVIDER`. `PROVIDER` scope requires System
  administrator privileges and `org = "System"`. Changing this forces a new resource.
* `attribute` - (Required) At least one attribute block, at most one per `type`. See
  [Attribute](#attribute) below.

<a id="attribute"></a>

## Attribute

* `type` - (Required) `APP_ID` or `DOMAIN_NAME`.
* `values` - (Required) A set of values:
  * For `APP_ID` - App IDs known to the backing NSX-T environment (e.g. `HTTP`, `SSL`, `CIFS`).
    ALG type App IDs (`FTP`, `TFTP`, `DCERPC`, `ORACLE`, `SUNRPC`) must be the only value of the
    attribute.
  * For `DOMAIN_NAME` - FQDNs known to the backing NSX-T environment (e.g. `*.office365.com`).
* `sub_attribute` - (Optional) Sub-attribute blocks. Only supported for `APP_ID` attributes with a
  single value (e.g. `SSL` or `CIFS`). See [Sub-attribute](#sub-attribute) below.

-> The backing NSX-T environment validates attribute values. The list of valid App IDs, FQDNs and
sub-attributes for a given context can be retrieved from the VCD API endpoint
`cloudapi/1.0.0/networkContextProfiles/attributes?filter=_context==<context-id>`.

<a id="sub-attribute"></a>

## Sub-attribute

* `type` - (Required) Sub-attribute type, one of `TLS_VERSION`, `TLS_CIPHER_SUITE`,
  `CIFS_SMB_VERSION`.
* `values` - (Required) A set of sub-attribute values (e.g. `TLS_V12`, `TLS_V13` for
  `TLS_VERSION`).

## Attribute Reference

The following attributes are exported on this resource:

* `id` - ID of the Network Context Profile

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing Network Context Profile can be [imported][docs-import] into this resource via
supplying its path. The path depends on profile scope:

* `PROVIDER` scoped profiles require NSX-T Manager name and profile name
* `TENANT` scoped profiles require Org name, VDC or VDC Group name, and profile name

```
terraform import vcd_nsxt_network_context_profile.imported nsxt-manager-name.profile-name
```

```
terraform import vcd_nsxt_network_context_profile.imported org-name.vdc-or-vdc-group-name.profile-name
```

NOTE: the default separator (.) can be changed using provider's `import_separator` or variable `VCD_IMPORT_SEPARATOR`

[docs-import]: https://www.terraform.io/docs/import/
