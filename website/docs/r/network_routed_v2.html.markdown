---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_network_routed_v2"
sidebar_current: "docs-vcd-resource-network-routed-v2"
description: |-
  Provides a VMware Cloud Director Org VDC routed Network. This can be used to create, modify, and
  delete routed VDC networks (backed by NSX-T or NSX-V).
---

# vcd\_network\_routed\_v2

Provides a VMware Cloud Director Org VDC routed Network. This can be used to create, modify, and
delete routed VDC networks (backed by NSX-T or NSX-V).

Supported in provider *v3.2+* for both NSX-T and NSX-V VDCs.

-> Starting with **v3.6.0** Terraform provider VCD supports NSX-T VDC Groups and `vdc` fields (in
resource and inherited from provider configuration) are deprecated. `vcd_network_routed_v2` will
inherit VDC or VDC Group membership from parent Edge Gateway specified in `edge_gateway_id` field.
More about VDC Group support in a [VDC Groups
guide](/providers/vmware/vcd/latest/docs/guides/vdc_groups).

## Example Usage (NSX-T backed routed Org VDC network)

```hcl
resource "vcd_network_routed_v2" "nsxt-backed" {
  org         = "my-org"
  name        = "nsxt-routed 1"
  description = "My routed Org VDC network backed by NSX-T"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  gateway       = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "1.1.1.10"
    end_address   = "1.1.1.20"
  }

  static_ip_pool {
    start_address = "1.1.1.100"
    end_address   = "1.1.1.103"
  }
}
```

## Example Usage (NSX-T backed routed Org VDC network and a DHCP pool)

```hcl
resource "vcd_network_routed_v2" "parent-network" {
  name = "nsxt-routed-dhcp"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  gateway       = "7.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "7.1.1.10"
    end_address   = "7.1.1.20"
  }
}

resource "vcd_nsxt_network_dhcp" "pools" {
  org_network_id = vcd_network_routed_v2.parent-network.id

  pool {
    start_address = "7.1.1.100"
    end_address   = "7.1.1.110"
  }

  pool {
    start_address = "7.1.1.111"
    end_address   = "7.1.1.112"
  }
}
```

## Example Usage (Dual-Stack mode with IPv4 and IPv6)

```hcl
resource "vcd_network_routed_v2" "ipv6-dualstack" {
  name = "Dual Stack Routed Network"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  gateway       = "192.168.1.1"
  prefix_length = 24
  static_ip_pool {
    start_address = "192.168.1.10"
    end_address   = "192.168.1.20"
  }

  dual_stack_enabled      = true
  secondary_gateway       = "2002:0:0:1234:abcd:ffff:c0a6:121"
  secondary_prefix_length = 124

  secondary_static_ip_pool {
    start_address = "2002:0:0:1234:abcd:ffff:c0a6:122"
    end_address   = "2002:0:0:1234:abcd:ffff:c0a6:124"
  }
}
```

## Example usage (IPv6 only)
```hcl
resource "vcd_network_routed_v2" "ipv6" {
  name = "IPv6 network"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  gateway       = "2002:0:0:1234:abcd:ffff:c0a7:121"
  prefix_length = 124

  static_ip_pool {
    start_address = "2002:0:0:1234:abcd:ffff:c0a7:122"
    end_address   = "2002:0:0:1234:abcd:ffff:c0a7:124"
  }
}
```

## Example Usage (NSX-V backed routed Org VDC network using `subinterface` NIC)

