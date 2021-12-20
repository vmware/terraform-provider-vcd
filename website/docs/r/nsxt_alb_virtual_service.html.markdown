---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_alb_virtual_service"
sidebar_current: "docs-vcd-resource-nsxt-alb-virtual-service"
description: |-
  Provides a resource to manage NSX-T ALB Virtual services for particular NSX-T Edge Gateway. A virtual service advertises
  an IP address and ports to the external world and listens for client traffic. When a virtual service receives traffic,
  it directs it to members in ALB Pool.
---

# vcd\_nsxt\_alb\_virtual\_service

Supported in provider *v3.5+* and VCD 10.2+ with NSX-T and ALB.

Provides a resource to manage NSX-T ALB Virtual services for particular NSX-T Edge Gateway. A virtual service advertises
an IP address and ports to the external world and listens for client traffic. When a virtual service receives traffic,
it directs it to members in ALB Pool.

## Example Usage (Adding HTTP NSX-T ALB Virtual Service)
```hcl
data "vcd_nsxt_edgegateway" "existing" {
  org = "my-org"
  vdc = "nsxt-test-vdc"

  name = "nsxt-edge-gateway"
}

data "vcd_nsxt_alb_edgegateway_service_engine_group" "assignment" {
  org = "my-org"
  vdc = "nsxt-test-vdc"

  edge_gateway_id           = data.vcd_nsxt_edgegateway.existing.id
  service_engine_group_name = "known_service_engine_group"

  # ID reference to service_engine_group_id can also be supplied by 
  # using `vcd_nsxt_alb_service_engine_group` data source
  # but it requires provider level access therefore tenant can use `service_engine_group_name` field.
  # service_engine_group_id = data.vcd_nsxt_alb_service_engine_group.existing.id
}

resource "vcd_nsxt_alb_pool" "test" {
  org = "my-org"
  vdc = "nsxt-test-vdc"

  name            = "test-pool"
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
}

resource "vcd_nsxt_alb_virtual_service" "test" {
  org = "my-org"
  vdc = "nsxt-test-vdc"

  name            = "new-virtual-service"
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  pool_id                  = vcd_nsxt_alb_pool.test.id
  service_engine_group_id  = vcd_nsxt_alb_edgegateway_service_engine_group.assignment.service_engine_group_id
  virtual_ip_address       = tolist(data.vcd_nsxt_edgegateway.existing.subnet)[0].primary_ip
  application_profile_type = "HTTP"
  service_port {
    start_port = 80
    type       = "TCP_PROXY"
  }
}
```

## Example Usage (Adding L4 TLS NSX-T ALB Virtual Service with certificate and multiple ports)
```hcl
data "vcd_nsxt_edgegateway" "existing" {
  org = "my-org"
  vdc = "nsxt-test-vdc"

  name = "nsxt-edge-gateway"
}

data "vcd_nsxt_alb_edgegateway_service_engine_group" "assignment" {
  org = "my-org"
  vdc = "nsxt-test-vdc"

  edge_gateway_id           = data.vcd_nsxt_edgegateway.existing.id
  service_engine_group_name = "known_service_engine_group"

  # ID reference to service_engine_group_id can also be supplied by 
  # using `vcd_nsxt_alb_service_engine_group` data source
  # but it requires provider level access therefore tenant can use `service_engine_group_name` field.
  # service_engine_group_id = data.vcd_nsxt_alb_service_engine_group.existing.id
}

data "vcd_library_certificate" "org-cert-1" {
  org   = "my-org"
  alias = "My-cert"

}

resource "vcd_nsxt_alb_pool" "test" {
  org = "my-org"
  vdc = "nsxt-test-vdc"

  name            = "test-pool"
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
}

resource "vcd_nsxt_alb_virtual_service" "test" {
  org = "my-org"
  vdc = "nsxt-test-vdc"

  name            = "new-virtual-service"
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  pool_id                  = vcd_nsxt_alb_pool.test.id
  service_engine_group_id  = vcd_nsxt_alb_edgegateway_service_engine_group.assignment.service_engine_group_id
  virtual_ip_address       = tolist(data.vcd_nsxt_edgegateway.existing.subnet)[0].primary_ip
  ca_certificate_id        = data.vcd_library_certificate.org-cert-1.id
  application_profile_type = "L4_TLS"

  service_port {
    start_port  = 80
    type        = "TCP_PROXY"
    ssl_enabled = true
  }

  service_port {
    start_port  = 84
    end_port    = 85
    type        = "TCP_PROXY"
    ssl_enabled = true
  }

  service_port {
    start_port = 87
    type       = "TCP_PROXY"
  }
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
  when connected as sysadmin working across different organisations.
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level.
* `name` - (Required) A name for NSX-T ALB Virtual Service
* `edge_gateway_id` - (Required) An ID of NSX-T Edge Gateway. Can be looked up using
  [vcd_nsxt_edgegateway](/providers/vmware/vcd/latest/docs/data-sources/nsxt_edgegateway) data source
* `description` - (Optional) An optional description NSX-T ALB Virtual Service
* `pool_id` - (Required) A reference to NSX-T ALB Pool. Can be looked up using `vcd_nsxt_alb_pool` resource or data
  source
* `service_engine_group_id` - (Required) A reference to NSX-T ALB Service Engine Group. Can be looked up using
  `vcd_nsxt_alb_edgegateway_service_engine_group` resource or data source
* `application_profile_type` - (Required) One of `HTTP`, `HTTPS`, `L4`, `L4_TLS`. 
* `virtual_ip_address` - (Required) IP Address for the service to listen on.
* `ca_certificate_id` - (Optional) ID reference of CA certificate. Required when `application_profile_type` is `HTTPS`
  or `L4_TLS`
* `service_port` - (Required) A block to define port, port range and traffic type. Multiple can be used. See
  [service_port](#service-port-block) and example for usage details.


<a id="service-port-block"></a>
## Service Port

* `start_port` (Required) Starting port in the range or exact port number
* `end_port` (Optional) Only required to specify port range and is not needed for single port values
* `type` (Required) One of `TCP_PROXY`, `TCP_FAST_PATH`, `UDP_FAST_PATH`
* `ssl_enabled` (Optional) Must be enabled if CA certificate is to be used for this port. Default `false`

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing NSX-T ALB Virtual Service configuration can be [imported][docs-import] into this resource
via supplying path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxt_alb_virtual_service.imported my-org.my-vdc.my-edge-gateway.my-virtual-service-name
```

The above would import the `my-virtual-service-name` NSX-T ALB Virtual Service that is defined in NSX-T Edge Gateway
`my-edge-gateway` inside Org `my-org` and VDC `my-vdc`.
