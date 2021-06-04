---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_firewall"
sidebar_current: "docs-vcd-data-source-nsxt-firewall"
description: |-
  Provides a resource to manage NSX-T Firewall. Firewalls allow user to control the incoming and 
  outgoing network traffic to and from an NSX-T Data Center edge gateway, you create firewall rules.
---

# vcd\_nsxt\_firewall

Supported in provider *v3.3+* and VCD 10.1+ with NSX-T backed Edge Gateways.

Provides a resource to manage NSX-T Firewall. Firewalls allow user to control the incoming and 
outgoing network traffic to and from an NSX-T Data Center edge gateway, you create firewall rules.

## Example Usage 1 (Single rule to allow all IPv4 traffic from anywhere to anywhere)
```hcl
resource "vcd_nsxt_firewall" "testing" {
  org  = "my-org"
  vdc  = "my-nsxt-vdc"

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  rule {
    name        = "allow all IPv4 traffic"
    direction   = "IN_OUT"
    ip_protocol = "IPV4"
  }
}
```

## Example Usage 2 (Multiple firewall rules - order matters)
```hcl
resource "vcd_nsxt_firewall" "testing" {
  org  = "my-org"
  vdc  = "my-nsxt-vdc"

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  # Rule #1 - Allows in IPv4 traffic from security group `vcd_nsxt_security_group.group1.id`
  rule {
    name        = "first rule"
    direction   = "IN"
    ip_protocol = "IPV4"
    sources     = [vcd_nsxt_security_group.frontend.id]
  }

  # Rule #2 - Drops and logs all outgoing IPv6 traffic to `vcd_nsxt_security_group.group.2.id`
  rule {
    name         = "drop IPv6 with destination to security group 2"
    direction    = "OUT"
    ip_protocol  = "IPV6"
    destinations = [vcd_nsxt_security_group.group2.id]
    action       = "DROP"
    logging      = true
  }
  
  # Rule #3 - Allows IPv4 and IPv6 traffic in both directions:
  # from vcd_nsxt_security_group.group.1.id to all list of security groups vcd_nsxt_security_group.group.*.id
  # from list of security groups vcd_nsxt_security_group.group.*.id to vcd_nsxt_security_group.group.1.id
  rule {
    name         = "test_rule-3"
    direction    = "IN_OUT"
    ip_protocol  = "IPV4_IPV6"
    sources      = [vcd_nsxt_security_group.group.1.id]
    destinations = vcd_nsxt_security_group.group.*.id
  }
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
  when connected as sysadmin working across different organisations.
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level.
* `edge_gateway_id` - (Required) The ID of the edge gateway (NSX-T only). Can be looked up using
  `vcd_nsxt_edgegateway` datasource
* `rule` (Required) One or more blocks with [Firewall Rule](#firewall-rule) definitions

<a id="firewall-rule"></a>
## Firwall Rule

Each Firewall Rule contains following attributes:

* `name` - (Required) Explanatory name for firewall rule (uniqueness not enforced)
* `direction` - (Required) One of `IN`, `OUT`, or `IN_OUT`
* `ip_protocol` - (Required) One of `IPV4`,  `IPV6`, or `IPV4_IPV6`
* `enabled` - (Optional) Defines if the rule is enabled (default `true`)
* `logging` - (Optional) Defines if logging for this rule is enabled (default `false`)
* `action` - (Optional) Defines if it should `ALLOW` or `DROP` traffic (default `ALLOW`)
* `sources` - (Optional) A set of source object Firewall Groups (`IP Sets` or `Security groups`). 
Leaving it empty matches `Any` (all)
* `destinations` - (Optional) A set of source object Firewall Groups (`IP Sets` or `Security groups`). 
Leaving it empty matches `Any` (all)
* `applications` - (Optional) A set of application port profiles. Leaving it empty matches  (all)

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

Existing Firewall Rules can be [imported][docs-import] into this resource
via supplying the full dot separated path for your Edge Gateway name. An example is
below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxt_firewall.imported my-org.my-org-vdc.my-nsxt-edge-gateway
```

The above would import all firewall rules defined on NSX-T Edge Gateway `my-nsxt-edge-gateway` which
is configured in organization named `my-org` and VDC named `my-org-vdc`.
