---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_rde"
sidebar_current: "docs-vcd-data-source-rde"
description: |-
   Provides the capability of reading an existing Runtime Defined Entity in VMware Cloud Director.
---

# vcd\_rde

Provides the capability of reading an existing Runtime Defined Entity in VMware Cloud Director.

-> VCD allows to have multiple RDEs of the same [RDE Type](/providers/vmware/vcd/latest/docs/resources/rde_type) with
the same name, meaning that the data source will not be able to fetch a RDE in this situation, as this data source
can only retrieve **unique RDEs**.

Supported in provider *v3.9+*

## Example Usage with a JSON file

```hcl
data "vcd_rde_type" "my_type" {
  vendor    = "bigcorp"
  namespace = "tech1"
  version   = "1.2.3"
}

data "vcd_rde" "my_rde" {
  org         = "my-org"
  rde_type_id = data.vcd_rde_type.my-type.id
  name        = "My custom RDE"
}

output "rde_output" {
  value = vcd_rde.my_rde.entity
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) Name of the [Organization](/providers/vmware/vcd/latest/docs/data-sources/org) that owns the RDE, optional if defined at provider level.
* `rde_type_id` - (Required) The ID of the [RDE Type](/providers/vmware/vcd/latest/docs/data-sources/rde_type) of the RDE to fetch.
* `name` - (Required) The name of the Runtime Defined Entity.

## Attribute Reference

The following attributes are supported:

* `entity` - The entity JSON.
* `owner_user_id` - The ID of the [Organization user](/providers/vmware/vcd/latest/docs/resources/org_user) that owns this Runtime Defined Entity.
* `org_id` - The ID of the [Organization](/providers/vmware/vcd/latest/docs/resources/org) to which the Runtime Defined Entity belongs.
* `state` - It can be `RESOLVED`, `RESOLUTION_ERROR` or `PRE_CREATED`.
