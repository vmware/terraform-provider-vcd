---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_distributed_firewall"
sidebar_current: "docs-vcd-resource-nsxt-distributed-firewall"
description: |-
  The Distributed Firewall allows user to segment organization virtual data center entities, such as
  virtual machines, based on virtual machine names and attributes. 
---

# vcd\_nsxt\_distributed\_firewall

The Distributed Firewall allows user to segment organization virtual data center entities, such as
virtual machines, based on virtual machine names and attributes. 

## Example Usage

```hcl
data "vcd_vdc_group" "existing" {
  org  = "my-org" # Optional, can be inherited from Provider configuration
  name = "main-vdc-group"
}

data "vcd_nsxt_network_context_profile" "cp1" {
  context_id = vcd_vdc_group.test1.id
  name       = "CTRXICA"
  scope      = "SYSTEM"
}

resource "vcd_nsxt_distributed_firewall" "t1" {
  org          = "my-org" # Optional, can be inherited from Provider configuration
  vdc_group_id = vcd_vdc_group.existing.id

  rule {
    name        = "rule1"
    action      = "ALLOW"
    description = "description"
    # 'comment' field is only supported in VCD 10.3.2+
    comment = "My first rule to allow everything"

    source_ids             = [data.vcd_nsxt_ip_set.set1.id, data.vcd_nsxt_ip_set.set2.id]
    source_groups_excluded = true # Negates value of 'source_id' (VCD 10.3.2+)
    app_port_profile_ids   = [data.vcd_nsxt_app_port_profile.WINS.id, data.vcd_nsxt_app_port_profile.FTP.id]
  }

  rule {
    name      = "rule2"
    action    = "DROP"
    enabled   = false
    logging   = true
    direction = "IN_OUT"

    network_context_profile_ids = [vcd_nsxt_network_context_profile.cp1.id]
  }

  rule {
    name = "rule3"
    # 'REJECT' is only supported in VCD 10.2.2+
    action      = "REJECT"
    ip_protocol = "IPV4"
  }

  rule {
    name        = "rule4"
    action      = "ALLOW"
    ip_protocol = "IPV6"
    direction   = "OUT"

    # Below two fields are supported in VCD 10.3.2+
    source_groups_excluded      = false
    destination_groups_excluded = false
  }

  rule {
    name        = "rule5"
    action      = "ALLOW"
    ip_protocol = "IPV6"
    direction   = "IN"
  }
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
  when connected as sysadmin working across different organisations.
* `vdc_group_id` - (Required) The ID of VDC Group to manage Distributed Firewall in. Can be looked
  up using `vcd_vdc_group` resource or data source.
* `rule` (Required) One or more blocks with [Firewall Rule](#firewall-rule) definitions. **Order**
  defines firewall rule precedence

<a id="firewall-rule"></a>
## Firwall Rule

Each Firewall Rule contains following attributes (**order defines rule precedence**):

* `name` - (Required) Explanatory name for firewall rule (uniqueness not enforced)
* `comment` - (Optional; *VCD 10.3.2+*)) Comment field shown in UI
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
* `source_groups_excluded` (Optional; VCD 10.3.2+) - reverses value of `source_ids` for the rule to
  match everything except specified IDs.
* `destination_groups_excluded` (Optional; VCD 10.3.2+) - reverses value of `destination_ids` for
  the rule to match everything except specified IDs.

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

Existing Distributed Firewall Rules can be [imported][docs-import] into this resource via supplying
the full dot separated path for your VDC Group Name. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxt_distributed_firewall.imported my-org-name.my-vdc-group-name
```

The above would import all firewall rules defined on VDC Group `my-vdc-group-name` which is
configured in organization named `my-org-name`.