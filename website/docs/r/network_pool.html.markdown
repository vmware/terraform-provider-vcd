---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_network_pool"
sidebar_current: "docs-vcd-resource-network-pool"
description: |-
  Provides a VMware Cloud Director Network Pool. This can be used to create, modify, and delete a VCD Network Pool
---

# vcd\_network\_pool

Provides a VMware Cloud Director VCD Network Pool. This can be used to create,
modify, and delete networks pools attached to a VCD.

Supported in provider *v3.11+*

~> Only `System Administrator` can create this resource.

## Example Usage 1 - Type "GENEVE"

```hcl
data "vcd_nsxt_manager" "mgr" {
  name = "mymanager"
}

resource "vcd_network_pool" "npool" {
  name                = "new-network-pool"
  description         = "New network pool"
  network_provider_id = data.vcd_nsxt_manager.mgr.id
  type                = "GENEVE"

  backing {
    transport_zone {
      name = "nsx-overlay-transportzone"
    }
  }
}
```

## Example Usage 2 - Type "VLAN"

```hcl
data "vcd_vcenter" "vc1" {
  name = "vc1"
}

resource "vcd_network_pool" "npool" {
  name                = "my-vlan-network-pool"
  description         = "New VLAN network pool"
  network_provider_id = data.vcd_vcenter.vc1.id
  type                = "VLAN"

  backing {
    distributed_switches {
      name = "NsxTDVS"
    }
    range_ids {
      start_id = 101
      end_id   = 200
    }
  }
}
```
## Example Usage 3 - Type "PORTGROUP_BACKED"

```hcl
data "vcd_vcenter" "vc1" {
  name = "vc1"
}

resource "vcd_network_pool" "npool" {
  name                = "my-pg-network-pool"
  description         = "New Port Group network pool"
  network_provider_id = data.vcd_vcenter.vc1.id
  type                = "PORTGROUP_BACKED"

  backing {
    port_groups {
      name = "TestbedPG"
    }
  }
}
```

## Example Usage 4 Retrieving backing elements

The elements needed as backing for a network pool can be retrieved using [`vcd_resource_list`](/providers/vmware/vcd/latest/docs/data_sources/resource_list), as in the example below

```hcl
data "vcd_nsxt_manager" "mgr" {
  name = "nsxManager1"
}

data "vcd_vcenter" "vc1" {
  name = "vc1"
}

data "vcd_resource_list" "tz" {
  name          = "tz"
  resource_type = "vcd_nsxt_transport_zone"
  parent        = data.vcd_nsxt_manager.mgr.name
}

data "vcd_resource_list" "pg" {
  name          = "pg"
  resource_type = "vcd_importable_port_group"
  parent        = data.vcd_vcenter.vc1.name
}

data "vcd_resource_list" "ds" {
  name          = "ds"
  resource_type = "vcd_distributed_switch"
  parent        = data.vcd_vcenter.vc1.name
}

output "tzs" {
  value = data.vcd_resource_list.tz.list
}

output "pgs" {
  value = data.vcd_resource_list.pg.list
}

output "ds" {
  value = data.vcd_resource_list.ds.list
}
```

-> Note: the lists provided as `vcd_resource_list` output are volatile: they only exist for items that have not been used
in a network pool. Once they have been assigned, they cease to be shown. As such, it is not a good idea to use
`vcd_resource_list` as direct source for one or more network pools: at the first `plan`, terraform would propose
to remove the network pool, as the element is not shown in the list anymore.

## Example Usage 5 - nameless backing retrieval

If we don't have preference about which of the elements we will use as backing for the network pool, we could let
the system pick the first available. This could be a good idea when we know that there is only one element available, or 
we know that all elements have similar capabilities.
If we are in these circumstances, we could avoid some details and skip the definition of the backing elements.

```hcl
data "vcd_nsxt_manager" "mgr" {
  name = "nsxManager1"
}

resource "vcd_network_pool" "npool" {
  name                = "new-network-pool"
  description         = "network pool without explicit transport zone"
  network_provider_id = data.vcd_nsxt_manager.mgr.id
  type                = "GENEVE"
  backing {
  }
}
```

The system will pick the first available transport zone, or fail if none was available. The name of the used transport
zone will be shown if we use an `output` for the network pool.

```hcl
output "pool" {
  value = vcd_network_pool.npool.backing
}
```

## Argument Reference

* `name` - (Required) Unique name of network pool
* `type` - (Required) Type of the network pool (one of `GENEVE`, `VLAN`, `PORTGROUP_BACKED`)
* `network_provider_id` - (Required) Id of the network provider (either vCenter or NSX-T manager)
* `description` - (Optional) Description of the network pool
* `backing` - (Optional) The components used by the network pool. See [Backing](#backing) below for details

## Attribute Reference

* `status` Status of the network pool
* `promiscuous_mode` Whether the network pool is in promiscuous mode
* `total_backings_count` Total number of backings
* `used_backings_count` Number of used backings
* `network_provider_name` Name of the network provider
* `network_provider_type` Type of network provider

### Backing
* `transport_zone` - (Optional) [backing structure](#backing-element) used for `GENEVE` network pool
* `distributed_switches` - (Optional) [backing structure](#backing-element) used for `VLAN` network pool
* `port_groups` - (Optional) [backing structure](#backing-element) used for `PORTGROUP_BACKED` network pool
* `range_ids` - (Optional) A list of range IDs, required with `VLAN` network pools
    * `start_id` - (Required) The first ID of the range
    * `end_id` - (Required) The last ID of the range


### Backing element
* `name` - (Optional) The name of the backing element (transport zone, distributed switch, importable port group). If omitted,
  the system will try to pick the first available.
* `type` - (Computed) The type of the backing element
* `id` - (Computed) The ID of the backing element

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. However, an experimental feature in Terraform 1.5+ allows also code generation.
See [Importing resources][importing-resources] for more information.

An existing network pool can be [imported][docs-import] into a resource via supplying its path for a
network pool. For example, using this structure, representing an existing network pool that was **not** created using Terraform:

```hcl
resource "vcd_network_pool" "net_pool" {
  name = "my-net-pool"
}
```

We can import such network pool into terraform state using this command

```bash
terraform import vcd_network_pool.net_pool my-net-pool
```

After that, we can expand the configuration file and either update or delete the network pool as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the network pool's stored properties.

[docs-import]:https://www.terraform.io/docs/import/
[importing-resources]:https://registry.terraform.io/providers/vmware/vcd/3.10.0/docs/guides/importing_resources