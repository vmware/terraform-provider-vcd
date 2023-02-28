---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxv_service_finder"
sidebar_current: "docs-vcd-data-source-nsxv-service-finder"
description: |-
  Provides a VMware Cloud Director data source for searching NSX-V distributed firewall services and service groups
---

# vcd\_nsxv\_service_finder

Provides a VMware Cloud Director NSX-V distributed firewall services and service groups finder
used to retrieve existing services by regular expressions

Supported in provider *v3.9+*

## Example usage 1

```hcl
data "vcd_odg_vdc" "my-vdc" {
  org  = "my-org"
  name = "my-vdc"
}

data "vcd_nsxv_service_finder" "my-SQL-services" {
  vdc_id            = data.vcd_odg_vdc.my-vdc.id
  search_expression = "sql"
  case_sensitive    = false
  type              = "service_group"
}
```

Sample output:

```
service_groups = {
  "case_sensitive" = false
  "id" = "urn:vcloud:vdc:e5680ceb-1c15-48a8-9a54-e0bbc6fe909f"
  "objects" = toset([
    {
      "name" = "MSSQL Integration Services"
      "type" = "ApplicationGroup"
      "value" = "applicationgroup-10"
    },
    {
      "name" = "MSSQL Reporting Services"
      "type" = "ApplicationGroup"
      "value" = "applicationgroup-11"
    },
    {
      "name" = "MSSQL Server Analysis Services"
      "type" = "ApplicationGroup"
      "value" = "applicationgroup-12"
    },
    {
      "name" = "MSSQL Server Database Engine"
      "type" = "ApplicationGroup"
      "value" = "applicationgroup-13"
    },
    {
      "name" = "Microsoft SQL Server"
      "type" = "ApplicationGroup"
      "value" = "applicationgroup-18"
    },
    {
      "name" = "Oracle i*SQLPlus"
      "type" = "ApplicationGroup"
      "value" = "applicationgroup-25"
    },
  ])
  "search_expression" = "sql"
  "type" = "service_group"
  "vdc_id" = "urn:vcloud:vdc:e5680ceb-1c15-48a8-9a54-e0bbc6fe909f"
}
```

## Example usage 2

```hcl
data "vcd_nsxv_service_finder" "services" {
  vdc_id            = data.vcd_org_vdc.my-vdc.id
  search_expression = "dns"
  case_sensitive    = false
  type              = "service"
}
```

Sample output:

```
services = {
  "case_sensitive" = false
  "id" = "urn:vcloud:vdc:e5680ceb-1c15-48a8-9a54-e0bbc6fe909f"
  "objects" = toset([
    {
      "name" = "APP_DNS"
      "type" = "Application"
      "value" = "application-297"
    },
    {
      "name" = "DNS"
      "type" = "Application"
      "value" = "application-136"
    },
    {
      "name" = "DNS-UDP"
      "type" = "Application"
      "value" = "application-286"
    },
  ])
  "search_expression" = "dns"
  "type" = "service"
  "vdc_id" = "urn:vcloud:vdc:e5680ceb-1c15-48a8-9a54-e0bbc6fe909f"
}
```

## Argument Reference

The following arguments are supported:

* `vdc_id` - (Required) The ID of VDC to use.
* `search_expression` - (Required) The regular expression that will be used to search the services
* `type` - (Required) What kind of service we seek. One of `service`, `service_group`
* `case_sensitive` (Optional) Makes the search case-sensitive. By default, it is false.

## Attribute Reference

* `objects` - A list of objects found by the search expression. Each one contains the following properties:
  * `name` - The name of the object
  * `value` - The identifier of the object
  * `type` - the type of the object (`Application` for services and `ApplicationGroup` for service groups)
