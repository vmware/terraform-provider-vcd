---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_resource_list"
sidebar_current: "docs-vcd-datasource-resource-list"
description: |-
  Provides lists of VCD resources
---

# vcd\_resource_list

Provides a VMware Cloud Director generic data source. It provides a list of existing resources in various formats.
Some of these lists are only informative (i.e. users will look at them to get general information about the existing
resources) and some are usable directly from Terraform, where the list can be re-used to create data sources, and the
data sources in turn can help generate more resources. 
The data created by this data source can also be reused by third party tools to create more complex tasks.

Supported in provider *v3.1+*

## Example 1 - List of organizations - name

```hcl
data "vcd_resource_list" "list_of_orgs" {
  name          = "list_of_orgs"
  resource_type = "vcd_org"
  list_mode     = "name"
}

# Shows the list of organizations
output "org_list" {
  value = data.vcd_resource_list.list_of_orgs.list
}
```
```
/* 
output:
  "org_list" = [
    "System",
    "my-org",
  ]
*/
```

## Example 2 - List of organizations - name and ID

```hcl
data "vcd_resource_list" "list_of_orgs" {
  name          = "list_of_orgs"
  resource_type = "org"
  list_mode     = "name_id"
}

# Shows the list of organizations with the corresponding ID
output "org_list" {
  value = data.vcd_resource_list.list_of_orgs.list
}
```
```
/* 
output:
  "org_list" = [
    "System  urn:vcloud:org:a93c9db9-7470-3191-8d19-a8f7eeda87f9",
    "my-org  urn:vcloud:org:92554cc7-6222-4102-af48-364c95ed1a35",
  ]
*/
```

## Example 3 - List of networks - output import

```hcl
data "vcd_resource_list" "list_of_nets" {
  name          = "list_of_nets"
  resource_type = "network" # Finds all networks, regardless of their type
  list_mode     = "import"
}

# Shows the list of all networks with the corresponding import command
output "net_list" {
  value = data.vcd_resource_list.list_of_nets.list
}

```
```
/*
output: 
list_networks_import = [
  "terraform import vcd_network_routed.net-datacloud-r my-org.my-vdcdatacloud.net-r",
  "terraform import vcd_network_isolated.net-datacloud-i my-org.my-vdcdatacloud.net-i",
  "terraform import vcd_network_routed.net-datacloud-r2 my-org.my-vdcdatacloud.net-r2",
  "terraform import vcd_network_direct.net-datacloud-d my-org.my-vdcdatacloud.net-d",
*/
```

## Example 4 - List of networks - hierarchy

```hcl
data "vcd_resource_list" "list_network_hierarchy" {
  name          = "list_of_nets"
  resource_type = "network" # Finds all networks, regardless of their type
  list_mode     = "hierarchy"
}

# Shows the list of all networks with their parent entities
output "net_network_hierarchy" {
  value = data.vcd_resource_list.list_network_hierarchy.list
}

```
```
/*
output: 
list_networks_hierarchy = [
  "datacloud  vdc-datacloud  net-datacloud-r",
  "datacloud  vdc-datacloud  net-datacloud-i",
  "datacloud  vdc-datacloud  net-datacloud-r2",
  "datacloud  vdc-datacloud  net-datacloud-d",
]
*/
```

## Example 5 - Data sources from list of networks - cascade to new networks

