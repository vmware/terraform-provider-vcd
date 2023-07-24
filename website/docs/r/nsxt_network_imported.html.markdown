---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_network_imported"
sidebar_current: "docs-vcd-resource-nsxt-network-imported"
description: |-
  Provides a VMware Cloud Director Org VDC NSX-T Imported Network type. This can be used to create, modify, and delete NSX-T VDC networks of Imported type (backed by NSX-T).
---

# vcd\_nsxt\_network\_imported

Provides a VMware Cloud Director Org VDC NSX-T Imported Network type. This can be used to create, modify, and delete NSX-T VDC networks of Imported type (backed by NSX-T).

Supported in provider *v3.2+* for NSX-T VDCs only.

-> Starting with **v3.6.0** Terraform provider VCD supports NSX-T VDC Groups and `vdc` fields (in
resource and inherited from provider configuration) are deprecated. New field `owner_id` supports
IDs of both VDC and VDC Groups. More about VDC Group support in a [VDC Groups
guide](/providers/vmware/vcd/latest/docs/guides/vdc_groups).

-> This is **not Terraform imported** resource, but a special **Imported** type of **Org VDC
network** in NSX-T VDC. Read more about Imported Network in [official VCD
documentation](https://docs.vmware.com/en/VMware-Cloud-Director/10.3/VMware-Cloud-Director-Tenant-Portal-Guide/GUID-FB303D62-67EA-4209-BE4D-C3746481BCC8.html).

## Example Usage (NSX-T backed imported Org VDC network backed by NSX-T logical switch)
```hcl
data "vcd_org_vdc" "main" {
  org  = "my-org"
  name = "my-nsxt-org-vdc"
}

resource "vcd_nsxt_network_imported" "nsxt-backed" {
  org      = "my-org"
  owner_id = data.vcd_org_vdc.main.id

  name        = "nsxt-imported"
  description = "My NSX-T VDC Imported network type"

  nsxt_logical_switch_name = "nsxt_segment_name"

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

## Example Usage (NSX-T backed imported Org VDC network in Dual-Stack mode)
```hcl
resource "vcd_nsxt_network_imported" "ipv6-dualstack" {
  vdc  = vcd_org_vdc.with-edge-cluster.name
  name = "dual-stack-imported"

  nsxt_logical_switch_name = "segment-cloud"

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

## Example Usage (NSX-T backed imported Org VDC network backed by Distributed Virtual Port Group - DVPG)

```hcl
data "vcd_org_vdc" "main" {
  org  = "my-org"
  name = "my-nsxt-org-vdc"
}

resource "vcd_nsxt_network_imported" "nsxt-backed" {
  org      = "my-org"
  owner_id = data.vcd_org_vdc.main.id

  name      = "nsxt-imported"
  dvpg_name = "vc-dvpg"

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


## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when
  connected as sysadmin working across different organisations
* `owner_id` - (Optional) VDC or VDC Group ID. Always takes precedence over `vdc` fields (in resource
and inherited from provider configuration)
* `vdc` - (Deprecated; Optional) The name of VDC to use. **Deprecated**  in favor of new field
  `owner_id` which supports VDC and VDC Group IDs.
* `name` - (Required) A unique name for the network
* `nsxt_logical_switch_name` - (Optional) Unique name of an existing NSX-T segment. 
  **Note** it will never be refreshed because API does not allow reading this name after it is
  consumed. Instead ID will be stored in `nsxt_logical_switch_id` attribute.
  
  This resource **will fail** if multiple segments with the same name are available. One can rename 
  them in NSX-T manager to make them unique.
* `dvpg_name` - (Optional) Unique name of an existing Distributed Virtual Port Group (DVPG). 
  **Note** it will never be refreshed because API does not allow reading this name after it is
  consumed. Instead ID will be stored in `dvpg_id` attribute.

-> One of `nsxt_logical_switch_name` or `dvpg_name` must be provided.

* `description` - (Optional) An optional description of the network
* `gateway` - (Required) The gateway for this network (e.g. 192.168.1.1, 2002:0:0:1234:abcd:ffff:c0a7:121)
* `prefix_length` - (Required) The prefix length for the new network (e.g. 24 for netmask 255.255.255.0).
* `dns1` - (Optional) First DNS server to use.
* `dns2` - (Optional) Second DNS server to use.
* `dns_suffix` - (Optional) A FQDN for the virtual machines on this network
* `static_ip_pool` - (Optional) A range of IPs permitted to be used as static IPs for
  virtual machines; see [IP Pools](#ip-pools) below for details.
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

Static IP Pools  support the following attributes:

* `start_address` - (Required) The first address in the IP Range
* `end_address` - (Required) The final address in the IP Range

<a id="secondary-ip-pools"></a>
## Secondary IP Pools (IPv6 only)

Static IP Pools support the following attributes:

* `start_address` - (Required) The first address in the IP Range
* `end_address` - (Required) The final address in the IP Range

## Attribute Reference
* `nsxt_logical_switch_id` - ID of NSX-T logical switch used by this network
* `dvpg_id` - ID of Distributed Virtual Port Group used by this network

## Importing

~> After import the fields `nsxt_logical_switch_name` and `dvpg_name` will remain empty because it
is impossible to read them in API once it is consumed by network.

~> The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]



An existing NSX-T VDC Imported network can be [imported][docs-import] into this Terraform resource via supplying its path.
The path for this resource is made of `org-name.vdc-or-vdc-group-name.network-name`.
For example, using this structure, representing an NSX-T Imported Network that was **not** created using Terraform:

```hcl
resource "vcd_nsxt_network_imported" "tf-mynet" {
  name = "my-net"
  org  = "my-org"
  vdc  = "my-vdc"
  # ...
}
```

You can import such NSX-T VDC Imported network type into terraform state using this command

```
terraform import vcd_nsxt_network_imported.tf-mynet my-org.my-vdc.my-net
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After importing, if you run `terraform plan` you will see the rest of the values and modify the script accordingly for
further operations.
