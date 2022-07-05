---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_dynamic_security_group"
sidebar_current: "docs-vcd-resource-nsxt-dynamic-security-group"
description: |-
  Provides a resource to manage NSX-T Dynamic Security Groups. Dynamic Security Groups group Virtual
  Machines based on specific criteria (VM Names or Security tags) to which Distributed Firewall Rules
  apply.
---

# vcd\_nsxt\_dynamic\_security\_group

Supported in provider *v3.7+* and VCD 10.3+ with NSX-T backed VDC Groups.

Provides a resource to manage NSX-T Dynamic Security Groups. Dynamic Security Groups group Virtual
Machines based on specific criteria (VM Names or Security tags) to which Distributed Firewall Rules
apply.

-> Dynamic Security Groups can be used in both - Edge Gateway Firewall Rules (`vcd_nsxt_firewall`)
and Distributed Firewall Rules (`vcd_nsxt_distributed_firewall`), however **it only works when Edge
Gateway belongs to a VDC Group**.

## Example Usage 1 (Dynamic Security Group with 3 criteria and 4 rules in each criteria)

```hcl
data "vcd_vdc_group" "group1" {
  org  = "cloud"
  name = "vdc-group-cloud"
}

resource "vcd_nsxt_dynamic_security_group" "group1" {
  org          = "cloud"
  vdc_group_id = data.vcd_vdc_group.group1.id

  name = "dynamic-security-group-example"

  criteria { # Boolean "OR"
    rule {   # Boolean "AND"
      type     = "VM_TAG"
      operator = "EQUALS"
      value    = "tag-equals"
    }

    rule { # Boolean "AND"
      type     = "VM_TAG"
      operator = "CONTAINS"
      value    = "tag-contains"
    }

    rule { # Boolean "AND"
      type     = "VM_TAG"
      operator = "STARTS_WITH"
      value    = "tag-starts-with"
    }

    rule { # Boolean "AND"
      type     = "VM_TAG"
      operator = "ENDS_WITH"
      value    = "tag-ends-with"
    }
  }

  criteria { # Boolean "OR" evaluation
    rule {   # Boolean "AND"
      type     = "VM_NAME"
      operator = "CONTAINS"
      value    = "name-contains-4"
    }

    rule { # Boolean "AND"
      type     = "VM_NAME"
      operator = "STARTS_WITH"
      value    = "starts_with2"
    }

    rule { # Boolean "AND"
      type     = "VM_NAME"
      operator = "CONTAINS"
      value    = "name-contains-22"
    }

    rule { # Boolean "AND"
      type     = "VM_NAME"
      operator = "STARTS_WITH"
      value    = "starts_with22"
    }
  }

  criteria { # Boolean "OR" evaluation
    rule {   # Boolean "AND"
      type     = "VM_NAME"
      operator = "CONTAINS"
      value    = "name-contains3"
    }

    rule { # Boolean "AND"
      type     = "VM_NAME"
      operator = "STARTS_WITH"
      value    = "starts_with3"
    }

    rule { # Boolean "AND"
      type     = "VM_NAME"
      operator = "CONTAINS"
      value    = "name-contains33"
    }

    rule { # Boolean "AND"
      type     = "VM_NAME"
      operator = "STARTS_WITH"
      value    = "starts_with33"
    }
  }
}
```

## Example Usage 2 (Empty Dynamic Security Group)
```hcl
data "vcd_vdc_group" "group1" {
  org  = "cloud"
  name = "vdc-group-cloud"
}

resource "vcd_nsxt_dynamic_security_group" "group1" {
  org          = "cloud"
  vdc_group_id = data.vcd_vdc_group.group1.id

  name = "empty-dynamic-security-group"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
  when connected as sysadmin working across different organisations.
* `vdc_group_id` - (Required) VDC Group ID for Dynamic Security Group creation.
* `name` - (Required) A unique name for Dynamic Security Group
* `description` - (Optional) An optional description of the Dynamic Security Group
* `criteria` (Optional) Up to 3 criteria for matching VMs. List of criteria is matched with boolean
  `OR` operation and matching any of defined criteria will include objects. Each `criteria` can
  contains up to 4 `rule` definitions.
* `rule` (Optional) Up to 4 rules for matching VMs. List of rules are matched with boolean `AND`
  operation and all defines rules must match to include object. See [Rule](#rule) for rule
  definition structure.


<a id="rule"></a>
## Rule

Each member Rule contains following attributes:

* `type` - (Required) `VM_NAME` or `VM_TAG`
* `value` - (Required) String to evaluate by given `type` and `operator`
* `operator` - (Required) Supported operators depend on `type`. `VM_TAG` supports 4 operator types
  with self explanatory names:
 * `CONTAINS`
 * `STARTS_WITH`
 * `EQUALS`
 * `ENDS_WITH`

`VM_NAME` supports two operator types:
 * `STARTS_WITH`
 * `CONTAINS`

## Attribute Reference
* `member_vms` A set of member VMs (if exist). see [Member VMs](#member-vms) below for details.

<a id="member-vms"></a>
## Member VMs

Each member VM contains following attributes:

* `vm_id` - Member VM ID
* `vm_name` - Member VM name
* `vapp_id` - Parent vApp ID for member VM (empty for standalone VMs)
* `vapp_name` - Parent vApp Name for member VM (empty for standalone VMs)

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing Security Group configuration can be [imported][docs-import] into this resource
via supplying the full dot separated path for your Security Group name. An example is
below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxt_dynamic_security_group.imported my-org.my-vdc-group.my-security-group-name
```

The above would import the `my-security-group-name` Dynamic Security Group config settings that are
defined in VDC Group `my-vdc-group` which is configured in organization named `my-org`.
