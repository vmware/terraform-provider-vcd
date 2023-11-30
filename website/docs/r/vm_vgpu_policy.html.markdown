---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_vm_vgpu_policy"
sidebar_current: "docs-vcd-resource-vm-vgpu-policy"
description: |-
  Provides a resource to manage vGPU policies for virtual machines in VMware Cloud Director.
---

# vcd\_vm\_vgpu\_policy

Supported in provider *3.11* and VCD *10.4.0+*.

Provides a resource to manage vGPU policies for virtual machines in VMware Cloud Director.

## Example Usage

```hcl
data "vcd_vgpu_profile" "example_vgpu_profile" {
  name = "grid_a100-10c"
}

data "vcd_provider_vdc" "example_provider_vdc" {
  name = "example_provider_vdc"
}

data "vcd_vm_group" "vm_group_example" {
  name = "vm-group-1"
}

resource "vcd_vm_vgpu_policy" "example_vgpu_policy" {
  name        = "example-vgpu-policy"
  description = "An example vGPU policy configuration"

  vgpu_profile {
    id    = data.vcd_vgpu_profile.example_vgpu_profile.id
    count = 1
  }

  cpu {
    shares                = "886"
    limit_in_mhz          = "2400"
    count                 = "9"
    speed_in_mhz          = "2500"
    cores_per_socket      = "3"
    reservation_guarantee = "0.55"
  }

  memory {
    shares                = "1580"
    size_in_mb            = "3200"
    limit_in_mb           = "2800"
  }

  provider_vdc_scope {
    provider_vdc_id = data.vcd_provider_vdc.example_provider_vdc.id
    cluster_names   = ["cluster1"]
    vm_group_id     = data.vcd_vm_group.vm_group_example.id
  }
}
```

## Example usage (without a sizing policy)

```hcl
resource "vcd_vm_vgpu_policy" "example_vgpu_policy_without_sizing" {
  name        = "example-vgpu-policy-without-sizing"
  description = "An example vGPU policy configuration"

  vgpu_profile {
    id    = data.vcd_vgpu_profile.example_vgpu_profile.id
    count = 1
  }

  provider_vdc_scope {
    provider_vdc_id = data.vcd_provider_vdc.example_provider_vdc.id
    cluster_names   = ["cluster1"]
    vm_group_id     = data.vcd_vm_group.vm_group_example.id
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The unique name assigned to the vGPU policy for a virtual machine.
* `description` - (Optional) A brief description of the vGPU policy.
* `vgpu_profile` - (Required) Defines the vGPU profile ID and count. 
* `cpu` - (Optional) Configuration options for CPU resources. See [cpu] for more details.
* `memory` - (Optional) Memory resource configuration settings. See [memory] for more details.
* `provider_vdc_scope` - (Optional) Defines the scope of the policy within 
  provider virtual data centers. If not provided, applies to all the current ant future PVDCs.
  See [`provider_vdc_scope`](#provider-vdc-scope) for more details.

### Provider VDC Scope
* `provider_vdc_id` - (Required) The ID of the provider VDC that should be in the scope.
* `cluster_names` - (Optional) A set of vCenter cluster names on which the provider VDC is hosted. 
  If none are provided, the provider attempts to find one automatically. Can be fetched using `data.vcd_resource_pool.cluster_moref` attribute.
* `vm_group_id` - (Optional) The ID of a VM group to which the policy is available. If not provided, the policy can be applied to all VMs created
  on the PVDC.

Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing vGPU Policy can be [imported][docs-import] into this resource
via supplying the path for it. An example is below:

```hcl
resource "vcd_vm_vgpu_policy" "imported_policy" {
  name = "existing-policy-name"
}
```

```sh
terraform import vcd_vm_vgpu_policy.imported_policy vgpu-policy-name
```

After that, you can expand the configuration file and either update or delete the VM vGPU policy as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the VM vGPU policy stored properties.

### Listing VM vGPU policies

If you want to list IDs there is a special command **`terraform import vcd_vm_vgpu_policy.imported list@`**. 
The output for this command should look similar to the one below:

```
terraform import vcd_vm_vgpu_policy.imported list@
Retrieving all VM vGPU policies
vcd_vm_vgpu_policy.import: Importing from ID "list@"...
No	ID									Name	
--	--									----	
1	urn:vcloud:vdcComputePolicy:100dc35a-572b-4876-a762-c734d67c56ef	tf_policy_3
2	urn:vcloud:vdcComputePolicy:446d623e-1eec-4c8c-8a14-2f7e6086546b	tf_policy_2

```

Now to import VM sizing policy with ID urn:vcloud:vdcComputePolicy:446d623e-1eec-4c8c-8a14-2f7e6086546b one could supply this command:

```shell
$ terraform import vcd_vm_vgpu_policy.imported urn:vcloud:vdcComputePolicy:446d623e-1eec-4c8c-8a14-2f7e6086546b
```

[docs-import]:https://www.terraform.io/docs/import/
[cpu]:/providers/vmware/vcd/latest/docs/resources/vm_sizing_policy#cpu
[memory]:/providers/vmware/vcd/latest/docs/resources/vm_sizing_policy#memory

