---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_edgegateway_dns"
sidebar_current: "docs-vcd-resource-nsxt-edgegateway-dns"
description: |-
  Provides a resource to manage NSX-T Edge Gateway DNS configuration.
---

# vcd\_nsxt\_edgegateway\_dns

Supported in provider *v3.11+* and VCD 10.4.0+ with NSX-T.

Provides a resource to manage NSX-T Edge Gateway DNS configuration.

## Example Usage

```hcl
data "vcd_org_vdc" "v1" {
  org  = "datacloud"
  name = "nsxt-vdc-datacloud"
}

data "vcd_nsxt_edgegateway" "testing-in-vdc" {
  org      = "datacloud"
  owner_id = data.vcd_org_vdc.v1.id

  name = "nsxt-gw-datacloud"
}

resource "vcd_nsxt_edgegateway_dns" "testing-in-vdc" {
  org             = "datacloud"
  edge_gateway_id = data.vcd_nsxt_edgegateway.testing-in-vdc.id
  enabled         = true

  default_forwarder_zone {
    name = "testing"

    upstream_servers = [
      "1.1.1.1",
      "2.2.2.2",
    ]
  }

  conditional_forwarder_zone {
    name = "conditional_testing"

    upstream_servers = [
      "3.3.3.3",
      "4.4.4.4",
    ]

    domain_names = [
      "example.org",
      "nonexistent.org",
    ]
  }
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) Org in which the NSX-T Edge Gateway is located, required
  if not set in the provider section.
* `edge_gateway_id` - (Required) NSX-T Edge Gateway ID.
* `enabled` - (Optional) Status of the DNS forwarding service. Defaults to `true`.
* `listener_ip` - (Optional) The IP on which the DNS forwarder listens. If the Edge Gateway 
  has a dedicated external network, this can be changed.
* `snat_rule_ip_address` - (Optional, VCD 10.5.1+) This property only applies if the Edge Gateway 
  is connected to a Provider Gateway using IP Spaces. If specified, VCD will 
  conveniently manage the SNAT rule with the specified IP address for the DNS forwarder. 
  The specified IP can be allocated using [`vcd_ip_space_ip_allocation`](/providers/vmware/vcd/latest/docs/resources/ip_space_ip_allocation) 
* `default_forwarder_zone` - (Required) The default forwarder zone to use if 
  there’s no matching domain in the conditional forwarder zones. See [`default_forwarder_zone`](#default-forwarder-zone)
* `conditional_forwarder_zone` - (Optional) A set (up to 5) of conditional forwarder zones that allows to define 
  specific forwarding routes based on the domain. See [`conditional_forwarder_zone`](#conditional-forwarder-zone)

## Attribute Reference

* `snat_rule_enabled` - Is `true` if there exists a SNAT rule for the DNS forwarder. 
  If the Edge Gateway is connected to a dedicated provider gateway and `listener_ip`
  is modified manually, this field will be set to `false`, otherwise `true`.

## Default forwarder zone

* `name` - (Required) Name of the forwarder zone.
* `upstream_servers` - (Required) A set (up to 3) of IP addresses that the unmatched DNS
  requests should be forwarded to.

## Conditional forwarder zone

* `name` - (Required) Name of the forwarder zone.
* `domain_names` - (Required) Set of domain names on which conditional forwarding is based.
* `upstream_servers` - (Required) A set (up to 3) of IP addresses that the matched DNS requests 
  should be forwarded to.

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing NSX-T Edge Gateway DNS forwarder configuration can be [imported][docs-import] into this
resource via supplying path for it. An example is below:

```tf
data "vcd_org" "my_org" {
  name = "my-org"
}

data "vcd_org_vdc" "my-vdc-or-vdc-group" {
  name = "my-vdc"
  org  = "my-org"
}

data "vcd_nsxt_edgegateway" "my-edge-gateway" {
  name     = "my-edge-gateway"
  owner_id = data.vcd_org_vdc.my-vdc-or-vdc-group.id
}

resource "vcd_nsxt_edgegateway_dns" "dns-imported" {
  edge_gateway_id = data.vcd_nsxt_edgegateway.my-edge-gateway.id
}
```

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxt_edgegateway_dns.dns-imported my-org.nsxt-vdc.nsxt-edge
```

The above would import the `dns-imported` Edge Gateway DNS forwarder configuration for this particular
Edge Gateway.
