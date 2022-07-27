---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_edgegateway_bgp_ip_prefix_list"
sidebar_current: "docs-vcd-resource-nsxt-edgegateway-bgp-ip-prefix-list"
description: |-
  Provides a resource to manage NSX-T Edge Gateway BGP IP Prefix Lists. IP prefix lists can contain 
  single or multiple IP addresses and can be used to assign BGP neighbors with access permissions 
  for route advertisement.
---

# vcd\_nsxt\_edgegateway\_bgp\_ip\_prefix\_list

Supported in provider *v3.7+* and VCD 10.2+ with NSX-T

Provides a resource to manage NSX-T Edge Gateway BGP IP Prefix Lists. IP prefix lists can contain
single or multiple IP addresses and can be used to assign BGP neighbors with access permissions for
route advertisement.


## Example Usage (BGP IP Prefix List with multiple entries)

```hcl
resource "vcd_nsxt_edgegateway_bgp_ip_prefix_list" "testing" {
  org = "cloud"

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  name        = "sample-ip-prefix-list"
  description = "This definition is meant only to demostrate capabilities"

  ip_prefix {
    network = "10.10.10.0/24"
    action  = "PERMIT"
  }

  ip_prefix {
    network = "20.10.10.0/24"
    action  = "DENY"
  }

  ip_prefix {
    network = "2001:db8::/48"
    action  = "DENY"
  }

  ip_prefix {
    network                  = "30.10.10.0/24"
    action                   = "DENY"
    greater_than_or_equal_to = "25"
    less_than_or_equal_to    = "27"
  }

  ip_prefix {
    network                  = "40.0.0.0/8"
    action                   = "PERMIT"
    greater_than_or_equal_to = "16"
    less_than_or_equal_to    = "24"
  }
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
  when connected as sysadmin working across different organisations
* `edge_gateway_id` - (Required) The ID of the Edge Gateway (NSX-T only). Can be looked up using
  `vcd_nsxt_edgegateway` datasource
* `name` - (Required) The Name of IP Prefix List
* `description` - (Optional) Description of IP Prefix List
* `ip_prefix` - (Required) At least one `ip_prefix` definition. See [IP Prefix](#ip-prefix) for
  definition structure.

<a id="ip-prefix"></a>
## IP Prefix

Each member ip_prefix contains following attributes:

* `network` - (Required) Network information should be in CIDR notation. (e.g. IPv4
  192.168.100.0/24, IPv6 2001:db8::/48)
* `action` - (Required) Can be `PERMIT` or `DENY`
* `greater_than_or_equal_to` - (Optional) Greater than or equal to (ge) subnet mask to match. For
  example, 192.168.100.3/27 ge 26 le 32 modifiers match subnet masks greater than or equal to
  26-bits and less than or equal to 32-bits in length.
* `less_than_or_equal_to` - (Optional) Less than or equal to (le) subnet mask to match. For example,
  192.168.100.3/27 ge 26 le 32 modifiers match subnet masks greater than or equal to 26-bits and
  less than or equal to 32-bits in length.

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing BGP IP Prefix List configuration can be [imported][docs-import] into this resource
via supplying path for it. An example is
below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxt_edgegateway_bgp_ip_prefix_list.imported `my-org.my-vdc-or-vdc-group.my-edge-gateway.ip-prefix-list-name`
```

The above would import the `ip-prefix-list-name` BGP IP Prefix List that is defined in
`my-edge-gateway` NSX-T Edge Gateway. Edge Gateway should be located in `my-vdc-or-vdc-group` VDC ir
VDC Group in Org `my-org`
