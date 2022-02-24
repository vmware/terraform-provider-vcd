---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_edgegateway"
sidebar_current: "docs-vcd-data-source-nsxt-edge-gateway"
description: |-
  Provides a VMware Cloud Director NSX-T edge gateway data source. This can be used to read NSX-T edge gateway configurations.
---

# vcd\_nsxt\_edgegateway

Provides a VMware Cloud Director NSX-T edge gateway data source. This can be used to read NSX-T edge gateway configurations.

-> **Note:** This data source uses new VMware Cloud Director
[OpenAPI](https://code.vmware.com/docs/11982/getting-started-with-vmware-cloud-director-openapi) and
requires at least VCD *10.1.1+* and NSX-T *3.0+*.

Supported in provider *v3.1+*.

## Example Usage

```hcl
data "vcd_nsxt_edgegateway" "t1" {
  org  = "myorg"
  vdc  = "my-nsxt-vdc"
  name = "nsxt-edge-gateway"
}
```

## Example Usage (Edge Gateway belonging to VDC group)

```hcl
data "vcd_vdc_group" "group1" {
  name = "existing-group"
}

data "vcd_nsxt_edgegateway" "t1" {
  org          = "myorg"
  vdc          = "my-nsxt-vdc"
  vdc_group_id = data.vcd_vdc_group.group1.id
  name = "nsxt-edge-gateway"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which the edge gateway belongs. Optional if defined at provider level.
* `vdc` - (Optional) The name of VDC that owns the edge gateway. Optional if defined at provider level.
* `vdc_group_id` - (Optional, *v3.6+*,*VCD 10.2+*) The ID of VDC group that this Edge Gateway belongs to.
**Note.** Data source [vcd_vdc_group](/providers/vmware/vcd/latest/docs/data-sources/vdc_group) can be used to lookup 
ID by name.
* `name` - (Required) NSX-T Edge Gateway name.

## Attribute reference

All properties defined in [vcd_nsxt_edgegateway](/providers/vmware/vcd/latest/docs/resources/nsxt_edgegateway)
resource are available.
