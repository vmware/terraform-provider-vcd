---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_vm_placement_policy"
sidebar_current: "docs-vcd-data-source-vm-placement-policy"
description: |-
  Provides a VMware Cloud Director VM Placement Policy data source. This can be
  used to read a VM placement policy.
---

# vcd\_vm\_placement\_policy

Provides a VMware Cloud Director VM Placement Policy data source. This can be used to read a VM Placement Policy.

Supported in provider *v3.8+* and requires VCD 10.2+

-> **Note:** This resource can be used by both system administrators and tenant users.

## Example Usage for System administrators

System administrators have full privileges to retrieve information of the Provider VDC to which the VM Placement Policy
belongs. The way to fetch a VM Placement Policy in this case would be:

```hcl
data "vcd_org_vdc" "my-vdc" {
  org  = "test"
  name = "vdc-test"
}

data "vcd_provider_vdc" "my-pvdc" {
  name = data.vcd_org_vdc.my-vdc.provider_vdc_name
}

data "vcd_vm_placement_policy" "tf-policy-name" {
  name            = "my-policy"
  provider_vdc_id = data.vcd_provider_vdc.my-pvdc.id
}

output "policyId" {
  value = data.vcd_vm_placement_policy.tf-policy-name.id
}
```

## Example Usage for tenant users

Tenant users don't have access to Provider VDC information so the only way to retrieve VM Placement Policies is to
fetch them using the VDC information. The only constraint is that the desired VM Placement Policy **must be assigned
to the VDC**.

```hcl
data "vcd_org_vdc" "my-vdc" {
  org  = "test"
  name = "vdc-test"
}

data "vcd_vm_placement_policy" "tf-policy-name" {
  name   = "my-policy"
  vdc_id = data.vcd_org_vdc.my-vdc.id
}

output "policyId" {
  value = data.vcd_vm_placement_policy.tf-policy-name.id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name VM Placement Policy.
* `provider_vdc_id` - (Required for System admins) The ID of the [Provider VDC](/providers/vmware/vcd/latest/docs/data-sources/provider_vdc) to which the VM Placement Policy belongs.
* `vdc_id` - (Required for tenant users; *v3.8.1+*) The ID of the [VDC](/providers/vmware/vcd/latest/docs/data-sources/org_vdc) to which the VM Placement Policy is assigned.

## Attribute Reference

All attributes defined in [`vcd_vm_placement_policy`](/providers/vmware/vcd/latest/docs/resources/vm_placement_policy#attribute-reference) resource are supported,
with a special casuistic to take into account:

* `vm_group_ids` - This attribute can't be retrieved if the data source is used by a tenant user when fetching by `vdc_id`.
* `logical_vm_group_ids` - This attribute can't be retrieved if the data source is used by a tenant user when fetching by `vdc_id`.