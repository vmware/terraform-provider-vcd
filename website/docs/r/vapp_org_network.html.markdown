---
layout: "vcd"
page_title: "vCloudDirector: vcd_vapp_org_network"
sidebar_current: "docs-vcd-resource-vapp-org-network"
description: |-
  Provides capability to attach existing Org VDC Network to vApp. This can be used to create and delete internal Org VDC networks for vApps to connect.
---

# vcd\_vapp\_org\_network

 Provides capability to attach existing Org VDC Network to vApp. This can be used to create and delete internal Org VDC networks for vApps to connect.

Supported in provider *v2.7+*

## Example Usage

```hcl
resource "vcd_vapp_org_network" "vappOrgNet" {
  org = "my-org" #Optional
  vdc = "my-vdc" #Optional

  vapp_name    = "my-vapp"
  org_network  = "my-org-network"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when 
  connected as sysadmin working across different organisations.
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level.
* `vapp_name` - (Required) The vApp this VM should belong to.
* `org_network` - (Required) An Org network name to which vApp network is connected to.
* `is_fenced` (Optional) "Fencing allows identical virtual machines in different vApp networks connect to organization VDC networks that are accessed in this vApp. Default is false.
* `firewall_enabled` - (Optional) Firewall service enabled or disabled. Configurable when `is_fenced` is true. Default is true. 
* `nat_enabled` - (Optional) NAT service enabled or disabled. Configurable when `is_fenced` is true. Default is true.
* `retain_ip_mac_enabled` - (Optional) Specifies whether the network resources such as IP/MAC of router will be retained across deployments. Configurable when `is_fenced` is true. Default is false.