```hcl
data "vcd_resource_list" "list_of_nets" {
  name          = "list_of_nets"
  resource_type = "network_routed"
  list_mode     = "name"
}

# Shows the list of routed networks
output "net_list" {
  value = data.vcd_resource_list.list_of_nets.list
}

# Uses the list of networks to get the data source of each
data "vcd_network_routed" "full_networks" {
  count = length(data.vcd_resource_list.list_of_nets.list)
  name  = data.vcd_resource_list.list_of_nets.list[count.index]
  org   = "datacloud"
  vdc   = "vdc-datacloud"
}

output "net" {
  value = data.vcd_network_routed.full_networks
}

# creates a new resource for each data source
resource "vcd_network_routed" "new_net" {
  count        = length(data.vcd_network_routed.full_networks)
  name         = "${data.vcd_network_routed.full_networks[count.index].name}-2"
  edge_gateway = data.vcd_network_routed.full_networks[count.index].edge_gateway
  gateway      = "192.168.${count.index + 10}.1"

  static_ip_pool {
    start_address = "10.10.${count.index + 10}.152"
    end_address   = "10.10.${count.index + 10}.254"
  }
}
```
```
/* 
full_networks = [
  {
    "description" = "net-datacloud-r"
    "dhcp_pool" = []
    "dns1" = "8.8.8.8"
    "dns2" = "8.8.4.4"
    "dns_suffix" = ""
    "edge_gateway" = "gw-datacloud"
    "gateway" = "192.168.2.1"
    "href" = "https://vcd.example.com/api/network/04915abf-0c91-4919-878e-0f292e032e2b"
    "id" = "urn:vcloud:network:04915abf-0c91-4919-878e-0f292e032e2b"
    "name" = "net-datacloud-r"
    "netmask" = "255.255.255.0"
    "org" = "datacloud"
    "shared" = false
    "static_ip_pool" = [
      {
        "end_address" = "192.168.2.100"
        "start_address" = "192.168.2.2"
      },
    ]
    "vdc" = "vdc-datacloud"
  }
  {
    "description" = ""
    "dhcp_pool" = []
    "dns1" = ""
    "dns2" = ""
    "dns_suffix" = ""
    "edge_gateway" = "gw-datacloud"
    "gateway" = "192.168.3.1"
    "href" = "https://vcd.example.com/api/network/2cc713b1-134f-4f21-9208-79f1e4f3ee36"
    "id" = "urn:vcloud:network:2cc713b1-134f-4f21-9208-79f1e4f3ee36"
    "name" = "net-datacloud-r2"
    "netmask" = "255.255.255.0"
    "org" = "datacloud"
    "shared" = false
    "static_ip_pool" = [
      {
        "end_address" = "192.168.3.50"
        "start_address" = "192.168.3.2"
      },
    ]
    "vdc" = "vdc-datacloud"
  }
]
*/
```

## Example 6 - List of resources

```hcl
data "vcd_resource_list" "list_of_resources" {
  name          = "list_of_resources"
  resource_type = "resources"
}

# Shows the list of resource types for VCD provider
output "resource_list" {
  value = data.vcd_resource_list.list_of_resources.list
}
```


## Example 7 - List of catalog items

```hcl
data "vcd_resource_list" "list_catalog_items" {
  name          = "list_of_catalog_items"
  resource_type = "vcd_catalog_item"
  parent        = "cat-datacloud" # name of the catalog to be listed
  list_mode     = "name"
}

# Shows the list of all catalog items
output "catalog_items" {
  value = data.vcd_resource_list.list_of_catalog_items.list
}

```
```
/*
output: 
catalog_items = [
  "photon-hw11",
  "Linux1",
  "debian-7",
]
*/
```

## Example 8 - List of LB virtual servers

```hcl
data "vcd_resource_list" "list_lb_virtual_servers" {
  name          = "list_of_virtual_servers"
  resource_type = "vcd_lb_virtual_server"
  parent        = "gw-datacloud" # name of the edge gateway to be listed
  list_mode     = "name"
}

# Shows the list of all LB virtual servers
output "lb_virtual_servers" {
  value = data.vcd_resource_list.list_lb_virtual_servers.list
}
```

## Example 9 - List of networks - output import with generated file

This example is similar to example n. 3, with the addition of a generated import file

```hcl
data "vcd_resource_list" "list_of_nets" {
  name             = "list_of_nets"
  resource_type    = "network" # Finds all networks, regardless of their type
  list_mode        = "import"
  import_file_name = "network-import.tf"
}

# Shows the list of all networks with the corresponding import command
output "net_list" {
  value = data.vcd_resource_list.list_of_nets.list
}

```
The output is the same as example n. 3, but we also have the generated file `network-import.tf`, containing the
import blocks needed to import resources.

```
$ cat network-import.tf
# Generated by vcd_resource_list - 2023-08-06T21:16:32+02:00
# import data for vcd_network_isolated datacloud.vdc-datacloud.net-datacloud-i
import {
  to = vcd_network_isolated.net-datacloud-i-b267e60747e7
  id = "datacloud.vdc-datacloud.net-datacloud-i"
}

# import data for vcd_network_direct datacloud.vdc-datacloud.net-datacloud-d
import {
  to = vcd_network_direct.net-datacloud-d-8b747c65f924
  id = "datacloud.vdc-datacloud.net-datacloud-d"
}

# import data for vcd_network_routed datacloud.vdc-datacloud.net-datacloud-r
import {
  to = vcd_network_routed.net-datacloud-r-ae303c5e35c1
  id = "datacloud.vdc-datacloud.net-datacloud-r"
}
```

