---
layout: "vcd"
page_title: "vCloudDirector: vcd_nsxt_edgegateway"
sidebar_current: "docs-vcd-resource-nsxt-edge-gatewa"
description: |-
  Provides a VMware Cloud Director NSX-T edge gateway. This can be used to create and delete NSX-T edge gateways connected to external networks.
---

# vcd\_nsxt\_edgegateway

Provides a VMware Cloud Director NSX-T edge gateway. This can be used to create and delete NSX-T edge gateways connected
to external networks.

Supported in provider *v3.1+*

~> **Note:** Only `System Administrator` can create an edge gateway.
You must use `System Adminstrator` account in `provider` configuration
and then provide `org` and `vdc` arguments for edge gateway to work.

## Example Usage (Simple case)

```hcl

data "vcd_external_network_v2" "nsxt-ext-net" {
  name = "nsxt-edge"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org                     = "dainius"
  vdc                     = "nsxt-vdc-dainius"
  name                    = "nsxt-edge"
  description             = "Description"

  external_network_id = data.vcd_external_network_v2.nsxt-ext-net.id

  subnet {
     gateway               = "10.150.191.253"
     prefix_length         = "19"
     primary_ip            = "10.150.160.137"
     allocated_ips {
       start_address = "10.150.160.137"
       end_address   = "10.150.160.137"
     }
  }
}
```

## Example Usage (Using custom Edge Cluster)

```hcl
data "vcd_nsxt_edge_cluster" "secondary" {
  name = "edge-cluster-two"
}


data "vcd_external_network_v2" "nsxt-ext-net" {
  name = "nsxt-edge"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org                     = "dainius"
  vdc                     = "nsxt-vdc-dainius"
  name                    = "nsxt-edge"
  description             = "Description"

  external_network_id = data.vcd_external_network_v2.nsxt-ext-net.id

  subnet {
     gateway               = "10.150.191.253"
     prefix_length         = "19"
     primary_ip            = "10.150.160.137"
     allocated_ips {
       start_address = "10.150.160.137"
       end_address   = "10.150.160.137"
     }
  }

  # Custom edge cluster reference
  edge_cluster_id = data.vcd_nsxt_edge_cluster.secondary.id
}
```


## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which the VDC belongs. Optional if defined at provider level.
* `vdc` - (Optional) The name of VDC that owns the edge gateway. Optional if defined at provider level.
* `name` - (Required) A unique name for the edge gateway.
* `description` - (Optional) A unique name for the edge gateway.
* `external_network_id` - (Required) A unique name for the edge gateway.
* `subnet` - (Required) One or more [subnets](#edgegateway-subnet) defined for edge gateway.
* `edge_cluster_id` - (Optional) Specific Edge Cluster ID if required

<a id="edgegateway-subnet"></a>
## Edge Gateway Subnet

* `gateway` (Required) - Gateway for a subnet in external network
* `netmask` (Required) - Netmask of a subnet in external network
* `ip_address` (Optional) - IP address to assign to edge gateway interface (will be auto-assigned if
  unspecified)
* `use_for_default_route` (Optional) - Should this network be used as default gateway on edge
  gateway. Default is `false`.
* `allocated_ips` (Required) - One or more blocks of [ip ranges](#edgegateway-subnet-ip-allocation) in the subnet to be
* sub-allocated

<a id="edgegateway-subnet-ip-allocation"></a>
## Edge Gateway Subnet IP Allocation

* `start_address` (Required) - Start IP address of a range
* `end_address` (Required) - End IP address of a range


## Attribute Reference

The following attributes are exported on this resource:

* `default_external_network_ip` (*v2.6+*) - IP address of edge gateway used for default network
* `external_network_ips` (*v2.6+*) - A list of IP addresses assigned to edge gateway interfaces
  connected to external networks.


## Importing

Supported in provider *v2.5+*

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing edge gateway can be [imported][docs-import] into this resource via supplying its path.
The path for this resource is made of org-name.vdc-name.edge-name
For example, using this structure, representing an edge gateway that was **not** created using Terraform:

```hcl
resource "vcd_edgegateway" "tf-edgegateway" {
  name              = "my-edge-gw"
  org               = "my-org"
  vdc               = "my-vdc"
  configuration     = "COMPUTE"

  external_network {
      name = "my-ext-net1"

      subnet {
        ip_address            = "192.168.30.51"
        gateway               = "192.168.30.49"
        netmask               = "255.255.255.240"
        use_for_default_route = true
      }
  }

}
```

You can import such resource into terraform state using one of the commands below

```
terraform import vcd_edgegateway.tf-egw my-org.my-vdc.my-edge-gw

terraform import vcd_edgegateway.tf-egw my-org.my-vdc.63ed92de-4001-450c-879f-deadbeef0123
```

* **Note 1**: the separator can be changed using `Provider.import_separator` or variable `VCD_IMPORT_SEPARATOR`
* **Note 2**: the identifier of the resource could be either the edge gateway name or the ID

[docs-import]:https://www.terraform.io/docs/import/

After importing, if you run `terraform plan` you will see the rest of the values and modify the script accordingly for
further operations.
