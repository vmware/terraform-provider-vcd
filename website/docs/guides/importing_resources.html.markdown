---
layout: "vcd"
page_title: "VMware Cloud Director: Importing resources"
sidebar_current: "docs-vcd-guides-importing-resources"
description: |-
 Provides guidance to VMware Cloud resources importing
---

# Importing resources

Supported in provider *v3.11+* with Terraform *v1.5.x+*.

-> Some parts of this document describe **EXPERIMENTAL** features.

## Overview

Importing is the process of bringing a resource under Terraform control, in those cases where the resource was created
from a different agent, or a resource was created using Terraform, but some or all of its contents were not.

When we create a resource using Terraform, the side effect of this action is that the resource is stored into [Terraform state][terraform-state],
which allows us to further manipulate the resource, such as changing its contents or deleting it.

There are cases when we create a resource (for example, a cloned vApp) which has the effect of creating other resources
(such as VMs) or making them available to another owner (for example, shared catalogs and their contents).
When we are in such situation, we need to [import][terraform-import] the resource, so that we can handle it with our
Terraform workflow.

## Importing terminology

In order to import a resource, we need to issue a command, containing several elements, which are explained below.

* A **command** is a keyword of the Terraform command line tool, like the `import` in the _import command_ above.
* The resource **type** is the type of resource, such as `vcd_org`, `vcd_provider_vdc`, etc.
* The **local name** (or resource **definer**) is the name given to the resource, right next to the resource type. Not to be confused with the `name`.
* The **import path** is the identification of the resource, given as the name of its ancestors, followed by the name of the resource itself.
  The import path may be different for each resource, and may include elements other than the name.
* The **id** is the same as the **import path** in Terraform parlance. When we see the `id` mentioned in Terraform import
  documentation, it is talking about the import path.
* The **resource block** is the portion of HCL code that defines the resource.
* The **name** is the unequivocal name of the resource, independently of the HCL script.
* The **separator** is a character used to separate the elements of the `import path`. The default is the dot (`.`).
* The **Terraform state** is the representation about our resources as understood by Terraform.
* A **stage** is a part of the Terraform workflow, which is mainly one of `plan`, `apply`, `destroy`, `import`. Each one of these
  has a corresponding `command`, but not all Terraform commands have a stage.

For example:

```hcl
resource "vcd_org_user" "admin-org-user" {
  org  = "my-org"
  name = "philip"
}
```

```shell
terraform import vcd_org_user.admin-org-user my-org.philip
```

In the two snippets above:

* `vcd_org_user` is the resource **type**
* `admin-org-user` is the resource **definer**
* `import` is the **command**
* `.` is the **separator**
* `philip` is the resource **name**
* `my-org.philip` is the **resource path** or **id**
* All 5 lines of HCL code starting from `resource` are the **resource block**
* The **Terraform state** (not visible in the above script) is collected in the file `terraform.tfstate`

## Basic importing

Up to Terraform 1.4.x, importing meant the conjunction of two operations:
1. Writing the resource definition into an HCL script
2. Running the command below, also known as "**the import command**"

```
terraform import vcd_resource_type.resource_definer path_to_resource
```

The effect of the above actions (which we can also perform in later versions of Terraform) is that the resource is
imported into the [state][terraform-state].
The drawback of this approach is that we need to write the HCL definition of the resource manually, which could result
in a very time-consuming operation.

## Import mechanics

When we run a `terraform import` command like the one in the previous section, Terraform will try to read all the
of the resource and fill the `state` with the resource information.
That completes the **import** stage, but it doesn't mean that the code is usable from now on.
In fact, running `terraform plan` after the import, would result in an error.

```
╷
│ Error: Missing required argument
│
│   on config.tf line 40, in resource "vcd_org_user" "admin-org-user":
│   40: resource "vcd_org_user" "admin-org-user" {
│
│ The argument "role" is required, but no definition was found.
```

Which means that we need to edit the HCL script, and add all the necessary elements that are missing. We may use the
data from the state file (`terraform.tfstate`) to supply the missing properties.

## Semi-Automated import (Terraform v1.5+)

~> Terraform warns that this procedure is considered **experimental**.

