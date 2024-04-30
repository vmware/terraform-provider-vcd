---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_org_oidc"
sidebar_current: "docs-vcd-resource-org-oidc"
description: |-
  Provides a VMware Cloud Director Organization OIDC resource. This can be used to create, delete, and update the OIDC configuration for an Organization.
---

# vcd\_org\_oidc

Provides a VMware Cloud Director Organization OIDC resource. This can be used to create, update, and delete OIDC configuration for an Organization.

Supported in provider *v3.13+*

## Example Usage

```hcl
data "vcd_org" "my_org" {
  name = "my-org"
}

resource "vcd_org_oidc" "my-org-oidc" {
  org_id = data.vcd_org.my_org.id
}
```

## Argument Reference

The following arguments are supported:

* `org_id` - (Required) Since there is only one OIDC configuration available for an organization, the resource can be identified by the Org itself
* `client_id` - (Required) Client ID to use with the OIDC provider
* `client_secret` - (Required) Client Secret to use with the OIDC provider
* `enabled` - (Required) Either `true` or `false`, specifies whether the OIDC authentication is enabled for the given organization
* `wellknown_endpoint` - (Optional) An endpoint that can be set to automatically retrieve the OIDC provider configuration and set
  the following arguments without setting them explicitly in HCL: `issuer_id`, `user_authorization_endpoint`, `access_token_endpoint`, 
  `userinfo_endpoint`, the `claims_mapping` block, any `key` block, and `scopes`. These mentioned attributes will be computed, and
  can be overridden by setting them explicitly in HCL configuration
* `issuer_id` - (Optional) The issuer ID for the OIDC provider.
  If `wellknown_endpoint` is **not** set, then this argument is **required**. Otherwise, it is **optional**.
  This allows administrators to override the configuration given by `wellknown_endpoint`
* `user_authorization_endpoint` - (Optional) The issuer ID for the OIDC provider.
  If `wellknown_endpoint` is **not** set, then this argument is **required**. Otherwise, it is **optional**.
  This allows administrators to override the configuration given by `wellknown_endpoint`
* `access_token_endpoint` - (Optional) The endpoint to use for access tokens.
  If `wellknown_endpoint` is **not** set, then this argument is **required**. Otherwise, it is **optional**.
  This allows administrators to override the configuration given by `wellknown_endpoint`
* `userinfo_endpoint` - (Optional) The endpoint to use for User Info.
  If `wellknown_endpoint` is **not** set, then this argument is **required**. Otherwise, it is **optional**.
  This allows administrators to override the configuration given by `wellknown_endpoint`
* `prefer_id_token` - (Required) If you want to combine claims from `userinfo_endpoint` and the ID Token, set this to `true`.
  The identity providers do not provide all the required claims set in `userinfo_endpoint`. By setting this argument to `true`,
  VMware Cloud Director can fetch and consume claims from both sources
* `max_clock_skew_seconds` - (Optional) The maximum clock skew is the maximum allowable time difference between the client and server.
  This time compensates for any small-time differences in the timestamps when verifying tokens. The **default** value is `60` seconds.
* `scopes` - (Optional) A set of scopes to use with the OIDC provider. They are used to authorize access to user details,
  by defining the permissions that the access tokens have to access user information.
  If `wellknown_endpoint` is **not**  set, then this argument is **required**. Otherwise, it is **optional**. This allows administrators
  to override the scopes given by `wellknown_endpoint`. Setting `scopes = []` will make Terraform to set the scopes provided originally
  by the `wellknown_endpoint`
* `claims_mapping` - (Optional) A single configuration block that specifies the claim mappings to use with the OIDC provider.
  If `wellknown_endpoint` is **not**  set, then this argument is **required**. Otherwise, it is **optional**. This allows administrators
  to override the claims given by `wellknown_endpoint`. Setting `claims_mapping {}` will make Terraform to set the claims provided originally
  by the `wellknown_endpoint`. The supported claims are:
  * `email` - Required if `wellknown_endpoint` doesn't give info about it
  * `subject` - Required if `wellknown_endpoint` doesn't give info about it
  * `last_name` - Required if `wellknown_endpoint` doesn't give info about it
  * `first_name` - Required if `wellknown_endpoint` doesn't give info about it
  * `full_name` - Required if `wellknown_endpoint` doesn't give info about it
  * `groups` - Optional
  * `roles` - Optional
* `key` - (Optional) One or more configuration blocks that specify the keys to use with the OIDC provider.
  If `wellknown_endpoint` is **not**  set, then this argument is **required**. Otherwise, it is **optional**. This allows administrators
  to override the keys given by `wellknown_endpoint`. Setting `key {}` will make Terraform to set the keys provided originally
  by the `wellknown_endpoint`. Each key requires the following:
  * `id` - Identifier of the key
  * `algorithm` - Algorithm used by the key. Can be `RSA` or `EC`
  * `pem_file` - PEM file to create/update the key
  * `pem` - (Computed) This attribute is read-only, and contains the PEM contents after the configuration is available in VCD
  * `expiration_date` - Expiration date for the key. The accepted format is the same used by [`timestamp`](https://developer.hashicorp.com/terraform/language/functions/timestamp)

## Attribute Reference

* `redirect_uri` - The client configuration redirect URI used to create a client application registration with an identity provider
  that complies with the OpenID Connect standard

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing OIDC configuration for an Org can be [imported][docs-import] into this resource via supplying the path for an Org. Since the Org is
at the top of the VCD hierarchy, the path corresponds to the Org name.
For example, using this structure, representing an existing OIDC configuration that was **not** created using Terraform:

```hcl
data "vcd_org" "my_org" {
  name = "my-org"
}

resource "vcd_org_oidc" "my_org_oidc" {
  org_id = data.vcd_org.my_org.id
}
```

You can import such OIDC configuration into terraform state using one of the following commands

```
terraform import vcd_org_oidc.my_org_oidc organization_name
# OR
terraform import vcd_org_oidc.my_org_oidc organization_id
```

After that, you must expand the configuration file before you can either update or delete the OIDC configuration. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the stored properties.
