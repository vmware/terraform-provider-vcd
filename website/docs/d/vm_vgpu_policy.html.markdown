---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_vm_vgpu_policy"
sidebar_current: "docs-vcd-data-source-vm-vgpu-policy"
description: |-
  Provides a data source to read vGPU policies in VMware Cloud Director.
---

# vcd\_vm\_vgpu\_policy

Experimental in provider *3.11*.

Provides a data source to read vGPU policies in VMware Cloud Director.

## Example Usage

```hcl
data "vcd_vm_vgpu_policy" "tf-policy-name" {
  name = "my-rule"
}

output "policyId" {
  value = data.vcd_vm_vgpu_policy.tf-policy-name.id
}
```
## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the VM vGPU policy.

All arguments defined in [`vcd_vm_vgpu_policy`](/providers/vmware/vcd/latest/docs/resources/vm_vgpu_policy#argument-reference) are supported.

