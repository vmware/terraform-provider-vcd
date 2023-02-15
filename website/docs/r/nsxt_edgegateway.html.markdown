---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_edgegateway"
sidebar_current: "docs-vcd-resource-nsxt-edge-gateway"
description: |-
  Provides a VMware Cloud Director NSX-T edge gateway. This can be used to create, update, and delete NSX-T edge gateways connected to external networks.
---

# vcd\_nsxt\_edgegateway

Provides a VMware Cloud Director NSX-T edge gateway. This can be used to create, update, and delete NSX-T edge gateways connected
to external networks.

~> **Note:** Only `System Administrator` can create an edge gateway.
You must use `System Adminstrator` account in `provider` configuration
and then provide `org` and `vdc` arguments for edge gateway to work.

-> **Note:** This resource uses new VMware Cloud Director
[OpenAPI](https://code.vmware.com/docs/11982/getting-started-with-vmware-cloud-director-openapi) and
requires at least VCD *10.1.1+* and NSX-T *3.0+*.

Supported in provider *v3.1+*.

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
## Example Usage (Automatic IP allocation from any subnet)

```hcl
resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org      = "my-org"
  owner_id = data.vcd_org_vdc.vdc1.id
  name     = "nsxt-edge"

  external_network_id = data.vcd_external_network_v2.ext-net-nsxt.id

  # 100 IPs will be allocated from any of `auto_subnet` defined blocks
  total_allocated_ip_count = 100

  auto_subnet {
    gateway       = "77.77.77.1"
    prefix_length = "24"
    primary_ip    = "77.77.77.254"
  }

  auto_subnet {
    gateway       = "88.77.77.1"
    prefix_length = "24"
  }
}
```

## Example Usage (Automatic IP allocation per subnet)

```hcl
resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org      = "my-org"
  owner_id = data.vcd_org_vdc.vdc1.id
  name     = "nsxt-edge"

  external_network_id = data.vcd_external_network_v2.ext-net-nsxt.id

  auto_allocated_subnet {
    gateway            = "77.77.77.1"
    prefix_length      = "24"
    primary_ip         = "77.77.77.10"
    allocated_ip_count = 9
  }

  auto_allocated_subnet {
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
  name

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
* `dedicate_external_network` - (Optional) Dedicating the External Network will enable Route Advertisement for this Edge Gateway. Default `false`.

* `subnet` - (Optional) One or more [subnets](#edgegateway-subnet) defined for edge gateway. One of
  `subnet`, `auto_subnet` or `auto_allocated_subnet` is **required**.
* `auto_subnet` - (Optional, *v3.9+*) One or more [subnets](#edgegateway-auto-subnet) defined for
  Edge Gateway. One of `subnet`, `auto_subnet` or `auto_allocated_subnet` is **required**.
* `auto_allocated_subnet` - (*v3.9+*) One or more [subnets](#edgegateway-auto-allocated-subnet)
  defined for Edge Gateway. One of `subnet`, `auto_subnet` or `auto_allocated_subnet` is
  **required**.
* `total_allocated_ip_count` - (Optional, *v3.9+*) Required with `auto_subnet`

~> Starting with v3.9 of this provider, NSX-T Edge Gateways can allocate IP addresses by using
different strategies: Manual IP allocation (`subnet`), automatic IP allocations in any of defined
subnets (`auto_subnet` with `total_allocated_ip_count`), automatic IP allocations per defined subnet
(`auto_allocated_subnet`). One of these is **required**. Different set definition structures are
required due to Terraform schema limitations. **Note**. Allocation modes are split due to Terraform 
schema limitations and migrations between configurations can only be done __manually__.


<a id="edgegateway-subnet"></a>

## Edge Gateway Subnet (manual IP allocation)

* `gateway` - (Required) - Gateway for a subnet in external network
* `prefix_length` - (Required) - Prefix length of a subnet in external network (e.g. 24 for netmask of 255.255.255.0)
* `primary_ip` - (Optional) - Primary IP address for edge gateway. **Note:** `primary_ip` must fall into `allocated_ips`
block range as otherwise `plan` will not be clean with a new range defined for that particular block. There __can only
be one__ `primary_ip` defined for edge gateway.
* `allocated_ips` (Required) - One or more blocks of [ip ranges](#edgegateway-subnet-ip-allocation) in the subnet to be
allocated

<a id="edgegateway-subnet-ip-allocation"></a>

## Edge Gateway Subnet IP Allocation

* `start_address` - (Required) - Start IP address of a range
* `end_address` - (Required) - End IP address of a range

<a id="edgegateway-auto-subnet"></a>

## Edge Gateway Automatic IP allocation (from *any* of the defined `auto_subnet` entries)

* `gateway` - (Required) - Gateway for a subnet in external network
* `prefix_length` - (Required) - Prefix length of a subnet in external network (e.g. 24 for netmask of 255.255.255.0)
* `primary_ip` (Required) - Is required, but only in one of defined `auto_subnet` block

~> Only network definitions are required and IPs are allocated automatically, based on
`total_allocated_ip_count` parameter


<a id="edgegateway-auto-allocated-subnet"></a>

## Automatic IP allocation (per defined `auto_allocated_subnet` entries)

~> Subnet definitions (with one of them having `primary_ip` defined) and `allocated_ip_count` are
required. Automatic allocation will be used 

* `gateway` - (Required) - Gateway for a subnet in external network
* `prefix_length` - (Required) - Prefix length of a subnet in external network (e.g. 24 for netmask of 255.255.255.0)
* `primary_ip` (Required) - Is required, but only in one of defined `auto_allocated_subnet` block
* `allocated_ip_count` (Required) - Number of allocated IPs from that particular subnet

## Attribute Reference

The following attributes are exported on this resource:

* `primary_ip` - Primary IP address exposed for an easy access without nesting.
* `used_ip_count` - Unused IP count in this Edge Gateway
* `unused_ip_count` Used IP count in this Edge Gateway


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
