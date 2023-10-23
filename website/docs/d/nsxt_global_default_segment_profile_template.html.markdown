---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_global_default_segment_profile_template"
sidebar_current: "docs-vcd-data-source-nsxt-segment-profile-template"
description: |-
  Provides a data source to read Global Default NSX-T Segment Profile Templates.
---

# vcd\_nsxt\_global\_default\_segment\_profile\_template

Provides a resource to manage Global Default NSX-T Segment Profile Templates.

Supported in provider *v3.11+* and VCD 10.4.0+ with NSX-T. Requires System Administrator privileges.

## Example Usage

```hcl
resource "vcd_nsxt_global_default_segment_profile_template" "singleton" {
}
```
## Argument Reference

No arguments are required because this is a global VCD configuration

## Attribute Reference

The following attributes are supported:

* `vdc_networks_default_segment_profile_template_id` - Global Default Segment Profile
  Template ID for all VDC Networks
* `vapp_networks_default_segment_profile_template_id` - Global Default Segment Profile
  Template ID for all vApp Networks


