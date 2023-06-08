---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_org_saml"
sidebar_current: "docs-vcd-resource-org-saml"
description: |-
  Provides a VMware Cloud Director Organization SAML resource. This can be used to create, delete, and update SAML configuration for an organization .
---

# vcd\_org\_saml

Provides a VMware Cloud Director Org SAML resource. This can be used to create, update, and delete SAML configuration for an organization.

Supported in provider *v3.10+*

-> **Note:** This resource requires system administrator privileges.

## Example Usage

```hcl
data "vcd_org" "my-org" {
  name = "my-org"
}

resource "vcd_org_saml" "my-org-saml" {
  org_id                          = data.vcd_org.my-org.id
  enabled                         = true
  entity_id                       = "my-entity"
  identity_provider_metadata_file = "idp-metadata.xml"
}
```

## Argument Reference

The following arguments are supported:

* `org_id` - (Required) Org ID: there is only one SAML configuration available for an organization. Thus, the resource can be identified by the Org.
* `enabled` - (Required) If true, the organization will use SAML for authentication.
* `entity_id` - (Optional) Your service provider entity ID. Once you set this field, it cannot be changed back to empty.
* `identity_provider_metadata_file` - (Required) Name of a file containing the metadata text from a SAML Identity Provider.

## Identity provider metadata

To configure a SAML service, we need to provide the metadata from an identity provider.
An easy way of getting such metadata into a file is by using Terraform itself:

```hcl
data "http" "example" {
  url = "https://samltest.id/saml/idp"
}

output "result" {
  value = data.http.example.response_body
}
```
To save into a file, we can run the following:

```
$ terraform apply -auto-approve
$ terraform output -raw result > data.xml
```

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing SAML configuration for an Org can be [imported][docs-import] into this resource via supplying the path for an Org. Since the Org is
at the top of the VCD hierarchy, the path corresponds to the Org name.
For example, using this structure, representing an existing SAML configuration that was **not** created using Terraform:

```hcl
data "vcd_org" "my-org" {
  name = "my-org"
}

resource "vcd_org_saml" "my-org-saml" {
  org_id                          = data.vcd_org.my-org.id
  enabled                         = true
  identity_provider_metadata_file = "somefile.xml"
}
```

You can import such SAML configuration into terraform state using one of the following commands

```
# EITHER
terraform import vcd_org_saml.my-org-saml organization_name
# OR
terraform import vcd_org_saml.my-org-saml organization_id
```

After that, you must expand the configuration file before you can either update or delete the SAML configuration. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the stored properties.
