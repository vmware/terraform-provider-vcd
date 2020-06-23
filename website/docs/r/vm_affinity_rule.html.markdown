---
layout: "vcd"
page_title: "vCloudDirector: vcd_vm_affinity_rule"
sidebar_current: "docs-vcd-resource-vm-affinity-rule"
description: |-
  Provides a vCloud Director VM affinity rule resource. This can be
  used to create, modify, and delete VM affinity and anti-affinity rules.
---

# vcd\_vm\_affinity\_rule

Provides a vCloud Director VM affinity rule resource. This can be
used to create, modify, and delete VM affinity and anti-affinity rules.

Supported in provider *v2.9+*

~> **Note:** The vCD UI defines two different entities (*Affinity Rules* and *Anti-Affinity Rules*). This resource combines both
entities: they are differentiated by the `polarity` property (see below).

## Example Usage

```hcl
data "vcd_vapp" "Test_EmptyVmVapp1" {
  name = "Test_EmptyVmVapp1"
}

data "vcd_vapp_vm" "Test_EmptyVm1a" {
  vapp_name = data.vcd_vapp.Test_EmptyVmVapp1.name
  name      = "Test_EmptyVm1a"
}

data "vcd_vapp_vm" "Test_EmptyVm1b" {
  vapp_name = data.vcd_vapp.Test_EmptyVmVapp1.name
  name      = "Test_EmptyVm1b"
}

data "vcd_vapp_vm" "Test_EmptyVm1c" {
  vapp_name = data.vcd_vapp.Test_EmptyVmVapp1.name
  name      = "Test_EmptyVm1c"
}

resource "vcd_vm_affinity_rule" "Test_VmAffinityRule1" {
  name     = "Test_VmAffinityRule1"
  required = true
  enabled  = true
  polarity = "Affinity"

  virtual_machine_ids = [
    data.vcd_vapp_vm.Test_EmptyVm1a.id,
    data.vcd_vapp_vm.Test_EmptyVm1b.id,
    data.vcd_vapp_vm.Test_EmptyVm1c.id
  ]
}
```
## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organizations
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level
* `name` - (Required) The name of VM affinity rule. Duplicates are allowed, although the name can be used to retrieve
  the rule (as data source or when importing) only if it is unique.
* `polarity` - (Required) One of `Affinity` or `Anti-Affinity`. This property cannot be changed. Once created, if we
   need to change polarity, we need to remove the rule and create a new one.
* `enabled` (Optional) True if this affinity rule is enabled. The default is `true`
* `required` (Optional) True if this affinity rule is required. When a rule is mandatory, a host failover will not 
   power on the VM if doing so would violate the rule. The default is `true`
* `virtual_machine_ids` (Required) A set of virtual machine IDs that compose this rule. At least 2 IDs must be provided.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing VM affinity rule can be [imported][docs-import] into this resource via supplying its path.
The path for this resource is made of orgName.vdcName.affinityRuleIdentifier.

The `affinityRuleIdentifier` can be either a name or an ID. If it is a name, it will succeed only if the name is unique.

For example, using this structure, representing a VM affinity rule that was **not** created using Terraform:

```hcl
resource "vcd_vm_affinity_rule" "tf-myar" {
}
```

You can import such VM affinity rule into terraform state using this command

```
terraform import vcd_vm_affinity_rule.tf-myar my-org.my-vdc.my-ar
```

### Dealing with duplicate or unknown names

If the name of the affinity rule you want to import is duplicated, when running the command above you will get an error,
containing the IDs of the rules, from which you can choose the one you need.

```
terraform  import vcd_vm_affinity_rule.unknown my-org.my-vdc.my-ar
vcd_vm_affinity_rule.unknown: Importing from ID "my-org.my-vdc.my-ar"...

Error: [VM affinity rule import] more than one VM affinity rule matches the name my-ar
  0 my-ar               Affinity      eda9011c-6841-4060-9336-d2f609c110c3
  1 my-ar               Anti-Affinity 390d737e-45ed-4fa0-86c5-2100efee7808
```

If you want to use an ID, but don't know any names, you can use the `list@` prefix to get a complete list of existing
rules.

```
terraform  import vcd_vm_affinity_rule.unknown list@my-org.my-vdc.my-ar
vcd_vm_affinity_rule.unknown: Importing from ID "my-org.my-vdc.my-ar"...

Error: [VM affinity rule import] list of all VM affinity rules:
  0 some-rule                      Affinity      a36855cc-5290-4d7f-a4a4-1d8d37b4d887
  1 my-ar                          Affinity      eda9011c-6841-4060-9336-d2f609c110c3
  2 my-ar                          Anti-Affinity 390d737e-45ed-4fa0-86c5-2100efee7808
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After importing, if you run `terraform plan` you will see the rest of the values and modify the script accordingly for
further operations.
