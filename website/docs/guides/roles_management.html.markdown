---
layout: "vcd"
page_title: "VMware Cloud Director: roles management"
sidebar_current: "docs-vcd-guides-roles"
description: |-
 Provides guidance to VMware Cloud roles management
---

# Roles management

Supported in provider *v3.3+*.

-> In this document, when we mention **tenants**, the term can be substituted with **organizations**.

## Overview

Roles management is a combination of four entities:

* **Rights**: read-only entities, available to both provider and tenants.
* **Roles**: a container of rights that defines the privileges that can be assigned to a user. It is available to both provider and tenants.
* **Global Role**: are blueprints for roles, created in the provider, which become available as _Roles_ in the tenant.
* **Rights Bundles**: are collections of rights that define which rights become available to one or more tenants.

There are similarities among Roles, Global Roles, and Rights Bundles: all three are collections of rights for different
purposes. The similarity is in the way we create and modify these resources. We can add and remove rights to obtain a
different resource. For the purpose of describing their common functionalities, we can call these three entities **Rights Containers**.

There are also similarities between Global Roles and Rights Bundles: both resources need to be published to one or more
tenants in order to be effective. Both can become isolated if we remove all tenants, or can be maximized if we publish
to all tenants without providing a list. In this later case, the resource will be also published to future tenants.

## Rights

**Rights**([`vcd_right`](/docs/providers/vcd/d/right.html)) are available as data sources. They can't be created by either provider or tenants.
They are building blocks for the other three entities (Roles, Global Roles, Rights Bundles), and can be used by simply
stating their name within the containing entity. You can also use data sources, but it would make for a crowded HCL
script, and would also increase the amount of computing needed to run a script. 

To see the list of available rights, you can do one of the following:

* make a data source of several existing Roles, Global Roles, or Rights Bundles, and use an `output` structure to show the contents;
* use a data source of [`vcd_resource_list`](/docs/providers/vcd/d/vcd_resource_list.html) to show the rights available to a given organization.

Examples:

```hcl
data "vcd_role" "vapp-author" {
  name = "vApp Author"
}

output "vapp-author" {
  value = data.vcd_role.vapp-author
}

data "vcd_respurce_list" "rights-list" {
  name          = "rights-list"
  resource_type = "rights"
}

output "rights-list" {
 value = data.vcd_respurce_list.rights-list
}
```

A right can have a list of **implied rights**. When such list exists, it means that, in addition to the main right, **you must
include all the implied rights** to the rights container (role, global role, rights bundle). If you don't include the
implied rights, you will get an error, listing all the rights that are missing from your entity.


## Roles

A **Role** ([`vcd_role`](/docs/providers/vcd/r/role.html)) is a set of rights that can be assigned to a user. When choosing a role for a user, we see a list of predefined
roles that are available to the organization. That list is the result of the **Global Roles** defined by the provider
and published to the tenant we are using, in addition to the roles that were created by the organization administrator.
As such, roles always belong to an organization. To define or use a role at provider level, we use the "System" organization.

## Global Roles

A **Global Role** ([`vcd_global_role`](/docs/providers/vcd/r/global_role.html)) is a definition of a role that is _published_ to one or more tenants, which in turn will see such global
roles converted into the roles they can use.
Provider can add, modify, and delete global roles. They can also alter the list of publication for each global role, to
make them available to a selected set of tenants.

## Rights Bundles

A **Rights Bundle** ([`vcd_rights_bundle`](/docs/providers/vcd/r/rights_bundle.html)) is a set of rights that can be made available to tenants. While global roles define tenant roles, a
rights bundle define which rights, independently of a global role listing, can be given to one or more tenants.

An example is necessary to understand the concept.
Let's say that, as a provider, you change the publishing of the rights bundle `Default Rights Bundle` and restrict its
usage to a single tenant (called `first-org`). Then, you create another rights bundle, similar to `Default Rights Bundle`, 
but with only _view_ rights, and publish this bundle to another tenant (`second-org`). With this change, an Org administrator
in `first-org` will see the usual roles, with the usual sets of rights. The Org administrator in `second-role`, meanwhile,
will see the same roles, but with only half the rights, as the _managing_ rights will be missing. While this is an extreme
example, it serves to illustrate the function of rights bundles. You can create general purpose global roles for several
tenants, and then limit their reach by adding or removing rights to the rights bundle governing different tenants.


## How to include rights and implied rights into a rights container

Adding rights to one of the rights containers (Role, Global Role, Rights Bundle) is a comparable operation that works
by the same principles:

* You add an array of rights names to the `rights` field of the entity;
* If you get an error about missing implied rights, you add them to the list

For example, lets say, for the sake of simplicity, that you want to create a role with just two rights, as listed below:

```hcl
resource "vcd_role" "new-role" {
  org         = "datacloud"
  name        = "new-role"
  description = "new role"
  rights = [
    "Catalog: Add vApp from My Cloud",
    "Catalog: Edit Properties",
  ]
}
```

When you run `terraform apply`, you get this error:

