---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxv_service"
sidebar_current: "docs-vcd-data-source-nsxv-service"
description: |-
  Provides a VMware Cloud Director data source for reading NSX-V distributed firewall services
---

# vcd\_nsxv\_service

Provides a VMware Cloud Director NSX-V distributed firewall service used to read an existing service

Supported in provider *v3.9+*

## Example Usage

```hcl
data "vcd_odg_vdc" "my-vdc" {
  org  = "my-org"
  name = "my-vdc"
}

data "vcd_nsxv_service" "pop3-service" {
  vdc_id = data.vcd_odg_vdc.my-vdc.id
  name   = "POP3"
}
```

Sample output:

```
pop3-service = {
  "app_guid" = tostring(null)
  "id" = "application-7"
  "name" = "POP3"
  "ports" = "110"
  "protocol" = "TCP"
  "source_port" = tostring(null)
  "vdc_id" = "urn:vcloud:vdc:e5680ceb-1c15-48a8-9a54-e0bbc6fe909f"
}
```

## Argument Reference

The following arguments are supported:

* `vdc_id` - (Required) The ID of VDC to use.
* `name` - (Required) The name of the service.

## Attribute Reference

* `id` - The identifier of the service
* `protocol` - The protocol used by the service
* `ports` - The ports used by the service. Could be a number, a list of numbers, or a range
* `source_port` - The source port used by this service. Not all services provide one
* `destination_port` - The destination port used by this service. Not all services provide one
* `app_guid` - The application Identifier, when available
