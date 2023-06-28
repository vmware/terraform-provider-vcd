---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_service_account"
sidebar_current: "docs-vcd-datasource-service-account"
description: |-
  Provides a data source to read VCD Service Accounts.
---

# vcd\_service\_account

Provides a data source to read VCD Service Accounts.

Supported in provider *v3.10+* and VCD 10.4+.

## Example Usage 1

```hcl
data "vcd_service_account" "example" {
  org  = "my-org"
  name = "my-parent-network"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
  when connected as sysadmin working across different organisations.
* `name` - (Required) Name of the Service Account in an organisation

## Attribute Reference

All the attributes defined in [`vcd_service_account`](/providers/vmware/vcd/latest/docs/resources/service_account)
resource except `file_name` and `allow_token_file` are available.
