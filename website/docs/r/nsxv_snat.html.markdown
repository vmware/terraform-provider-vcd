---
layout: "vcd"
page_title: "vCloudDirector: vcd_nsxv_snat"
sidebar_current: "docs-vcd-resource-nsxv-snat"
description: |-
  Provides a vCloud Director SNAT resource for advanced edge gateways (NSX-V). This
  can be used to create, modify, and delete source NATs to allow vApps to send external traffic.
---

# vcd\_nsxv\_snat

Provides a vCloud Director SNAT resource for advanced edge gateways (NSX-V). This
can be used to create, modify, and delete source NATs to allow vApps to send external traffic. Replaces
[`vcd_snat`](/docs/providers/vcd/r/snat.html) resource.

~> **Note:** This resource requires advanced edge gateway. For non-advanced edge gateways please
use the [`vcd_snat`](/docs/providers/vcd/r/snat.html) resource.

## Example Usage

```hcl
resource "vcd_nsxv_snat" "web" {
  org = "my-org" # Optional
  vdc = "my-vdc" # Optional

  edge_gateway = "Edge Gateway Name"
  network_type = "org"
  network_name = "my-org-network"

  original_address   = "10.10.10.15/24"
  translated_address = "78.101.10.20"
}
```


## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
when connected as sysadmin working across different organisations.
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level.
* `edge_gateway` - (Required) The name of the edge gateway on which to apply the SNAT rule.
* `network_type` - (Required) Type of the network on which to apply the DNAT rule. Possible values
`org` or `ext`.
* `network_name` - (Required) The name of the network on which to apply the SNAT rule.
* `enabled` - (Optional) Defines if the rule is enabaled. Default `true`.
* `logging_enabled` - (Optional) Defines if the logging for this rule is enabaled. Default `false`.
* `description` - (Optional) Free text description.
* `rule_tag` - (Optional) This can be used to specify user-controlled rule tag. If not specified,
it will report rule ID after creation. Must be between 65537-131072.
* `original_address` - (Required) IP address, range or subnet. These addresses are the IP addresses
of one or more virtual machines for which you are configuring the SNAT rule so that they can send
traffic to the external network. 
* `translated_address` - (Required) IP address, range or subnet. This address is always the public
IP address of the gateway for which you are configuring the SNAT rule. Specifies the IP address to
which source addresses (the virtual machines) on outbound packets are translated to when they send
traffic to the external network. 

## Attribute Reference

The following additional attributes are exported:

* `rule_type` - Possible values - `user`, `internal_high`.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing dnat rule can be [imported][docs-import] into this resource
via supplying the full dot separated path for SNAT rule. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxv_dnat.imported my-org.my-org-vdc.my-edge-gw.my-snat-rule-id
```

The above would import the application rule named `my-snat-rule-id` that is defined on edge
gateway `my-edge-gw` which is configured in organization named `my-org` and vDC named `my-org-vdc`.
