---
layout: "vcd"
page_title: "vCloudDirector: vcd_vm_affinity_rule"
sidebar_current: "docs-vcd-datasource-vm-affinity-rule"
description: |-
  Provides a vCloud Director VM affinity rule data source. This can be
  used to read VM affinity and anti-affinity rules.
---

# vcd\_vm\_affinity\_rule

Provides a vCloud Director VM affinity rule data source. This can be
used to read VM affinity and anti-affinity rules.

Supported in provider *v2.9+*

~> **Note:** The vCD UI defines two different entities (*Affinity Rules* and *Anti-Affinity Rules*). This data source combines both
entities: they are differentiated by the `polarity` property (See below).

## Example Usage

```hcl
data "vcd_vm_affinity_rule" "tf-rule-by-name" {
  name = "my-rule"
}

data "vcd_vm_affinity_rule" "tf-rule-by-id" {
  rule_id = "eda9011c-6841-4060-9336-d2f609c110c3"
}
```
## Argument Reference

The following arguments are supported:

* `org` - The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `vdc` - The name of VDC to use, optional if defined at provider level
* `name` - The name of VM affinity rule. Needed if we don't provide `rule_id`
* `rule_id` - Is the ID of the affinity rule. It's the preferred way to retrieve the affinity
rule, especially if the rule name could have duplicates
 
## Attribute reference

* `polarity` - One of `Affinity` or `Anti-Affinity`. This property cannot be changed. Once created, if we
   need to change polarity, we need to remove the rule and create a new one.
* `enabled` True if this affinity rule is enabled.
* `required` True if this affinity rule is required. When a rule is mandatory, a host failover will not 
   power on the VM if doing so would violate the rule.
* `virtual_machine_ids` A set of virtual machine IDs that compose this rule.

