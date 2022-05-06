---
layout: "vcd"
page_title: "VMware Cloud Director: security_tag"
sidebar_current: "docs-vcd-resource-security-tag"
description: |-
Provides a VMware Cloud Director Security Tag resource. This can be
used to assign security tag to VMs.
---

# vcd\_security\_tag

Provides a VMware Cloud Director Security Tag resource. This can be
used to assign security tag to VMs.

Supported in provider *v3.7+* and requires VCD 10.3.0+

-> **Note:** This resource requires either system or org administrator privileges.

## Example Usage

```hcl
resource "vcd_security_tag" "my_tag" {
  name   = "test-tag"
  vm_ids = [vcd_vm.my-vm-one.id, vcd_vm.my-vm-two.id]
}
```
## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `name` - (Required) The name of the security tag.
* `vm_ids` - (Required) List of VM IDs that the security tag is going to be applied to.

-> The ID of `vcd_security_tag` is set to its name since VCD behind the scenes doesn't create an ID.

# Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing security tag can be [imported][docs-import] into this resource. An example is below:

```
terraform import vcd_security_tag.my-tag my-org.my-security-tag-name
```

[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the security tag as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the security tag stored properties.
