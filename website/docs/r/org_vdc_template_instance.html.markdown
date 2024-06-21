---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_org_vdc_template_instance"
sidebar_current: "docs-vcd-resource-org-vdc-template-instance"
description: |-
  Provides a resource to instantiate VDCs with a VDC Template in VMware Cloud Director.
---

# vcd\_org\_vdc\_template\_instance

Provides a resource to instantiate VDCs with a [VDC Template](/providers/vmware/vcd/latest/docs/resources/org_vdc_template) in VMware Cloud Director.
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

resource "vcd_org_vdc_template" "tmpl" {
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

resource "vcd_org_vdc_template_instance" "my_instance" {
  org_vdc_template_id = vcd_org_vdc_template.tmpl.id
  name                = "myInstantiatedVdc"
  description         = "A new VDC"
  org_id              = data.vcd_org.org.id
}
```

## Argument Reference

The following arguments are supported:

* `org_vdc_template_id` - (Required) The ID of the VDC Template to instantiate
* `name` - (Required) Name to give to the instantiated Organization VDC
* `description` - (Optional) Description of the instantiated Organization VDC
* `org_id` - (Required) ID of the Organization where the VDC will be instantiated

## How to manage the instantiated VDC using Terraform

After the `vcd_org_vdc_template_instance` resource is created successfully,
the instantiated VDC identifier is saved in Terraform state as the `vcd_org_vdc_template_instance` resource ID.

If users want to modify the new VDC, they can [import](/providers/vmware/vcd/latest/docs/guides/importing_resources#semi-automated-import-terraform-v15) it.
In the same `.tf` file (once the VDC has been instantiated), or in a new one, we can place the following snippet: 

```hcl
import {
  to = vcd_org_vdc.imported
  id = "my_org.myInstantiatedVdc" # Using the same names from the example
}
```

Note that this importing mechanism still does not support `${}` placeholders, so the Organization and VDC name must be explicitly
written. When running the `terraform plan -generate-config-out=generated_resources.tf`, Terraform will generate the new file
`generated_resources.tf` with the instantiated VDC code.

With a subsequent `terraform apply`, the instantiated VDC will be managed by Terraform as a normal `vcd_org_vdc` resource.

Please take into account that deleting the `vcd_org_vdc_template_instance` resource will attempt to **delete the instantiated VDC** it created,
which require the `Organization vDC: Delete` right and other implicit rights for tenant users.
If you would like to avoid this behavior, you can run `terraform state rm vcd_org_vdc_template_instance.my_instance` so the VDC stops being managed by this resource
(it is removed from Terraform state). This way you can remove the `vcd_org_vdc_template_instance` resource without any other effect.

## Importing

There is no importing for this resource, as it should be used only on creation.
The instantiated VDC can be imported using `vcd_org_vdc` by following the steps of the section above.
