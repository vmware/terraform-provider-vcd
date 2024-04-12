---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_solution_landing_zone"
sidebar_current: "docs-vcd-resource-solution-landing-zone"
description: |-
  Provides a resource to configure VCD Solution Add-on Landing Zone
---

# vcd\_solution\_landing\_zone

Supported in provider *v3.13+* and VCD 10.4+.

Provides a resource to configure VCD Solution Add-on Landing Zone.

~> Only `System Administrator` can create this resource.

## Example Solution Landing Zone configuration

```hcl
data "vcd_catalog" "nsxt" {
  org  = "datacloud"
  name = "cat-datacloud-nsxt-backed"
}

data "vcd_org_vdc" "vdc1" {
  org  = "datacloud"
  name = "nsxt-vdc-datacloud"
}

data "vcd_network_routed_v2" "r1" {
  org  = "datacloud"
  vdc  = "nsxt-vdc-datacloud"
  name = "nsxt-net-datacloud-r"
}

data "vcd_storage_profile" "sp" {
  org  = "datacloud"
  vdc  = "nsxt-vdc-datacloud"
  name = "*"
}

resource "vcd_solution_landing_zone" "slz" {
  org = "datacloud"

  catalog {
    id = data.vcd_catalog.nsxt.id
  }

  vdc {
    id         = data.vcd_org_vdc.vdc1.id
    is_default = true

    org_vdc_network {
      id = data.vcd_network_routed_v2.r1.id
    }

    compute_policy {
      id = data.vcd_org_vdc.vdc1.default_compute_policy_id
    }

    storage_policy {
      id = data.vcd_storage_profile.sp.id
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Required) Destination Organization name for Solution Add-ons
* `catalog` - (Required) This catalog stores all executable .ISO files for solution add-ons. There
  can be a single `catalog` element and the required field is `id`.
* `vdc` - (Required)  A single [vdc](#vdc) block that defines landing VDC configuration

<a id="vdc"></a>
## vdc configuration block

* `id` - (Required) Destination VDC ID for Solution Add-ons
* `org_vdc_network` - (Required) At least one Org VDC Network is required. See [vdc
  child](#vdc-child) block description for possible values.
* `compute_policy` - (Required) At least Compute Policy is required. See [vdc child](#vdc-child)
  block description for possible values.
* `storage_policy` - (Required) At least Storage Policy is required. See [vdc child](#vdc-child)
  block description for possible values.


<a id="vdc-child"></a>
## vdc child configuration block

* `id` - (Required) ID of child entity (Org VDC Network, Compute Policy, Storage Policy)
* `is_default` - (Optional) Defines which of the child entities is default (only one default is
  possible)

## Attribute Reference

The following attributes are exported on this resource:

* `state` - reports the state of parent [Runtime Defined
  Entity](/providers/vmware/vcd/latest/docs/resources/rde)

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

Only a single entry should exist in a valid scenario, but :

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_solution_landing_zone.imported
```
