---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxv_application_finder"
sidebar_current: "docs-vcd-data-source-nsxv-application-finder"
description: |-
  Provides a VMware Cloud Director data source for searching NSX-V distributed firewall applications and application groups
---

# vcd\_nsxv\_application_finder

Provides a VMware Cloud Director NSX-V distributed firewall applications and application groups finder
used to retrieve existing ones by regular expressions.

Supported in provider *v3.9+*

## Example usage 1

```hcl
data "vcd_odg_vdc" "my-vdc" {
  org  = "my-org"
  name = "my-vdc"
}

data "vcd_nsxv_application_finder" "my-SQL-services" {
  vdc_id            = data.vcd_odg_vdc.my-vdc.id
  search_expression = "sql"
  case_sensitive    = false
  type              = "application_group"
}
```

Sample output:

```
application_groups = {
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
  "type" = "application_group"
  "vdc_id" = "urn:vcloud:vdc:e5680ceb-1c15-48a8-9a54-e0bbc6fe909f"
}
```

## Example usage 2

```hcl
data "vcd_nsxv_application_finder" "applications" {
  vdc_id            = data.vcd_org_vdc.my-vdc.id
  search_expression = "dns"
  case_sensitive    = false
  type              = "application"
}
```

Sample output:

```
applications = {
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
  "type" = "application"
  "vdc_id" = "urn:vcloud:vdc:e5680ceb-1c15-48a8-9a54-e0bbc6fe909f"
}
```

## Argument Reference

The following arguments are supported:

* `vdc_id` - (Required) The ID of VDC to use
* `search_expression` - (Required) The regular expression that will be used to search the applications. See [Search Expressions](#search-expressions) below
* `type` - (Required) What kind of application we seek. One of `application`, `application_group`
* `case_sensitive` (Optional) Makes the search case-sensitive. By default, it is false

## Attribute Reference

* `objects` - A list of objects found by the search expression. Each one contains the following properties:
  * `name` - The name of the object
  * `type` - the type of the object (`Application` or `ApplicationGroup`)
  * `value` - The identifier of the object


## Search expressions

To search for an application or application group, we can use simple or complex [regular expressions](https://en.wikipedia.org/wiki/Regular_expression).
The expressions in this data source follow the [PCRE](https://en.wikipedia.org/wiki/Perl_Compatible_Regular_Expressions) standard.

A **simple** regular expression is a (short) text that we expect to find within the application name. For example, the
expression `sql` will find, among others, `Oracle i*SQLPlus` and `MSSQL Server Database Engine`, because the search, by default,
ignores the case of the searched text.

A more complex regular expression could use meta-characters and regular expression directives to search more precisely.
For example, the expression `^server` tells the search to find a name that starts (`^`) with "server", thus finding
"Server Message Block (SMB)" (starts with `server`), but not  "SAP MDM Server" (where `server` is not at the beginning
of the name).

If we want to search with even more accuracy, we could set the property `case_sensitive = true`, where the case of the
text matters. Thus, searching for `VMware` would find `VMware-SRM-Replication` and `VMware-VCO-Messaging`, but not
`Vmware-VC-WebAccess` (lowercase `m` after `V`).