```
vcd_role.new-role: Creating...
╷
│ Error: The rights set for this role require the following implied rights to be added:
│ "vApp Template / Media: Edit",
│ "vApp Template / Media: View",
│ "Catalog: View Private and Shared Catalogs",
│
│
│   with vcd_role.new-role,
│   on config.tf line 91, in resource "vcd_role" "new-role":
│   91: resource "vcd_role" "new-role" {
│
```
Thus, you update the script to include the rights mentioned in the error message

```hcl
resource "vcd_role" "new-role" {
  org         = "datacloud"
  name        = "new-role"
  description = "new role"
  rights = [
    "Catalog: Add vApp from My Cloud",
    "Catalog: Edit Properties",
    "vApp Template / Media: Edit",
    "vApp Template / Media: View",
    "Catalog: View Private and Shared Catalogs",
  ]
}
```

Then repeat `terraform apply`. This time the operation succeeds.

The corresponding structure for global role and rights bundle are almost the same. You just need to add the tenants
management fields.

```hcl
resource "vcd_global_role" "new-global-role" {
  name        = "new-global-role"
  description = "new global role"
  rights = [
    "Catalog: Add vApp from My Cloud",
    "Catalog: Edit Properties",
    "vApp Template / Media: Edit",
    "vApp Template / Media: View",
    "Catalog: View Private and Shared Catalogs",
  ]
  publish_to_all_tenants = true
}

resource "vcd_rights_bundle" "new-rights-bundle" {
 name        = "new-rights-bundle"
 description = "new rights bundle"
 rights = [
  "Catalog: Add vApp from My Cloud",
  "Catalog: Edit Properties",
  "vApp Template / Media: Edit",
  "vApp Template / Media: View",
  "Catalog: View Private and Shared Catalogs",
 ]
 publish_to_all_tenants = true
}
```

## Tenant management

Rights Bundle and Global Roles have a `tenants` section where you can list to which tenants the resource should be
published, meaning which tenants can feel the effects of this resource.

There are two fields related to managing tenants:

* `publish_to_all_tenants` with value "true" or "false".
    * If true, the resource will be published to all tenants, even if they don't exist yet. All future organizations will get to feel the benefits or restrictions published by the resource
    * If false, then we take into account the `tenants` field.
* `tenants` is a list of organizations (tenants) to which we want the effects of this resource to apply.

Examples:

```hcl
resource "vcd_global_role" "new-global-role" {
  name        = "new-global-role"
  description = "new global role"
  rights = [ /* rights list goes here */ ]
  publish_to_all_tenants = true
}
```
This global role will be published to all tenants, including the ones that will be created after this resource.

Now we modify it:

```hcl
resource "vcd_global_role" "new-global-role" {
  name        = "new-global-role"
  description = "new global role"
  rights = [ /* rights list goes here */ ]
  publish_to_all_tenants = false
  tenants = [ "org1", "org2" ]
```

The effects of this global role are only propagated to `org1` and `org2`. Other organizations cease to see the role that
was instantiated by thsi global role.

Let's do another change:

```hcl
resource "vcd_global_role" "new-global-role" {
  name        = "new-global-role"
  description = "new global role"
  rights = [ /* rights list goes here */ ]
  publish_to_all_tenants = false
```

The `tenants` field is removed, meaning that we don't publish to anyone. And since `publish_to_all_tenants` is false,
the tenants previously in the list are removed from publishing, making the global role isolated. It won't have
any effect on any organization until we update its tenants list.

## How to change an existing rights container

If you want to modify a Role, Global Role, or Rights Bundle that is already in your system, you need first to import
it into Terraform state, and only then you can apply your changes.

Let's say, for example, that you want to change a rights bundle `Default Rights Bundle`, to publish it only to a limited
set of tenants, while you will create a separate rights bundle for other tenants that need a different set of rights.

The import procedure works in three steps:

(1)<br>
Create a data source for the rights bundle, and a resource that takes all its attributes from the data source:

```hcl

data "vcd_rights_bundle" "old-rb" {
  name = "Default Rights Bundle"
}

resource "vcd_rights_bundle" "new-rb" {
  name                   = "Default Rights Bundle"
  rights                 = data.vcd_rights_bundle.old-rb.rights
  tenants                = [ "first-org" ]
  publish_to_all_tenants = false
}
```

Using the data source will free you from the need of listing all the rights contained in the bundle (113 in VCD 10.2).
It will also make the script work across different versions, where the list of rights may differ. If you were interested
in changing the rights themselves, you could add an `output` block for the data source, copy the rights to the resource
definition, and then remove or add what you need.

(2)<br>
Import the rights bundle into terraform:

```
$ terraform import vcd_rights_bundle.new-rb "Default Rights Bundle"
```

(3)<br>
Now you can run `terraform apply`, which will remove the default condition of "publish to all tenants", replacing it
with "publish to a single tenant".


## References

* [Managing Rights and Roles](https://docs.vmware.com/en/VMware-Cloud-Director/10.2/VMware-Cloud-Director-Service-Provider-Admin-Portal-Guide/GUID-816FBBBC-2CDA-4B1D-9B1A-C22BC31B46F2.html)
* [VMware Cloud Director – Simple Rights Management with Bundles](https://blogs.vmware.com/cloudprovider/2019/12/effective-rights-bundles.html)