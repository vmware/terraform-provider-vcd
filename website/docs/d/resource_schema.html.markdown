---
layout: "vcd"
page_title: "vCloudDirector: vcd_resource_schema"
sidebar_current: "docs-vcd-datasource-resource-schema"
description: |-
  Provides information about a vCD resource structure
---

# vcd\_resource_schema

Provides a vCloud Director generic structure data source. It shows the structure of any VCD resource.

Supported in provider *v3.1+*

## Example Usage 1

Showing a structure with simple attributes only 

```hcl
data "vcd_resource_schema" "org_struct" {
  name          = "org_struct"
  resource_type = "vcd_org"
}

// Shows the organization attributes
output "org_struct" {
  value = data.vcd_resource_schema.org_struct.attributes
}
/* 
output:
org_struct = [
  {
    "computed" = false
    "description" = "When destroying use delete_force=True with delete_recursive=True to remove an org and any objects it contains, regardless of their state."
    "name" = "delete_force"
    "optional" = false
    "required" = true
    "sensitive" = false
    "type" = "bool"
  },
  {
    "computed" = false
    "description" = ""
    "name" = "full_name"
    "optional" = false
    "required" = true
    "sensitive" = false
    "type" = "string"
  },
  {
    "computed" = false
    "description" = ""
    "name" = "name"
    "optional" = false
    "required" = true
    "sensitive" = false
    "type" = "string"
  },
  {
    "computed" = false
    "description" = "Specifies this organization's default for virtual machine boot delay after power on."
    "name" = "delay_after_power_on_seconds"
    "optional" = true
    "required" = false
    "sensitive" = false
    "type" = "number"
  },
  {
    "computed" = false
    "description" = "When destroying use delete_recursive=True to remove the org and any objects it contains that are in a state that normally allows removal."
    "name" = "delete_recursive"
    "optional" = false
    "required" = true
    "sensitive" = false
    "type" = "bool"
  },
  {
    "computed" = false
    "description" = "True if this organization is enabled (allows login and all other operations)."
    "name" = "is_enabled"
    "optional" = true
    "required" = false
    "sensitive" = false
    "type" = "bool"
  },
  {
    "computed" = false
    "description" = "Maximum number of virtual machines that can be deployed simultaneously by a member of this organization. (0 = unlimited)"
    "name" = "deployed_vm_quota"
    "optional" = true
    "required" = false
    "sensitive" = false
    "type" = "number"
  },
  {
    "computed" = false
    "description" = "True if this organization is allowed to share catalogs."
    "name" = "can_publish_catalogs"
    "optional" = true
    "required" = false
    "sensitive" = false
    "type" = "bool"
  },
  {
    "computed" = true
    "description" = ""
    "name" = "id"
    "optional" = true
    "required" = false
    "sensitive" = false
    "type" = "string"
  },
  {
    "computed" = false
    "description" = ""
    "name" = "description"
    "optional" = true
    "required" = false
    "sensitive" = false
    "type" = "string"
  },
  {
    "computed" = false
    "description" = "Maximum number of virtual machines in vApps or vApp templates that can be stored in an undeployed state by a member of this organization. (0 = unlimited)"
    "name" = "stored_vm_quota"
    "optional" = true
    "required" = false
    "sensitive" = false
    "type" = "number"
  },
]
*/
```

## Example Usage 2

Showing a structure with both simple and compound attributes

