---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_edgegateway"
sidebar_current: "docs-vcd-resource-nsxt-edge-gateway"
description: |-
  Provides a VMware Cloud Director NSX-T edge gateway. This can be used to create, update, and delete NSX-T edge gateways connected to external networks.
---

# vcd\_nsxt\_edgegateway

Provides a VMware Cloud Director NSX-T Edge Gateway. This can be used to create, update, and delete
NSX-T edge gateways connected to external networks.

~> **Note:** Only `System Administrator` can create an edge gateway.
You must use `System Adminstrator` account in `provider` configuration
and then provide `org` and `owner_id` arguments for Edge Gateway to work.

This resource supports **IP Spaces** - read [IP Spaces guide
page](https://registry.terraform.io/providers/vmware/vcd/latest/docs/guides/ip_spaces) for more
information.

## Example Usage (Simple case)

```hcl
data "vcd_external_network_v2" "nsxt-ext-net" {
  name = "nsxt-edge"
}

data "vcd_org_vdc" "vdc1" {
  name = "existing-vdc"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org         = "my-org"
  owner_id    = data.vcd_org_vdc.vdc1.id
  name        = "nsxt-edge"
  description = "Description"

  external_network_id = data.vcd_external_network_v2.nsxt-ext-net.id

  subnet {
    gateway       = "10.150.191.253"
    prefix_length = "19"
    # primary_ip should fall into defined "allocated_ips" 
    # range as otherwise next apply will report additional
    # range of "allocated_ips" with the range containing 
    # single "primary_ip" and will cause non-empty plan.
    primary_ip = "10.150.160.137"
    allocated_ips {
      start_address = "10.150.160.137"
      end_address   = "10.150.160.138"
    }
  }
}
```

## Example Usage (NSX-T Segment backed external networks attachment)

```hcl
resource "vcd_nsxt_edgegateway" "with-external-networks" {
  org      = "my-org"
  owner_id = data.vcd_org_vdc.vdc1.id
  name     = "edge-with-segment-uplinks"

  external_network_id = data.vcd_external_network_v2.existing-extnet.id

  subnet {
    gateway       = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].gateway
    prefix_length = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].prefix_length

    primary_ip = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
    allocated_ips {
      start_address = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
      end_address   = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
    }
  }

  # NSX-T Segment backed uplink with Automatic Primary IP allocation
  external_network {
    external_network_id = data.vcd_external_network_v2.segment-backed.id
    gateway             = tolist(data.vcd_external_network_v2.segment-backed.ip_scope)[0].gateway
    prefix_length       = tolist(data.vcd_external_network_v2.segment-backed.ip_scope)[0].prefix_length
    allocated_ip_count  = 4
  }

  # NSX-T Segment backed uplink with Primary IP
  external_network {
    external_network_id = data.vcd_external_network_v2.segment-backed2.id
    gateway             = tolist(data.vcd_external_network_v2.segment-backed2.ip_scope)[0].gateway
    prefix_length       = tolist(data.vcd_external_network_v2.segment-backed2.ip_scope)[0].prefix_length
    allocated_ip_count  = 4
    primary_ip          = "15.14.14.12"
  }
}
```

## Example Usage (IP Space backed Provider Gateway)

```hcl
data "vcd_external_network_v2" "ip-space-provider-gw" {
  name = "nsxt-edge"
}

data "vcd_org_vdc" "vdc1" {
  name = "existing-vdc"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org                 = "my-org"
  owner_id            = data.vcd_org_vdc.vdc1.id
  name                = "ip-space-backed-edge"
  external_network_id = data.vcd_external_network_v2.ip-space-provider-gw.id
}
```

<a id="subnet-example"></a>
## Example Usage (Using custom Edge Cluster and multiple subnets)

```hcl
data "vcd_nsxt_edge_cluster" "secondary" {
  name = "edge-cluster-two"
}

data "vcd_external_network_v2" "nsxt-ext-net" {
  name = "nsxt-edge"
}

data "vcd_org_vdc" "vdc1" {
  name = "existing-vdc"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org         = "my-org"
  owner_id    = data.vcd_org_vdc.vdc1.id
  name        = "nsxt-edge"
  description = "Description"

  external_network_id       = data.vcd_external_network_v2.nsxt-ext-net.id
  dedicate_external_network = true

  # Custom edge cluster reference
  edge_cluster_id = data.vcd_nsxt_edge_cluster.secondary.id

  subnet {
    gateway       = "10.150.191.253"
    prefix_length = "19"
    # primary_ip should fall into defined "allocated_ips" 
    # range as otherwise next apply will report additional
    # range of "allocated_ips" with the range containing 
    # single "primary_ip" and will cause non-empty plan.
    primary_ip = "10.150.160.137"
    allocated_ips {
      start_address = "10.150.160.137"
      end_address   = "10.150.160.137"
    }
  }

  subnet {
    gateway       = "77.77.77.1"
    prefix_length = "26"

    allocated_ips {
      start_address = "77.77.77.10"
      end_address   = "77.77.77.12"
    }
  }

  subnet {
    gateway       = "88.88.88.1"
    prefix_length = "24"

    allocated_ips {
      start_address = "88.88.88.91"
      end_address   = "88.88.88.92"
    }

    allocated_ips {
      start_address = "88.88.88.94"
      end_address   = "88.88.88.95"
    }

    allocated_ips {
      start_address = "88.88.88.97"
      end_address   = "88.88.88.98"
    }
  }
}
```


## Example Usage (Assigning NSX-T Edge Gateway to VDC Group)

```hcl
data "vcd_nsxt_edge_cluster" "secondary" {
  name = "edge-cluster-two"
}

data "vcd_external_network_v2" "nsxt-ext-net" {
  name = "nsxt-edge"
}

data "vcd_vdc_group" "group1" {
  name = "existing-group"
}

data "vcd_org_vdc" "vdc-1" {
  name = "existing-group"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org      = "my-org"
  owner_id = data.vcd_vdc_group.group1.id

  # VDC Group cannot be created directly in VDC Group - 
  # it must originate in some VDC (belonging to 
  # destination VDC Group)
  #
  # `starting_vdc_id` field is optional. If only VDC Group 
  # ID is specified in `owner_id` field - this resource will
  # will pick a random member VDC to precreate it and will 
  # move to destination VDC Group in a single apply cycle
  starting_vdc_id = data.vcd_org_vdc.vdc-1.id

  name        = "nsxt-edge"
  description = "Description"

  external_network_id       = data.vcd_external_network_v2.nsxt-ext-net.id
  dedicate_external_network = true

  # Custom edge cluster reference
  edge_cluster_id = data.vcd_nsxt_edge_cluster.secondary.id

  subnet {
    gateway       = "10.150.191.253"
    prefix_length = "19"
    primary_ip    = "10.150.160.137"
    allocated_ips {
      start_address = "10.150.160.137"
      end_address   = "10.150.160.137"
    }
  }

  subnet {
    gateway       = "77.77.77.1"
    prefix_length = "26"

    allocated_ips {
      start_address = "77.77.77.10"
      end_address   = "77.77.77.12"
    }
  }
}
```
<a id="subnet-with-total-ip-count-example"></a>
## Example Usage (Automatic IP allocation from any subnet)

```hcl
resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org      = "my-org"
  owner_id = data.vcd_org_vdc.vdc1.id
  name     = "nsxt-edge"

  external_network_id = data.vcd_external_network_v2.ext-net-nsxt.id

  # 100 IPs will be allocated from any of `subnet_with_total_ip_count` defined blocks
  total_allocated_ip_count = 100

  subnet_with_total_ip_count {
    gateway       = "77.77.77.1"
    prefix_length = "24"
    primary_ip    = "77.77.77.254"
  }

  subnet_with_total_ip_count {
    gateway       = "88.77.77.1"
    prefix_length = "24"
  }
}
```
<a id="subnet-with-ip-count-example"></a>
## Example Usage (Automatic IP allocation per subnet)

```hcl
resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org      = "my-org"
  owner_id = data.vcd_org_vdc.vdc1.id
  name     = "nsxt-edge"

  external_network_id = data.vcd_external_network_v2.ext-net-nsxt.id

  subnet_with_ip_count {
    gateway            = "77.77.77.1"
    prefix_length      = "24"
    primary_ip         = "77.77.77.10"
    allocated_ip_count = 9
  }

  subnet_with_ip_count {
    gateway            = "88.77.77.1"
    prefix_length      = "24"
    allocated_ip_count = 15
  }
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which the VDC belongs. Optional if defined at provider level.
* `vdc` - (Optional) **Deprecated** in favor of `owner_id`. The name of VDC that owns the edge
  gateway. Can be inherited from `provider` configuration if not defined here.
* `owner_id` - (Optional, *v3.6+*,*VCD 10.2+*) The ID of VDC or VDC Group. **Note:** Data sources
  [vcd_vdc_group](/providers/vmware/vcd/latest/docs/data-sources/vdc_group) or
  [vcd_org_vdc](/providers/vmware/vcd/latest/docs/data-sources/org_vdc) can be used to lookup IDs by
  name.

~> Only one of `vdc` or `owner_id` can be specified. `owner_id` takes precedence over `vdc`
definition at provider level.

~> When a VDC Group ID is specified in `owner_id` field, the Edge Gateway will be created in VDC
  (random member of VDC Group or specified in `starting_vdc_id`). Main use case of `starting_vdc_id`
  is to pick egress traffic origin for multi datacenter VDC Groups.

* `starting_vdc_id` - (Optional, *v3.6+*,*VCD 10.2+*)  If `owner_id` is a VDC Group, by default Edge
  Gateway will be created in random member VDC and moved to destination VDC Group. This field allows
  to specify initial VDC for Edge Gateway (this can define Egress location of traffic in the VDC
  Group) **Note:** It can only be used when `owner_id` is a VDC Group. 

* `name` - (Required) A unique name for the edge gateway.
* `description` - (Optional) A unique name for the edge gateway.
* `external_network_id` - (Required) An external network ID. **Note:** Data source [vcd_external_network_v2](/providers/vmware/vcd/latest/docs/data-sources/external_network_v2)
can be used to lookup ID by name.
* `edge_cluster_id` - (Optional) Specific Edge Cluster ID if required
* `dedicate_external_network` - (Optional) Dedicating the external network will enable Route Advertisement for this Edge Gateway. Default `false`.
* `non_distributed_routing_enabled` - (Optional, *v3.14+*) Non-Distributed routing will allow
  tenants the [option of connecting Org Vdc networks to the Service
  Router](providers/vmware/vcd/latest/docs/resources/network_routed_v2#interface_type). This would
  force all VM traffic through the service router for that network. Default `false`.

* `subnet` - (Optional) One or more [subnets](#edgegateway-subnet) defined for Edge Gateway. One of
  `subnet`, `subnet_with_total_ip_count` or `subnet_with_ip_count` is **required unless** parent
  network is backed by *IP Spaces*. Read more in [IP allocation modes](#ip-allocation-modes) section.
* `subnet_with_total_ip_count` - (Optional, *v3.9+*) One or more
  [subnets](#edgegateway-total-ip-count-allocation) defined for Edge Gateway. One of `subnet`,
  `subnet_with_total_ip_count` or `subnet_with_ip_count` is **required unless** parent network is
  backed by *IP Spaces*. Read more in [IP allocation modes](#ip-allocation-modes) section.
* `subnet_with_ip_count` - (*v3.9+*) One or more
  [subnets](#edgegateway-per-subnet-ip-count-allocation) defined for Edge Gateway. One of `subnet`,
  `subnet_with_total_ip_count` or `subnet_with_ip_count` is **required unless** parent network is
  backed by *IP Spaces*. Read more in [IP allocation modes](#ip-allocation-modes) section.
* `total_allocated_ip_count` - (Optional, *v3.9+*) Required with `subnet_with_total_ip_count`. It is
  **read-only** attribute with other other allocation models `subnet` and `subnet_with_ip_count`.
  **Note**. It sets or reports IP count *only for NSX-T Tier 0 backed external network Uplink*.
* `external_network` - (Optional, *VCD 10.4.1+*, *v3.11+*) attaches NSX-T Segment backed External
  Networks with a given [configuration block](#edgegateway-subnet-external-network). It *does not
  support IP Spaces*.
* `ip_count_read_limit` - (Optional, *v3.13+*) Sets a limit of IPs to count for
  `used_ip_count` and `unused_ip_count` attributes to avoid exhausting compute resource while
  counting IPs in large IPv6 subnets. It does not affect operation of Edge Gateway configuration,
  only IP count reporting. Defaults to `1000000`, update is a no-op, but will affect newly read
  data. While it is unlikely that a single Edge Gateway can effectively manage more IPs, one can
  specify `0` for *unlimited* value. 

<a id="ip-allocation-modes"></a>

## IP allocation modes

~> Starting with `v3.9.0` of Terraform Provider for VCD, NSX-T Edge Gateway supports automatic IP
allocations. More details in this section. **None** of these methods **should be applied** when Edge
Gateway is using **IP Spaces**

There are three ways to handle IP allocations, but it's important to note that _due to Terraform
schema limitations_, only **one of the three** methods can be utilized:

* [_Manual_ IP allocation using `subnet` and `allocated_ips` ranges](#edgegateway-subnet)
* [_Automatic_ IP allocation mode with _total IP count_ using `subnet_with_total_ip_count` and
  `total_allocated_ip_count`](#edgegateway-total-ip-count-allocation)
* [_Automatic_ IP allocation mode with _per subnet IP count_  using
  `subnet_with_ip_count`](#edgegateway-per-subnet-ip-count-allocation)
<a id="edgegateway-subnet"></a>

### Manual IP allocation using `subnet` and `allocated_ips` ranges

Each `subnet` block has these attributes ([complete example](#subnet-example)):

* `gateway` - (Required) - Gateway for a subnet in external network
* `prefix_length` - (Required) Prefix length of a subnet in external network (e.g. 24 for netmask
  of 255.255.255.0)
* `primary_ip` - (Optional) - Primary IP address for edge gateway. **Note:** `primary_ip` must fall
into `allocated_ips` block range as otherwise `plan` will not be clean with a new range defined for
that particular block. There __can only be one__ `primary_ip` defined for edge gateway.
* `allocated_ips` (Required) - One or more blocks of [ip ranges](#edgegateway-subnet-ip-allocation)
in the subnet to be allocated
<a id="edgegateway-subnet-ip-allocation"></a>

Each `subnet` must have on or more `allocated_ips` blocks that consist of IP range definitions:

* `start_address` - (Required) - Start IP address of a range
* `end_address` - (Required) - End IP address of a range
<a id="edgegateway-total-ip-count-allocation"></a>

### Automatic IP allocation mode with total IP count using `subnet_with_total_ip_count` and `total_allocated_ip_count`

This mode handles IP allocations automatically by leveraging `subnet_with_total_ip_count` blocks and
total IP count definition using `total_allocated_ip_count` field ([complete
example](#subnet-with-total-ip-count-example)).

Each `subnet_with_total_ip_count` has these attributes:

* `gateway` - (Required) - Gateway for a subnet in external network
* `prefix_length` - (Required) Prefix length of a subnet in external network (e.g. 24 for netmask of 255.255.255.0)
* `primary_ip` (Required) - Exactly one Primary IP is required for an Edge Gateway. It should be
  defined in any of the `subnet_with_total_ip_count` blocks

~> Only network definitions are required and IPs are allocated automatically, based on
`total_allocated_ip_count` parameter
<a id="edgegateway-per-subnet-ip-count-allocation"></a>

### Automatic IP allocation mode with per subnet IP count using `subnet_with_ip_count`

This mode allows defining multiple subnets through `the subnet_with_ip_count` block, with the
added capability of automatically allocating a specific number of IPs from the block based on the
`allocated_ip_count` field ([complete example](#subnet-with-ip-count-example)).

* `gateway` - (Required) - Gateway for a subnet in external network
* `prefix_length` - (Required) Prefix length of a subnet in external network (e.g. 24 for netmask of 255.255.255.0)
* `primary_ip` (Required) - Exactly one Primary IP is required for an Edge Gateway. It should be
  defined in any of the `subnet_with_ip_count` blocks
* `allocated_ip_count` (Required) - Number of allocated IPs from that particular subnet

<a id="edgegateway-subnet-external-network"></a>

### NSX-T Segment backed network attachment using `external_network` block

This option can attach NSX-T Segment backed external networks

* `external_network_id` - (Required) - ID of NSX-T Segment backed external network
* `gateway` - (Required) - Gateway for a subnet in external network
* `prefix_length` - (Required) Prefix length of a subnet in external network (e.g. 24)
* `primary_ip` (Optional) - Exactly one Primary IP is required for an Edge Gateway. It should fall
  in provided `gateway` and `prefix_length`
* `allocated_ip_count` (Required) - Number of allocated IPs

## Attribute Reference

The following attributes are exported on this resource:

* `primary_ip` - Primary IP address exposed for an easy access without nesting.
* `used_ip_count` - Used IP count in this Edge Gateway (for all uplinks). **Note**: it is not
  exposed when using IP Spaces.
* `unused_ip_count` - Unused IP count in this Edge Gateway (for all uplinks). **Note**: it is not
  exposed when using IP Spaces.
* `use_ip_spaces` - Boolean value that hints if the NSX-T Edge Gateway uses IP Spaces
* `external_network_allocated_ip_count` - Total allocated IP count in attached NSX-T Segment backed
  external networks

~> `primary_ip`, `used_ip_count` and `unused_ip_count` will not be populated when using **IP Spaces**

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. [More information.][docs-import]

An existing edge gateway can be [imported][docs-import] into this resource via supplying its path.
The path for this resource is made of `org-name.vdc-name.nsxt-edge-name` or
`org-name.vdc-group-name.nsxt-edge-name` For example, using this structure, representing an edge
gateway that was **not** created using Terraform:

```hcl
data "vcd_org_vdc" "vdc-1" {
  name = "vdc-name"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org         = "my-org"
  owner_id    = data.vcd_org_vdc.vdc-1.id
  name        = "nsxt-edge"
  description = "Description"

  external_network_id = data.vcd_external_network_v2.nsxt-ext-net.id

  subnet {
    gateway       = "10.10.10.1"
    prefix_length = "24"
    primary_ip    = "10.10.10.10"
    allocated_ips {
      start_address = "10.10.10.10"
      end_address   = "10.10.10.30"
    }
  }
}
```

You can import such resource into terraform state using the command below:

```
terraform import vcd_nsxt_edgegateway.nsxt-edge my-org.nsxt-vdc.nsxt-edge
```

* **Note 1**: the separator can be changed using `Provider.import_separator` or variable `VCD_IMPORT_SEPARATOR`
* **Note 2**: it is possible to list all available NSX-T edge gateways using data source [vcd_resource_list](/providers/vmware/vcd/latest/docs/data-sources/resource_list#vcd_nsxt_edgegateway)

[docs-import]:https://www.terraform.io/docs/import/

After importing, if you run `terraform plan` you will see the rest of the values and modify the script accordingly for
further operations.
