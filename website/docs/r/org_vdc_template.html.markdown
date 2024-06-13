---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_org_vdc_template"
sidebar_current: "docs-vcd-resource-org-vdc-template"
description: |-
  Provides a resource to create Organization VDC Templates in VMware Cloud Director. This can be used to create, delete, and update a Organization VDC Template.
---

# vcd\_org\_vdc\_template

Provides a resource to create Organization VDC Templates in VMware Cloud Director. This can be used to create, delete, and update a Organization VDC Template.
Requires system administrator privileges.

~> Only supports NSX-T network provider

Supported in provider *v3.13+*

## Example Usage

```hcl
data "vcd_org" "org" {
   name = "my_org"
}

data "vcd_provider_vdc" "pvdc1" {
   name = "nsxTPvdc1"
}

data "vcd_provider_vdc" "pvdc2" {
   name = "nsxTPvdc2"
}

data "vcd_external_network_v2" "ext_net" {
   name = "nsxt-extnet"
}

data "vcd_network_pool" "np1" {
   name = "NSX-T Overlay 1"
}

resource "vcd_org_vdc_template" "adam" {
   name               = "myTemplate"
   tenant_name        = "myAwesomeTemplate"
   description        = "Requires System privileges"
   tenant_description = "Any tenant can use this"
   allocation_model   = "AllocationVApp"

   compute_configuration {
      cpu_limit         = 0
      cpu_guaranteed    = 20
      cpu_speed         = 256
      memory_limit      = 1024
      memory_guaranteed = 30
   }

   provider_vdc {
      id                  = data.vcd_provider_vdc.pvdc1.id
      external_network_id = data.vcd_external_network_v2.ext_net.id
   }

   provider_vdc {
      id                  = data.vcd_provider_vdc.pvdc2.id
      external_network_id = data.vcd_external_network_v2.ext_net.id
   }

   storage_profile {
      name    = "*"
      default = true
      limit   = 1024
   }

   network_pool_id = data.vcd_network_pool.np1.id

   readable_by_org_ids = [
      data.vcd_org.org.id
   ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name to give to the Organization VDC Template, as seen by System administrators
* `description` - (Optional) Description of the Organization VDC Template, as seen by System administrators
* `tenant_name` - (Required) Name to give to the Organization VDC Template, as seen by the allowed tenants
* `tenant_description` - (Optional) Description of the Organization VDC Template, as seen by the allowed tenants
* `provider_vdc` - (Required) A block that defines a candidate location for the instantiated VDCs. There must be **at least one**, which has the following properties:
  * `id` - (Required) ID of the Provider VDC, can be obtained with
  [`vcd_provider_vdc` data source](/providers/vmware/vcd/latest/docs/data-sources/provider_vdc)
  * `external_network_id` - (Required) ID of the Provider Gateway to use, can be obtained with
  [`vcd_external_network_v2` data source](/providers/vmware/vcd/latest/docs/data-sources/external_network_v2)
  * `gateway_edge_cluster_id` - (Optional) ID of the Edge Cluster that the VDCs instantiated from this template will use with the Edge Gateway.
  Can be obtained with [`vcd_nsxt_edge_cluster` data source](/providers/vmware/vcd/latest/docs/data-sources/nsxt_edge_cluster).
  If set, a `edge_gateway` block **must** be present in the VDC Template configuration (see below).
  * `services_edge_cluster_id` - (Optional) ID of the Edge Cluster that the VDCs instantiated from this template will use for services.
  Can be obtained with [`vcd_nsxt_edge_cluster` data source](/providers/vmware/vcd/latest/docs/data-sources/nsxt_edge_cluster)
* `allocation_model` - (Required) Allocation model that the VDCs instantiated from this template will use.
  Must be one of: `AllocationVApp`, `AllocationPool`, `ReservationPool` or  `Flex`
* `compute_configuration`: The compute configuration for the VDCs instantiated from this template:
  * `cpu_allocated` - (Required for `AllocationPool`, `ReservationPool` or `Flex`) The maximum amount of CPU, in MHz, available to the VMs running within the VDC that is instantiated from this template. Minimum is 256MHz
  * `cpu_limit` - (Required for `AllocationVApp`, `ReservationPool` or `Flex`) The limit amount of CPU, in MHz, of the VDC that is instantiated from this template. Minimum is 256MHz. 0 means unlimited
  * `cpu_guaranteed` - (Required for `AllocationVApp`, `AllocationPool` or `Flex`) The percentage of the CPU guaranteed to be available to VMs running within the VDC instantiated from this template
  * `cpu_speed` - (Required for `AllocationVApp`, `AllocationPool` or `Flex`) Specifies the clock frequency, in MHz, for any virtual CPU that is allocated to a VM. Minimum is 256MHz
  * `memory_allocated` - (Required for `AllocationPool`, `ReservationPool` or `Flex`) The maximum amount of Memory, in MB, available to the VMs running within the VDC that is instantiated from this template
  * `memory_limit` - (Required for `AllocationVApp`, `ReservationPool` or `Flex`) The limit amount of Memory, in MB, of the VDC that is instantiated from this template. Minimum is 1024MB. 0 means unlimited
  * `memory_guaranteed` - (Required for `AllocationVApp`, `AllocationPool` or `Flex`) The percentage of the Memory guaranteed to be available to VMs running within the VDC instantiated from this template
  * `elasticity` - (Required for `Flex`) True if compute capacity can grow or shrink based on demand
  * `include_vm_memory_overhead` - (Required for `Flex`) True if the instantiated VDC includes memory overhead into its accounting for admission control
* `storage_profile` - (Required) A block that defines a storage profile that the VDCs instantiated from this template will use. Must be **at least one**, which has the following properties:
  * `name` - (Required) Name of Provider VDC storage profile to use for the VDCs instantiated from this template
  * `default` - (Required) True if this is default storage profile for the VDCs instantiated from this template. Only **one** block should have this set to `true`
  * `limit` - (Required) Storage limit for the VDCs instantiated from this template, in Megabytes. 0 means unlimited
* `enable_fast_provisioning` - (Optional) If `true`, the VDCs instantiated from this template will have Fast provisioning enabled. Defaults to `false`
* `thin_provisioning` - (Optional) If `true`, the VDCs instantiated from this template will have Thin provisioning enabled. Defaults to `false`
* `edge_gateway` - (Optional) VDCs instantiated from this template will create a new Edge Gateway with the provided setup. Required if any `provider_vdc` block
  has defined a `gateway_edge_cluster_id`. This **unique** block has the following properties:
  * `name` - (Required) Name of the Edge Gateway
  * `description` - (Optional) Description of the Edge Gateway
  * `ip_allocation_count` - (Optional) Allocated IPs for the Edge Gateway. Defaults to 0
  * `network_name` - (Required) Name of the network to create with the Edge Gateway
  * `network_description` - (Optional) Description of the network to create with the Edge Gateway
  * `network_gateway_cidr` - (Required) CIDR of the Edge Gateway for the created network
  * `static_ip_pool` - (Required) **One block** with a single IP range (this is a constraint due to a bug in VCD) that have the following properties:
    * `start_address` - (Required) Start address of the IP range
    * `end_address` - (Required) End address of the IP range
* `network_pool_id` - (Optional) If set, specifies the Network pool for the instantiated VDCs. Otherwise, it is automatically chosen
* `nic_quota` - (Optional) Quota for the NICs of the instantiated VDCs. 0 means unlimited. Defaults to 0
* `vm_quota` - (Optional) Quota for the VMs of the instantiated VDCs. 0 means unlimited. Defaults to 0
* `provisioned_network_quota` - (Optional) Quota for the provisioned networks of the instantiated VDCs. 0 means unlimited. Defaults to 0
* `readable_by_org_ids` - (Optional) A set of Organization IDs that will be able to view and read this VDC template, they can be obtained with
  [`vcd_org` data source](/providers/vmware/vcd/latest/docs/data-sources/org)

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing Organization VDC Template can be [imported][docs-import] into this resource via supplying its System name (`name`).
For example, using this structure, representing an existing Organization VDC Template that was **not** created using Terraform:

```hcl
resource "vcd_org_vdc_template" "an_existing_vdc_template" {
  # ...
}
```

You can import such Organization VDC Template into Terraform state using one of the following commands

```
terraform import vcd_org_vdc_template.an_existing_vdc_template "MyTemplate"
```

After that, you must expand the configuration file before you can either update or delete the Organization VDC Template. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the stored properties.

[docs-import]:https://www.terraform.io/docs/import/