# Example 10 - List of roles with filter

```hcl
data "vcd_resource_list" "role-filter1" {
  vdc           = "nsxt-vdc-datacloud"
  name          = "role-filter1"
  resource_type = "vcd_role"
  # name_regex  = "Author|vApp"
}

output "vapp_roles" {
  value = data.vcd_resource_list.role-filter1.list
}
```

Without the filter, the listing would be all the roles available to the current organization.

```
vapp_roles = tolist([
  "Organization Administrator",
  "Catalog Author",
  "vApp Author",
  "vApp User",
  "Console Access Only",
  "Defer to Identity Provider",
])
```

If we remove the comment on `name_regex`, the listing will be filtered by the regular expression, matching only roles
with a name including either `Author` or `vApp`.

```
vapp_roles = tolist([
  "Catalog Author",
  "vApp Author",
  "vApp User",
])
```


See [Importing resources][import-resources] for more information on how to leverage `vcd_resource_list` functionality
to import resources.

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level
* `name` - (Required) An unique name to identify the data source
* `resource_type` (Required) Which resource we want to list. Supported keywords are:
    * `resources`  (list the resource types in the provider)
    * `vcd_org`
    * `vcd_external_network`
    * `vcd_org_vdc`
    * `vcd_catalog`
    * `vcd_catalog_item`
    * `vcd_catalog_vapp_template`
    * `vcd_catalog_media`
    * `vcd_vapp`
    * `vcd_vapp_vm` (only VMs within a vApp)
    * `vcd_all_vm`  (both standalone VMs and VMs within a vApp)
    * `vcd_vm`      (only standalone VMs)
    * `vcd_org_user`
    * `vcd_edgegateway`
    * `vcd_independent_disk`
    * `vcd_nsxt_edgegateway`
    * `vcd_lb_server_pool`
    * `vcd_lb_service_monitor`
    * `vcd_lb_virtual_server`
    * `vcd_lb_app_rule`
    * `vcd_lb_app_profile`
    * `vcd_nsxv_firewall_rule`
    * `vcd_ipset`
    * `vcd_nsxv_dnat`
    * `vcd_nsxv_snat`
    * `vcd_network_isolated`
    * `vcd_network_direct`
    * `vcd_network_routed`
    * `vcd_network_routed_v2`
    * `vcd_network_isolated_v2`
    * `vcd_nsxt_network_imported`
    * `vcd_library_certificate`
    * `vcd_provider_vdc`
    * `vcd_network_pool`
    * `vcd_vcenter`
    * `vcd_nsxt_manager`
    * `vcd_nsxt_transport_zone`
    * `vcd_distributed_switch`
    * `vcd_importable_port_group`
    * `vcd_multisite_site_association`
    * `vcd_multisite_org_association`
    * `vcd_vapp_network`
    * `vcd_vapp_org_network`
    * `vcd_vapp_all_network`
* `list_mode` (Optional) How the list should be built. One of:
    * `name` (default): Only the resource name
    * `id`: Only the resource ID
    * `href`: Only the resource HREF
    * `name_id`: Both the resource name and ID separated by `name_id_separator`
    * `hierarchy`: All the ancestor names (if any) followed by the resource name, separated by `name_id_separator`
    * `import`: A terraform client command to import the resource
* `name_id_separator` (Optional) A string separating name and ID in the list. Default is "  " (two spaces)
* `parent` (Optional) The resource parent, such as vApp, catalog, or edge gateway name, when needed. 
* `name_regex` (Optional; *v3.11+*) If set, will restrict the list of resources to the ones whose name matches the given regular expression.
* `import_file_name` (Optional; *v3.11+*; EXPERIMENTAL) Name of the file containing the import block. (Requires `list_mode = "import"`).
  See [Importing resources][import-resources] for more information on importing.

## Attribute Reference

* `list` - (Computed) The list of requested resources in the chosen format.

[Import-resources]:https://registry.terraform.io/providers/vmware/vcd/latest/docs/guides/importing_resources
