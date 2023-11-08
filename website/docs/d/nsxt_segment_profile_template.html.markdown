---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_segment_profile_template"
sidebar_current: "docs-vcd-data-source-nsxt-segment-profile-template"
description: |-
  Provides a data source to read NSX-T Segment Profile Templates.
---

# vcd\_nsxt\_segment\_profile\_template

Provides a data source to read NSX-T Segment Profile Templates.

Supported in provider *v3.11+* and VCD 10.4.0+ with NSX-T. Requires System Administrator privileges.

## Example Usage (Complete example with all Segment Profiles)

```hcl
data "vcd_nsxt_segment_profile_template" "complete" {
  name = "my-segment-profile-template-name"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of existing Segment Profile Template

## Attribute reference

All properties defined in [vcd_nsxt_segment_profile_template](/providers/vmware/vcd/latest/docs/resources/nsxt_segment_profile_template)
resource are available.
