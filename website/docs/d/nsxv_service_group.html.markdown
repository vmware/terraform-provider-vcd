---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxv_service_group"
sidebar_current: "docs-vcd-data-source-nsxv-service-group"
description: |-
  Provides a VMware Cloud Director data source for reading NSX-V distributed firewall service groups
---

# vcd\_nsxv\_service\_group

Provides a VMware Cloud Director NSXV distributed firewall service used to read an existing service group

Supported in provider *v3.9+*

## Example Usage

```hcl
data "vcd_odg_vdc" "my-vdc" {
  org  = "my-org"
  name = "my-vdc"
}

data "vcd_nsxv_service_group" "reporting-services" {
  vdc_id = data.vcd_odg_vdc.my-vdc.id
  name   = "MSSQL Reporting Services"
}
```

Sample output:

```
reporting-services = {
  "id" = "applicationgroup-11"
  "name" = "MSSQL Reporting Services"
  "services" = toset([
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

* `vdc_id` - (Required) The ID of VDC to use.
* `name` - (Required) The name of the service group.

## Attribute Reference

* `id` - The identifier of the service groups
* `services` - The list of the services belonging to this group. For each one we get the following:
  * `name` - The name of the service
  * `value` - The identifier of the service
