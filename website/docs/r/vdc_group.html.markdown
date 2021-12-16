---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_vdc_group"
sidebar_current: "docs-vcd-resource-vdc-group"
description: |-
  Provides a VDC group resource.
---

# vcd\_vdc\_group
Supported in provider *v3.5+* and VCD 10.2+.

Provides a resource to manage NSX-T VDC groups.

~> Only `System Administrator` and `Org Users` with rights `View VDC Group`, `Configure VDC Group`, `vDC Group: Configure Logging`, `Organization vDC Distributed Firewall: Enable/Disable` can manage VDC groups using this resource.

## Example Usage

```hcl
data "vcd_org_vdc" "startVdc" {
  name = "existingVdc"
}

data "vcd_org_vdc" "additionalVdc" {
  name = "oneMoreVdc"
}

resource "vcd_vdc_group" "new-vdc-group" {
  org                   = "myOrg"
  name                  = "newVdcGroup"
  description           = "my description"
  starting_vdc_id       = data.vcd_org_vdc.startVdc.id
  participating_vdc_ids = [data.vcd_org_vdc.startVdc.id, data.vcd_org_vdc.additionalVdc.id]
  dfw_enabled           = true
  default_policy_status = true
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organizations
* `name` - (Required)  - The name for VDC group
* `description` - (Optional)  - VDC group description
* `starting_vdc_id` - (Required)  - With selecting a starting VDC you will be able to create a group in which this VDC can participate. **Note**. It must be included in `participating_vdc_ids` to participate in this group.
* `participating_vdc_ids` - (Required)  - The list of organization vDCs that are participating in this group. **Note**. `starting_vdc_id` isn't automatically included.
* `dfw_enabled` - (Optional)  - Whether Distributed Firewall is enabled for this vDC Group.
* `default_policy_status` - (Optional)  - Whether this security policy is enabled. `dfw_enabled` must be `true`.

## Attribute Reference

The following attributes are exported on this resource:

* `id` - The VDC group ID
* `error_message` - More detailed error message when VDC group has error status
* `local_egress` - Status whether local egress is enabled for a universal router belonging to a universal vDC group.
* `network_pool_id` - ID of used network pool.
* `network_pool_universal_id` - The network provider’s universal id that is backing the universal network pool.
* `network_provider_type` - Defines the networking provider backing the vDC Group.
* `status` - The status that the group can be in (e.g. 'SAVING', 'SAVED', 'CONFIGURING', 'REALIZED', 'REALIZATION_FAILED', 'DELETING', 'DELETE_FAILED', 'OBJECT_NOT_FOUND', 'UNCONFIGURED').
* `type` - Defines the group as LOCAL or UNIVERSAL.
* `universal_networking_enabled` - True means that a vDC group router has been created.
* `participating_org_vdcs` - A list of blocks providing organization vDCs that are participating in this group details. See [Participating Org VDCs](#participatingOrgVdcs) below for details.

<a id="participatingOrgVdcs"></a>
## Participating Org VDCs

* `vdc_id` - VDC ID.
* `vdc_name` - VDC name.
* `site_id` - Site ID.
* `site_name` - Site name.
* `org_id` - Organization ID.
* `org_name` - Organization Name.
* `status` - "The status that the vDC can be in e.g. 'SAVING', 'SAVED', 'CONFIGURING', 'REALIZED', 'REALIZATION_FAILED', 'DELETING', 'DELETE_FAILED', 'OBJECT_NOT_FOUND', 'UNCONFIGURED')."
* `remote_org` - Specifies whether the vDC is local to this VCD site.
* `network_provider_scope` - Specifies the network provider scope of the vDC.
* `fault_domain_tag` - Represents the fault domain of a given organization vDC.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing VDC group can be [imported][docs-import] into this resource
via supplying the full dot separated path VDC group. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_vdc_group.imported my-org.my-vdc-group
```

The above would import the VDC group named `my-vdc-group` which is configured in organization named `my-org`.
