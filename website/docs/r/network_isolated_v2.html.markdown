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

## Example Usage (NSX-T backed isolated Org VDC network)

```hcl
resource "vcd_network_isolated_v2" "nsxt-backed" {
  org = "my-org"
  vdc = "my-nsxt-org-vdc"

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

## Example Usage (NSX-V backed isolated Org VDC network shared with other VDCs)

```hcl
resource "vcd_network_isolated_v2" "nsxv-backed" {
  org = "my-org"
  vdc = "my-nsxv-org-vdc"

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
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level
* `name` - (Required) A unique name for the network
* `description` - (Optional) An optional description of the network
* `is_shared` - (Optional) **NSX-V only.** Defines if this network is shared between multiple VDCs
  in the Org.  Defaults to `false`.
* `gateway` (Required) The gateway for this network (e.g. 192.168.1.1)
* `prefix_length` - (Required) The prefix length for the new network (e.g. 24 for netmask 255.255.255.0).
* `dns1` - (Optional) First DNS server to use.
* `dns2` - (Optional) Second DNS server to use.
* `dns_suffix` - (Optional) A FQDN for the virtual machines on this network
* `static_ip_pool` - (Optional) A range of IPs permitted to be used as static IPs for
  virtual machines; see [IP Pools](#ip-pools) below for details.

<a id="ip-pools"></a>
## IP Pools

Static IP Pools support the following attributes:

* `start_address` - (Required) The first address in the IP Range
* `end_address` - (Required) The final address in the IP Range

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing isolated network can be [imported][docs-import] into this resource via supplying its path.
The path for this resource is made of orgName.vdcName.networkName.
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
