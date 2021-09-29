---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_app_port_profile"
sidebar_current: "docs-vcd-data-source-nsxt-app-port-profile"
description: |-
  Provides a data source to read NSX-T Application Port Profiles. Application Port Profiles include 
  a combination of a protocol and a port, or a group of ports, that is used for Firewall and NAT
  services on the Edge Gateway.
---

# vcd\_nsxt\_app\_port\_profile

Supported in provider *v3.3+* and VCD 10.1+ with NSX-T backed VDCs.

Provides a data source to read NSX-T Application Port Profiles. Application Port Profiles include a
combination of a protocol and a port, or a group of ports, that is used for Firewall and NAT
services on the Edge Gateway.

## Example Usage 1 (Find an Application Port Profile defined by Provider)

```hcl
data "vcd_nsxt_app_port_profile" "custom" {
  org   = "my-org"
  vdc   = "my-nsxt-vdc"
  name  = "WINS"
  scope = "PROVIDER"
}
```

## Example Usage 2 (Find an Application Port Profile defined by Tenant)

```hcl
data "vcd_nsxt_app_port_profile" "custom" {
  org   = "my-org"
  vdc   = "my-nsxt-vdc"
  name  = "SSH"
  scope = "TENANT"
}
```

## Example Usage 3 (Find a System defined  Application Port Profile)

```hcl
data "vcd_nsxt_app_port_profile" "custom" {
  scope = "SYSTEM"
  name  = "SSH"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
  when connected as sysadmin working across different organisations.
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level.
* `name` - (Required)  - Unique name of existing Security Group.
* `scope` - (Required)  - `SYSTEM`, `PROVIDER`, or `TENANT`.

## Attribute Reference

All the arguments and attributes defined in
[`vcd_nsxt_app_port_profile`](/providers/vmware/vcd/latest/docs/resources/nsxt_app_port_profile.html) resource
are available.