```hcl
data "vcd_resource_schema" "network_isolated_struct" {
  name          = "net_struct"
  resource_type = "vcd_network_isolated"
}

output "net_struct" {
  value = data.vcd_resource_schema.net_struct
}
/* 
output:
struct_network_isolated = {
  "attributes" = [
    {
      "computed" = true
      "description" = "Network Hyper Reference"
      "name" = "href"
      "optional" = false
      "required" = false
      "sensitive" = false
      "type" = "string"
    },
    {
      "computed" = true
      "description" = ""
      "name" = "id"
      "optional" = true
      "required" = false
      "sensitive" = false
      "type" = "string"
    },
    {
      "computed" = false
      "description" = "First DNS server to use"
      "name" = "dns1"
      "optional" = true
      "required" = false
      "sensitive" = false
      "type" = "string"
    },
    {
      "computed" = false
      "description" = "The name of VDC to use, optional if defined at provider level"
      "name" = "vdc"
      "optional" = true
      "required" = false
      "sensitive" = false
      "type" = "string"
    },
    {
      "computed" = false
      "description" = "Optional description for the network"
      "name" = "description"
      "optional" = true
      "required" = false
      "sensitive" = false
      "type" = "string"
    },
    {
      "computed" = false
      "description" = "The gateway for this network"
      "name" = "gateway"
      "optional" = true
      "required" = false
      "sensitive" = false
      "type" = "string"
    },
    {
      "computed" = false
      "description" = "The netmask for the new network"
      "name" = "netmask"
      "optional" = true
      "required" = false
      "sensitive" = false
      "type" = "string"
    },
    {
      "computed" = false
      "description" = "A unique name for this network"
      "name" = "name"
      "optional" = false
      "required" = true
      "sensitive" = false
      "type" = "string"
    },
    {
      "computed" = false
      "description" = "The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organizations"
      "name" = "org"
      "optional" = true
      "required" = false
      "sensitive" = false
      "type" = "string"
    },
    {
      "computed" = false
      "description" = "Second DNS server to use"
      "name" = "dns2"
      "optional" = true
      "required" = false
      "sensitive" = false
      "type" = "string"
    },
    {
      "computed" = false
      "description" = "A FQDN for the virtual machines on this network"
      "name" = "dns_suffix"
      "optional" = true
      "required" = false
      "sensitive" = false
      "type" = "string"
    },
    {
      "computed" = false
      "description" = "Defines if this network is shared between multiple VDCs in the Org"
      "name" = "shared"
      "optional" = true
      "required" = false
      "sensitive" = false
      "type" = "bool"
    },
  ]
  "block_attributes" = [
    {
      "attributes" = [
        {
          "computed" = false
          "description" = "The first address in the IP Range"
          "name" = "start_address"
          "optional" = false
          "required" = true
          "sensitive" = false
          "type" = "string"
        },
        {
          "computed" = false
          "description" = "The final address in the IP Range"
          "name" = "end_address"
          "optional" = false
          "required" = true
          "sensitive" = false
          "type" = "string"
        },
      ]
      "name" = "static_ip_pool"
      "nesting_mode" = "NestingSet"
    },
    {
      "attributes" = [
        {
          "computed" = false
          "description" = "The first address in the IP Range"
          "name" = "start_address"
          "optional" = false
          "required" = true
          "sensitive" = false
          "type" = "string"
        },
        {
          "computed" = false
          "description" = "The final address in the IP Range"
          "name" = "end_address"
          "optional" = false
          "required" = true
          "sensitive" = false
          "type" = "string"
        },
        {
          "computed" = false
          "description" = "The default DHCP lease time to use"
          "name" = "default_lease_time"
          "optional" = true
          "required" = false
          "sensitive" = false
          "type" = "number"
        },
        {
          "computed" = false
          "description" = "The maximum DHCP lease time to use"
          "name" = "max_lease_time"
          "optional" = true
          "required" = false
          "sensitive" = false
          "type" = "number"
        },
      ]
      "name" = "dhcp_pool"
      "nesting_mode" = "NestingSet"
    },
  ] 
  "id" = "struct_network_isolated"
  "name" = "struct_network_isolated"
  "resource_type" = "vcd_network_isolated"
}
*/
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) An unique name to identify the data source
* `resource_type` (Required) Which resource we want to list. It needs to use the full name of the resource (i.e. "vcd_org",
not simply "org")

## Attribute Reference

* `attributes` - (Computed) The list of attributes for the resource.
  Each attribute is made of:
  * `name` - Name of the attribute
  * `description` - an optional description of the attribute
  * `required` - whether the attribute is required
  * `optional` - whether the attribute is optional
  * `computed` - whether the attribute is computed
  * `sensitive` - whether the attribute is sensitive
  
* `block_attributes` - (Computed) The list of compound attributes
    Each bock attribute is made of:
    * `name` the attribute name
    * `nesting_type` - (Computed) How the block is organized (one of `NestingSet`, `NestingList`)
    * `attributes` - (Computed) Same composition of the simple `attributes` above.

