---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_edgegateway_bgp_configuration"
sidebar_current: "docs-vcd-resource-nsxt-edgegateway-bgp-configuration"
description: |-
  Provides a resource to manage BGP configuration on NSX-T Edge Gateway that has a dedicated Tier-0 
  Gateway or VRF.
---

# vcd\_nsxt\_edgegateway\_bgp\_configuration

Provides a resource to manage BGP configuration on NSX-T Edge Gateway that has a dedicated Tier-0
Gateway or VRF. BGP makes core routing decisions by using a table of IP networks, or prefixes, which
designate multiple routes between autonomous systems (AS).

~> Only `System Administrator` can create this resource.

~> In an NSX-T Edge Gateway that is connected to an external network backed by a VRF gateway, the
local AS number and graceful restart settings are **read only**. Your system administrator can edit
these settings on the parent Tier-0 gateway in NSX-T Data Center. 

## Example Usage (Using BGP configuration for a dedicated Tier-0 gateway backed Edge Gateway)

```hcl
resource "vcd_nsxt_edgegateway_bgp_configuration" "testing" {
  org = "my-org"

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  enabled                = false
  local_as_number        = "65430"
  graceful_restart_mode  = "HELPER_ONLY"
  graceful_restart_timer = 190
  stale_route_timer      = 600
  ecmp_enabled           = true
}
```

## Example Usage (Using BGP configuration for a dedicated VRF backed Edge Gateway)
```hcl
resource "vcd_nsxt_edgegateway_bgp_configuration" "testing" {
  org = "my-org"

  edge_gateway_id = vcd_nsxt_edgegateway.vrf-backed.id

  enabled      = true
  ecmp_enabled = true
}
```


## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
  when connected as sysadmin working across different organisations
* `edge_gateway_id` - (Required) The ID of the edge gateway (NSX-T only). Can be looked up using
  `vcd_nsxt_edgegateway` datasource
* `enabled` - (Required) Defines if BGP service is enabled or not
* `ecmp_enabled` (Optional) - A flag indicating whether ECMP is enabled or not
* `local_as_number` - (Optional) BGP autonomous systems (AS) number to advertise to BGP peers. BGP
  AS number can be specified in either ASPLAIN or ASDOT formats, like ASPLAIN format : '65546',
  ASDOT format : '1.10'. **Read only** for VRF backed Edge Gateways
* `graceful_restart_mode` - (Optional) - Describes Graceful Restart configuration Modes for BGP
  configuration on an Edge Gateway. **Read only** for VRF backed Edge Gateways. Possible options are:
 * `DISABLE` - Both graceful restart and helper modes are disabled
 * `HELPER_ONLY` - The ability for a BGP speaker to indicate its ability to preserve forwarding
   state during BGP restart
 * `GRACEFUL_AND_HELPER` - The ability of a BGP speaker to advertise its restart to its peers
* `graceful_restart_timer` (Optional) Maximum time taken (in seconds) for a BGP session to be
  established after a restart. If the session is not re-established within this timer, the receiving
  speaker will delete all the stale routes from that peer. **Read only** for VRF backed Edge Gateways.
* `stale_route_timer` (Optional) - Maximum time (in seconds) before stale routes are removed when
  BGP restarts. **Read only** for VRF backed Edge Gateways

More information about settings can be found in VMware Cloud Director [BGP Configuration
documentation](https://docs.vmware.com/en/VMware-Cloud-Director/10.3/VMware-Cloud-Director-Tenant-Portal-Guide/GUID-EB585DDC-9F1C-4971-A4AD-44C239E6E822.html)

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

Existing BGP Configuration can be [imported][docs-import] into this resource
via supplying the full dot separated path for your Edge Gateway name. An example is
below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxt_edgegateway_bgp_configuration.imported my-org.my-org-vdc-org-vdc-group-name.my-nsxt-edge-gateway
```

The above would import BGP configuration defined on NSX-T Edge Gateway `my-nsxt-edge-gateway` which
is configured in organization named `my-org` and VDC or VDC Group named
`my-org-vdc-org-vdc-group-name`.