```hcl
resource "vcd_network_routed_v2" "nsxv-backed" {
  org         = "my-org"
  name        = "nsxv-routed-network"
  description = "NSX-V routed network"

  interface_type = "subinterface"

  edge_gateway_id = data.vcd_edgegateway.existing.id

  gateway       = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "1.1.1.10"
    end_address   = "1.1.1.20"
  }
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when
  connected as sysadmin working across different organisations
* `vdc` - (Deprecated; Optional) The name of VDC to use. *v3.6+* inherits parent VDC or VDC Group
  from `edge_gateway_id`)
* `name` - (Required) A unique name for the network
* `description` - (Optional) An optional description of the network
* `interface_type` - (Optional) An interface for the network. One of `internal` (default), `subinterface`, 
  `distributed` (requires the edge gateway to support distributed networks). NSX-T supports only `internal`
* `edge_gateway_id` - (Required) The ID of the Edge Gateway (NSX-V or NSX-T)
* `gateway` - (Required) The gateway for this network (e.g. 192.168.1.1, 2002:0:0:1234:abcd:ffff:c0a7:121)
* `prefix_length` - (Required) The prefix length for the new network (e.g. 24 for netmask 255.255.255.0).
* `dns1` - (Optional) First DNS server to use.
* `dns2` - (Optional) Second DNS server to use.
* `dns_suffix` - (Optional) A FQDN for the virtual machines on this network
* `static_ip_pool` - (Optional) A range of IPs permitted to be used as static IPs for
  virtual machines; see [IP Pools](#ip-pools) below for details.
* `metadata` - (Deprecated; *v3.6+*) Use `metadata_entry` instead. Key value map of metadata to assign to this network. **Not supported** if the owner edge gateway belongs to a VDC Group.
* `metadata_entry` - (Optional; *v3.8+*) A set of metadata entries to assign. See [Metadata](#metadata) section for details.
* `dual_stack_enabled` - (Optional; *v3.10+*) Enables Dual-Stack mode so that one can configure one
  IPv4 and one IPv6 networks. **Note** In such case *IPv4* addresses must be used in `gateway`,
  `prefix_length` and `static_ip_pool` while *IPv6* addresses in `secondary_gateway`,
  `secondary_prefix_length` and `secondary_static_ip_pool` fields.
* `secondary_gateway` - (Optional; *v3.10+*) IPv6 gateway *when Dual-Stack mode is enabled*
* `secondary_prefix_length` - (Optional; *v3.10+*) IPv6 prefix length *when Dual-Stack mode is
  enabled*
* `secondary_static_ip_pool` - (Optional; *v3.10+*) One or more [IPv6 static
  pools](#secondary-ip-pools) *when Dual-Stack mode is enabled*

-> When using IPv6, VCD API will expand IP Addresses if they are specified using *double colon*
notation and it will cause inconsistent plan. (e.g. `2002::1234:abcd:ffff:c0a6:121` will be
converted to `2002:0:0:1234:abcd:ffff:c0a6:121`)

<a id="ip-pools"></a>
## IP Pools

~> `static_ip_pool` can be either *IPv4* or *IPv6* in non Dual-Stack mode (when
`dual_stack_enabled=false`) and only IPv4 if Dual-Stack mode is enabled. Please use *secondary*
fields `secondary_gateway`, `secondary_prefix_length` and `secondary_static_ip_pool` in such case 

Static IP Pools support the following attributes:

* `start_address` - (Required) The first address in the IP Range
* `end_address` - (Required) The final address in the IP Range

<a id="secondary-ip-pools"></a>
## Secondary IP Pools (IPv6 only)

Static IP Pools support the following attributes:

* `start_address` - (Required) The first address in the IP Range
* `end_address` - (Required) The final address in the IP Range

<a id="metadata"></a>
## Metadata

The `metadata_entry` (*v3.8+*) is a set of metadata entries that have the following structure:

* `key` - (Required) Key of this metadata entry.
* `value` - (Required) Value of this metadata entry.
* `type` - (Required) Type of this metadata entry. One of: `MetadataStringValue`, `MetadataNumberValue`, `MetadataDateTimeValue`, `MetadataBooleanValue`.
* `user_access` - (Required) User access level for this metadata entry. One of: `PRIVATE` (hidden), `READONLY` (read only), `READWRITE` (read/write).
* `is_system` - (Required) Domain for this metadata entry. true if it belongs to `SYSTEM`, false if it belongs to `GENERAL`.

~> Note that `is_system` requires System Administrator privileges, and not all `user_access` options support it.
   You may use `is_system = true` with `user_access = "PRIVATE"` or `user_access = "READONLY"`.

Example:

```hcl
resource "vcd_network_routed_v2" "example" {
  # ...
  metadata_entry {
    key         = "foo"
    type        = "MetadataStringValue"
    value       = "bar"
    user_access = "PRIVATE"
    is_system   = true # Requires System admin privileges
  }

  metadata_entry {
    key         = "myBool"
    type        = "MetadataBooleanValue"
    value       = "true"
    user_access = "READWRITE"
    is_system   = false
  }
}
```

To remove all metadata one needs to specify an empty `metadata_entry`, like:

```
metadata_entry {}
```

The same applies also for deprecated `metadata` attribute:

```
metadata = {}
```

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing routed network can be [imported][docs-import] into this resource via supplying its path.
The path for this resource is made of `OrgName.vdc-or-vdc-group-name.NetworkName`.
For example, using this structure, representing a routed network that was **not** created using Terraform:

```hcl
resource "vcd_network_routed_v2" "tf-mynet" {
  name = "my-net"
  org  = "my-org"
  # ...
}
```

You can import such routed network into terraform state using this command

```
terraform import vcd_network_routed_v2.tf-mynet my-org.my-vdc.my-net
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After importing, if you run `terraform plan` you will see the rest of the values and modify the script accordingly for
further operations.
