---
layout: "vcd"
page_title: "vCloudDirector: vcd_lb_application_profile"
sidebar_current: "docs-vcd-data-source-lb-application-profile"
description: |-
  Provides a NSX load balancer application profile data source.
---

# vcd\_lb\_application\_profile

Provides a vCloud Director Edge Gateway Load Balancer Application Profile data source. An
application profile defines the behavior of the load balancer for a particular type of network
traffic. After configuring a profile, you associate it with a virtual server. The virtual server
then processes traffic according to the values specified in the profile.

~> **Note:** See additional support notes in [application profile resource page]
(/docs/providers/vcd/r/lb_application_profile.html).

Supported in provider *v2.4+*

## Example Usage

```hcl
data "vcd_lb_application_profile" "my-profile" {
  org                 = "my-org"
  vdc                 = "my-org-vdc"
  edge_gateway        = "my-edge-gw"

  name = "not-managed"
}
```

## Argument Reference

The following arguments are supported:

* `edge_gateway` - (Required) The name of the edge gateway on which the service monitor is defined
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level
* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
when connected as sysadmin working across different organisations
* `name` - (Required) Application profile name for identifying the exact application profile

## Attribute Reference

All the attributes defined in `vcd_lb_application_profile` resource are be available.
