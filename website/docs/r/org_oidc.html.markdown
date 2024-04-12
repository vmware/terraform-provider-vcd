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