Terraform v1.5 introduces the concept of an [import block][terraform-import], which replaces the `import` command.
Instead of 

```shell
terraform import vcd_org_user.admin-org-user my-org.philip
```

we would put the import instructions directly into the HCL script

```hcl
import {
  to = vcd_org_user.admin-org-user
  id = "my-org.philip"
}
```

There are two differences between the old and new import methods:
* the import happens as part of the `apply` stage, rather than on a separate command;
* Although we could write the resource block ourselves, we can now generate the HCL code using a Terraform command.

To generate the source HCL, we issue the command

```shell
terraform plan -generate-config-out=generated_resources.tf
```

The above command, and the next `terraform plan` or `terraform apply` will show one more set of actions to perform

```
 10 to import, 0 to add, 9 to change, 0 to destroy.
```

Here we see that the import is an operation that will happen during `apply`.

## More automated import operations

Compared to full manual importing, the procedure of writing an *import block* instead of the full HCL *resource block* is
a huge improvement. Even so, if we need to import all the vApp templates from a catalog, or the VMs from a vApp, this
reduced task can still be time-consuming and error-prone.
To reduce the manual effort and minimize errors, `terraform-provider-vcd` **v3.11+** offers a new functionality, embedded
in [`vcd_resource_list`][vcd-resource-list].

Let's suppose we want to import all VMs from a vApp *SampleClonedVapp*, wich was either created outside of Terraform,
or was created using [`vcd_cloned_vapp`][vcd-cloned-vapp] . We could write the *import block* for every VM manually, or 
we can ask `vcd_resource_list` to do that for us:

```hcl
data "vcd_resource_list" "import_vms_from_vapp" {
  resource_type    = "vcd_vapp_vm"
  name             = "import_vms_from_vapp"
  parent           = "SampleClonedVapp"
  list_mode        = "import"
  import_file_name = "import-vms_from_vapp.tf"
}
```

The result of this operation is a file (`import-vms_from_vapp.tf`) containing something like the following:

```hcl
# Generated by vcd_resource_list - 2023-08-08T15:00:10+02:00
# import data for vcd_vapp_vm datacloud.nsxt-vdc-datacloud.SampleClonedVapp.secondVM
import {
  to = vcd_vapp_vm.secondVM-b198139dab00
  id = "datacloud.nsxt-vdc-datacloud.SampleClonedVapp.secondVM"
}

# import data for vcd_vapp_vm datacloud.nsxt-vdc-datacloud.SampleClonedVapp.firstVM
import {
  to = vcd_vapp_vm.firstVM-2a0390eecf13
  id = "datacloud.nsxt-vdc-datacloud.SampleClonedVapp.firstVM"
}

# import data for vcd_vapp_vm datacloud.nsxt-vdc-datacloud.SampleClonedVapp.thirdVM
import {
  to = vcd_vapp_vm.thirdVM-23b9e15086c4
  id = "datacloud.nsxt-vdc-datacloud.SampleClonedVapp.thirdVM"
}
```

