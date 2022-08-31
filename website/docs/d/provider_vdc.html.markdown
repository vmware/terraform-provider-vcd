---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_provider_vdc"
sidebar_current: "docs-vcd-data-source-provider-vdc"
description: |-
  Provides an Provider VDC data source.
---

# vcd\_provider\_vdc

Provides a VMware Cloud Director Provider VDC data source. A Provider VDC can be used to reference a Provider VDC and use its 
data within other resources or data sources.

Supported in provider *v3.8+*

## Example Usage

```hcl
data "vcd_org_vdc" "my-org-vdc" {
  org  = "my-org"
  name = "my-vdc"
}

output "provider_vdc" {
  value = data.vcd_org_vdc.my-org-vdc.provider_vdc_name
}

```

## Argument Reference

The following arguments are supported:

* `org` - (Optional, but required if not set at provider level) Org name 
* `name` - (Required) Organization VDC name

## Attribute reference

All attributes defined in [organization VDC resource](/providers/vmware/vcd/latest/docs/resources/org_vdc#attribute-reference) are supported.

