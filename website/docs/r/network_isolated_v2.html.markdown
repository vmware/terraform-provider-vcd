---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_network_isolated_v2"
sidebar_current: "docs-vcd-resource-network-isolated-v2"
description: |-
  Provides a VMware Cloud Director Org VDC isolated Network. This can be used to create, modify, and
  delete isolated VDC networks (backed by NSX-T or NSX-V).
---

# vcd\_network\_isolated\_v2

Provides a VMware Cloud Director Org VDC isolated Network. This can be used to create, modify, and
delete isolated VDC networks (backed by NSX-T or NSX-V).

Supported in provider *v3.2+* for both NSX-T and NSX-V VDCs.

-> Starting with **v3.6.0** Terraform provider VCD supports NSX-T VDC Groups and `vdc` fields (in
resource and inherited from provider configuration) are deprecated. New field `owner_id` supports
IDs of both VDC and VDC Groups. More about VDC Group support in a [VDC Groups
guide](/providers/vmware/vcd/latest/docs/guides/vdc_groups).

## Example Usage (NSX-T backed isolated Org VDC network)

```hcl
data "vcd_org_vdc" "main" {
  org  = "my-org"
  name = "my-nsxt-org-vdc"
}

resource "vcd_network_isolated_v2" "nsxt-backed" {
  org      = "my-org"
  owner_id = data.vcd_org_vdc.main.id

  name        = "nsxt-isolated 1"
  description = "My isolated Org VDC network backed by NSX-T"

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

## Example Usage (NSX-T backed isolated IPv6 Org VDC network)

```hcl
resource "vcd_network_isolated_v2" "ipv6" {
  owner_id = vcd_org_vdc.with-edge-cluster.id

  name = "isolated-ipv6"

  gateway       = "2002:0:0:1234:abcd:ffff:c0a8:121"
  prefix_length = 124

  static_ip_pool {
    start_address = "2002:0:0:1234:abcd:ffff:c0a8:122"
    end_address   = "2002:0:0:1234:abcd:ffff:c0a8:123"
  }
}
```

## Example Usage (Dual-Stack mode with IPv4 and IPv6)

```hcl
resource "vcd_network_isolated_v2" "ipv6-dualstack" {
  owner_id = vcd_org_vdc.with-edge-cluster.id

  name = "Isolated Dual Stack"

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

## Example Usage (NSX-V backed isolated Org VDC network shared with other VDCs)

```hcl
data "vcd_org_vdc" "main" {
  org  = "my-org"
  name = "my-nsxt-org-vdc"
}

resource "vcd_network_isolated_v2" "nsxv-backed" {
  org      = "my-org"
  owner_id = data.vcd_org_vdc.main.id

  name        = "nsxv-isolated-network"
  description = "NSX-V isolated network"

  is_shared = true

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

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful 
  when connected as sysadmin working across different organisations
* `owner_id` - (Optional) VDC or VDC Group ID. Always takes precedence over `vdc` fields (in resource
and inherited from provider configuration)
* `vdc` - (Deprecated; Optional) The name of VDC to use. **Deprecated**  in favor of new field
  `owner_id` which supports VDC and VDC Group IDs.
* `name` - (Required) A unique name for the network
* `description` - (Optional) An optional description of the network
* `is_shared` - (Optional) **NSX-V only.** Defines if this network is shared between multiple VDCs
  in the Org.  Defaults to `false`.
* `gateway` - (Required) The gateway for this network (e.g. 192.168.1.1, 2002:0:0:1234:abcd:ffff:c0a7:121)
* `prefix_length` - (Required) The prefix length for the new network (e.g. 24 for netmask 255.255.255.0).
* `dns1` - (Optional) First DNS server to use.
* `dns2` - (Optional) Second DNS server to use.
* `dns_suffix` - (Optional) A FQDN for the virtual machines on this network
* `static_ip_pool` - (Optional) A range of IPs permitted to be used as static IPs for
  virtual machines; see [IP Pools](#ip-pools) below for details.
* `metadata` - (Deprecated; *v3.6+*) Use `metadata_entry` instead. Key value map of metadata to assign to this network. **Not supported** if the network belongs to a VDC Group.
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
resource "vcd_network_isolated_v2" "example" {
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

An existing isolated network can be [imported][docs-import] into this resource via supplying its path.
The path for this resource is made of `org-name.vdc-or-vdc-group-name.network-name`.
For example, using this structure, representing a isolated network that was **not** created using Terraform:

```hcl
resource "vcd_network_isolated_v2" "tf-mynet" {
  name = "my-net"
  org  = "my-org"
  vdc  = "my-vdc"
  # ...
}
```

You can import such isolated network into terraform state using this command

```
terraform import vcd_network_isolated_v2.tf-mynet my-org.my-vdc.my-net
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After importing, if you run `terraform plan` you will see the rest of the values and modify the script accordingly for
further operations.
