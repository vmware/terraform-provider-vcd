---
layout: "vcd"
page_title: "vCloudDirector: vcd_lb_service_monitor"
sidebar_current: "docs-vcd-resource-lb-service-monitor"
description: |-
  Provides an NSX load balancer service monitor resource.
---

# vcd\_lb\_service\_monitor

Provides a vCloud Director Edge Gateway Load Balancer Service Monitor resource. A service monitor 
defines health check parameters for a particular type of network traffic. It can be associated with
a pool. Pool members are monitored according to the service monitor parameters. 

~> **Note:** To make load balancing work one must ensure that load balancing is enabled on edge gateway (edge gateway must be advanced).
This depends on NSX version to work properly. Please refer to [VMware Product Interoperability Matrices](https://www.vmware.com/resources/compatibility/sim/interop_matrix.php#interop&29=&93=) 
to check supported vCloud director and NSX for vSphere configurations.

~> **Note:** The vCloud Director API for NSX supports a subset of the operations and objects defined in the NSX vSphere 
API Guide. The API supports NSX 6.2, 6.3, and 6.4.

Supported in provider *v2.4+*

## Example Usage

```hcl
provider "vcd" {
  user     = "${var.admin_user}"
  password = "${var.admin_password}"
  org      = "System"
  url      = "https://AcmeVcd/api"
}

resource "vcd_lb_service_monitor" "monitor" {
  org          = "my-org"
  vdc          = "my-org-vdc"
  edge_gateway = "my-edge-gw"

  name        = "http-monitor"
  interval    = "5"
  timeout     = "20"
  max_retries = "3"
  type        = "http"
  method      = "GET"
  url         = "/health"
  send        = "{\"key\": \"value\"}"
  extension = {
    content-type = "application/json"
    linespan     = ""
  }
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level
* `edge_gateway` - (Required) The name of the edge gateway on which the service monitor is to be created
* `name` - (Required) Service Monitor name
* `interval` - (Required) Interval in seconds at which a server is to be monitored using the specified Method.
* `timeout` - (Required) Maximum time in seconds within which a response from the server must be received
* `max_retries` - (Required) Number of times the specified monitoring Method must fail sequentially before the server is declared down
* `type` - (Required) Select the way in which you want to send the health check request to the server — `http`, `https`, 
`tcp`, `icmp`, or `udp`. Depending on the type selected, the remaining attributes are allowed or not
* `method` - (Optional) For types `http` and `https`. Select http method to be used to detect server status. One of OPTIONS, GET, HEAD, POST, PUT, DELETE, TRACE, or CONNECT
* `url` - (Optional) For types `http` and `https`. URL to be used in the server status request
* `send` - (Optional) For types `http`,  `https`, and `udp`. The data to be sent.
* `expected` - (Optional) For types `http` and `https`. String that the monitor expects to match in the status line of 
the HTTP or HTTPS response (for example, `HTTP/1.1`)
* `receive` - (Optional) For types `http`,  `https`, and `udp`. The string to be matched in the response content.
**Note**: When `expected` is not matched, the monitor does not try to match the Receive content
* `extension` - (Required) A map of advanced monitor parameters as key=value pairs (i.e. `max-age=SECONDS`, `invert-regex`)
**Note**: When you need a value of `key` only format just set value to empty string (i.e. `linespan = ""`)

## Attribute Reference

The following attributes are exported on the base level of this resource:

* `id` - The NSX ID of the load balancer service monitor

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.](https://www.terraform.io/docs/import/)

An existing load balancer service monitor can be [imported][docs-import] into this resource
via supplying the full dot separated path for load balancer service monitor. An example is below:

[docs-import]: /docs/import/index.html

```
terraform import vcd_lb_service_monitor.imported my-org.my-org-vdc.my-edge-gw.my-lb-service-monitor
```

The above would import the service monitor named `my-lb-service-monitor` that is defined on edge gateway
`my-edge-gw` which is configured in organization named `my-org` and vDC named `my-org-vdc`.
