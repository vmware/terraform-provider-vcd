---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_ip_space"
sidebar_current: "docs-vcd-data-source-ip-space"
description: |-
  Provides a data source to read IP Spaces. IP Spaces provide 
  structured approach to allocating public and private IP addresses by preventing the use of 
  overlapping IP addresses across organizations and organization VDCs.
---

# vcd\_ip\_space

Provides a data source to read IP Spaces. IP Spaces provide structured approach to allocating public
and private IP addresses by preventing the use of overlapping IP addresses across organizations and
organization VDCs.

IP Spaces require VCD 10.4.1+ with NSX-T.

## Example Usage (Private IP Space within an Org)

```hcl
data "vcd_ip_space" "space1" {
  org_id = data.vcd_org.org1.id
  name   = "private-ip-space"
}
```

## Example Usage (Public or Shared IP Space)
```hcl
data "vcd_ip_space" "space1" {
  name = "private-ip-space"
}
```

## Argument Reference

The following arguments are supported:

* `org_id` - (Optional) Org ID for Private IP Space.
* `name` - (Required) The name of IP Space.

## Attribute Reference

All the arguments and attributes defined in
[`vcd_ip_space`](/providers/vmware/vcd/latest/docs/resources/ip_space) resource are available.
