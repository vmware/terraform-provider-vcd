---
layout: "vcd"
page_title: "vCloudDirector: vcd_org"
sidebar_current: "docs-vcd-resource-org"
description: |-
  Provides a vCloud Director Organization resource. This can be used to create  delete, and update an organization.
---

# vcd\_org

Provides a vCloud Director Org resource. This can be used to create, update, and delete an organization.
Requires system administrator privileges.

Supported in provider *v2.0+*

!> **Warning:** Up to version 2.4, there were two bugs in the handling of this resource. If you have existing resources
created in versions 2.0 to 2.4, you should re-create them by following *Upgrading Org resources to 2.5* below.


## Example Usage

```hcl
provider "vcd" {
  user     = "${var.admin_user}"
  password = "${var.admin_password}"
  org      = "System"
  url      = "https://AcmeVcd/api"
}

resource "vcd_org" "my-org" {
  name             = "my-org"
  full_name        = "My organization"
  description      = "The pride of my work"
  is_enabled       = "true"
  delete_recursive = "true"
  delete_force     = "true"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Org name
* `full_name` - (Required) Org full name
* `delete_recursive` - (Required) - pass `delete_recursive`=true as query parameter to remove an organization or VDC and any objects it contains that are in a state that normally allows removal.
* `delete_force` - (Required) - pass `delete_force=true` and `delete_recursive=true` to remove an organization or VDC and any objects it contains, regardless of their state.
* `is_enabled` - (Optional) - True if this organization is enabled (allows login and all other operations). Default is `true`.
* `description` - (Optional) - Org description. Default is empty.
* `deployed_vm_quota` - (Optional) - Maximum number of virtual machines that can be deployed simultaneously by a member of this organization. Default is unlimited (0)
* `stored_vm_quota` - (Optional) - Maximum number of virtual machines in vApps or vApp templates that can be stored in an undeployed state by a member of this organization. Default is unlimited (0)
* `can_publish_catalogs` - (Optional) - True if this organization is allowed to share catalogs. Default is `true`.
* `delay_after_power_on_seconds` - (Optional) - Specifies this organization's default for virtual machine boot delay after power on. Default is `0`.

## Importing

Supported in provider *v2.5+*

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing Org can be [imported][docs-import] into this resource via supplying the path for an Org. Since the Org is
at the top of the vCD hierarchy, the path corresponds to the Org name.
For example, using this structure, representing an existing Org that was **not** created using Terraform:

```hcl
resource "vcd_org" "my-orgadmin" {
  name             = "my-org"
  full_name        = "guessing"
  delete_recursive = "true"
  delete_force     = "true"
}
```

You can import such organization into terraform state using this command

```
terraform import vcd_org.my-org my-org
```

[docs-import]:https://www.terraform.io/docs/import/

The state (in `terraform.tfstate`) would look like this:

```json
{
  "version": 4,
  "terraform_version": "0.12.0",
  "serial": 1,
  "lineage": "4f328a1d-3ac3-a1be-b739-c1edde689335",
  "outputs": {},
  "resources": [
    {
      "mode": "managed",
      "type": "vcd_org",
      "name": "my-org",
      "provider": "provider.vcd",
      "instances": [
        {
          "schema_version": 0,
          "attributes": {
            "can_publish_catalogs": true,
            "delay_after_power_on_seconds": null,
            "delete_force": null,
            "delete_recursive": null,
            "deployed_vm_quota": 50,
            "description": "",
            "full_name": "my-org",
            "id": "urn:vcloud:org:875e81c4-3d7a-4bf4-b7db-9d0abe0f0b0d",
            "is_enabled": true,
            "name": "my-org",
            "stored_vm_quota": 50
          }
        }
      ]
    }
  ]
}
```

After that, you can expand the configuration file and either update or delete the org as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the Org's stored properties.

## Upgrading Org resources to 2.5

If you have resources that were created with earlier versions, they may not work correctly in 2.5+, due to a few bugs
in the handling of the resource ID and the default values for VM quotas.

Running a plan on such resource, terraform would want to re-deploy the resource, which is a consequence of the bug fix
that now gives the correct ID to the resource.

In this scenario, the safest approach is to remove the resource from terraform state and import it, using these steps.
Let's assume your org `my-org` was created in 2.4.

1. `terraform state list` (it will show `vcd_org.my-org`)
2. `terraform state rm vcd_org.my-org`
3. `terraform import vcd_org.my-org my-org`

At this point, the org will have the correct information.

## Sources

* [OrgType](https://code.vmware.com/apis/287/vcloud#/doc/doc/types/OrgType.html)
* [ReferenceType](https://code.vmware.com/apis/287/vcloud#/doc/doc/types/ReferenceType.html)
* [Org deletion](https://code.vmware.com/apis/287/vcloud#/doc/doc/operations/DELETE-Organization.html)

