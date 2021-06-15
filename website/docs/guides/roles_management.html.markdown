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
different resource.

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
* use a data source of `vcd_resource_list` to show the rights available to a given organization.

A right can have a list of **implied rights**. When such list exists, it means that, in addition to the main right, you must
include all the implied rights to the rights container. If you don't include the implied rights, you will get an error, listing
all the rights that are missing from your entity.


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

* [Predefined Roles and Their Rights](https://docs.vmware.com/en/VMware-Cloud-Director/9.5/com.vmware.vcloud.spportal.doc/GUID-BC504F6B-3D38-4F25-AACF-ED584063754F.html)
* [VMware Cloud Director – Simple Rights Management with Bundles](https://blogs.vmware.com/cloudprovider/2019/12/effective-rights-bundles.html)