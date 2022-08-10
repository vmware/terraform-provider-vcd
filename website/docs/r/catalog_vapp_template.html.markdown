---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_catalog_vapp_template"
sidebar_current: "docs-vcd-resource-catalog-vapp_template"
description: |-
  Provides a VMware Cloud Director vApp Template resource. This can be used to upload and delete OVA files inside a catalog.
---

# vcd\_catalog\_vapp\_template

Provides a VMware Cloud Director vApp Template resource. This can be used to upload OVA to catalog and delete it.

Supported in provider *v3.8+*

## Example Usage

```hcl
resource "vcd_catalog_vapp_template" "myNewVappTemplate" {
  org     = "my-org"
  catalog = "my-catalog"

  name                 = "my ova"
  description          = "new vapp template"
  ova_path             = "/home/user/file.ova"
  upload_piece_size    = 10
  show_upload_progress = true

  metadata = {
    license = "public"
    version = "v1"
  }
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `catalog` - (Required) The name of the catalog where to upload OVA file
* `name` - (Required) vApp Template name in catalog
* `description` - (Optional) Description of the vApp Template
* `ova_path` - (Optional) Absolute or relative path to file to upload
* `ovf_url` - (Optional) URL to OVF file. Only OVF (not OVA) files are supported by VCD uploading by URL
* `upload_piece_size` - (Optional) - Size in MB for splitting upload size. It can possibly impact upload performance. Default 1MB.
* `show_upload_progress` - (Optional) - Default false. Allows seeing upload progress. (See note below)
* `metadata` - (Optional) Key value map of metadata to assign to the associated vApp Template

### A note about upload progress

Until version 3.5.0, the progress was optionally shown on the screen. Due to changes in the terraform tool, such operation
is no longer possible. The progress messages are thus written to the log file (`go-vcloud-director.log`) using a special
tag `[SCREEN]`. To see the progress at run time, users can run the command below in a separate terminal window while 
`terraform apply` is working:

```
$ tail -f go-vcloud-director.log | grep '\[SCREEN\]'
```

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing vApp Template can be [imported][docs-import] into this resource via supplying the full dot separated path for a
vApp Template. For example, using this structure, representing an existing vAppTemplate that was **not** created using Terraform:

```hcl
resource "vcd_catalog_vapp_template" "my-vapp-template" {
  org      = "my-org"
  catalog  = "my-catalog"
  name     = "my-vapp-template"
  ova_path = "guess"
}
```

You can import such vApp Template into terraform state using this command

```
terraform import vcd_catalog_vapp_template.my-vapp-template my-org.my-catalog.my-vapp-template
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the vApp Template as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the vApp Template's stored properties.
