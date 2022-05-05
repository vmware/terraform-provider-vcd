---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_vm_sizing_policy"
sidebar_current: "docs-vcd-datasource-vm-sizing-policy"
description: |-
  Provides a VMware Cloud Director VM sizing policy data source. This can be
  used to read VM sizing policy.
---

# vcd\_vm\_sizing\_policy

Provides a VMware Cloud Director VM sizing policy data source. This can be
used to read VM sizing policy.

Supported in provider *v3.0+* and requires VCD 10.0+

## Example Usage

```hcl
data "vcd_vm_sizing_policy" "tf-policy-name" {
  name = "my-rule"
}
output "policyId" {
  value = data.vcd_vm_sizing_policy.tf-policy-name.id
}
```
## Argument Reference

The following arguments are supported:

* `name` - (Required) The name VM sizing policy

-> **Note:**  
Previously, it was incorrectly stated that the `org` argument was required. In fact, it is not, and it has been deprecated in the resource schema.
To preserve compatibility until the next release, though, the parameter is still parsed, but ignored.

All arguments defined in [`vcd_vm_sizing_policy`](/providers/vmware/vcd/latest/docs/resources/vm_sizing_policy#argument-reference) are supported.

