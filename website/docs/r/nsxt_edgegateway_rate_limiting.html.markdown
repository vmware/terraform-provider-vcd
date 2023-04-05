---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_edgegateway_rate_limiting"
sidebar_current: "docs-vcd-resource-nsxt-edge-rate-limiting"
description: |-
  Provides a resource to manage NSX-T Edge Gateway Rate Limiting configuration.
---

# vcd\_nsxt\_edgegateway\_rate\_limiting

Supported in provider *v3.9+* and VCD 10.3.2+ with NSX-T.

Provides a resource to manage NSX-T Edge Gateway Rate Limiting configuration.

~> Only `System Administrator` can create this resource.

## Example Usage

```hcl
data "vcd_nsxt_manager" "nsxt" {
  name = "nsxManager1"
}

data "vcd_nsxt_edgegateway_qos_profile" "qos-1" {
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
  name = "qos-policy-1"
}

data "vcd_org_vdc" "v1" {
  org  = "datacloud"
  name = "nsxt-vdc-datacloud"
}

data "vcd_nsxt_edgegateway" "testing-in-vdc" {
  org      = "datacloud"
  owner_id = data.vcd_org_vdc.v1.id

  name = "nsxt-gw-datacloud"
}

resource "vcd_nsxt_edgegateway_rate_limiting" "testing-in-vdc" {
  org             = "datacloud"
  edge_gateway_id = data.vcd_nsxt_edgegateway.testing-in-vdc.id

  ingress_profile_id = data.vcd_nsxt_edgegateway_qos_profile.qos-1.id
  egress_profile_id  = data.vcd_nsxt_edgegateway_qos_profile.qos-1.id
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Required) Org in which the NSX-T Edge Gateway is located
* `edge_gateway_id` - (Required) NSX-T Edge Gateway ID
* `ingress_profile_id` - (Optional) A QoS profile to apply for ingress traffic. *Note* it will be
  `unlimited` if not set.
* `egress_profile_id` - (Optional) A QoS profile to apply for egress traffic. *Note* it will be
  `unlimited` if not set.

-> Ingress and Egress profile IDs can be looked up using
  [`vcd_nsxt_edgegateway_qos_profile`](/providers/vmware/vcd/latest/docs/resources/nsxt_edgegateway_qos_profile)
  data source. 

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing NSX-T Edge Gateway Rate Limiting configuration can be [imported][docs-import] into this
resource via supplying path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxt_edgegateway_rate_limiting.imported my-org.nsxt-vdc.nsxt-edge
```

The above would import the `nsxt-edge` Edge Gateway Rate Limiting configuration for this particular
Edge Gateway.
