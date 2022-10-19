---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_vapp"
sidebar_current: "docs-vcd-resource-vapp"
description: |-
  Provides a VMware Cloud Director vApp resource. This can be used to create, modify, and delete vApps.
---

# vcd\_vapp

Provides a VMware Cloud Director vApp resource. This can be used to create, modify, and delete vApps.

## Example of vApp with 2 VMs

Example with more than one VM under a vApp.

```hcl
resource "vcd_network_direct" "direct-network" {
  name             = "net"
  external_network = "my-ext-net"
}

resource "vcd_vapp" "web" {
  name = "web"

  metadata_entry {
    key   = "CostAccount"
    value = "Marketing Department"
  }
}

resource "vcd_vapp_org_network" "direct-network" {
  vapp_name        = vcd_vapp.web.name
  org_network_name = vcd_network_direct.direct-network.name
}

resource "vcd_vapp_vm" "web1" {
  vapp_name = vcd_vapp.web.name
  name      = "web1"

  catalog_name  = "my-catalog"
  template_name = "photon-os"

  memory = 2048
  cpus   = 1

  network {
    type               = "org"
    name               = vcd_vapp_org_network.direct-network.org_network_name
    ip_allocation_mode = "POOL"
  }

  guest_properties = {
    "vapp.property1" = "value1"
    "vapp.property2" = "value2"
  }

  lease {
    runtime_lease_in_sec = 60 * 60 * 24 * 30 # extends the runtime lease to 30 days
    storage_lease_in_sec = 60 * 60 * 24 * 7  # extends the storage lease to 7 days
  }
}

resource "vcd_vapp_vm" "web2" {
  vapp_name = vcd_vapp.web.name
  name      = "web2"

  catalog_name  = "my-catalog"
  template_name = "photon-os"

  memory = 2048
  cpus   = 1

  network {
    type               = "org"
    name               = vcd_vapp_org_network.direct-network.org_network_name
    ip_allocation_mode = "POOL"
  }
}
```

## Example of Empty vApp with no VMs

```hcl
resource "vcd_vapp" "web" {
  name = "web"

  metadata_entry {
    key   = "boss"
    value = "Why is this vApp empty?"
  }

  metadata_entry {
    key   = "john"
    value = "I don't really know. Maybe somebody did forget to clean it up."
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the vApp
* `org` - (Optional; *v2.0+*) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `vdc` - (Optional; *v2.0+*) The name of VDC to use, optional if defined at provider level
* `description` (Optional; *v3.3*) An optional description for the vApp, up to 256 characters.
* `power_on` - (Optional) A boolean value stating if this vApp should be powered on. Default is `false`. Works only on update when vApp already has VMs.
* `metadata` - (Deprecated) Use `metadata_entry` instead. Key value map of metadata to assign to this vApp. Key and value can be any string. (Since *v2.2+* metadata is added directly to vApp instead of first VM in vApp)
* `metadata_entry` - (*v3.8+*) A set of metadata entries to assign. See [Metadata](#metadata) section for details.
* `guest_properties` - (Optional; *v2.5+*) Key value map of vApp guest properties
* `lease` - (Optional *v3.5+*) the information about the vApp lease. It includes the fields below. When this section is 
   included, both fields are mandatory. If lease values are higher than the ones allowed for the whole Org, the values
   are **silently** reduced to the highest value allowed.
  * `runtime_lease_in_sec` - How long any of the VMs in the vApp can run before the vApp is automatically powered off or suspended. 0 means never expires (or maximum allowed by Org). Regular values accepted from 3600+.
  * `storage_lease_in_sec` - How long the vApp is available before being automatically deleted or marked as expired. 0 means never expires (or maximum allowed by Org). Regular values accepted from 3600+.

## Attribute reference

* `href` - (Computed) The vApp Hyper Reference.
* `status` - (Computed; *v2.5+*) The vApp status as a numeric code.
* `status_text` - (Computed; *v2.5+*) The vApp status as text.

<a id="metadata"></a>
## Metadata

The `metadata_entry` (*v3.8+*) is a set of metadata entries that have the following structure:

* `key` - Key of this metadata entry.
* `value` - Value of this metadata entry.
* `type` - Type of this metadata entry. One of: `MetadataStringValue`, `MetadataNumberValue`, `MetadataDateTimeValue`, `MetadataBooleanValue`. Defaults to `MetadataStringValue`.
* `user_access` - User access level for this metadata entry. One of: `PRIVATE` (hidden), `READONLY` (read only), `READWRITE` (read/write). Defaults to `READWRITE`.
* `is_system` - Domain for this metadata entry. true if it belongs to `SYSTEM`, false if it belongs to `GENERAL`. Defaults to `false`.

## Importing

Supported in provider *v2.5+*

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing vApp can be [imported][docs-import] into this resource via supplying its path.
The path for this resource is made of org-name.vdc-name.vapp-name
For example, using this structure, representing a vApp that was **not** created using Terraform:

```hcl
resource "vcd_vapp" "tf-vapp" {
  name = "my-vapp"
  org  = "my-org"
  vdc  = "my-vdc"
}
```

You can import such vapp into terraform state using this command

```
terraform import vcd_vapp.tf-vapp my-org.my-vdc.my-vapp
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After importing, if you run `terraform plan` you will see the rest of the values and modify the script accordingly for
further operations.

