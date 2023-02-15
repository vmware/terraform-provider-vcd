---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_alb_settings"
sidebar_current: "docs-vcd-resource-nsxt-alb-settings"
description: |-
  Provides a resource to manage NSX-T ALB General Settings for particular NSX-T Edge Gateway. One can activate or
  deactivate NSX-T ALB for a defined Edge Gateway.
---

# vcd\_nsxt\_alb\_settings

Supported in provider *v3.5+* and VCD 10.2+ with NSX-T and ALB.

Provides a resource to manage NSX-T ALB General Settings for particular NSX-T Edge Gateway. One can activate or
deactivate NSX-T ALB for a defined Edge Gateway.

~> Only `System Administrator` can create this resource.

## Example Usage (Enabling NSX-T ALB on NSX-T Edge Gateway)

```hcl
data "vcd_nsxt_edgegateway" "existing" {
  org = "my-org"
  vdc = "nsxt-vdc"

  name = "nsxt-gw"
}

resource "vcd_nsxt_alb_settings" "org1" {
  org = "my-org"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  is_active       = true

  # Optional definition of service network for the ALB. "192.168.255.125/25" is the default one.
  # service_network_specification = "192.168.255.125/25"
}
```

## Example Usage (Enabling NSX-T ALB with IPv6 service network and transparent mode on NSX-T Edge Gateway)

```hcl
data "vcd_nsxt_edgegateway" "existing" {
  org = "my-org"
  vdc = "nsxt-vdc"

  name = "nsxt-gw"
}

resource "vcd_nsxt_alb_settings" "org1" {
  org = "my-org"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  is_active       = true

  service_network_specification      = "10.10.255.225/27"
  ipv6_service_network_specification = "2001:0db8:85a3:0000:0000:8a2e:0370:7334/120"
  is_transparent_mode_enabled        = true
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which the edge gateway belongs. Optional if defined at provider level.
* `edge_gateway_id` - (Required) An ID of NSX-T Edge Gateway. Can be looked up using
  [vcd_nsxt_edgegateway](/providers/vmware/vcd/latest/docs/data-sources/nsxt_edgegateway) data source
* `is_active` - (Required) Boolean value `true` or `false` if ALB is enabled. **Note** Delete operation of this resource
  will set it to `false`
* `supported_feature_set` - (Optional; *v3.7+*) Feature set of this Edge Gateway if ALB is enabled (`STANDARD` or `PREMIUM`)
* `is_transparent_mode_enabled` - (Optional; *v3.9+*, *VCD 10.4.1+*) When enabled, it allows to
  configure Preserve Client IP on a Virtual Service
* `service_network_specification` - (Optional) Gateway CIDR format which will be used by Load Balancer service. All the
  load balancer service engines associated with the Service Engine Group will be attached to this network. The subnet
  prefix length must be 25. If nothing is set, the **default is 192.168.255.125/25**. This field cannot be updated
* `ipv6_service_network_specification` (Optional; *v3.9+*, *VCD 10.4.0+*) The IPv6 network definition in Gateway
  CIDR format which will be used by Load Balancer service on Edge. All the load balancer service
  engines associated with the Service Engine Group will be attached to this network. This field
  cannot be updated

~> IPv4 service network will be used if both the `service_network_specification` and
`ipv6_service_network_specification` properties are unset. If both are set, it will still be one
service network with a dual IPv4 and IPv6 stack.

~> The attribute `supported_feature_set` must not be used in VCD versions lower than 10.4. Starting with 10.4, it replaces `license_type` field in [nsxt_alb_controller](/providers/vmware/vcd/latest/docs/resources/nsxt_alb_controller).

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing NSX-T ALB General Settings configuration can be [imported][docs-import] into this resource via supplying
path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxt_alb_settings.imported my-org.my-org-vdc-org-vdc-group-name.my-nsxt-edge-gateway-name
```

The above would import the NSX-T ALB General Settings for Edge Gateway named
`my-nsxt-edge-gateway-name` in Org `my-org` and VDC or VDC Group `my-org-vdc-org-vdc-group-name`.