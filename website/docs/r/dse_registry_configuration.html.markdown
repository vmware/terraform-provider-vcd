---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_dse_registry_configuration"
sidebar_current: "docs-vcd-resource-dse-registry-configuration"
description: |-
  Provides a resource to manage Data Solution Extension (DSE) registry configuration.
---

# vcd\_dse\_registry\_configuration

Supported in provider *v3.13+* with Data Solution Extension.

Provides a resource to manage Data Solution Extension (DSE) registry configuration.

~> Only `System Administrator` can create this resource.

## About Data Solution structure

There are 2 types of Data Solution configurations:
* Helm based
* Container based

All of Data Solutions provide default versions and repositories that can be used by setting
`use_default_value` flag that will set these defaults in configuration.

One can also set custom repositories, package names and their versions. For Helm based configuration
one should specify `chart_repository`, `version` and `package_name`. For container based
configuration, `package_repository` and `version` should be set.

Helm based configurations apply to:

* `MongoDB Community`
* `Confluent Platform`
* `MongoDB`


Container bases configurations apply to:

* `VMware SQL with Postgres`
* `VMware RabbitMQ`
* `VMware SQL with MySQL`
* `VCD Data Solutions`

Additionally, `VCD Data Solutions` provide [Container Registry Configuration ](#container-registry).

## Example Usage (Configure package repository)

```hcl
resource "vcd_dse_registry_configuration" "mongodb-community" {
  name             = "MongoDB Community"
  chart_repository = "https://mongodb.github.io/helm-charts"
  version          = "0.9.0"
  package_name     = "community-operator"

  depends_on = [vcd_solution_add_on_instance_publish.public]
}
```

## Example Usage (Configure chart repository)

```hcl
resource "vcd_dse_registry_configuration" "mysql" {
  name               = "VMware SQL with MySQL"
  package_repository = "registry.tanzu.vmware.com/packages-for-vmware-tanzu-data-services/tds-packages:1.13.0"
  version            = "1.10.1"

  depends_on = [vcd_solution_add_on_instance_publish.public]
}
```

## Example Usage (Configure container registries)

```hcl
resource "vcd_dse_registry_configuration" "dso" {
  name               = "VCD Data Solutions"
  # Using default versions for packages
  use_default_value = true

  container_registry {
    host        = "first-host.sample"
    description = "host2"
    username    = "user1"
    password    = "pass1"
  }

  container_registry {
    host        = "another-host.sample"
    description = "Test Host that does not work"
    username    = "user2"
    password    = "pass2"
  }

  depends_on = [vcd_solution_add_on_instance_publish.public]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of Data Solution as it appears in repository configuration
* `use_default_value` - (Optional) Defines if repository settings should be inherited from Data
  Solution itself. Default `false`
* `version` - (Optional) Version of package to use. Required when `use_default_value` is not used.
* `package_repository` - (Optional) Package repository for container based images
* `chart_repository` - (Optional) Chart repository for Helm based images
* `package_name` - (Optional) Helm package name. Only for Helm based images
* `container_registry` - (Optional) Only applies to `VCD Data Solutions` configuration. Specifies
  credentials that can be used to authenticate to repositories. See [Container Registry
  Configuration ](#container-registry)


<a id="container-registry"></a>
## Container Registry Configuration 

* `host` - (Required) Host of container registry
* `description` - (Required) Description of container registry entry
* `username` - (Required) Username for authentication
* `password` - (Required) Password for authentication



## Attribute Reference

The following attributes are exported on this resource:

* `type` - Type of repository settings. It can be one of `PackageRepository`, `ChartRepository`
* `default_package_name` - Default package name as provided by Data Solution
* `default_version` - Default package version as provided by Data Solution
* `default_chart_repository` - Default chart repository as provided by Data Solution
* `default_package_name` - Default package name as provided by Data Solution
* `default_repository` - Default container repository as provided by Data Solution
* `compatible_version_constraints` - A set of version constrains that this Data Solution defines
* `requires_version_compatibility` - Boolean flag as defined in Data Solution
* `rde_state` - State of parent Runtime Defined Entity (RDE)

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing Data Solution registry configuration can be [imported][docs-import] into this resource
via supplying its name. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_dse_registry_configuration.imported "VCD Data Solutions"
```

The above would import the `VCD Data Solutions` Data Solution registry configuration.
