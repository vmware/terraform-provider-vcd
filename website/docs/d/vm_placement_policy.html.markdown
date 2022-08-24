---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_vm_placement_policy"
sidebar_current: "docs-vcd-datasource-vm-placement-policy"
description: |-
  Provides a VMware Cloud Director VM Placement Policy data source. This can be
  used to read a VM placement policy.
---

# vcd\_vm\_placement\_policy

Provides a VMware Cloud Director VM Placement Policy data source. This can be
used to read a VM Placement Policy.

Supported in provider *v3.8+* and requires VCD 10.0+

## Example Usage

```hcl
data "vcd_provider_vdc" "my-pvdc" {
  name = "my-pVDC"
}

data "vcd_vm_placement_policy" "tf-policy-name" {
  name = "my-policy"
  provider_vdc_id = data.vcd_provider_vdc.my-pvdc.id
}

output "policyId" {
  value = data.vcd_vm_placement_policy.tf-policy-name.id
}
```
## Argument Reference

The following arguments are supported:

* `name` - (Required) The name VM Placement Policy
* `provider_vdc_id` - (Required) The URN of the Provider VDC to which the VM Placement Policy belongs.

## Attribute Reference

TODO