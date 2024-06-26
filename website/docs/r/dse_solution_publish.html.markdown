---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_dse_solution_publish"
sidebar_current: "docs-vcd-resource-dse-solution-publish"
description: |-
  Provides a resource to manage Data Solution Extension (DSE) publishing settings.
---

# vcd\_dse\_registry\_configuration

Supported in provider *v3.13+* and VCD 10.5.0+ with Data Solution Extension.

Provides a resource to manage Data Solution Extension (DSE) registry configuration.

~> Only `System Administrator` can create this resource.

## Example Usage

```hcl
resource "vcd_dse_solution_publish" "mongodb-community" {
  data_solution_id = data.vcd_dse_registry_configuration.mongodb-community.id

  org_id = data.vcd_org.tenant-org.id
}

data "vcd_dse_registry_configuration" "mongodb-community" {
  name = "MongoDB Community"
}

data "vcd_org" "tenant-org" {
  name = "tenant_org"
}
```

## Example Usage (Confluent Platform with Licensing)

```hcl
resource "vcd_dse_solution_publish" "confluent-platform" {
  data_solution_id = data.vcd_dse_registry_configuration.confluent-platform.id

  confluent_license_type = "With License"
  confluent_license_key  = "XXXXXXXXXX"
  
  org_id = data.vcd_org.tenant-org.id
}

data "vcd_dse_registry_configuration" "confluent-platform" {
  name = "Confluent Platform"
}

data "vcd_org" "tenant-org" {
  name = "tenant_org"
}
```

## Argument Reference

The following arguments are supported:

* `data_solution_id` - (Required) ID of Data Solution
* `org_id` - (Required) Organization ID
* `confluent_license_type` - (Optional) Required for `Confluent Platform` Data Solution. One of
  `With License`, `No License`.
* `confluent_license_key` - (Optional) Required for `Confluent Platform` Data Solution if . One of
  `confluent_license_type` is set to `With License`. 

## Attribute Reference

The following attributes are exported on this resource:

* `dso_acl_id` - Data Solutions Operator ACL ID
* `template_acl_ids` - A set of Data Solution Instance Template ACL IDs
* `ds_org_config_id` - Data Solution Org Configuration ID (only available for `Confluent Platform`
  which has additional licensing configuration)

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing Data Solution publishing configuration can be [imported][docs-import] into this resource
via supplying its name and tenant. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_dse_solution_publish.imported "MongoDB Community".tenant_org
```

The above would import the `MongoDB Community` Data Solution publishing configuration for
`tenant_org`.
