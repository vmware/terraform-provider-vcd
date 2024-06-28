---
layout: "vcd"
page_title: "VMware Cloud Director: Data Solution Extension guide"
sidebar_current: "docs-vcd-guides-data-solution-extension"
description: |-
 Provides guidance to Cloud Director Solution Landing Zone, Solution Add-On and Data Solution
 Extension management using Terraform
---

# About

Supported in provider *v3.13+*.

This is a guide page that introduces Terraform users to managing Solution Landing Zone, Solution
Add-Ons and Data Solution Extension (DSE) Add-On.

## Solution Landing Zone and Solution Add-Ons

Solution Add-Ons extend Cloud Director offering with value-added functionalities. One can manage
multiple Solution Add-Ons within a Cloud Director Solution Landing Zone.

The Solution Add-Ons come packed as `.iso` files and Terraform Provider for VCD v3.13+ is capable of
leveraging them to configure Solution Add-Ons within VCD.

*Note:* For a more hands-on experience, one can check [DSE deployment
examples](https://github.com/vmware/terraform-provider-vcd/tree/main/examples/data-solution-extension/).

## Data Solution Extension (DSE)

Data Solution Extension is one of the Solution Add-Ons available for VCD. It provides capability to
extend VCD and deliver a portfolio of on-demand caching, messaging and database software.

Terraform provider VCD v3.13 added initial support for configuring DSE and publishing it to tenants.

## Terraform resources and data sources

Terraform provider VCD v3.13 adds support for Solution Landing Zone, Solution Add-On management and
Data Solution Extension configuration resources with their respective data sources.

### Solution Landing Zone and Add-On resources

* [`vcd_solution_landing_zone`](/providers/vmware/vcd/latest/docs/resources/solution_landing_zone)
* [`vcd_solution_add_on`](/providers/vmware/vcd/latest/docs/resources/solution_add_on)
* [`vcd_solution_add_on_instance`](/providers/vmware/vcd/latest/docs/resources/solution_add_on_instance)

### Data Solution Extension resources

* [`vcd_dse_registry_configuration`](/providers/vmware/vcd/latest/docs/resources/dse_registry_configuration)
* [`vcd_dse_solution_publish`](/providers/vmware/vcd/latest/docs/resources/dse_solution_publish)

### Rights management resources

Additionally, after deploying a Solution Add-On, one can leverage resources and data sources for role
management to provision access to new Add-On features:

* [`vcd_rights_bundle`](/providers/vmware/vcd/latest/docs/resources/rights_bundle)
* [`vcd_global_role`](/providers/vmware/vcd/latest/docs/resources/global_role)

[Read more about role and rights management.](https://registry.terraform.io/providers/vmware/vcd/latest/docs/guides/roles_management)

## Solution Landing Zone configuration (Step 1)

The first step for deploying a Solution Add-On is to have a configured Solution Landing Zone and
[`vcd_solution_landing_zone`](/providers/vmware/vcd/latest/docs/resources/solution_landing_zone)
does that. It requires specifying an Organization, Catalog, VDC, Routed Org VDC network and both -
Storage and Compute policies. There can be only *one Solution Landing Zone per VCD*.

```hcl
resource "vcd_catalog" "solution_add_ons" {
  org = var.vcd_solutions_org

  name             = "solution_add_ons"
  description      = "Catalog host Data Solution Add-Ons"
  delete_recursive = true
  delete_force     = true
}

data "vcd_org_vdc" "solutions_vdc" {
  org  = var.vcd_solutions_org
  name = var.vcd_solutions_vdc
}

data "vcd_network_routed_v2" "solutions" {
  org  = var.vcd_solutions_org
  vdc  = var.vcd_solutions_vdc
  name = var.vcd_solutions_vdc_routed_network
}

data "vcd_storage_profile" "solutions" {
  org  = var.vcd_solutions_org
  vdc  = var.vcd_solutions_vdc
  name = var.vcd_solutions_vdc_storage_profile_name
}

resource "vcd_solution_landing_zone" "slz" {
  org = var.vcd_solutions_org

  catalog {
    id = vcd_catalog.solution_add_ons.id
  }

  vdc {
    id         = data.vcd_org_vdc.solutions_vdc.id
    is_default = true

    org_vdc_network {
      id         = data.vcd_network_routed_v2.solutions.id
      is_default = true
    }

    compute_policy {
      id         = data.vcd_org_vdc.solutions_vdc.default_compute_policy_id
      is_default = true
    }

    storage_policy {
      id         = data.vcd_storage_profile.solutions.id
      is_default = true
    }
  }
}
```

## Solution Add-On configuration (Step 2)

Once the Solution Landing Zone is set up, the next step is creating a Solution Add-On. This requires
having a [Solution Add-On `.iso`
file](https://docs.vmware.com/en/VMware-Cloud-Director/10.6/VMware-Cloud-Director-Service-Provider-Admin-Guide/GUID-8633A240-DC05-438B-96C1-4A3762B55ED0.html).
Due to the deployment process, Solution Add-On `.iso` image must be present both - locally and in
the catalog defined in Solution Landing Zones.

Each Solution Add-On image file contains a certificate that must be trusted so that a Solution
Add-On can be used. To do that automatically, one can leverage `auto_trust_certificate` within
[`vcd_solution_add_on`](/providers/vmware/vcd/latest/docs/resources/solution_add_on) resource.

```hcl
resource "vcd_catalog_media" "dse14" {
  org        = var.vcd_solutions_org
  catalog_id = vcd_catalog.solution_add_ons.id

  name              = basename(var.vcd_dse_add_on_iso_path)
  description       = "DSE Solution Add-On"
  media_path        = var.vcd_dse_add_on_iso_path
  upload_piece_size = 10
}

resource "vcd_solution_add_on" "dse14" {
  catalog_item_id        = data.vcd_catalog_media.dse14.catalog_item_id
  add_on_path            = var.vcd_dse_add_on_iso_path
  auto_trust_certificate = true

  depends_on = [vcd_solution_landing_zone.slz]
}
```

## Solution Add-On instantiation (Step 3)

After deployment, the Solution Add-On must be instantiated with correct `input` parameters. More
details about setting `input` and `delete_input` values below.

EULA must be accepted with `accept_eula` field. If it isn't - instantiating an add-on will fail
with an error message that contains EULA.

```hcl
resource "vcd_solution_add_on_instance" "dse14" {
  add_on_id   = vcd_solution_add_on.dse14.id
  accept_eula = true
  name        = "dse-14"

  input = {
    delete-previous-uiplugin-versions = true
  }

  delete_input = {
    force-delete = true
  }
}
```

### About dynamic Solution Add-On instantiation input validation

Each Solution Add-On comes with its own input values used for instantiation and removal. UI
renders these values as an input form. It is not that trivial to provide such option for CLI
applications, like Terraform. Terraform provider VCD attempts to present as much convenience as
possible by providing dynamic input validation in
[`vcd_solution_add_on_instance`](/providers/vmware/vcd/latest/docs/resources/solution_add_on_instance)
resource.

It works by reading the provided input schema of a Solution Add-On and dynamically validating
(during `apply` operation) if the provided inputs match the requested ones. If they don't - it will
print all the missing inputs with an error message that contains details for each of the missing
fields (example below).

In the printed error message, each field has an `IsDelete` flag which defines whether it should be
specified in `input` or `delete_input` value in
[`vcd_solution_add_on_instance`](/providers/vmware/vcd/latest/docs/resources/solution_add_on_instance)
resource.

All fields also have a `Required` flag which hints if they are mandatory or not. By default,
[`vcd_solution_add_on_instance`](/providers/vmware/vcd/latest/docs/resources/solution_add_on_instance)
resource requires providing all `input` and `delete_input` values. If one doesn't want to specify
some of the non-mandatory fields, it is possible to disable validation for the non required fields by
setting `validate_only_required_inputs = true`.

-> The `delete_input` fields are validated during removal (`destroy` operation). It may occur
that these values have to be adjusted during removal phase - for that reason it is safe to update
`delete_input`, perform `apply` (update operation will be no-op) and then retry `destroy` operation.


```shell
...
╷
│ Error: dynamic creation input field validation error: 
│ -----------------
│ Field: delete-previous-uiplugin-versions
│ Title: Delete Previous UI Plugin Versions
│ Type: Boolean
│ Required: true
│ IsDelete: false
│ Description: If setting true, the installation will delete all previous versions of this ui plugin. If setting false, the installation will just disable previous versions
│ Default: false
│ -----------------
│
│ 
│ ERROR: Missing fields 'delete-previous-uiplugin-versions' for Solution Add-On 'vmware.ds-1.4.0-23376809'
│ 
│   with vcd_solution_add_on_instance.dse14,
│   on vcd.TestAccSolutionAddonInstanceAndPublishingstep1.tf line 96, in resource "vcd_solution_add_on_instance" "dse14":
│   96: resource "vcd_solution_add_on_instance" "dse14" {
...
```

## Publishing a Solution Add-On Instance (Step 4)

The last step for making the Solution Add-On available is publishing it to tenants.

```hcl
resource "vcd_solution_add_on_instance_publish" "public" {
  add_on_instance_id     = vcd_solution_add_on_instance.dse14.id
  org_ids                = [data.vcd_org.recipient.id]
  publish_to_all_tenants = false
}
```

~> Clients must logout and login back to VCD so that newly published Solution Add-On can
be managed.

## Configuring Data Solution Extension (DSE) and publishing Data Solutions (Step 5)

Once DSE is deployed, the first step for a provider is to configure registry information for each
Data Solution. Below is a minimized example that takes default registry values that come with Data
Solution itself, but [resource
docs](/providers/vmware/vcd/latest/docs/resources/dse_registry_configuration) have examples how to
set up custom values.

The last step of Data Solution configuration is publishing it to a given tenant.

```hcl
resource "vcd_dse_registry_configuration" "dso" {
  name               = "VCD Data Solutions"
  use_default_values = true
}

resource "vcd_dse_registry_configuration" "mongodb-community" {
  name               = "MongoDB Community"
  use_default_values = true
}

# Publish Mongo DB Data Solution to tenant
resource "vcd_dse_solution_publish" "mongodb-community" {
  data_solution_id = vcd_dse_registry_configuration.mongodb-community.id

  org_id = data.vcd_org.dse-consumer.id
}
```

## Creating new tenant user with required rights (Step 6)

Solutions Add-On brings additional rights to VCD. Usually, to leverage new functionalities introduced
by a Solution Add-On, one should have those new rights. This functionality has been long present in
Terraform provider VCD, but this is just a tiny example on how one can combine multiples rights
bundles to create a new role and user.

Read more about [roles and rights in a designated guide page](https://registry.terraform.io/providers/vmware/vcd/latest/docs/guides/roles_management).

```hcl
data "vcd_rights_bundle" "dse-rb" {
  name = "vmware:dataSolutionsRightsBundle"
}

data "vcd_rights_bundle" "k8s-rights" {
  name = "Kubernetes Clusters Rights Bundle"
}

resource "vcd_global_role" "dse" {
  name                   = "DSE Role"
  description            = "Global role for consuming DSE"
  rights                 = setunion(data.vcd_rights_bundle.k8s-rights.rights, data.vcd_rights_bundle.dse-rb.rights)
  publish_to_all_tenants = false
  tenants = [
    data.vcd_org.dse-consumer.name
  ]
}

resource "vcd_org_user" "my-org-admin" {
  org = data.vcd_org.dse-consumer.name

  name        = var.vcd_tenant_user
  description = "DSE User"
  role        = vcd_global_role.dse.name
  password    = var.vcd_tenant_password

  depends_on = [vcd_global_role.dse]
}
```

After executing this last step - one should be able to login to tenant Organization with newly
created user and find Data Solution *"MongoDB Community"* available.

## References

* [Deployment HCL example in Terraform provider VCD
  repository](https://github.com/vmware/terraform-provider-vcd/tree/main/examples/data-solution-extension/)
* [Roles and Rights guide for Terraform provider VCD](https://registry.terraform.io/providers/vmware/vcd/latest/docs/guides/roles_management)
* [Official Solution Add-On documentation](https://docs.vmware.com/en/VMware-Cloud-Director/10.5/VMware-Cloud-Director-Service-Provider-Admin-Guide/GUID-4F12C8F7-7CD3-44E8-9711-A5F43F8DCEB5.html)
* [Data Solution Extension documentation](https://www.vmware.com/products/cloud-director/data-solutions.html)
