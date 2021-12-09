---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_alb_pool"
sidebar_current: "docs-vcd-resource-nsxt-alb-pool"
description: |-
  Provides a resource to manage NSX-T ALB Pools for particular NSX-T Edge Gateway. Pools maintain the list of servers
  assigned to them and perform health monitoring, load balancing, persistence. A pool may only be used or referenced by
  only one virtual service at a time.
---

# vcd\_nsxt\_alb\_pool

Supported in provider *v3.5+* and VCD 10.2+ with NSX-T and ALB.

Provides a resource to manage NSX-T ALB Pools for particular NSX-T Edge Gateway. Pools maintain the list of servers
assigned to them and perform health monitoring, load balancing, persistence. A pool may only be used or referenced by
only one virtual service at a time.

## Example Usage 1 (tiny example with defaults and single pool member)

```hcl
resource "vcd_nsxt_alb_pool" "first-pool" {
  org = "sample"
  vdc = "nsxt-vdc-sample"

  name            = "tiny-pool"
  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id

  member {
    ip_address = "192.168.1.1"
  }
}
```

## Example Usage 2 (more complex example with health monitors and persistence profile)

```hcl
resource "vcd_nsxt_alb_pool" "first-pool" {
  org = "sample"
  vdc = "nsxt-vdc-sample"

  name            = "configured-pool"
  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id

  algorithm                  = "LEAST_LOAD"
  default_port               = 9000
  graceful_timeout_period    = "0"
  passive_monitoring_enabled = false

  health_monitor {
    type = "PING"
  }

  persistence_profile {
    type = "CLIENT_IP"
  }

  member {
    enabled    = false
    ip_address = "192.168.1.1"
    port       = 8000
    ratio      = 2
  }

  member {
    ip_address = "192.168.1.2"
    ratio      = 4
  }
}
```

## Example Usage 3 (Using CA certificates)

```hcl
data "vcd_library_certificate" "sample-cert" {
  alias = "Sample-cert-1"
}

resource "vcd_nsxt_alb_pool" "sample-pool" {
  org = "sample"
  vdc = "nsxt-vdc-sample"

  name            = "sample-cert-pool"
  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id

  ca_certificate_ids = [data.vcd_library_certificate.sample-cert.id]
  cn_check_enabled   = true
  domain_names       = ["domain1", "domain2"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for NSX-T ALB Pool
* `description` - (Optional) An optional description NSX-T ALB Pool
* `enabled` - (Optional) Boolean value if NSX-T ALB Pool should be enabled (default `true`)
* `edge_gateway_id` - (Required) An ID of NSX-T Edge Gateway. Can be looked up using
  [vcd_nsxt_edgegateway](/providers/vmware/vcd/latest/docs/data-sources/nsxt_edgegateway) data source
* `algorithm` - (Optional) Optional algorithm for choosing pool members (default `LEAST_CONNECTIONS`). Other options
  contain `ROUND_ROBIN`, `CONSISTENT_HASH` (uses Source IP Address hash), `FASTEST_RESPONSE`, `LEAST_LOAD`,
  `FEWEST_SERVERS`, `RANDOM`, `FEWEST_TASKS`, `CORE_AFFINITY`
* `default_port` - (Optional) Default Port defines destination server port used by the traffic sent to the member
  (default `80`)
* `graceful_timeout_period` (Optional) Maximum time in minutes to gracefully disable pool member (default `1`). Special
  values are `0` (immediate) and `-1` (infinite)
* `passive_monitoring_enabled` (Optional) defines if client traffic should be used to check if pool member is up or down
  (default `true`)
* `ca_certificate_ids` - (Optional) A set of CA Certificates to be used when validating certificates presented by the
  pool members. Can be looked up using
  [vcd_library_certificate](/providers/vmware/vcd/latest/docs/data-sources/library_certificate) data source
* `cn_check_enabled` - (Optional) Specifies whether to check the common name of the certificate presented by the pool
  member
* `domain_names` - (Optional) A set of domain names which will be used to verify the common names or subject alternative
  names presented by the pool member certificates. It is performed only when common name check `cn_check_enabled` is
  enabled
* `member` - (Optional) A block to define pool members. Multiple can be used. See [Member](#member-block) and example
  for usage details.
* `persistence_profile` - (Optional) Persistence profile will ensure that the same user sticks to the same server for a
  desired duration of time. If the persistence profile is unmanaged by Cloud Director, updates that leave the values
  unchanged will continue to use the same unmanaged profile. Any changes made to the persistence profile will cause
  Cloud Director to switch the pool to a profile managed by Cloud Director. See [Persistence
  profile](#persistence-profile-block) and example for usage details.
* `health_monitor` - (Optional) A block to define health monitor. Multiple can be used. See [Health
  monitor](#health-monitor-block) and example for usage details.

<a id="member-block"></a>
## Member

* `ip_address` (Required) IP address of pool member. 
* `enabled` (Optional) defines if the pool member is enabled to receive traffic (default `true`)
* `port` (Optional) Port for receiving traffic - overrides the root value `default_port` for individual members
* `ratio` (Optional) Ratio of selecting eligible servers in the pool. (default `1`)

### Attributes of members

* `health_status` - one of `UP`, `DOWN`, `DISABLED`.
* `detailed_health_message` - human-readable member health description. 
* `marked_down_by` - A set of health monitors that marked the member as `DOWN` 

<a id="persistence-profile-block"></a>
## Persistence profile

* `type` (Required) Type of persistence profile. One of:

  * `CLIENT_IP` - The clients IP is used as the identifier and mapped to the server

  * `HTTP_COOKIE` - Load Balancer inserts a cookie into HTTP responses. Cookie name must be provided as `value`

  * `CUSTOM_HTTP_HEADER` - Custom, static mappings of header values to specific servers are used. Header name must be provided as `value`

  * `APP_COOKIE` - Load Balancer reads existing server cookies or URI embedded data such as JSessionID. Cookie name must be provided as `value`

  * `TLS` - Information is embedded in the client's SSL/TLS ticket ID. This will use default system profile System-Persistence-TLS

* `value` (Optional) is required for some `type`s: ``HTTP_COOKIE`, `CUSTOM_HTTP_HEADER`, `APP_COOKIE`

### Attributes of persistence profile

* `name` - System generated name of Persistence Profile

<a id="health-monitor-block"></a>
## Health monitor

* `type` (Required) Type of health monitor. One of `HTTP`, `HTTPS`, `TCP`, `UDP`, `PING`

### Attributes of health monitor

* `name` - System generated name of Health monitor
* `system_defined` - A boolean flag if the Health monitor is system defined.

## Attribute Reference

The following attributes are exported on this resource:

* `associated_virtual_service_ids` - A set of associated Virtual Service IDs
* `associated_virtual_services` - A set of associated Virtual Service names
* `member_count` - Total number of members defined in the Pool
* `up_member_count` - Number of members defined in the Pool that are accepting traffic
* `enabled_member_count` - Number of enabled members defined in the Pool
* `health_message` - Health message of NSX-T ALB Pool 

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing NSX-T ALB pool configuration can be [imported][docs-import] into this resource
via supplying path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxt_alb_pool.imported my-org.my-vdc.my-edge-gateway.my-alb-pool
```

The above would import the `vcd_nsxt_alb_pool` NSX-T ALB Pool that is defined in VDC `my-vdc` of Org `my-org` for NSX-T
Edge Gateway `my-edge-gateway`
