---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_edgegateway"
sidebar_current: "docs-vcd-data-source-nsxt-edge-gateway"
description: |-
  Provides a VMware Cloud Director NSX-T edge gateway data source. This can be used to read NSX-T edge gateway configurations.
---

# vcd\_nsxt\_edgegateway

Provides a VMware Cloud Director NSX-T edge gateway data source. This can be used to read NSX-T edge gateway configurations.

Supported in provider *v3.1+*.

## Example Usage (NSX-T Edge Gateway belonging to VDC Group)

```hcl
data "vcd_vdc_group" "group1" {
  name = "existing-group"
}

data "vcd_nsxt_edgegateway" "t1" {
  org      = "myorg"
  owner_id = data.vcd_vdc_group.group1.id
  name     = "nsxt-edge-gateway"
}
```

## Example Usage (NSX-T Edge Gateway belonging to VDC)

```hcl
data "vcd_org_vdc" "vdc1" {
  name = "existing-vdc"
}

data "vcd_nsxt_edgegateway" "t1" {
  org      = "myorg"
  owner_id = data.vcd_org_vdc.vdc1.id
  name     = "nsxt-edge-gateway"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which the NSX-T Edge Gateway belongs. Optional if
  defined at provider level.
* `vdc` - (Optional)  **Deprecated** - please use `owner_id` field. The name of VDC that owns the
  NSX-T Edge Gateway. Optional if defined at provider level.
* `owner_id` - (Optional, *v3.6+*,*VCD 10.2+*) **Replaces** `vdc` field. The ID of VDC or VDC Group
that this Edge Gateway belongs to. **Note:** Data source
[vcd_vdc_group](/providers/vmware/vcd/latest/docs/data-sources/vdc_group) can be used to lookup ID
by name.

~> Only one of `vdc` or `owner_id` can be specified. `owner_id` takes precedence over `vdc`
definition at provider level.

* `name` - (Required) NSX-T Edge Gateway name.

## Attribute reference

All properties defined in [vcd_nsxt_edgegateway](/providers/vmware/vcd/latest/docs/resources/nsxt_edgegateway)
resource are available.
