---
layout: "vcd"
page_title: "vCloudDirector: vcd_nsxv_dnat"
sidebar_current: "docs-vcd-resource-nsxv-dnat"
description: |-
  Provides a vCloud Director DNAT resource. This can be used to create, modify, and delete destination NATs to map external IPs to a VM.
---

# vcd\_nsxv\_dnat

Provides a vCloud Director DNAT resource. This can be used to create, modify,
and delete destination NATs to map an external IP/port to an internal IP/port.

~> **Note:** This resource requires advanced edge gateway. For non-advanced edge gateways please
use the [`vcd_dnat`](/docs/providers/vcd/r/dnat.html) resource.

## Example Usage 1 (Minimal input)

```hcl
resource "vcd_nsxv_dnat" "web" {
  org = "my-org" # Optional
  vdc = "my-vdc" # Optional

  edge_gateway = "Edge Gateway Name"
  network_type = "ext"
  network_name = "my-external-network"

  original_address   = "1.1.1.1"
  translated_address = "10.10.10.15"
}
```

## Example Usage 2 (ICMP)

```hcl
resource "vcd_nsxv_dnat" "forIcmp" {
  org = "my-org" # Optional
  vdc = "my-vdc" # Optional
  
  edge_gateway = "Edge Gateway Name"
  network_name = "my-external-network"
  network_type = "ext"

  original_address   = "78.101.10.20-78.101.10.30"
  translated_address = "10.10.0.5"
  protocol           = "icmp"
  icmp_sub_type      = "router-advertisement"
}
```

## Example Usage 3 (All settings)

```hcl
resource "vcd_nsxv_dnat" "forIcmp" {
  org = "my-org" # Optional
  vdc = "my-vdc" # Optional
  
  edge_gateway = "Edge Gateway Name"
  network_name = "my-external-network"
  network_type = "ext"

  enabled = false
  logging_enabled = true
  description = "My dnat rule"

  original_address   = "78.101.10.20"
  original_port      = 443

  translated_address = "10.10.0.5"
  translated_port    = 8443
  protocol           = "tcp"

  dnat_match_source_address = "192.168.1.1/24"
  dnat_match_source_port    = "1-65535"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
when connected as sysadmin working across different organisations.
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level.
* `edge_gateway` - (Required) The name of the edge gateway on which to apply the DNAT rule.
* `network_type` - (Optional) Type of the network on which to apply the DNAT rule. Possible values
`org` or `ext`. Default is `org`.
* `network_name` - (Required) The name of the network on which to apply the DNAT rule.
* `enabled` - (Optional) Defines if the rule is enabaled. Default `true`.
* `logging_enabled` - (Optional) Defines if the logging for this rule is enabaled. Default `false`.
* `description` - (Optional) Free text description.
* `original_address` - (Required) IP address, range or subnet. This address must be the public IP
address of the edge gateway for which you are configuring the DNAT rule. In the packet being
inspected, this IP address or range would be those that appear as the destination IP address of the
packet. These packet destination addresses are the ones translated by this DNAT rule. 
* `original_port` - (Optional) Select the port or port range that the incoming traffic uses on the
edge gateway to connect to the internal network on which the virtual machines are connected. This
selection is not available when the Protocol is set to `icmp` or `any`. Default `any`.
* `translated_address` - (Required) IP address, range or subnet. IP addresses to which destination
addresses on inbound packets will be translated. These addresses are the IP addresses of the one or
more virtual machines for which you are configuring DNAT so that they can receive traffic from the
external network. 
* `translated_port` - (Optional) Select the port or port range that inbound traffic is connecting
to on the virtual machines on the internal network. These ports are the ones into which the DNAT
rule is translating for the packets inbound to the virtual machines.
* `protocol` - (Optional) Select the protocol to which the rule applies. One of `tcp`, `udp`,
`icmp`, `any`. Default `any`
protocols, select Any.
* `icmp_type` - (Optional) Only when `protocol` is set to `icmp`. One of `any`,
`address-mask-request`, `address-mask-reply`, `destination-unreachable`, `echo-request`,
`echo-reply`, `parameter-problem`, `redirect`, `router-advertisement`, `router-solicitation`,
`source-quench`, `time-exceeded`, `timestamp-request`, `timestamp-reply`. Default `any`
* `rule_tag` - (Optional) This can be used to specifyuser-controlled ruleId. If not specified,
NSX Manager will generate rule ID. Must be between 65537-131072.
* `dnat_match_source_address` - (Optional) Source address to match in DNAT rule.
* `dnat_match_source_port` - (Optional) Source port to match in DNAT rule.

## Attribute Reference

The following additional attributes are exported:

* `rule_type` - Possible values - `user`, `internal_high`.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing dnat rule can be [imported][docs-import] into this resource
via supplying the full dot separated path for DNAT rule. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxv_dnat.imported my-org.my-org-vdc.my-edge-gw.my-dnat-rule-id
```

The above would import the application rule named `my-dnat-rule-id` that is defined on edge
gateway `my-edge-gw` which is configured in organization named `my-org` and vDC named `my-org-vdc`.
