---
layout: "vcd"
page_title: "vCloudDirector: vcd_lb_service_monitor"
sidebar_current: "docs-vcd-data-source-lb-service-monitor"
description: |-
  Provides a NSX load balancer service monitor data source.
---

# vcd\_lb\_service\_monitor

Provides a vCloud Director Edge Gateway Load Balancer Service Monitor data source. A service monitor 
defines health check parameters for a particular type of network traffic. It can be associated with
a pool. Pool members are monitored according to the service monitor parameters. 

~> **Note:** To make load balancing work one must ensure that load balancing is enabled on edge gateway. This depends 
on NSX version to work properly. Please refer to [VMware Product Interoperability Matrices](https://www.vmware.com/resources/compatibility/sim/interop_matrix.php#interop&29=&93=) 
to check supported vCloud director and NSX for vSphere configurations.

~> **Note:** The vCloud Director API for NSX supports a subset of the operations and objects defined in the NSX vSphere 
API Guide. The API supports NSX 6.2, 6.3, and 6.4.

Supported in provider *v2.4+*

## Example Usage

```hcl
data "vcd_lb_service_monitor" "my-monitor" {
  org                 = "my-org"
  vdc                 = "my-org-vdc"
  edge_gateway        = "my-edge-gw"

  name = "not-managed"
}
```

## Argument Reference

The following arguments are supported:

* `edge_gateway` - (Required) The name of the edge gateway on which the service monitor is defined
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level
* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `name` - (Required) Service Monitor name for identifying the exact service monitor

## Attribute Reference

All the attributes defined in `vcd_lb_service_monitor` resource are be available.
