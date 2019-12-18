--
layout: "vcd"
page_title: "vCloudDirector: vcd_info"
sidebar_current: "docs-vcd-datasource-info"
description: |-
  Provides information about vCD resources
---

# vcd\_info

Provides a vCloud Director generic data source. It provides a list of existing resources in various formats.

Supported in provider *v2.7+*

## Example Usage 1

```hcl
data "vcd_info" "list_of_orgs" {
  name          = "list_of_orgs"
  resource_type = "org"
  list_mode     = "name"
}

// Shows the list of organizations
output "org_list" {
  value = data.vcd_info.list_of_orgs.list
}
/* 
output:
  "org_list" = [
    "System",
    "my-org",
  ]
*/
```

## Example Usage 2

```hcl
data "vcd_info" "list_of_orgs" {
  name          = "list_of_orgs"
  resource_type = "org"
  list_mode     = "name_id"
}

// Shows the list of organizations with the corresponding ID
output "org_list" {
  value = data.vcd_info.list_of_orgs.list
}
/* 
output:
  "org_list" = [
    "System  urn:vcloud:org:a93c9db9-7470-3191-8d19-a8f7eeda87f9",
    "my-org  urn:vcloud:org:92554cc7-6222-4102-af48-364c95ed1a35",
  ]
*/
```

## Example Usage 3
```hcl
data "vcd_info" "list_of_nets" {
  name          = "list_of_nets"
  resource_type = "network" // Finds all networks, regardless of their type
  list_mode     = "import"
}

// Shows the list of all networks with the corresponding import command
output "net_list" {
  value = data.vcd_info.list_of_nets.list
}

/*
output: 
list_networks_import = [
  "terraform import vcd_network_routed.net-datacloud-r my-org.my-vdcdatacloud.net-r",
  "terraform import vcd_network_isolated.net-datacloud-i my-org.my-vdcdatacloud.net-i",
  "terraform import vcd_network_routed.net-datacloud-r2 my-org.my-vdcdatacloud.net-r2",
  "terraform import vcd_network_direct.net-datacloud-d my-org.my-vdcdatacloud.net-d",
*/
```

## Example Usage 4
```hcl
data "vcd_info" "list_network_hierarchy" {
  name          = "list_of_nets"
  resource_type = "network" // Finds all networks, regardless of their type
  list_mode     = "hierarchy"
}

// Shows the list of all networks with their parent entities
output "net_network_hierarchy" {
  value = data.vcd_info.list_network_hierarchy.list
}

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

## Example Usage 5
```hcl
data "vcd_info" "list_of_nets" {
  name          = "list_of_nets"
  resource_type = "network_routed"
  list_mode     = "name"
}

// Shows the list of routed networks
output "net_list" {
  value = data.vcd_info.list_of_nets.list
}

// Uses the list of networks to get the data source of each
data "vcd_network_routed" "full_networks" {
  for_each = toset(data.vcd_info.net_list.list)
  name     = each.value
  org      = "my-org"
  vdc      = "my-vdc"
}
/* 
full_networks = {
  "net-datacloud-r  urn:vcloud:network:04915abf-0c91-4919-878e-0f292e032e2b" = {
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
  "net-datacloud-r2  urn:vcloud:network:2cc713b1-134f-4f21-9208-79f1e4f3ee36" = {
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
}
*/
```

## Example Usage 6

```hcl
data "vcd_info" "list_of_resources" {
  name          = "list_of_resources"
  resource_type = "resources"
}

// Shows the list of resource types for VCD provider
output "resource_list" {
  value = data.vcd_info.list_of_resources.list
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level
* `name` - (Required) An unique name to identify the data source
* `resource_type` (Required) Which resource we want to list
* `list_mode` (Optional) How the list should be built. One of:
    * `name` (default): Only the resource name
    * `id`: Only the resource ID
    * `href`: Only the resource HREF
    * `name_id`: Both the resource name and ID separated by `name_id_separator`
    * `hierarchy`: All the ancestor names (if any) followed by the resource name, separated by `name_id_separator`
    * `import`: A terraform client command to import the resource
* `name_id_separator` (Optional) A string separating name and ID in the list. Default is "  " (two spaces)
* `parent` (Optional) The resource parent, such as "vapp" or "catalog", when needed. If not provided, all available
resources will be listed. E.g.: for a "vm", if no vApp name is provided as parent, all VMs are listed.

## Attribute Reference

* `list` - (Computed) The list of requested resources in the chosen format.
