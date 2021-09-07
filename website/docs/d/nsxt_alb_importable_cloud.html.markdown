---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_alb_importable_cloud"
sidebar_current: "docs-vcd-datasource-nsxt-alb-importable-cloud"
description: |-
  Provides a data source to reference existing NSX-T ALB Importable Clouds. An NSX-T Importable Cloud is a reference to a
  Cloud configured in ALB Controller.
---

# vcd\_nsxt\_alb\_importable\_cloud

Supported in provider *v3.4+* and VCD 10.2+ with NSX-T and ALB.

Provides a data source to reference existing NSX-T ALB Importable Clouds. An NSX-T Importable Cloud is a reference to a
Cloud configured in ALB Controller.

~> Only `System Administrator` can use this data source.

## Example Usage

```hcl
data "vcd_nsxt_alb_controller" "first" {
  name = "aviController1-renamed"
}

data "vcd_nsxt_alb_importable_cloud" "cld" {
  name          = "NSXT bos1-vcloud-static-171-68.eng.vmware.com"
  controller_id = data.vcd_nsxt_alb_controller.first.id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required)  - Name of NSX-T ALB Importable Cloud
* `controller_id` - (Required)  - NSX-T ALB Controller ID

## Attribute Reference

* `already_imported` - boolean value which displays if the ALB Importable Cloud is already consumed
* `network_pool_id` - backing network pool ID 
* `network_pool_name` - backing network pool ID
* `transport_zone_name` - backing transport zone name
