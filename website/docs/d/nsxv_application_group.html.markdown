---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxv_application_group"
sidebar_current: "docs-vcd-data-source-nsxv-application-group"
description: |-
  Provides a VMware Cloud Director data source for reading NSX-V Distributed Firewall application groups
---

# vcd\_nsxv\_application\_group

Provides a VMware Cloud Director NSX-V Distributed Firewall data source used to read an existing application group.

Supported in provider *v3.9+*

## Example Usage

```hcl
data "vcd_odg_vdc" "my-vdc" {
  org  = "my-org"
  name = "my-vdc"
}

data "vcd_nsxv_application_group" "reporting-applications" {
  vdc_id = data.vcd_odg_vdc.my-vdc.id
  name   = "MSSQL Reporting Services"
}
```

Sample output:

```
reporting-applications = {
  "id" = "applicationgroup-11"
  "name" = "MSSQL Reporting Services"
  "applications" = toset([
    {
      "name" = "HTTP"
      "value" = "application-93"
    },
    {
      "name" = "HTTPS"
      "value" = "application-105"
    },
  ])
  "vdc_id" = "urn:vcloud:vdc:e5680ceb-1c15-48a8-9a54-e0bbc6fe909f"
}
```

## Argument Reference

The following arguments are supported:

* `vdc_id` - (Required) The ID of VDC to use
* `name` - (Required) The name of the application group

## Attribute Reference

* `id` - The identifier of the application groups
* `applications` - The list of the applications belonging to this group. For each one we get the following:
  * `name` - The name of the application
  * `value` - The identifier of the application
