---
layout: "vcd"
page_title: "vCloudDirector: vcd_external_network"
sidebar_current: "docs-vcd-resource-external-network"
description: |-
  TODO
---

# vcd\_external\_network

TODO

Supported in provider *v2.2+*

~> **Note:** Only `System Administrator` can create an organization virtual datacenter network that connects 
directly to an external network. You must use `System Adminstrator` account in `provider` configuration
and then provide `org` and `vdc` arguments for direct networks to work.

## Example Usage

```hcl
resource "vcd_external_network" "net" {
  org = "my-org" # Optional

  name             = "my-ext-net"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional; *v2.0+*) The name of organization to use, optional if defined at provider level. Useful when 
  connected as sysadmin working across different organisations
* `name` - (Required) A unique name for the network
