---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_dynamic_security_group"
sidebar_current: "docs-vcd-data-source-nsxt-dynamic-security-group"
description: |-
  Provides a data source to read NSX-T Dynamic Security Groups. Dynamic Security Groups group virtual
  machines based on specific criteria (VM Names or Security tags) to which distributed firewall rules
  applies.
---

# vcd\_nsxt\_dynamic\_security\_group

Supported in provider *v3.7+* and VCD 10.3+ with NSX-T backed VDC Groups.

Provides a data source to read NSX-T Dynamic Security Groups. Dynamic Security Groups group virtual
machines based on specific criteria (VM Names or Security tags) to which distributed firewall rules
applies.

## Example Usage 1 (Existing Dynamic Security Group lookup)

```hcl
data "vcd_vdc_group" "group1" {
  org  = "cloud"
  name = "vdc-group-cloud"
}

data "vcd_nsxt_dynamic_security_group" "group1" {
  org          = "cloud"
  vdc_group_id = data.vcd_vdc_group.group1.id

  name = "cloud-dynamic-security-group"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
  when connected as sysadmin working across different organisations.
* `vdc_group_id` - (Required) VDC Group ID hosting existing Dynamic Security Group.
* `name` - (Required) A unique name for existing Dynamic Security Group

All the arguments and attributes defined in
[`vcd_nsxt_dynamic_security_group`](/providers/vmware/vcd/latest/docs/resources/nsxt_dynamic_security_group) resource are available.
