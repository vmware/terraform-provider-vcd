---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_org_vdc_nsxt_network_profile
sidebar_current: "docs-vcd-resource-vcd-org-vdc-nsxt-network-profile"
description: |-
  Provides a resource to manage NSX-T Org VDC Network Profile.
---

# vcd\_org\_vdc\_nsxt\_network\_profile

Provides a resource to manage NSX-T Org VDC Network Profile.

Supported in provider *v3.11+* and VCD 10.4.0+ with NSX-T.

## Example Usage

```hcl
data "vcd_org_vdc" "nsxt" {
  org  = "my-org"
  name = "my-vdc"
}

data "vcd_nsxt_edge_cluster" "first" {
  org    = "my-org"
  vdc_id = data.vcd_org_vdc.nsxt.id
  name   = "my-edge-cluster"
}

resource "vcd_org_vdc_nsxt_network_profile" "nsxt" {
  org = "my-org"
  vdc = "my-vdc"

  edge_cluster_id                                   = data.vcd_nsxt_edge_cluster.first.id
  vdc_networks_default_segment_profile_template_id  = vcd_nsxt_segment_profile_template.complete.id
  vapp_networks_default_segment_profile_template_id = vcd_nsxt_segment_profile_template.complete.id
}
```

## Argument Reference

The following arguments are supported:

* `edge_cluster_id` - (Optional) - Edge Cluster ID to be used for this VDC
* `vdc_networks_default_segment_profile_template_id` - (Optional) - Default Segment Profile
  Template ID for all VDC Networks in a VDC
* `vapp_networks_default_segment_profile_template_id` - (Optional) - Default Segment Profile
  Template ID for all vApp Networks in a VDC


## Importing


~> **Note:** The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing an organization VDC NSX-T Network Profile configuration can be [imported][docs-import] into
this resource via supplying the full dot separated path to VDC. An example is below:

```
terraform import vcd_org_vdc_nsxt_network_profile.my-cfg my-org.my-vdc
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the VDC Network
Profile as needed. Running `terraform plan` at this stage will show the difference between the
minimal configuration file and the VDC's stored properties.

