---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_ip_space_ip_allocation"
sidebar_current: "docs-vcd-data-source-ip-space-ip-allocation"
description: |-
  Provides a data source to read IP IP Allocations within IP Spaces. It supports both - Floating IPs 
  (IPs from IP Ranges) and IP Prefix (subnet) allocations with manual and automatic reservations.
---

# vcd\_ip\_space\_ip\_allocation

Provides a data source to read IP IP Allocations within IP Spaces. It supports both - Floating IPs
(IPs from IP Ranges) and IP Prefix (subnet) allocations with manual and automatic reservations.

IP Spaces require VCD 10.4.1+ with NSX-T.

## Example Usage (IP Space IP Prefix Allocation)

```hcl
data "vcd_ip_space_ip_allocation" "ip-prefix" {
  org_id      = data.vcd_org.org1.id
  ip_space_id = vcd_ip_space.space1.id
  type        = "IP_PREFIX"
  ip_address  = "192.168.1.1/24"
}
```

## Example Usage (IP Space Floating IP Allocation)
```hcl
data "vcd_ip_space_ip_allocation" "floating-ip" {
  org_id      = data.vcd_org.org1.id
  ip_space_id = vcd_ip_space.space1.id
  type        = "FLOATING_IP"
  ip_address  = "192.168.1.1"
}
```

## Argument Reference

The following arguments are supported:

* `ip_space_id` - (Required) Parent IP Space ID of IP Allocation
* `org_id` - (Required) Parent Org ID of IP Allocation
* `type` - (Required) Type of IP Allocation. One of `FLOATING_IP` or `IP_PREFIX`
* `ip_address` - (Required) IP Address or CIDR of IP allocation (e.g. "192.168.1.1/24", "192.168.1.1")

## Attribute Reference

All the arguments and attributes defined in
[`vcd_ip_space_ip_allocation`](/providers/vmware/vcd/latest/docs/resources/ip_space_ip_allocation)
resource are available.