If we adopt this workflow, we will end up at the same place described in [the previous section](#semi-automated-import-terraform-v15),
where we wrote the import blocks manually. Here we are ready to run the code generation command:

```shell
terraform plan -generate-config-out=generated_resources.tf
```

Notice a few points in the above code:
* The *resource definer* is not just the resource name. To avoid duplications within the HCL code (two VMs having the
  same name in two different vApps), `vcd_resource_list` adds the rightmost portion of the ID to the definer.
* The file generation happens at every read operation of the data source. The file will be overwritten at every `plan`,
  `apply`, `refresh`. Thus, if we need to modify something, we should remove the `vcd_resource_list` data source.
  
## Troubleshooting

-> Since we refer to an experimental feature, issues and relative advice given in this section may change in future
  releases or get fixed due to upstream improvements.

### Required field not found

Some resources require several properties to be filled. For example, when creating a VDC group , we need to indicate
which of the participating VDCs is the starting one.

Let's try to import an existing VDC group:

```hcl
import {
  to = vcd_vdc_group.vdc-group-datacloud
  id = "datacloud.vdc-group-datacloud"
}
```
In this file we are saying that we want to import the VDC group `vdc-group-datacloud`, belonging to the organization `datacloud`.

```shell
terraform plan -generate-config-out=generated_resources.tf
```
```
data.vcd_resource_list.vdc-groups: Reading...
vcd_vdc_group.vdc-group-datacloud: Preparing import... [id=datacloud.vdc-group-datacloud]
data.vcd_resource_list.vdc-groups: Read complete after 2s [id=list-vdc-groups]
vcd_vdc_group.vdc-group-datacloud: Refreshing state... [id=urn:vcloud:vdcGroup:db815539-c885-4d9b-9992-aac82dce89d0]

Planning failed. Terraform encountered an error while generating this plan.
[...]
╷
│ Error: Missing required argument
│
│   with vcd_vdc_group.vdc-group-datacloud,
│   on generated_resources.tf line 8:
│   (source code not available)
│
│ The argument "starting_vdc_id" is required, but no definition was found.
╵
```

The Terraform interpreter signals that there is one missing property. Since the current syntax of import blocks does
not allow any adjustments, the only possible workaround is to update the generated HCL code. Fortunately, the above error
does not prevent the generation of the code.
If we edit the file `generated_resources.tf`, changing the value for `starting_vdc_id` from
`null` to the ID of the first VDC, the import will succeed.

### Phantom updates

In addition to missing required properties, we may have the problem of properties that are needed during creation, but
their values are not stored in the VCD, and consequently can't be retrieved and used to populate the importing HCL code.
For example, the [`accept_all_eulas`][accept-all-eulas] property is only used during VM creation, but we can't retrieve it
from the VM data in VCD.
When we have such fields, Terraform will signal that the resource needs to be updated, and it will do so at the next
occurrence of `terraform apply`. This is a minor annoyance, which will delay the operation by a few seconds, but which
won't actually change anything in the resource. What this update means is that Terraform is trying to match the HCL data
with the resource stored data. We won't be making any real changes in the VM: nonetheless, we should be vigilant and
make sure that the updates being proposed don't touch important data. If they do, we should probably edit the generated
code and set the record straight.

### Lack of dependency ties

The code generation is good enough to put the resource information into Terraform state, but it won't write the HCL code
the same way we would. Most importantly, the names or IDs of other resources will be placed verbatim into the resource
definition, rather than using a reference.

For example, when creating a vApp template, we may write the following:

```hcl
resource "vcd_catalog_vapp_template" "my_template" {
  catalog_id  = vcd_catalog.mycatalog.id
  name        = "my_template"
  description = "my template"
}
```
However, the corresponding generated code would be:

```hcl
resource "vcd_catalog_vapp_template" "my_template" {
  catalog_id  = "urn:vcloud:catalog:59b15c74-8dea-4331-ae2c-4fc4217c4191"
  name        = "my_template"
  description = "my template"
}
```

This code would work well when we use it to update the vApp template, but it may become a problem when we want to delete
all the resources. The lack of dependency information will cause the removal to happen in a random order, and we may
see "entity not found" errors during such operations. For example, if the catalog deletion happens before the vApp template
deletion, the template will not exist by the time Terraform attempts to retrieve it for removal.

There is no simple solution to this issue, other than manually editing the generated HCL code to add dependency instructions.

## Examples

There are two complete examples of multiple resource imports in the [`terraform-provider-vcd` repository][examples].
They show how we can import multiple VMs, or multiple catalog items, with step-by-step instructions.

[terraform-state]:https://developer.hashicorp.com/terraform/language/state
[terraform-import]:https://developer.hashicorp.com/terraform/language/import
[vcd-resource-list]:https://registry.terraform.io/providers/vmware/vcd/latest/docs/data-sources/resource_list
[vcd-cloned-vapp]:https://registry.terraform.io/providers/vmware/vcd/3.10.0/docs/resources/cloned_vapp
[accept-all-eulas]:https://registry.terraform.io/providers/vmware/vcd/3.10.0/docs/resources/vapp_vm#accept_all_eulas
[examples]:https://github.com/dataclouder/terraform-provider-vcd/tree/import-compound-resources/examples/importing