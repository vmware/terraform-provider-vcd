---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_ip_space_custom_quota"
sidebar_current: "docs-vcd-resource-ip-space-custom-quota"
description: |-
  Provides a data source to read Custom Quotas for a given Org in a particular IP Space.
---

# vcd\_ip\_space\_custom\_quota

Provides a data source to read Custom Quotas for a given Org in a particular IP Space.

## Example Usage

```hcl
data "vcd_ip_space_custom_quota" "org1" {
  org_id      = data.vcd_org.org1.id
  ip_space_id = vcd_ip_space.space1.id
}
```

## Argument Reference

The following arguments are supported:

* `ip_space_id` - (Required) - IP Space ID to read Custom Quotas
* `org_id` - (Required) Organization ID, for which the Custom Quota should be read

## Attribute Reference

All the arguments and attributes defined in
[`vcd_ip_space_custom_quota`](/providers/vmware/vcd/latest/docs/resources/ip_space_custom_quota)
resource are available.
