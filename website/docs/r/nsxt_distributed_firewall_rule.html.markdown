---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_distributed_firewall_rule"
sidebar_current: "docs-vcd-resource-nsxt-distributed-firewall-rule"
description: |-
  The Distributed Firewall rule allows user to segment organization network entities by creating firewall rules.
---

# vcd\_nsxt\_distributed\_firewall\_rule

The Distributed Firewall rule allows user to segment organization network entities by creating
firewall rules.

Multiple rules defined with this resource will **not be created in parallel** because Cloud Director
API provides no direct endpoint to create a single rule. To overcome this,
`vcd_nsxt_distributed_firewall_rule` calls an "update all rules" API endpoint for each single rule,
and this call is serialized to avoid data conflicts.

!> There is a different resource
[`vcd_nsxt_distributed_firewall`](/providers/vmware/vcd/latest/docs/resources/nsxt_distributed_firewall)
that can manage all firewall rules in one resource. One should use **only one of**
`vcd_nsxt_distributed_firewall` or `vcd_nsxt_distributed_firewall_rule` as using both will result in
unexpected firewall configuration.

## Example Usage

```hcl
data "vcd_vdc_group" "test1" {
  org  = "my-org" # Optional, can be inherited from Provider configuration
  name = "main-vdc-group"
}

data "vcd_nsxt_app_port_profile" "FTP" {
  context_id = data.vcd_nsxt_manager.main.id
  name       = "FTP"
  scope      = "SYSTEM"
}

resource "vcd_nsxt_distributed_firewall_rule" "r1" {
  org          = "my-org"
  vdc_group_id = data.vcd_vdc_group.test1.id

  name        = "rule1"
  action      = "ALLOW"
  description = "description"

  source_ids           = [vcd_nsxt_ip_set.set1.id, vcd_nsxt_ip_set.set2.id]
  destination_ids      = [vcd_nsxt_security_group.g1-empty.id, vcd_nsxt_security_group.g2.id]
  app_port_profile_ids = [data.vcd_nsxt_app_port_profile.FTP.id]
}

resource "vcd_nsxt_distributed_firewall_rule" "r2" {
  org          = "{{.Org}}"
  vdc_group_id = data.vcd_vdc_group.test1.id

  # Specifying a particular ID of other firewall rule will ensure that the current one is placed above
  above_rule_id = vcd_nsxt_distributed_firewall_rule.r1.id

  name        = "rule4"
  action      = "ALLOW"
  ip_protocol = "IPV6"
  direction   = "OUT"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
  when connected as sysadmin working across different organisations.
* `above_rule_id` - (Optional) ID of an existing `vcd_nsxt_distributed_firewall_rule` entry, above
  which the newly created firewall rule will be positioned. **Note.** By default, new rule will be
  created at the bottom of the list

-> When activating Distributed Firewall with resource
[`vcd_vdc_group`](/providers/vmware/vcd/latest/docs/resources/vdc_group), there is a default firewall
rule created which can make inconvenient to use this resource. For that reason, resource
[`vcd_vdc_group`](/providers/vmware/vcd/latest/docs/resources/vdc_group) has a parameter
`remove_default_firewall_rule` which can remove default firewall rule.

* `vdc_group_id` - (Required) The ID of VDC Group to manage Distributed Firewall in. Can be looked
  up using `vcd_vdc_group` resource or data source.
* `name` - (Required) Explanatory name for firewall rule (uniqueness not enforced)
* `comment` - (Optional; *VCD 10.3.2+*) Comment field shown in UI
* `description` - (Optional) Description of firewall rule (not shown in UI)
* `direction` - (Optional) One of `IN`, `OUT`, or `IN_OUT`. (default `IN_OUT`)
* `ip_protocol` - (Optional) One of `IPV4`,  `IPV6`, or `IPV4_IPV6` (default `IPV4_IPV6`)
* `action` - (Required) Defines if it should `ALLOW`, `DROP`, `REJECT` traffic. `REJECT` is only
  supported in VCD 10.2.2+
* `enabled` - (Optional) Defines if the rule is enabled (default `true`)
* `logging` - (Optional) Defines if logging for this rule is enabled (default `false`)
* `source_ids` - (Optional) A set of source object Firewall Groups (`IP Sets` or `Security groups`).
Leaving it empty matches `Any` (all)
* `destination_ids` - (Optional) A set of source object Firewall Groups (`IP Sets` or `Security
groups`). Leaving it empty matches `Any` (all)
* `app_port_profile_ids` - (Optional) An optional set of Application Port Profiles.
* `network_context_profile_ids` - (Optional) An optional set of Network Context Profiles. Can be
  looked up using `vcd_nsxt_network_context_profile` data source.
* `source_groups_excluded` - (Optional; VCD 10.3.2+) - reverses value of `source_ids` for the rule to
  match everything except specified IDs.
* `destination_groups_excluded` - (Optional; VCD 10.3.2+) - reverses value of `destination_ids` for
  the rule to match everything except specified IDs.


## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

Existing Distributed Firewall Rules can be [imported][docs-import] into this resource via supplying
the full dot separated path for your VDC Group Name. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxt_distributed_firewall_rule.imported my-org-name.my-vdc-group-name.my-rule-name
```

The above would import firewall rule with name `my-rule-name` defined on VDC Group
`my-vdc-group-name` which is configured in organization named `my-org-name`.